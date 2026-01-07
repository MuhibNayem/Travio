package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// Handler processes events of a specific type
type Handler func(ctx context.Context, event *Event) error

// Consumer consumes events from Kafka topics
type Consumer struct {
	client   sarama.ConsumerGroup
	handlers map[string]Handler
	topics   []string
	groupID  string
	mu       sync.RWMutex
	ready    chan bool
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewConsumer creates a new Kafka consumer group
func NewConsumer(brokers []string, groupID string, topics []string) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	client, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Consumer{
		client:   client,
		handlers: make(map[string]Handler),
		topics:   topics,
		groupID:  groupID,
		ready:    make(chan bool),
		ctx:      ctx,
		cancel:   cancel,
	}, nil
}

// RegisterHandler registers a handler for a specific event type
func (c *Consumer) RegisterHandler(eventType string, handler Handler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[eventType] = handler
}

// Start begins consuming messages
func (c *Consumer) Start() error {
	go func() {
		for {
			if err := c.client.Consume(c.ctx, c.topics, c); err != nil {
				logger.Error("consumer error", "error", err)
			}
			if c.ctx.Err() != nil {
				return
			}
			c.ready = make(chan bool)
		}
	}()

	<-c.ready // Wait until consumer is set up
	logger.Info("consumer started", "group", c.groupID, "topics", c.topics)
	return nil
}

// Stop stops the consumer
func (c *Consumer) Stop() error {
	c.cancel()
	return c.client.Close()
}

// Setup is run at the beginning of a new session
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim processes messages from partitions
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var event Event
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			logger.Error("failed to unmarshal event", "error", err)
			session.MarkMessage(msg, "")
			continue
		}

		c.mu.RLock()
		handler, ok := c.handlers[event.Type]
		c.mu.RUnlock()

		if !ok {
			logger.Info("no handler for event type", "type", event.Type)
			session.MarkMessage(msg, "")
			continue
		}

		if err := handler(session.Context(), &event); err != nil {
			logger.Error("handler error",
				"event_type", event.Type,
				"event_id", event.ID,
				"error", err,
			)
			// Don't mark as processed - will be retried
			continue
		}

		session.MarkMessage(msg, "")
		logger.Info("event processed",
			"event_type", event.Type,
			"event_id", event.ID,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)
	}
	return nil
}
