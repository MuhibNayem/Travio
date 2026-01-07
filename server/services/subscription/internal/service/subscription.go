package service

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/subscription/internal/repository"
)

type SubscriptionService struct {
	repo repository.Repository
}

func NewSubscriptionService(repo repository.Repository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// Plans
func (s *SubscriptionService) CreatePlan(ctx context.Context, name, description string, price int64, interval string, features map[string]string) (*repository.Plan, error) {
	plan := &repository.Plan{
		Name:        name,
		Description: description,
		PricePaisa:  price,
		Interval:    interval,
		Features:    features,
		IsActive:    true,
	}
	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *SubscriptionService) ListPlans(ctx context.Context, includeInactive bool) ([]*repository.Plan, error) {
	return s.repo.ListPlans(ctx, includeInactive)
}

func (s *SubscriptionService) GetPlan(ctx context.Context, id string) (*repository.Plan, error) {
	return s.repo.GetPlan(ctx, id)
}

func (s *SubscriptionService) UpdatePlan(ctx context.Context, id, name, description string, price int64, isActive bool, features map[string]string) (*repository.Plan, error) {
	plan, err := s.repo.GetPlan(ctx, id)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	if name != "" {
		plan.Name = name
	}
	if description != "" {
		plan.Description = description
	}
	if price != 0 {
		plan.PricePaisa = price
	}
	plan.IsActive = isActive
	if features != nil {
		plan.Features = features
	}

	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// Subscriptions
func (s *SubscriptionService) CreateSubscription(ctx context.Context, orgID, planID string) (*repository.Subscription, error) {
	// 1. Check if Plan exists
	plan, err := s.repo.GetPlan(ctx, planID)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return nil, errors.New("plan not found")
	}

	// 2. Check if Org already has active subscription
	existing, err := s.repo.GetSubscription(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("organization already has an active subscription")
	}

	// 3. Calculate Period
	now := time.Now()
	var end time.Time
	if plan.Interval == "month" {
		end = now.AddDate(0, 1, 0)
	} else if plan.Interval == "year" {
		end = now.AddDate(1, 0, 0)
	} else {
		// Default 30 days
		end = now.AddDate(0, 0, 30)
	}

	// 4. Create Subscription
	sub := &repository.Subscription{
		OrganizationID:     orgID,
		PlanID:             planID,
		Status:             "active", // Assume paid/active immediately for now
		CurrentPeriodStart: now,
		CurrentPeriodEnd:   end,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	// 5. Generate Initial Invoice
	invoice := &repository.Invoice{
		SubscriptionID: sub.ID,
		AmountPaisa:    plan.PricePaisa,
		Status:         "paid", // Pre-paid model assumption
		IssuedAt:       now,
		DueDate:        now,
		PaidAt:         &now,
	}
	// Best effort invoice creation (or use tx)
	_ = s.repo.CreateInvoice(ctx, invoice)

	return sub, nil
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, planID, status string) ([]*repository.Subscription, error) {
	return s.repo.ListSubscriptions(ctx, planID, status)
}

func (s *SubscriptionService) GetSubscription(ctx context.Context, orgID string) (*repository.Subscription, error) {
	return s.repo.GetSubscription(ctx, orgID)
}

func (s *SubscriptionService) CancelSubscription(ctx context.Context, orgID string) (*repository.Subscription, error) {
	if err := s.repo.CancelSubscription(ctx, orgID); err != nil {
		return nil, err
	}
	return nil, nil
}

// Billing
func (s *SubscriptionService) ListInvoices(ctx context.Context, subscriptionID string) ([]*repository.Invoice, error) {
	return s.repo.ListInvoices(ctx, subscriptionID)
}
