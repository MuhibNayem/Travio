package worker

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/gateway"
	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
)

type Reconciler struct {
	repo       *repository.TransactionRepository
	configRepo *repository.PaymentConfigRepository
	registry   *gateway.Registry
	interval   time.Duration
}

func NewReconciler(repo *repository.TransactionRepository, configRepo *repository.PaymentConfigRepository, registry *gateway.Registry, interval time.Duration) *Reconciler {
	return &Reconciler{
		repo:       repo,
		configRepo: configRepo,
		registry:   registry,
		interval:   interval,
	}
}

func (r *Reconciler) Start(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	logger.Info("Starting Payment Reconciler", "interval", r.interval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping Payment Reconciler")
			return
		case <-ticker.C:
			r.reconcile(ctx)
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) {
	// Find transactions pending for more than 5 minutes
	txs, err := r.repo.FindsPending(ctx, 5)
	if err != nil {
		logger.Error("Reconciler failed to fetch pending transactions", "error", err)
		return
	}

	if len(txs) > 0 {
		logger.Info("Reconciler found pending transactions", "count", len(txs))
	}

	for _, tx := range txs {
		if tx.GatewayTxID == "" {
			// No Gateway Session ID, maybe failed before gateway call or strict pending?
			// If too old, mark failed?
			if time.Since(tx.CreatedAt) > 30*time.Minute {
				_ = r.repo.UpdateStatus(ctx, tx.ID, "FAILED", "")
			}
			continue
		}

		providerName := r.registry.ResolveProvider(tx.Gateway)
		factory, err := r.registry.GetFactory(providerName)
		if err != nil {
			logger.Error("Reconciler unknown gateway", "gateway", tx.Gateway, "tx_id", tx.ID)
			continue
		}

		// Load Config
		payConfig, err := r.configRepo.GetConfig(ctx, tx.OrganizationID)
		if err != nil {
			logger.Error("Reconciler failed to get config", "org_id", tx.OrganizationID, "error", err)
			continue
		}

		// Create Gateway
		gw, err := factory.Create(payConfig.Credentials, false)
		if err != nil {
			logger.Error("Reconciler failed to create gateway", "gateway", tx.Gateway, "error", err)
			continue
		}

		// Verify with Gateway
		status, err := gw.VerifyPayment(ctx, tx.GatewayTxID)
		if err != nil {
			logger.Error("Reconciler failed to verify payment", "tx_id", tx.ID, "error", err)
			continue
		}

		// Mapping generic gateway status to our internal status
		// Assuming status.Status returns strings like "VALID", "SUCCESS", "FAILED"
		// This mapping depends on gateway impl.
		internalStatus := "PENDING"
		switch status.Status {
		case "VALID", "SUCCESS", "COMPLETED":
			internalStatus = "SUCCESS"
		case "FAILED", "CANCELLED", "EXPIRED":
			internalStatus = "FAILED"
		case "PENDING", "INITIATED":
			internalStatus = "PENDING"
		default:
			logger.Info("Reconciler unknown gateway status", "status", status.Status) // Warn not available
			internalStatus = "PENDING"
		}

		if internalStatus != "PENDING" {
			logger.Info("Reconciler updating status", "tx_id", tx.ID, "old", "PENDING", "new", internalStatus)
			if err := r.repo.UpdateStatus(ctx, tx.ID, internalStatus, ""); err != nil {
				logger.Error("Reconciler failed to update DB", "tx_id", tx.ID, "error", err)
			}
		}
	}
}
