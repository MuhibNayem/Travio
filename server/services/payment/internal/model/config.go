package model

import (
	"encoding/json"
	"time"
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
