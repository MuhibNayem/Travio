package repository

import (
	"context"
	"fmt"
)

// InitSchema ensures the database schema is up to date for Catalog Service
func (r *PostgresStationRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		// 1. Base Tables (from init-db.sql + 001 fix)
		`CREATE TABLE IF NOT EXISTS stations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID,
			code VARCHAR(10) NOT NULL,
			name VARCHAR(255) NOT NULL,
			city VARCHAR(100) NOT NULL,
			state VARCHAR(100),
			country VARCHAR(100) DEFAULT 'Bangladesh',
			latitude DOUBLE PRECISION,
			longitude DOUBLE PRECISION,
			timezone VARCHAR(50) DEFAULT 'Asia/Dhaka',
			address TEXT,
			amenities JSONB DEFAULT '[]',
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS routes (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			code VARCHAR(50) NOT NULL,
			name VARCHAR(255) NOT NULL,
			origin_station_id UUID REFERENCES stations(id),
			destination_station_id UUID REFERENCES stations(id),
			intermediate_stops JSONB DEFAULT '[]',
			distance_km INTEGER,
			estimated_duration_minutes INTEGER,
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS trips (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			schedule_id UUID,
			service_date DATE,
			route_id UUID REFERENCES routes(id),
			vehicle_id VARCHAR(100),
			vehicle_type VARCHAR(50) NOT NULL,
			vehicle_class VARCHAR(50),
			departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
			arrival_time TIMESTAMP WITH TIME ZONE,
			total_seats INTEGER NOT NULL,
			available_seats INTEGER NOT NULL,
			pricing JSONB DEFAULT '{}',
			status VARCHAR(50) DEFAULT 'scheduled',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// 2. Indexes (Critical ones)
		`CREATE INDEX IF NOT EXISTS idx_stations_city ON stations(city)`,
		`CREATE INDEX IF NOT EXISTS idx_trips_departure ON trips(departure_time)`,
		`CREATE INDEX IF NOT EXISTS idx_trips_route_id ON trips(route_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS trips_schedule_date_idx ON trips (schedule_id, service_date) WHERE status != 'cancelled'`,

		// 3. Migrations
		// 001: Ensure amenities is JSONB (If table existed from old schema)
		// Note: CREATE TABLE using JSONB above handles fresh install.
		// This handles migration from TEXT[] if needed, or verifies type.
		// For simplicity/idempotency in this context, we'll assume the CREATE handled it or
		// user manually fixes strictly typed conflicts.
		// Ideally: ALTER TABLE stations ALTER COLUMN amenities TYPE JSONB USING ...
		// But fail-safe:
		// `ALTER TABLE stations ALTER COLUMN amenities TYPE JSONB USING to_jsonb(amenities)` is risky if already jsonb.

		// 002: Add Audit Logs
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID NOT NULL,
			action VARCHAR(50) NOT NULL,
			actor_id UUID,
			changes JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_logs_entity ON audit_logs(entity_type, entity_id)`,

		// 003: Outbox Table for Transactional Messaging
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
		`CREATE INDEX IF NOT EXISTS idx_outbox_unprocessed ON event_outbox(processed_at) WHERE processed_at IS NULL`,
	}

	for _, query := range queries {
		if _, err := r.DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("schema init failed: %w query: %s", err, query)
		}
	}

	return nil
}
