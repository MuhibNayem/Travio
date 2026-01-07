package consumer

import (
	"context"
	"encoding/json"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/notification/internal/service"
)

// EventConsumer consumes events and sends notifications
type EventConsumer struct {
	consumer            *kafka.Consumer
	notificationService *service.NotificationService
}

// NewEventConsumer creates a new notification event consumer
func NewEventConsumer(brokers []string, notificationSvc *service.NotificationService) (*EventConsumer, error) {
	// Subscribe to multiple topics
	topics := []string{
		kafka.TopicOrders,
		kafka.TopicFulfillment,
	}

	consumer, err := kafka.NewConsumer(brokers, "notification-service", topics)
	if err != nil {
		return nil, err
	}

	c := &EventConsumer{
		consumer:            consumer,
		notificationService: notificationSvc,
	}

	// Register handlers for different event types
	consumer.RegisterHandler(kafka.EventOrderConfirmed, c.handleOrderConfirmed)
	consumer.RegisterHandler(kafka.EventOrderCancelled, c.handleOrderCancelled)
	consumer.RegisterHandler(kafka.EventTicketGenerated, c.handleTicketGenerated)

	return c, nil
}

// OrderConfirmedPayload from order service
type OrderConfirmedPayload struct {
	OrderID      string `json:"order_id"`
	UserID       string `json:"user_id"`
	TripID       string `json:"trip_id"`
	BookingID    string `json:"booking_id"`
	TotalPaisa   int64  `json:"total_paisa"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}

// handleOrderConfirmed sends confirmation notification
func (c *EventConsumer) handleOrderConfirmed(ctx context.Context, event *kafka.Event) error {
	logger.Info("sending order confirmation notification", "event_id", event.ID)

	payloadBytes, _ := json.Marshal(event.Payload)
	var payload OrderConfirmedPayload
	json.Unmarshal(payloadBytes, &payload)

	// Send email notification
	err := c.notificationService.SendEmail(ctx, &service.EmailRequest{
		To:       payload.ContactEmail,
		Subject:  "Order Confirmed - Booking #" + payload.BookingID,
		Template: "order_confirmed",
		Data: map[string]interface{}{
			"order_id":   payload.OrderID,
			"booking_id": payload.BookingID,
			"total":      float64(payload.TotalPaisa) / 100,
		},
	})
	if err != nil {
		logger.Error("failed to send order confirmation email", "error", err)
		return err
	}

	// Send SMS notification
	if payload.ContactPhone != "" {
		c.notificationService.SendSMS(ctx, &service.SMSRequest{
			To:      payload.ContactPhone,
			Message: "Your booking #" + payload.BookingID + " is confirmed!",
		})
	}

	return nil
}

// OrderCancelledPayload from order service
type OrderCancelledPayload struct {
	OrderID      string `json:"order_id"`
	BookingID    string `json:"booking_id"`
	RefundAmount int64  `json:"refund_amount"`
	Reason       string `json:"reason"`
}

// handleOrderCancelled sends cancellation notification
func (c *EventConsumer) handleOrderCancelled(ctx context.Context, event *kafka.Event) error {
	logger.Info("sending order cancellation notification", "event_id", event.ID)

	payloadBytes, _ := json.Marshal(event.Payload)
	var payload OrderCancelledPayload
	json.Unmarshal(payloadBytes, &payload)

	// Get contact from event payload (simplified)
	email, _ := event.Payload["contact_email"].(string)
	phone, _ := event.Payload["contact_phone"].(string)

	if email != "" {
		c.notificationService.SendEmail(ctx, &service.EmailRequest{
			To:       email,
			Subject:  "Order Cancelled - #" + payload.OrderID,
			Template: "order_cancelled",
			Data: map[string]interface{}{
				"order_id":      payload.OrderID,
				"refund_amount": float64(payload.RefundAmount) / 100,
				"reason":        payload.Reason,
			},
		})
	}

	if phone != "" {
		c.notificationService.SendSMS(ctx, &service.SMSRequest{
			To:      phone,
			Message: "Order #" + payload.OrderID + " cancelled. Refund: " + formatMoney(payload.RefundAmount),
		})
	}

	return nil
}

// TicketGeneratedPayload from fulfillment service
type TicketGeneratedPayload struct {
	TicketID     string `json:"ticket_id"`
	OrderID      string `json:"order_id"`
	BookingID    string `json:"booking_id"`
	QRCodeURL    string `json:"qr_code_url"`
	ContactEmail string `json:"contact_email"`
	ContactPhone string `json:"contact_phone"`
}

// handleTicketGenerated sends ticket notification
func (c *EventConsumer) handleTicketGenerated(ctx context.Context, event *kafka.Event) error {
	logger.Info("sending ticket notification", "event_id", event.ID)

	payloadBytes, _ := json.Marshal(event.Payload)
	var payload TicketGeneratedPayload
	json.Unmarshal(payloadBytes, &payload)

	if payload.ContactEmail != "" {
		c.notificationService.SendEmail(ctx, &service.EmailRequest{
			To:       payload.ContactEmail,
			Subject:  "Your Ticket is Ready - #" + payload.TicketID,
			Template: "ticket_ready",
			Data: map[string]interface{}{
				"ticket_id":  payload.TicketID,
				"booking_id": payload.BookingID,
				"qr_url":     payload.QRCodeURL,
			},
		})
	}

	return nil
}

// Start begins consuming events
func (c *EventConsumer) Start() error {
	return c.consumer.Start()
}

// Stop stops the consumer
func (c *EventConsumer) Stop() error {
	return c.consumer.Stop()
}

func formatMoney(paisa int64) string {
	return "BDT " + string(rune(paisa/100))
}
