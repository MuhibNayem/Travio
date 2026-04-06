package consumer

import (
	"context"
	"encoding/json"
	"fmt"
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
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	OrganizationID string    `json:"organization_id"`
	TripID         string    `json:"trip_id"`
	FromStationID  string    `json:"from_station_id"`
	ToStationID    string    `json:"to_station_id"`
	BookingID      string    `json:"booking_id"`
	PaymentID      string    `json:"payment_id"`
	TotalPaisa     int64     `json:"total_paisa"`
	Currency       string    `json:"currency"`
	ContactEmail   string    `json:"contact_email"`
	ContactPhone   string    `json:"contact_phone"`
	Passengers     []PassengerInfo `json:"passengers"`
	RouteName      string    `json:"route_name"`
	DepartureTime  string    `json:"departure_time"`
	ArrivalTime    string    `json:"arrival_time"`
	VehicleType    string    `json:"vehicle_type"`
}

// PassengerInfo contains passenger details from order
type PassengerInfo struct {
	NID        string `json:"nid"`
	Name       string `json:"name"`
	SeatID     string `json:"seat_id"`
	SeatNumber string `json:"seat_number"`
	SeatClass  string `json:"seat_class"`
	Gender     string `json:"gender"`
	Age        int    `json:"age"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
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

	// Validate passenger data
	if len(payload.Passengers) == 0 {
		logger.Error("Order confirmed with no passengers", "order_id", payload.OrderID)
		return fmt.Errorf("order %s has no passengers", payload.OrderID)
	}

	// Parse departure/arrival times
	var departureTime, arrivalTime time.Time
	if payload.DepartureTime != "" {
		departureTime, _ = time.Parse(time.RFC3339, payload.DepartureTime)
	}
	if departureTime.IsZero() {
		departureTime = time.Now().Add(24 * time.Hour)
	}

	if payload.ArrivalTime != "" {
		arrivalTime, _ = time.Parse(time.RFC3339, payload.ArrivalTime)
	}
	if arrivalTime.IsZero() {
		arrivalTime = departureTime.Add(4 * time.Hour)
	}

	// Build ticket generation request with real passenger data
	req := &service.GenerateTicketsReq{
		BookingID:      payload.BookingID,
		OrderID:        payload.OrderID,
		OrganizationID: payload.OrganizationID,
		TripID:         payload.TripID,
		RouteName:      payload.RouteName,
		FromStation:    payload.FromStationID,
		ToStation:      payload.ToStationID,
		DepartureTime:  departureTime,
		ArrivalTime:    arrivalTime,
		ContactEmail:   payload.ContactEmail,
		ContactPhone:   payload.ContactPhone,
	}

	for _, p := range payload.Passengers {
		req.Passengers = append(req.Passengers, service.PassengerSeat{
			NID:        p.NID,
			Name:       p.Name,
			SeatID:     p.SeatID,
			SeatNumber: p.SeatNumber,
			SeatClass:  p.SeatClass,
			PricePaisa: payload.TotalPaisa / int64(len(payload.Passengers)),
		})
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
		"passengers", len(req.Passengers),
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
