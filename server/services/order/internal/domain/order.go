package domain

import (
	"time"
)

// Order represents a booking order
type Order struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
	TripID         string `json:"trip_id"`
	FromStationID  string `json:"from_station_id"`
	ToStationID    string `json:"to_station_id"`

	// Passengers
	Passengers []OrderPassenger `json:"passengers"`

	// Pricing
	SubtotalPaisa   int64  `json:"subtotal_paisa"`
	TaxPaisa        int64  `json:"tax_paisa"`
	BookingFeePaisa int64  `json:"booking_fee_paisa"`
	DiscountPaisa   int64  `json:"discount_paisa"`
	TotalPaisa      int64  `json:"total_paisa"`
	Currency        string `json:"currency"`

	// Payment
	PaymentID     string `json:"payment_id"`
	PaymentStatus string `json:"payment_status"`
	PaymentMethod string `json:"payment_method"`

	// Booking
	BookingID string       `json:"booking_id"`
	HoldID    string       `json:"hold_id"`
	Seats     []BookedSeat `json:"seats"`

	// Status
	Status OrderStatus `json:"status"`
	SagaID string      `json:"saga_id"`

	// Contact
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ExpiresAt time.Time `json:"expires_at"`

	// Idempotency
	IdempotencyKey string `json:"idempotency_key"`
}

type OrderStatus string

const (
	OrderStatusPending       OrderStatus = "pending"
	OrderStatusConfirmed     OrderStatus = "confirmed"
	OrderStatusFailed        OrderStatus = "failed"
	OrderStatusCancelled     OrderStatus = "cancelled"
	OrderStatusExpired       OrderStatus = "expired"
	OrderStatusRefundPending OrderStatus = "refund_pending"
	OrderStatusRefunded      OrderStatus = "refunded"
)

// OrderPassenger is a passenger on an order (distinct from anti-scalp Passenger)
type OrderPassenger struct {
	NID         string `json:"nid"`
	Name        string `json:"name"`
	SeatID      string `json:"seat_id"`
	SeatNumber  string `json:"seat_number"`
	SeatClass   string `json:"seat_class"`
	Gender      string `json:"gender"`
	Age         int    `json:"age"`
	NIDVerified bool   `json:"nid_verified"`
}

type BookedSeat struct {
	SeatID     string `json:"seat_id"`
	SeatNumber string `json:"seat_number"`
	SeatClass  string `json:"seat_class"`
	TicketID   string `json:"ticket_id"`
	PricePaisa int64  `json:"price_paisa"`
}

// PaymentStatus constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusAuthorized = "authorized"
	PaymentStatusCaptured   = "captured"
	PaymentStatusFailed     = "failed"
	PaymentStatusRefunded   = "refunded"
)

// CalculateTotals calculates order totals
func (o *Order) CalculateTotals(basePrices map[string]int64, taxRate float64, bookingFee int64) {
	o.SubtotalPaisa = 0
	for _, p := range o.Passengers {
		if price, ok := basePrices[p.SeatID]; ok {
			o.SubtotalPaisa += price
		}
	}

	o.TaxPaisa = int64(float64(o.SubtotalPaisa) * taxRate)
	o.BookingFeePaisa = bookingFee * int64(len(o.Passengers))
	o.TotalPaisa = o.SubtotalPaisa + o.TaxPaisa + o.BookingFeePaisa - o.DiscountPaisa

	if o.TotalPaisa < 0 {
		o.TotalPaisa = 0
	}
}
