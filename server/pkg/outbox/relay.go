package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/kafka"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// Relay polls the outbox and publishes events to Kafka
type Relay struct {
	publisher    *Publisher
	producer     *kafka.Producer
	pollInterval time.Duration
	batchSize    int
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewRelay creates a new outbox relay
func NewRelay(db *sql.DB, brokers []string, pollInterval time.Duration, batchSize int) (*Relay, error) {
	producer, err := kafka.NewProducer(brokers)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Relay{
		publisher:    NewPublisher(db),
		producer:     producer,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

// Start begins the outbox relay polling loop
func (r *Relay) Start() {
	go func() {
		ticker := time.NewTicker(r.pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				r.processBatch()
			}
		}
	}()

	logger.Info("outbox relay started", "interval", r.pollInterval, "batch_size", r.batchSize)
}

// processBatch fetches and publishes pending outbox entries
func (r *Relay) processBatch() {
	entries, err := r.publisher.GetPending(r.ctx, r.batchSize)
	if err != nil {
		logger.Error("failed to get pending outbox entries", "error", err)
		return
	}

	for _, entry := range entries {
		event := &kafka.Event{
			ID:          entry.ID,
			Type:        entry.EventType,
			AggregateID: entry.AggregateID,
			Timestamp:   entry.CreatedAt,
			Version:     1,
			Payload:     make(map[string]interface{}),
		}

		// Parse payload
		if err := json.Unmarshal(entry.Payload, &event.Payload); err != nil {
			logger.Error("failed to unmarshal outbox payload", "id", entry.ID, "error", err)
			r.publisher.IncrementRetry(r.ctx, entry.ID)
			continue
		}

		// Publish to Kafka
		if err := r.producer.Publish(r.ctx, entry.Topic, event); err != nil {
			logger.Error("failed to publish outbox entry", "id", entry.ID, "error", err)
			r.publisher.IncrementRetry(r.ctx, entry.ID)
			continue
		}

		// Mark as processed
		if err := r.publisher.MarkProcessed(r.ctx, entry.ID); err != nil {
			logger.Error("failed to mark outbox entry as processed", "id", entry.ID, "error", err)
		}
	}
}

// Stop stops the relay
func (r *Relay) Stop() error {
	r.cancel()
	return r.producer.Close()
}
