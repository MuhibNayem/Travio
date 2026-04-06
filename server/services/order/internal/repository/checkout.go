package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
	"gorm.io/gorm"
)

// CheckoutRepository handles checkout session persistence
type CheckoutRepository struct {
	db *gorm.DB
}

// NewCheckoutRepository creates a new checkout repository
func NewCheckoutRepository(db *gorm.DB) *CheckoutRepository {
	return &CheckoutRepository{db: db}
}

// Create creates a new checkout session
func (r *CheckoutRepository) Create(ctx context.Context, session *domain.CheckoutSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a checkout session by ID
func (r *CheckoutRepository) GetByID(ctx context.Context, id, userID string) (*domain.CheckoutSession, error) {
	var session domain.CheckoutSession
	err := r.db.WithContext(ctx).
		Preload("Passengers").
		Where("id = ? AND user_id = ?", id, userID).
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("checkout session not found: %w", err)
		}
		return nil, err
	}
	return &session, nil
}

// GetByHoldID retrieves a checkout session by hold ID
func (r *CheckoutRepository) GetByHoldID(ctx context.Context, holdID, userID string) (*domain.CheckoutSession, error) {
	var session domain.CheckoutSession
	err := r.db.WithContext(ctx).
		Preload("Passengers").
		Where("hold_id = ? AND user_id = ?", holdID, userID).
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Not found is not an error for this use case
		}
		return nil, err
	}
	return &session, nil
}

// Update updates a checkout session
func (r *CheckoutRepository) Update(ctx context.Context, session *domain.CheckoutSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// UpdateStatus updates the status of a checkout session
func (r *CheckoutRepository) UpdateStatus(ctx context.Context, id string, status domain.CheckoutStatus) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}
	if status == domain.CheckoutStatusConfirmed {
		now := time.Now()
		updates["completed_at"] = now
	}
	return r.db.WithContext(ctx).
		Model(&domain.CheckoutSession{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// UpdateOrderID sets the order ID on a checkout session
func (r *CheckoutRepository) UpdateOrderID(ctx context.Context, id, orderID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.CheckoutSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"order_id":   orderID,
			"status":     domain.CheckoutStatusConfirmed,
			"completed_at": time.Now(),
			"updated_at": time.Now(),
		}).Error
}

// ListByUser retrieves checkout sessions for a user with pagination
func (r *CheckoutRepository) ListByUser(ctx context.Context, userID, status string, limit, offset int) ([]*domain.CheckoutSession, int64, error) {
	var sessions []*domain.CheckoutSession
	var total int64

	query := r.db.WithContext(ctx).Model(&domain.CheckoutSession{}).Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.
		Preload("Passengers").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

// ExpireOldSessions marks expired checkout sessions as expired
func (r *CheckoutRepository) ExpireOldSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&domain.CheckoutSession{}).
		Where("status = ? AND expires_at < ?", domain.CheckoutStatusPending, time.Now()).
		Updates(map[string]interface{}{
			"status":     domain.CheckoutStatusExpired,
			"updated_at": time.Now(),
		}).Error
}

// DeleteExpired deletes expired checkout sessions older than the given duration
func (r *CheckoutRepository) DeleteExpired(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	return r.db.WithContext(ctx).
		Where("expires_at < ?", cutoff).
		Delete(&domain.CheckoutSession{}).Error
}

// DeletePassengers deletes all passengers for a checkout session
func (r *CheckoutRepository) DeletePassengers(ctx context.Context, sessionID string) error {
	return r.db.WithContext(ctx).
		Where("checkout_session_id = ?", sessionID).
		Delete(&domain.CheckoutPassenger{}).Error
}

// AutoMigrate creates/updates the checkout tables
func (r *CheckoutRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&domain.CheckoutSession{},
		&domain.CheckoutPassenger{},
	)
}
