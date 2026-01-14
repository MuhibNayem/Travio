package events

import (
	"context"
	"database/sql"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/outbox"
	"github.com/MuhibNayem/Travio/server/services/order/internal/domain"
)

// Publisher handles event publishing for order domain events
type Publisher struct {
	outbox *outbox.Publisher
}

// NewPublisher creates a new event publisher
func NewPublisher(db *sql.DB) *Publisher {
	return &Publisher{
		outbox: outbox.NewPublisher(db),
	}
}

// OrderCreatedPayload is the event payload for order created
type OrderCreatedPayload struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	TripID         string `json:"trip_id"`
	RouteID        string `json:"route_id"`
	FromStationID  string `json:"from_station_id"`
	ToStationID    string `json:"to_station_id"`
	HoldID         string `json:"hold_id"`
	TotalPaisa     int64  `json:"total_paisa"`
	Currency       string `json:"currency"`
	ContactEmail   string `json:"contact_email"`
	ContactPhone   string `json:"contact_phone"`
	PassengerCount int    `json:"passenger_count"`
}

// OrderConfirmedPayload is the event payload for order confirmed
type OrderConfirmedPayload struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	TripID         string `json:"trip_id"`
	BookingID      string `json:"booking_id"`
	PaymentID      string `json:"payment_id"`
	TotalPaisa     int64  `json:"total_paisa"`
	ContactEmail   string `json:"contact_email"`
	ContactPhone   string `json:"contact_phone"`
}

// OrderCancelledPayload is the event payload for order cancelled
type OrderCancelledPayload struct {
	OrderID        string `json:"order_id"`
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	BookingID      string `json:"booking_id"`
	PaymentID      string `json:"payment_id"`
	RefundID       string `json:"refund_id"`
	RefundAmount   int64  `json:"refund_amount"`
	Reason         string `json:"reason"`
}

// OrderFailedPayload is the event payload for order failed
type OrderFailedPayload struct {
	OrderID   string `json:"order_id"`
	UserID    string `json:"user_id"`
	Reason    string `json:"reason"`
	SagaState string `json:"saga_state"`
}

// PublishOrderCreated publishes order created event within a transaction
func (p *Publisher) PublishOrderCreated(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	payload := OrderCreatedPayload{
		OrderID:        order.ID,
		UserID:         order.UserID,
		OrganizationID: order.OrganizationID,
		TripID:         order.TripID,
		RouteID:        order.RouteID,
		FromStationID:  order.FromStationID,
		ToStationID:    order.ToStationID,
		HoldID:         order.HoldID,
		TotalPaisa:     order.TotalPaisa,
		Currency:       order.Currency,
		ContactEmail:   order.ContactEmail,
		ContactPhone:   order.ContactPhone,
		PassengerCount: len(order.Passengers),
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicOrders, kafka.EventOrderCreated, order.ID, payload)
}

// PublishOrderConfirmed publishes order confirmed event within a transaction
func (p *Publisher) PublishOrderConfirmed(ctx context.Context, tx *sql.Tx, order *domain.Order) error {
	payload := OrderConfirmedPayload{
		OrderID:        order.ID,
		UserID:         order.UserID,
		OrganizationID: order.OrganizationID,
		TripID:         order.TripID,
		BookingID:      order.BookingID,
		PaymentID:      order.PaymentID,
		TotalPaisa:     order.TotalPaisa,
		ContactEmail:   order.ContactEmail,
		ContactPhone:   order.ContactPhone,
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicOrders, kafka.EventOrderConfirmed, order.ID, payload)
}

// PublishOrderCancelled publishes order cancelled event within a transaction
func (p *Publisher) PublishOrderCancelled(ctx context.Context, tx *sql.Tx, order *domain.Order, refundID string, refundAmount int64, reason string) error {
	payload := OrderCancelledPayload{
		OrderID:        order.ID,
		UserID:         order.UserID,
		OrganizationID: order.OrganizationID,
		BookingID:      order.BookingID,
		PaymentID:      order.PaymentID,
		RefundID:       refundID,
		RefundAmount:   refundAmount,
		Reason:         reason,
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicOrders, kafka.EventOrderCancelled, order.ID, payload)
}

// PublishOrderFailed publishes order failed event within a transaction
func (p *Publisher) PublishOrderFailed(ctx context.Context, tx *sql.Tx, order *domain.Order, reason, sagaState string) error {
	payload := OrderFailedPayload{
		OrderID:   order.ID,
		UserID:    order.UserID,
		Reason:    reason,
		SagaState: sagaState,
	}
	return p.outbox.Publish(ctx, tx, kafka.TopicOrders, kafka.EventOrderFailed, order.ID, payload)
}
