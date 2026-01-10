package repository

import (
	"context"

	"github.com/MuhibNayem/Travio/server/services/payment/internal/model"
	"gorm.io/gorm"
)

type PaymentConfigRepository struct {
	db *gorm.DB
}

func NewPaymentConfigRepository(db *gorm.DB) *PaymentConfigRepository {
	return &PaymentConfigRepository{db: db}
}

func (r *PaymentConfigRepository) SaveConfig(ctx context.Context, config *model.PaymentConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *PaymentConfigRepository) GetConfig(ctx context.Context, orgID string) (*model.PaymentConfig, error) {
	var config model.PaymentConfig
	err := r.db.WithContext(ctx).Where("organization_id = ? AND is_active = ?", orgID, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
