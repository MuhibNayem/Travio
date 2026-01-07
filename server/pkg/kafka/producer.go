package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// Event represents a domain event
type Event struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	AggregateID string                 `json:"aggregate_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Version     int                    `json:"version"`
	Payload     map[string]interface{} `json:"payload"`
}

// Producer publishes events to Kafka
type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new Kafka producer
func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{producer: producer}, nil
}

// Publish sends an event to the specified topic
func (p *Producer) Publish(ctx context.Context, topic string, event *Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Key:       sarama.StringEncoder(event.AggregateID),
		Value:     sarama.ByteEncoder(data),
		Timestamp: event.Timestamp,
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_type"), Value: []byte(event.Type)},
			{Key: []byte("event_id"), Value: []byte(event.ID)},
		},
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	logger.Info("event published",
		"topic", topic,
		"partition", partition,
		"offset", offset,
		"event_type", event.Type,
		"aggregate_id", event.AggregateID,
	)

	return nil
}

// Close closes the producer
func (p *Producer) Close() error {
	return p.producer.Close()
}

// Topics defines standard event topics
const (
	TopicOrders        = "travio.orders"
	TopicPayments      = "travio.payments"
	TopicInventory     = "travio.inventory"
	TopicFulfillment   = "travio.fulfillment"
	TopicNotifications = "travio.notifications"
	TopicCatalog       = "travio.catalog"
)

// Event types
const (
	EventOrderCreated      = "order.created"
	EventOrderConfirmed    = "order.confirmed"
	EventOrderCancelled    = "order.cancelled"
	EventOrderFailed       = "order.failed"
	EventPaymentAuthorized = "payment.authorized"
	EventPaymentCaptured   = "payment.captured"
	EventPaymentFailed     = "payment.failed"
	EventPaymentRefunded   = "payment.refunded"
	EventSeatsHeld         = "inventory.seats_held"
	EventSeatsReleased     = "inventory.seats_released"
	EventSeatsBooked       = "inventory.seats_booked"
	EventTicketGenerated   = "fulfillment.ticket_generated"
	EventNotificationSent  = "notification.sent"
	EventTripCreated       = "trip.created"
	EventStationCreated    = "station.created"
)
