package consumer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/realtime"
	"github.com/google/uuid"
)

// EventConsumer handles Kafka events for realtime updates
type EventConsumer struct {
	consumer *kafka.Consumer
	realtime *realtime.Manager
}

// New creates a new event consumer with a UNIQUE group ID for broadcast
func New(brokers []string, realtimeMgr *realtime.Manager) (*EventConsumer, error) {
	// Generate unique Group ID to ensure Fan-Out (Broadcast) behavior
	// Every gateway instance must receive every seat update
	uniqueGroupID := fmt.Sprintf("gateway-realtime-%s", uuid.New().String())

	// Topics to listen to
	topics := []string{kafka.TopicInventory}

	consumer, err := kafka.NewConsumer(brokers, uniqueGroupID, topics)
	if err != nil {
		return nil, err
	}

	ec := &EventConsumer{
		consumer: consumer,
		realtime: realtimeMgr,
	}

	// Register Handlers
	consumer.RegisterHandler(kafka.EventSeatsHeld, ec.handleSeatUpdate)
	consumer.RegisterHandler(kafka.EventSeatsReleased, ec.handleSeatUpdate)
	consumer.RegisterHandler(kafka.EventSeatsBooked, ec.handleSeatUpdate)

	return ec, nil
}

// Start starts the consumer
func (c *EventConsumer) Start() error {
	return c.consumer.Start()
}

// Stop stops the consumer
func (c *EventConsumer) Stop() error {
	return c.consumer.Stop()
}

// handleSeatUpdate processes seat events and broadcasts to SSE clients
func (c *EventConsumer) handleSeatUpdate(ctx context.Context, event *kafka.Event) error {
	// We use AggregateID as TripID based on producer logic
	tripID := event.AggregateID
	if tripID == "" {
		logger.Warn("Received seat update without AggregateID (TripID)")
		return nil
	}

	// Marshal payload back to JSON to send to client
	// event.Payload is map[string]interface{} after unmarshal
	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		logger.Error("Failed to marshal event payload for broadcast", "error", err)
		return nil
	}

	// Construct SSE message data
	// Format: { "type": "inventory.seats_held", "data": { ... } }
	sseMessage := fmt.Sprintf(`{"type":"%s","data":%s}`, event.Type, string(payloadBytes))

	c.realtime.Broadcast(tripID, sseMessage)

	return nil
}
