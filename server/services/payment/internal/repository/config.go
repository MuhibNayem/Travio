package repository

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type PaymentConfig struct {
	OrganizationID string          `gorm:"primaryKey;type:uuid"`
	Gateway        string          `gorm:"not null"`
	Credentials    json.RawMessage `gorm:"type:jsonb;not null"`
	IsSandbox      bool            `gorm:"default:true"`
	IsActive       bool            `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type PaymentConfigRepository struct {
	db *gorm.DB
}

func NewPaymentConfigRepository(db *gorm.DB) *PaymentConfigRepository {
	return &PaymentConfigRepository{db: db}
}

func (r *PaymentConfigRepository) SaveConfig(ctx context.Context, config *PaymentConfig) error {
	return r.db.WithContext(ctx).Save(config).Error
}

func (r *PaymentConfigRepository) GetConfig(ctx context.Context, orgID string) (*PaymentConfig, error) {
	var config PaymentConfig
	err := r.db.WithContext(ctx).Where("organization_id = ? AND is_active = ?", orgID, true).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}
