// Package clickhouse provides ClickHouse database client.
package clickhouse

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/domain"
)

// Config holds ClickHouse connection configuration.
type Config struct {
	Host          string
	Port          int
	Database      string
	Username      string
	Password      string
	MaxConns      int
	BatchSize     int
	FlushInterval time.Duration
}

// LoadConfigFromEnv loads ClickHouse config from environment.
func LoadConfigFromEnv() Config {
	port := 9000
	if p := os.Getenv("CLICKHOUSE_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	maxConns := 10
	if m := os.Getenv("CLICKHOUSE_MAX_CONNS"); m != "" {
		if v, err := strconv.Atoi(m); err == nil {
			maxConns = v
		}
	}

	batchSize := 10000
	if b := os.Getenv("REPORT_BATCH_SIZE"); b != "" {
		if v, err := strconv.Atoi(b); err == nil {
			batchSize = v
		}
	}

	flushInterval := 5 * time.Second
	if f := os.Getenv("REPORT_FLUSH_INTERVAL"); f != "" {
		if d, err := time.ParseDuration(f); err == nil {
			flushInterval = d
		}
	}

	return Config{
		Host:          getEnv("CLICKHOUSE_HOST", "localhost"),
		Port:          port,
		Database:      getEnv("CLICKHOUSE_DATABASE", "travio_analytics"),
		Username:      getEnv("CLICKHOUSE_USER", "default"),
		Password:      os.Getenv("CLICKHOUSE_PASSWORD"),
		MaxConns:      maxConns,
		BatchSize:     batchSize,
		FlushInterval: flushInterval,
	}
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// Client is the ClickHouse database client.
type Client struct {
	conn       driver.Conn
	config     Config
	eventBatch []domain.Event
	lastFlush  time.Time
	flushChan  chan struct{}
	stopChan   chan struct{}
}

// NewClient creates a new ClickHouse client.
func NewClient(cfg Config) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		MaxOpenConns: cfg.MaxConns,
		MaxIdleConns: cfg.MaxConns / 2,
		DialTimeout:  10 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	client := &Client{
		conn:       conn,
		config:     cfg,
		eventBatch: make([]domain.Event, 0, cfg.BatchSize),
		lastFlush:  time.Now(),
		flushChan:  make(chan struct{}, 1),
		stopChan:   make(chan struct{}),
	}

	// Start background flusher
	go client.backgroundFlusher()

	return client, nil
}

// Close closes the ClickHouse connection.
func (c *Client) Close() error {
	close(c.stopChan)
	c.Flush(context.Background()) // Final flush
	return c.conn.Close()
}

// Ping checks if ClickHouse is reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// InitSchema creates the required tables and materialized views.
func (c *Client) InitSchema(ctx context.Context) error {
	schemas := []string{
		// Raw events table
		`CREATE TABLE IF NOT EXISTS events (
			event_id UUID,
			event_type LowCardinality(String),
			organization_id UUID,
			user_id UUID,
			timestamp DateTime64(3),
			order_id Nullable(UUID),
			payment_id Nullable(UUID),
			trip_id Nullable(UUID),
			amount_paisa Int64,
			status LowCardinality(String),
			metadata String,
			event_date Date MATERIALIZED toDate(timestamp)
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(event_date)
		ORDER BY (organization_id, event_type, timestamp)
		TTL event_date + INTERVAL 2 YEAR`,

		// Daily revenue materialized view
		`CREATE MATERIALIZED VIEW IF NOT EXISTS daily_revenue_mv
		ENGINE = SummingMergeTree()
		PARTITION BY toYYYYMM(date)
		ORDER BY (organization_id, date)
		AS SELECT
			organization_id,
			toDate(timestamp) AS date,
			count() AS order_count,
			sum(amount_paisa) AS total_revenue_paisa,
			avg(amount_paisa) AS avg_order_value
		FROM events
		WHERE event_type = 'order_completed'
		GROUP BY organization_id, date`,

		// Hourly bookings materialized view
		`CREATE MATERIALIZED VIEW IF NOT EXISTS hourly_bookings_mv
		ENGINE = SummingMergeTree()
		ORDER BY (organization_id, hour)
		AS SELECT
			organization_id,
			toStartOfHour(timestamp) AS hour,
			count() AS booking_count,
			countIf(status = 'completed') AS completed_count,
			countIf(status = 'cancelled') AS cancelled_count
		FROM events
		WHERE event_type IN ('order_created', 'order_completed', 'order_cancelled')
		GROUP BY organization_id, hour`,

		// Top routes materialized view
		`CREATE MATERIALIZED VIEW IF NOT EXISTS top_routes_mv
		ENGINE = SummingMergeTree()
		ORDER BY (organization_id, trip_id)
		AS SELECT
			organization_id,
			trip_id,
			count() AS booking_count,
			sum(amount_paisa) AS revenue
		FROM events
		WHERE event_type = 'order_completed' AND trip_id IS NOT NULL
		GROUP BY organization_id, trip_id`,
	}

	for _, schema := range schemas {
		if err := c.conn.Exec(ctx, schema); err != nil {
			logger.Warn("Schema creation warning", "error", err)
			// Continue - might already exist
		}
	}

	logger.Info("ClickHouse schema initialized")
	return nil
}

// InsertEvent adds an event to the batch.
func (c *Client) InsertEvent(ctx context.Context, event domain.Event) error {
	c.eventBatch = append(c.eventBatch, event)

	if len(c.eventBatch) >= c.config.BatchSize {
		return c.Flush(ctx)
	}

	// Trigger flush check
	select {
	case c.flushChan <- struct{}{}:
	default:
	}

	return nil
}

// InsertEvents adds multiple events to the batch.
func (c *Client) InsertEvents(ctx context.Context, events []domain.Event) error {
	for _, event := range events {
		if err := c.InsertEvent(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// Flush writes buffered events to ClickHouse.
func (c *Client) Flush(ctx context.Context) error {
	if len(c.eventBatch) == 0 {
		return nil
	}

	batch, err := c.conn.PrepareBatch(ctx, `INSERT INTO events (
		event_id, event_type, organization_id, user_id, timestamp,
		order_id, payment_id, trip_id, amount_paisa, status, metadata
	)`)
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, e := range c.eventBatch {
		if err := batch.Append(
			e.EventID,
			e.EventType,
			e.OrganizationID,
			e.UserID,
			e.Timestamp,
			nullableString(e.OrderID),
			nullableString(e.PaymentID),
			nullableString(e.TripID),
			e.AmountPaisa,
			e.Status,
			e.Metadata,
		); err != nil {
			logger.Warn("Failed to append event to batch", "error", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to send batch: %w", err)
	}

	logger.Debug("Flushed events to ClickHouse", "count", len(c.eventBatch))

	c.eventBatch = c.eventBatch[:0]
	c.lastFlush = time.Now()

	return nil
}

// backgroundFlusher periodically flushes events.
func (c *Client) backgroundFlusher() {
	ticker := time.NewTicker(c.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			if time.Since(c.lastFlush) >= c.config.FlushInterval && len(c.eventBatch) > 0 {
				if err := c.Flush(context.Background()); err != nil {
					logger.Error("Background flush failed", "error", err)
				}
			}
		case <-c.flushChan:
			// Check if time-based flush needed
			if time.Since(c.lastFlush) >= c.config.FlushInterval && len(c.eventBatch) > 0 {
				if err := c.Flush(context.Background()); err != nil {
					logger.Error("Triggered flush failed", "error", err)
				}
			}
		}
	}
}

// Query executes a query and returns rows.
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}

// Exec executes a statement.
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
