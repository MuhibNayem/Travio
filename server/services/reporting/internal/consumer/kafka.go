// Package consumer provides Kafka event consumer.
package consumer

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/clickhouse"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/domain"
	"github.com/segmentio/kafka-go"
)

// Config holds Kafka consumer configuration.
type Config struct {
	Brokers       []string
	ConsumerGroup string
	Topics        []string
}

// LoadConfigFromEnv loads Kafka config from environment.
func LoadConfigFromEnv() Config {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092"
	}

	topics := os.Getenv("KAFKA_TOPICS")
	if topics == "" {
		topics = "travio.events"
	}

	return Config{
		Brokers:       strings.Split(brokers, ","),
		ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "reporting-service"),
		Topics:        strings.Split(topics, ","),
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// Consumer consumes events from Kafka and writes to ClickHouse.
type Consumer struct {
	readers      []*kafka.Reader
	chClient     *clickhouse.Client
	config       Config
	stopChan     chan struct{}
	wg           sync.WaitGroup
	processedIDs sync.Map // For deduplication
}

// NewConsumer creates a new Kafka consumer.
func NewConsumer(cfg Config, chClient *clickhouse.Client) *Consumer {
	return &Consumer{
		chClient: chClient,
		config:   cfg,
		stopChan: make(chan struct{}),
	}
}

// Start begins consuming messages from Kafka.
func (c *Consumer) Start(ctx context.Context) error {
	for _, topic := range c.config.Topics {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        c.config.Brokers,
			Topic:          topic,
			GroupID:        c.config.ConsumerGroup,
			MinBytes:       1e3,  // 1KB
			MaxBytes:       10e6, // 10MB
			MaxWait:        1 * time.Second,
			StartOffset:    kafka.LastOffset,
			CommitInterval: 1 * time.Second,
		})
		c.readers = append(c.readers, reader)

		c.wg.Add(1)
		go c.consumeTopic(ctx, reader, topic)
	}

	logger.Info("Kafka consumer started", "topics", c.config.Topics, "group", c.config.ConsumerGroup)
	return nil
}

// Stop gracefully stops the consumer.
func (c *Consumer) Stop() error {
	close(c.stopChan)
	c.wg.Wait()

	for _, reader := range c.readers {
		if err := reader.Close(); err != nil {
			logger.Warn("Failed to close Kafka reader", "error", err)
		}
	}

	return nil
}

// consumeTopic consumes messages from a single topic.
func (c *Consumer) consumeTopic(ctx context.Context, reader *kafka.Reader, topic string) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ctx.Done():
			return
		default:
		}

		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			logger.Warn("Failed to fetch message", "topic", topic, "error", err)
			time.Sleep(1 * time.Second)
			continue
		}

		// Parse event envelope
		var envelope struct {
			ID          string          `json:"id"`
			Type        string          `json:"type"`
			AggregateID string          `json:"aggregate_id"`
			Timestamp   time.Time       `json:"timestamp"`
			Version     int             `json:"version"`
			Payload     json.RawMessage `json:"payload"`
		}

		if err := json.Unmarshal(msg.Value, &envelope); err != nil {
			logger.Warn("Failed to parse event envelope", "topic", topic, "error", err)
			_ = reader.CommitMessages(ctx, msg)
			continue
		}

		// Map to domain.Event
		event := domain.Event{
			EventID:        envelope.ID,
			EventType:      envelope.Type,
			Timestamp:      envelope.Timestamp,
			Metadata:       string(envelope.Payload), // Store raw payload as metadata
			OrganizationID: "",                       // Extracted below
		}

		// Extract fields based on type
		switch envelope.Type {
		case domain.EventOrderCreated:
			var p struct {
				OrganizationID string `json:"organization_id"`
				UserID         string `json:"user_id"`
				OrderID        string `json:"order_id"`
				TripID         string `json:"trip_id"`
				RouteID        string `json:"route_id"`
				TotalPaisa     int64  `json:"total_paisa"`
			}
			if err := json.Unmarshal(envelope.Payload, &p); err == nil {
				event.OrganizationID = p.OrganizationID
				event.UserID = p.UserID
				event.OrderID = p.OrderID
				event.TripID = p.TripID
				event.RouteID = p.RouteID
				event.AmountPaisa = p.TotalPaisa
				event.Status = "created"
			}

		case domain.EventOrderCompleted:
			var p struct {
				OrganizationID string `json:"organization_id"`
				OrderID        string `json:"order_id"`
				TripID         string `json:"trip_id"`
				TotalPaisa     int64  `json:"total_paisa"`
			}
			if err := json.Unmarshal(envelope.Payload, &p); err == nil {
				event.OrganizationID = p.OrganizationID
				event.OrderID = p.OrderID
				event.TripID = p.TripID
				event.AmountPaisa = p.TotalPaisa
				event.Status = "completed"
			}

		case domain.EventOrderCancelled:
			var p struct {
				OrganizationID string `json:"organization_id"`
				OrderID        string `json:"order_id"`
				TripID         string `json:"trip_id"`
			}
			if err := json.Unmarshal(envelope.Payload, &p); err == nil {
				event.OrganizationID = p.OrganizationID
				event.OrderID = p.OrderID
				event.TripID = p.TripID
				event.Status = "cancelled"
			}
			// Ideally event should contain OrgID.
			// I'll skip OrgID extraction for others if missing, but TripCreated is vital.

		case domain.EventTripCreated:
			var p struct {
				OrganizationID string `json:"organization_id"`
				TripID         string `json:"trip_id"`
				Status         string `json:"status"`
			}
			if err := json.Unmarshal(envelope.Payload, &p); err == nil {
				event.OrganizationID = p.OrganizationID
				event.TripID = p.TripID
				event.Status = p.Status
			}
		}

		// Deduplicate by event ID
		if event.EventID != "" {
			if _, exists := c.processedIDs.LoadOrStore(event.EventID, true); exists {
				logger.Debug("Duplicate event skipped", "event_id", event.EventID)
				_ = reader.CommitMessages(ctx, msg)
				continue
			}
		}

		// Insert to ClickHouse
		if err := c.chClient.InsertEvent(ctx, event); err != nil {
			logger.Error("Failed to insert event", "event_id", event.EventID, "error", err)
			continue
		}

		// Commit offset
		if err := reader.CommitMessages(ctx, msg); err != nil {
			logger.Warn("Failed to commit message", "error", err)
		}

		logger.Debug("Processed event",
			"event_id", event.EventID,
			"event_type", event.EventType,
			"org_id", event.OrganizationID,
		)
	}
}

// CleanupDeduplicationCache periodically cleans old entries from dedup cache.
func (c *Consumer) CleanupDeduplicationCache() {
	// Simple cleanup - clear every hour
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.processedIDs = sync.Map{}
			logger.Debug("Deduplication cache cleared")
		}
	}
}
