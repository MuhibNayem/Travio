package consumer

import (
	"context"
	"encoding/json"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/search/internal/indexer"
)

type EventConsumer struct {
	consumer *kafka.Consumer
	indexer  *indexer.Indexer
}

func New(brokers []string, groupID string, indexer *indexer.Indexer) (*EventConsumer, error) {
	topics := []string{kafka.TopicCatalog, kafka.TopicEvents}
	consumer, err := kafka.NewConsumer(brokers, groupID, topics)
	if err != nil {
		return nil, err
	}

	c := &EventConsumer{
		consumer: consumer,
		indexer:  indexer,
	}

	consumer.RegisterHandler(kafka.EventTripCreated, c.handleTripCreated)
	consumer.RegisterHandler(kafka.EventTripUpdated, c.handleTripUpdated)
	consumer.RegisterHandler(kafka.EventStationCreated, c.handleStationCreated)
	consumer.RegisterHandler(kafka.EventEventCreated, c.handleEventUpsert)
	consumer.RegisterHandler(kafka.EventEventUpdated, c.handleEventUpsert)
	consumer.RegisterHandler(kafka.EventEventPublished, c.handleEventUpsert)

	return c, nil
}

func (c *EventConsumer) Start() error {
	return c.consumer.Start()
}

func (c *EventConsumer) Stop() error {
	return c.consumer.Stop()
}

func (c *EventConsumer) handleTripCreated(ctx context.Context, event *kafka.Event) error {
	return c.upsertTrip(ctx, event)
}

func (c *EventConsumer) handleTripUpdated(ctx context.Context, event *kafka.Event) error {
	return c.upsertTrip(ctx, event)
}

func (c *EventConsumer) upsertTrip(ctx context.Context, event *kafka.Event) error {
	logger.Info("Indexing trip", "id", event.AggregateID, "type", event.Type)

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	return c.indexer.IndexDocument(ctx, "trips", event.AggregateID, string(payloadBytes))
}

func (c *EventConsumer) handleStationCreated(ctx context.Context, event *kafka.Event) error {
	logger.Info("Indexing station", "id", event.AggregateID)

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	return c.indexer.IndexDocument(ctx, "stations", event.AggregateID, string(payloadBytes))
}

func (c *EventConsumer) handleEventUpsert(ctx context.Context, event *kafka.Event) error {
	logger.Info("Indexing event", "id", event.AggregateID)

	payloadBytes, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	return c.indexer.IndexDocument(ctx, "events", event.AggregateID, string(payloadBytes))
}
