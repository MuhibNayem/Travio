package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Transaction struct {
	ID             string `gorm:"primaryKey;type:uuid"`
	OrganizationID string `gorm:"type:uuid;not null"`
	OrderID        string `gorm:"uniqueIndex:idx_order_attempt;not null"`
	Attempt        int    `gorm:"uniqueIndex:idx_order_attempt;default:1"`
	Amount         int64  `gorm:"not null"`
	Currency       string `gorm:"size:3;not null"`
	Gateway        string `gorm:"size:50;not null"`
	GatewayTxID    string `gorm:"index"`                  // External Transaction ID from Gateway
	Status         string `gorm:"size:20;index;not null"` // PENDING, SUCCESS, FAILED
	IdempotencyKey string `gorm:"uniqueIndex;type:uuid"`  // Deterministic Key
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return
}
