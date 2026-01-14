package service

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/repository"
)

type ReconciliationService struct {
	repo *repository.TransactionRepository
}

func NewReconciliationService(repo *repository.TransactionRepository) *ReconciliationService {
	return &ReconciliationService{repo: repo}
}

type ReconciliationReport struct {
	OrganizationID   string
	Date             string
	TotalCollected   int64
	TotalRefunded    int64
	NetAmount        int64
	TransactionCount int
	RefundCount      int
}

func (s *ReconciliationService) GenerateReport(ctx context.Context, orgID string, date time.Time) (*ReconciliationReport, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	txs, err := s.repo.GetTransactionsByDateRange(ctx, orgID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	report := &ReconciliationReport{
		OrganizationID: orgID,
		Date:           date.Format("2006-01-02"),
	}

	for _, tx := range txs {
		// Transaction status is SUCCESS at this point (filtered in repo)
		// We need to differentiate PAYMENT vs REFUND.
		// Current model doesn't have a 'type' field.
		// Convention: Refunds might have negative amount or a separate status.
		// For now, assume all are payments (positive). Refunds would need schema update.
		// TODO: Add 'tx_type' (PAYMENT, REFUND) to Transaction model for proper reconciliation.

		if tx.Amount > 0 {
			report.TotalCollected += tx.Amount
			report.TransactionCount++
		} else {
			// Negative amounts treated as refunds (if that's the convention)
			report.TotalRefunded += -tx.Amount
			report.RefundCount++
		}
	}

	report.NetAmount = report.TotalCollected - report.TotalRefunded
	return report, nil
}
