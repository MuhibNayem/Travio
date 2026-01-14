package repository

import (
	"context"
	"fmt"
)

// InitSchema ensures the Postgres schema is up to date
func (r *AssetRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS assets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id UUID NOT NULL,
			name VARCHAR(255),
			license_plate VARCHAR(50),
			vin VARCHAR(100),
			make VARCHAR(100),
			model VARCHAR(255),
			year INTEGER,
			type VARCHAR(50) NOT NULL,
			status VARCHAR(50) DEFAULT 'active',
			config JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		// Postgres location table (if used alongside Scylla, or as fallback)
		`CREATE TABLE IF NOT EXISTS asset_locations (
			asset_id UUID PRIMARY KEY REFERENCES assets(id),
			latitude DOUBLE PRECISION,
			longitude DOUBLE PRECISION,
			speed DOUBLE PRECISION,
			heading DOUBLE PRECISION,
			timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_assets_org ON assets(organization_id)`,
	}

	for _, query := range queries {
		if _, err := r.DB.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("postgres schema init failed: %w query: %s", err, query)
		}
	}
	return nil
}
