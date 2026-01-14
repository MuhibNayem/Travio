package repository

import (
	"context"
	"fmt"
)

// InitSchema ensures the Subscription database schema is up to date
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS plans (
			id VARCHAR(100) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price_paisa BIGINT NOT NULL,
			interval VARCHAR(50) NOT NULL,
			features JSONB DEFAULT '{}',
			is_active BOOLEAN DEFAULT true,
			usage_price_paisa BIGINT DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS subscriptions (
			id VARCHAR(100) PRIMARY KEY,
			organization_id VARCHAR(100) NOT NULL,
			plan_id VARCHAR(100) NOT NULL REFERENCES plans(id),
			status VARCHAR(50) DEFAULT 'active',
			current_period_start TIMESTAMP WITH TIME ZONE,
			current_period_end TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS invoices (
			id VARCHAR(100) PRIMARY KEY,
			subscription_id VARCHAR(100) NOT NULL REFERENCES subscriptions(id),
			amount_paisa BIGINT NOT NULL,
			status VARCHAR(50) DEFAULT 'open',
			issued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			due_date TIMESTAMP WITH TIME ZONE,
			paid_at TIMESTAMP WITH TIME ZONE,
			line_items JSONB,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS usage_events (
			id VARCHAR(100) PRIMARY KEY,
			subscription_id VARCHAR(100) NOT NULL, 
			event_type VARCHAR(255),
			units BIGINT,
			idempotency_key VARCHAR(255) UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`,

		`CREATE INDEX IF NOT EXISTS idx_plans_active ON plans(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_org_id ON subscriptions(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices(subscription_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_events_sub_id ON usage_events(subscription_id)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("subscription schema init failed: %w query: %s", err, query)
		}
	}
	return nil
}
