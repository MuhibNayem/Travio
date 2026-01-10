package domain

import (
	"time"
)

type DiscountType int32

const (
	DiscountTypeUnspecified DiscountType = 0
	DiscountTypePercentage  DiscountType = 1
	DiscountTypeFixedAmount DiscountType = 2
)

type Coupon struct {
	ID                string       `json:"id"`
	OrganizationID    string       `json:"organization_id"`
	Code              string       `json:"code"`
	DiscountType      DiscountType `json:"discount_type"`
	DiscountValue     float64      `json:"discount_value"`
	MinPurchaseAmount int64        `json:"min_purchase_amount"` // In smallest currency unit (e.g. paisa)
	MaxDiscountAmount int64        `json:"max_discount_amount"`
	StartDate         time.Time    `json:"start_date"`
	EndDate           time.Time    `json:"end_date"`
	UsageLimit        int32        `json:"usage_limit"`
	UsageCount        int32        `json:"usage_count"`
	IsActive          bool         `json:"is_active"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

type SupportTicket struct {
	ID             string    `json:"id"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Subject        string    `json:"subject"`
	Status         string    `json:"status"`   // OPEN, CLOSED, PENDING
	Priority       string    `json:"priority"` // LOW, MEDIUM, HIGH
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TicketMessage struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	SenderID  string    `json:"sender_id"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
