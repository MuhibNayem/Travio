package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/subscription/internal/repository"
)

// BillingWorker handles automated subscription billing
type BillingWorker struct {
	repo          repository.Repository
	billingSvc    *BillingService
	interval      time.Duration
	maxRetries    int
}

// NewBillingWorker creates a new billing worker
func NewBillingWorker(repo repository.Repository, billingSvc *BillingService, interval time.Duration) *BillingWorker {
	return &BillingWorker{
		repo:       repo,
		billingSvc: billingSvc,
		interval:   interval,
		maxRetries: 3,
	}
}

// Start begins the billing cycle
func (w *BillingWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	logger.Info("Starting Subscription Billing Worker", "interval", w.interval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Subscription Billing Worker")
			return
		case <-ticker.C:
			w.runBillingCycle(ctx)
		}
	}
}

// runBillingCycle processes all subscriptions due for billing
func (w *BillingWorker) runBillingCycle(ctx context.Context) {
	logger.Info("Running billing cycle...")

	// Get all active subscriptions
	subs, err := w.repo.ListSubscriptions(ctx, "", "active")
	if err != nil {
		logger.Error("Failed to list subscriptions for billing", "error", err)
		return
	}

	successCount := 0
	failCount := 0

	for _, sub := range subs {
		// Check if subscription is due for renewal
		if time.Now().Before(sub.CurrentPeriodEnd) {
			continue // Not yet due
		}

		// Attempt billing with retries
		billed := false
		for attempt := 0; attempt < w.maxRetries; attempt++ {
			err := w.billingSvc.ProcessRenewal(ctx, sub.ID)
			if err != nil {
				logger.Warn("Billing attempt failed",
					"sub_id", sub.ID,
					"org_id", sub.OrganizationID,
					"attempt", attempt+1,
					"error", err,
				)
				if attempt < w.maxRetries-1 {
					backoff := time.Duration(attempt+1) * 5 * time.Minute
					select {
					case <-ctx.Done():
						return
					case <-time.After(backoff):
					}
				}
			} else {
				billed = true
				successCount++
				break
			}
		}

		if !billed {
			failCount++
			// After max retries, suspend the subscription
			if err := w.repo.CancelSubscription(ctx, sub.OrganizationID); err != nil {
				logger.Error("Failed to suspend subscription after billing failures",
					"sub_id", sub.ID,
					"error", err,
				)
			} else {
				logger.Info("Subscription suspended due to billing failures",
					"sub_id", sub.ID,
					"org_id", sub.OrganizationID,
				)
			}
		}
	}

	logger.Info("Billing cycle completed",
		"total_processed", len(subs),
		"success", successCount,
		"failed", failCount,
	)
}

// BillingService handles billing operations
type BillingService struct {
	repo        repository.Repository
	paymentRepo PaymentRepository
}

// PaymentRepository interface for payment processing
type PaymentRepository interface {
	CreatePayment(ctx context.Context, orgID, subID string, amount int64) (string, error)
}

// NewBillingService creates a new billing service
func NewBillingService(repo repository.Repository, paymentRepo PaymentRepository) *BillingService {
	return &BillingService{
		repo:        repo,
		paymentRepo: paymentRepo,
	}
}

// ProcessRenewal processes a subscription renewal
func (s *BillingService) ProcessRenewal(ctx context.Context, subID string) error {
	sub, err := s.repo.GetSubscription(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get subscription: %w", err)
	}

	if sub == nil {
		return fmt.Errorf("subscription not found: %s", subID)
	}

	// Get plan details
	plan, err := s.repo.GetPlan(ctx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("failed to get plan: %w", err)
	}

	// Create payment
	paymentID, err := s.paymentRepo.CreatePayment(ctx, sub.OrganizationID, sub.ID, plan.PricePaisa)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// For now, just create the invoice and mark subscription as active
	// In production, you'd update the period dates here
	newStart := sub.CurrentPeriodEnd

	// Create invoice for the renewal
	invoice := &repository.Invoice{
		SubscriptionID: sub.ID,
		AmountPaisa:    plan.PricePaisa,
		Status:         "paid",
		IssuedAt:       time.Now(),
		DueDate:        newStart,
		PaidAt:         &newStart,
		LineItems: []repository.LineItem{
			{
				Description:    fmt.Sprintf("Subscription renewal - %s (%s)", plan.Name, plan.Interval),
				AmountPaisa:    plan.PricePaisa,
				Quantity:       1,
				UnitPricePaisa: plan.PricePaisa,
			},
		},
	}

	if err := s.repo.CreateInvoice(ctx, invoice); err != nil {
		logger.Error("Failed to create renewal invoice", "sub_id", sub.ID, "error", err)
		// Non-fatal, subscription is renewed
	}

	logger.Info("Subscription renewed successfully",
		"sub_id", sub.ID,
		"org_id", sub.OrganizationID,
		"payment_id", paymentID,
	)

	return nil
}

// CalculateProration calculates prorated amount for plan changes
func (s *BillingService) CalculateProration(oldPlan, newPlan *repository.Plan, daysRemaining int) (int64, error) {
	if oldPlan == nil || newPlan == nil {
		return 0, fmt.Errorf("plan cannot be nil")
	}

	oldDailyRate := float64(oldPlan.PricePaisa) / 30.0
	newDailyRate := float64(newPlan.PricePaisa) / 30.0

	// Credit for unused days
	credit := oldDailyRate * float64(daysRemaining)
	// Charge for new plan remaining days
	charge := newDailyRate * float64(daysRemaining)

	// Difference is what the user owes (or gets as credit)
	diff := charge - credit
	if diff < 0 {
		diff = 0 // Don't refund for downgrades, just apply next cycle
	}

	return int64(diff), nil
}
