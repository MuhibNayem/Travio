package repository

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/model"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// CreateIdempotent atomically creates a pending transaction if it doesn't exist.
// Returns existing transaction if found (Idempotency Hit).
func (r *TransactionRepository) CreateIdempotent(ctx context.Context, tx *model.Transaction) (*model.Transaction, bool, error) {
	var existing model.Transaction

	// Check by Idempotency Key first
	err := r.db.WithContext(ctx).Where("idempotency_key = ?", tx.IdempotencyKey).First(&existing).Error
	if err == nil {
		return &existing, true, nil // Already exists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, err // DB Error
	}

	// Create new
	if err := r.db.WithContext(ctx).Create(tx).Error; err != nil {
		// Fallback check for race condition unique violation
		if r.db.WithContext(ctx).Where("idempotency_key = ?", tx.IdempotencyKey).First(&existing).Error == nil {
			return &existing, true, nil
		}
		return nil, false, err
	}

	return tx, false, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id, status, gatewayTxID string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": r.db.NowFunc(),
	}
	if gatewayTxID != "" {
		updates["gateway_tx_id"] = gatewayTxID
	}
	return r.db.WithContext(ctx).Model(&model.Transaction{}).Where("id = ?", id).Updates(updates).Error
}

func (r *TransactionRepository) FindsPending(ctx context.Context, olderThanMinutes int) ([]model.Transaction, error) {
	var txs []model.Transaction
	cutoff := time.Now().Add(time.Duration(-olderThanMinutes) * time.Minute)

	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "PENDING", cutoff).
		Find(&txs).Error
	return txs, err
}
