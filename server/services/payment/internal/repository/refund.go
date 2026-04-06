package repository

import (
	"context"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/domain"
	"gorm.io/gorm"
)

var ErrRefundNotFound = errors.New("refund not found")

// RefundRepository handles refund persistence
type RefundRepository struct {
	db *gorm.DB
}

// NewRefundRepository creates a new refund repository
func NewRefundRepository(db *gorm.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

// Create creates a new refund record
func (r *RefundRepository) Create(ctx context.Context, refund *domain.Refund) error {
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(refund).Error
}

// GetByID retrieves a refund by ID
func (r *RefundRepository) GetByID(ctx context.Context, id string) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&refund).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return &refund, nil
}

// GetByOrderID retrieves a refund by order ID
func (r *RefundRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Refund, error) {
	var refund domain.Refund
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).First(&refund).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return &refund, nil
}

// GetByTransactionID retrieves refunds by transaction ID (multiple refunds possible)
func (r *RefundRepository) GetByTransactionID(ctx context.Context, transactionID string) ([]*domain.Refund, error) {
	var refunds []*domain.Refund
	err := r.db.WithContext(ctx).
		Where("transaction_id = ?", transactionID).
		Order("created_at DESC").
		Find(&refunds).Error
	return refunds, err
}

// UpdateStatus updates the status of a refund
func (r *RefundRepository) UpdateStatus(ctx context.Context, id, status, failureReason string) error {
	updates := map[string]interface{}{
		"status":   status,
		"updated_at": time.Now(),
	}
	if failureReason != "" {
		updates["failure_reason"] = failureReason
	}
	if status == domain.RefundStatusCompleted {
		updates["refunded_at"] = time.Now()
	}
	return r.db.WithContext(ctx).
		Model(&domain.Refund{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListByOrganization lists refunds for an organization
func (r *RefundRepository) ListByOrganization(ctx context.Context, orgID string, limit, offset int) ([]*domain.Refund, int64, error) {
	var refunds []*domain.Refund
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.Refund{}).Where("organization_id = ?", orgID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&refunds).Error; err != nil {
		return nil, 0, err
	}

	return refunds, total, nil
}

// AutoMigrate creates/updates the refund table
func (r *RefundRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.Refund{})
}
