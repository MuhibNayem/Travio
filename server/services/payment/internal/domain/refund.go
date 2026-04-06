package domain

import "time"

// Refund represents a payment refund record
type Refund struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	TransactionID   string    `json:"transaction_id" gorm:"index"`
	OrderID         string    `json:"order_id" gorm:"index"`
	OrganizationID  string    `json:"organization_id"`
	Gateway         string    `json:"gateway"`
	GatewayRefundID string    `json:"gateway_refund_id"`
	AmountPaisa     int64     `json:"amount_paisa"`
	Currency        string    `json:"currency"`
	Reason          string    `json:"reason"`
	Status          string    `json:"status"` // PENDING, PROCESSING, COMPLETED, FAILED
	FailureReason   string    `json:"failure_reason,omitempty"`
	RefundedAt      time.Time `json:"refunded_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// RefundStatus constants
const (
	RefundStatusPending     = "pending"
	RefundStatusProcessing  = "processing"
	RefundStatusCompleted   = "completed"
	RefundStatusFailed      = "failed"
)
