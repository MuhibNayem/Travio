package repository

import (
	"context"
	"fmt"
)

// InitSchema ensures the database schema is up to date
func (r *OrderRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// 001_initial_schema (reconstructed/simplified as idempotent)
		`CREATE TABLE IF NOT EXISTS orders (
			id UUID PRIMARY KEY,
			organization_id UUID NOT NULL,
			user_id UUID NOT NULL,
			trip_id UUID NOT NULL,
			from_station_id UUID NOT NULL,
			to_station_id UUID NOT NULL,
			
			passengers JSONB NOT NULL,
			
			subtotal_paisa BIGINT NOT NULL,
			tax_paisa BIGINT NOT NULL,
			booking_fee_paisa BIGINT NOT NULL,
			discount_paisa BIGINT NOT NULL,
			total_paisa BIGINT NOT NULL,
			currency VARCHAR(10) NOT NULL,
			
			payment_id VARCHAR(255),
			payment_status VARCHAR(50),
			payment_method VARCHAR(50),
			
			booking_id VARCHAR(255),
			hold_id VARCHAR(255),
			seats JSONB,
			
			status VARCHAR(50) NOT NULL,
			saga_id UUID,
			
			contact_email VARCHAR(255),
			contact_phone VARCHAR(50),
			
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			expires_at TIMESTAMP WITH TIME ZONE,
			idempotency_key VARCHAR(255)
		)`,

		// 002_add_outbox
		`CREATE TABLE IF NOT EXISTS event_outbox (
			id UUID PRIMARY KEY,
			aggregate_id VARCHAR(255) NOT NULL,
			event_type VARCHAR(100) NOT NULL,
			topic VARCHAR(100) NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			processed_at TIMESTAMP WITH TIME ZONE,
			retries INT DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON event_outbox (created_at) WHERE processed_at IS NULL`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_processed ON event_outbox (processed_at) WHERE processed_at IS NOT NULL`,

		// 003_add_route_id
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS route_id UUID`,
		`CREATE INDEX IF NOT EXISTS idx_orders_route_id ON orders(route_id)`,
	}

	for _, query := range queries {
		if _, err := r.DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("schema init failed: %w query: %s", err, query)
		}
	}

	return nil
}
