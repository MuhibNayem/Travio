package domain

import (
	"time"
)

// CheckoutSession represents a checkout session created from a seat hold
type CheckoutSession struct {
	ID             string             `json:"id" gorm:"primaryKey"`
	OrganizationID string             `json:"organization_id" gorm:"index"`
	UserID         string             `json:"user_id" gorm:"index"`
	TripID         string             `json:"trip_id"`
	HoldID         string             `json:"hold_id" gorm:"uniqueIndex"`
	FromStationID  string             `json:"from_station_id"`
	ToStationID    string             `json:"to_station_id"`

	// Passenger details (filled during checkout)
	Passengers     []CheckoutPassenger `json:"passengers" gorm:"foreignKey:CheckoutSessionID"`

	// Pricing
	BasePricePaisa    int64            `json:"base_price_paisa"`
	TaxPaisa          int64            `json:"tax_paisa"`
	BookingFeePaisa   int64            `json:"booking_fee_paisa"`
	DiscountPaisa     int64            `json:"discount_paisa"`
	TotalPaisa        int64            `json:"total_paisa"`
	Currency          string           `json:"currency"`

	// Coupon
	CouponCode        string           `json:"coupon_code"`
	CouponDiscount    int64            `json:"coupon_discount"`
	CouponValidated   bool             `json:"coupon_validated"`

	// Payment
	PaymentMethod     string           `json:"payment_method"`
	PaymentGateway    string           `json:"payment_gateway"`

	// Status
	Status            CheckoutStatus   `json:"status" gorm:"index"`

	// Order reference (set after checkout confirmation)
	OrderID           string           `json:"order_id" gorm:"index"`

	// Timestamps
	CreatedAt         time.Time        `json:"created_at"`
	UpdatedAt         time.Time        `json:"updated_at"`
	ExpiresAt         time.Time        `json:"expires_at"`
	CompletedAt       *time.Time       `json:"completed_at,omitempty"`
}

// CheckoutStatus represents the lifecycle of a checkout session
type CheckoutStatus string

const (
	CheckoutStatusPending    CheckoutStatus = "pending"
	CheckoutStatusInProgress CheckoutStatus = "in_progress"
	CheckoutStatusConfirmed  CheckoutStatus = "confirmed"
	CheckoutStatusFailed     CheckoutStatus = "failed"
	CheckoutStatusExpired    CheckoutStatus = "expired"
)

// CheckoutPassenger represents a passenger in a checkout session
type CheckoutPassenger struct {
	ID                    uint   `json:"id" gorm:"primaryKey"`
	CheckoutSessionID     string `json:"-" gorm:"index"`
	NID                   string `json:"nid"`
	Name                  string `json:"name"`
	SeatID                string `json:"seat_id"`
	SeatNumber            string `json:"seat_number"`
	SeatClass             string `json:"seat_class"`
	SeatType              string `json:"seat_type"`
	DateOfBirth           string `json:"date_of_birth"`
	Gender                string `json:"gender"`
	Age                   int    `json:"age"`
	Phone                 string `json:"phone"`
	Email                 string `json:"email"`
	EmergencyContactName  string `json:"emergency_contact_name"`
	EmergencyContactPhone string `json:"emergency_contact_phone"`
}

// PricingBreakdown represents the pricing calculation result
type PricingBreakdown struct {
	BasePricePaisa    int64            `json:"base_price_paisa"`
	TaxPaisa          int64            `json:"tax_paisa"`
	BookingFeePaisa   int64            `json:"booking_fee_paisa"`
	DiscountPaisa     int64            `json:"discount_paisa"`
	TotalPaisa        int64            `json:"total_paisa"`
	Currency          string           `json:"currency"`
	AppliedRules      []PricingRule    `json:"applied_rules,omitempty"`
	CouponApplied     bool             `json:"coupon_applied"`
	CouponDiscount    int64            `json:"coupon_discount"`
	CouponCode        string           `json:"coupon_code,omitempty"`
}

// PricingRule represents an applied pricing rule
type PricingRule struct {
	RuleID     string  `json:"rule_id"`
	RuleName   string  `json:"rule_name"`
	Multiplier float64 `json:"multiplier"`
}

// CouponValidation represents the result of coupon validation
type CouponValidation struct {
	Valid          bool   `json:"valid"`
	DiscountPaisa  int64  `json:"discount_paisa"`
	Message        string `json:"message"`
	CouponCode     string `json:"coupon_code"`
	DiscountType   string `json:"discount_type"`
	DiscountValue  float64 `json:"discount_value"`
}
