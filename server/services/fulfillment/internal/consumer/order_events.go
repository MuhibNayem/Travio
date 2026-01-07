package consumer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/fulfillment/internal/service"
)

// OrderEventConsumer consumes order events and triggers ticket generation
type OrderEventConsumer struct {
	consumer           *kafka.Consumer
	fulfillmentService *service.FulfillmentService
}

// NewOrderEventConsumer creates a new consumer for order events
func NewOrderEventConsumer(brokers []string, fulfillmentSvc *service.FulfillmentService) (*OrderEventConsumer, error) {
	consumer, err := kafka.NewConsumer(brokers, "fulfillment-service", []string{kafka.TopicOrders})
	if err != nil {
		return nil, err
	}

	c := &OrderEventConsumer{
		consumer:           consumer,
		fulfillmentService: fulfillmentSvc,
	}

	// Register handlers
	consumer.RegisterHandler(kafka.EventOrderConfirmed, c.handleOrderConfirmed)

	return c, nil
}

// OrderConfirmedPayload matches the event structure from order service
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

// handleOrderConfirmed processes OrderConfirmed events
func (c *OrderEventConsumer) handleOrderConfirmed(ctx context.Context, event *kafka.Event) error {
	logger.Info("received OrderConfirmed event",
		"event_id", event.ID,
		"order_id", event.AggregateID,
	)

	// Parse payload
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		logger.Error("failed to marshal payload", "error", err)
		return err
	}

	var payload OrderConfirmedPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		logger.Error("failed to unmarshal OrderConfirmed payload", "error", err)
		return err
	}

	// In a real implementation, we would fetch trip/passenger details from catalog
	// For now, create a placeholder request
	req := &service.GenerateTicketsReq{
		BookingID:      payload.BookingID,
		OrderID:        payload.OrderID,
		OrganizationID: payload.OrganizationID,
		TripID:         payload.TripID,
		RouteName:      "Express Route", // Would come from catalog
		FromStation:    "Origin",        // Would come from catalog
		ToStation:      "Destination",   // Would come from catalog
		DepartureTime:  time.Now().Add(24 * time.Hour),
		ArrivalTime:    time.Now().Add(28 * time.Hour),
		Passengers: []service.PassengerSeat{
			{
				NID:        "PLACEHOLDER",
				Name:       "Passenger",
				SeatID:     "seat-1",
				SeatNumber: "A1",
				SeatClass:  "AC",
				PricePaisa: payload.TotalPaisa,
			},
		},
		ContactEmail: payload.ContactEmail,
		ContactPhone: payload.ContactPhone,
	}

	// Generate tickets
	result, err := c.fulfillmentService.GenerateTickets(ctx, req)
	if err != nil {
		logger.Error("failed to generate tickets",
			"order_id", payload.OrderID,
			"error", err,
		)
		return err
	}

	logger.Info("tickets generated successfully",
		"order_id", payload.OrderID,
		"ticket_count", len(result.Tickets),
	)

	return nil
}

// Start begins consuming events
func (c *OrderEventConsumer) Start() error {
	return c.consumer.Start()
}

// Stop stops the consumer
func (c *OrderEventConsumer) Stop() error {
	return c.consumer.Stop()
}
