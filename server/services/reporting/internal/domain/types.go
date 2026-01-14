// Package domain defines reporting types.
package domain

import "time"

// Event represents a generic analytics event.
type Event struct {
	EventID        string    `json:"event_id"`
	EventType      string    `json:"event_type"`
	OrganizationID string    `json:"organization_id"`
	UserID         string    `json:"user_id"`
	Timestamp      time.Time `json:"timestamp"`

	// Optional fields based on event type
	OrderID   string `json:"order_id,omitempty"`
	PaymentID string `json:"payment_id,omitempty"`
	TripID    string `json:"trip_id,omitempty"`
	RouteID   string `json:"route_id,omitempty"`

	AmountPaisa int64  `json:"amount_paisa"`
	Status      string `json:"status"`
	Metadata    string `json:"metadata"` // JSON string
}

// EventType constants
const (
	EventOrderCreated     = "order.created"
	EventOrderCompleted   = "order.confirmed" // Mapped from order.confirmed
	EventOrderCancelled   = "order.cancelled"
	EventPaymentInitiated = "payment.authorized"
	EventPaymentCompleted = "payment.captured"
	EventPaymentFailed    = "payment.failed"
	EventPaymentRefunded  = "payment.refunded"
	EventTripCreated      = "trip.created" // Added
	// Legacy or Internal ?
	EventTicketGenerated = "fulfillment.ticket_generated"
	EventTicketScanned   = "ticket_scanned"
	EventSeatHeld        = "inventory.seats_held"
	EventSeatReleased    = "inventory.seats_released"
	EventSeatBooked      = "inventory.seats_booked"
)

// RevenueReport represents aggregated revenue data.
type RevenueReport struct {
	OrganizationID    string    `json:"organization_id"`
	Date              time.Time `json:"date"`
	OrderCount        int64     `json:"order_count"`
	TotalRevenuePaisa int64     `json:"total_revenue_paisa"`
	AvgOrderValue     float64   `json:"avg_order_value"`
	Currency          string    `json:"currency"`
}

// BookingTrend represents booking trends over time.
type BookingTrend struct {
	OrganizationID string    `json:"organization_id"`
	Period         time.Time `json:"period"` // Hour/Day/Week start
	BookingCount   int64     `json:"booking_count"`
	CompletedCount int64     `json:"completed_count"`
	CancelledCount int64     `json:"cancelled_count"`
	ConversionRate float64   `json:"conversion_rate"`
}

// TopRoute represents a popular route.
type TopRoute struct {
	OrganizationID string  `json:"organization_id"`
	TripID         string  `json:"trip_id"`
	RouteName      string  `json:"route_name"`
	BookingCount   int64   `json:"booking_count"`
	Revenue        int64   `json:"revenue"`
	AvgOccupancy   float64 `json:"avg_occupancy"`
}

// OrganizationMetrics represents overall metrics for an organization.
type OrganizationMetrics struct {
	OrganizationID     string  `json:"organization_id"`
	TotalOrders        int64   `json:"total_orders"`
	TotalRevenue       int64   `json:"total_revenue"`
	AvgOrderValue      float64 `json:"avg_order_value"`
	TotalCustomers     int64   `json:"total_customers"`
	RepeatCustomerRate float64 `json:"repeat_customer_rate"`
	AvgBookingsPerDay  float64 `json:"avg_bookings_per_day"`
	CancellationRate   float64 `json:"cancellation_rate"`
	RefundRate         float64 `json:"refund_rate"`
}

// ReportQuery represents query parameters for reports.
type ReportQuery struct {
	OrganizationID string
	StartDate      time.Time
	EndDate        time.Time
	Granularity    string // "hour", "day", "week", "month"
	Limit          int
	Offset         int
	SortBy         string
	SortOrder      string // "asc", "desc"
}

// ExportFormat represents export file formats.
type ExportFormat string

const (
	ExportCSV     ExportFormat = "csv"
	ExportJSON    ExportFormat = "json"
	ExportParquet ExportFormat = "parquet"
)
