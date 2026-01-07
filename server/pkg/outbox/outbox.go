package outbox

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// OutboxEntry represents a pending event in the outbox
type OutboxEntry struct {
	ID          string
	AggregateID string
	EventType   string
	Topic       string
	Payload     []byte
	CreatedAt   time.Time
	ProcessedAt *time.Time
	Retries     int
}

// Publisher provides reliable event publishing via outbox pattern
type Publisher struct {
	db *sql.DB
}

// NewPublisher creates a new outbox publisher
func NewPublisher(db *sql.DB) *Publisher {
	return &Publisher{db: db}
}

// CreateTable creates the outbox table if it doesn't exist
func (p *Publisher) CreateTable(ctx context.Context) error {
	_, err := p.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS event_outbox (
			id UUID PRIMARY KEY,
			aggregate_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			topic VARCHAR(100) NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			processed_at TIMESTAMP WITH TIME ZONE,
			retries INT DEFAULT 0,
			INDEX idx_outbox_unprocessed (processed_at) WHERE processed_at IS NULL
		)
	`)
	return err
}

// Publish adds an event to the outbox within a transaction
func (p *Publisher) Publish(ctx context.Context, tx *sql.Tx, topic, eventType, aggregateID string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	id := uuid.New().String()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO event_outbox (id, aggregate_id, event_type, topic, payload)
		VALUES ($1, $2, $3, $4, $5)
	`, id, aggregateID, eventType, topic, data)
	return err
}

// GetPending returns unprocessed outbox entries
func (p *Publisher) GetPending(ctx context.Context, limit int) ([]OutboxEntry, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT id, aggregate_id, event_type, topic, payload, created_at, retries
		FROM event_outbox
		WHERE processed_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []OutboxEntry
	for rows.Next() {
		var e OutboxEntry
		if err := rows.Scan(&e.ID, &e.AggregateID, &e.EventType, &e.Topic, &e.Payload, &e.CreatedAt, &e.Retries); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// MarkProcessed marks an entry as successfully published
func (p *Publisher) MarkProcessed(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE event_outbox
		SET processed_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

// IncrementRetry increments the retry counter
func (p *Publisher) IncrementRetry(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE event_outbox
		SET retries = retries + 1
		WHERE id = $1
	`, id)
	return err
}

// CleanupOld removes old processed entries
func (p *Publisher) CleanupOld(ctx context.Context, olderThan time.Duration) error {
	_, err := p.db.ExecContext(ctx, `
		DELETE FROM event_outbox
		WHERE processed_at IS NOT NULL
		AND processed_at < $1
	`, time.Now().Add(-olderThan))
	return err
}
