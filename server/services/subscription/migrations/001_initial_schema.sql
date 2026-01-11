-- Initial Schema for Subscription Service
-- Fixes missing tables and mismatched ID types (plans uses string slugs)

CREATE TABLE IF NOT EXISTS plans (
    id VARCHAR(100) PRIMARY KEY, -- Changed from UUID to support 'plan_free', 'plan_pro'
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_paisa BIGINT NOT NULL,
    interval VARCHAR(50) NOT NULL, -- 'month', 'year'
    features JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    usage_price_paisa BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id VARCHAR(100) PRIMARY KEY,
    organization_id VARCHAR(100) NOT NULL,
    plan_id VARCHAR(100) NOT NULL REFERENCES plans(id),
    status VARCHAR(50) DEFAULT 'active', -- active, canceled, past_due, trialing
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS invoices (
    id VARCHAR(100) PRIMARY KEY,
    subscription_id VARCHAR(100) NOT NULL REFERENCES subscriptions(id),
    amount_paisa BIGINT NOT NULL,
    status VARCHAR(50) DEFAULT 'open', -- paid, open, void, uncollectible
    issued_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    due_date TIMESTAMP WITH TIME ZONE,
    paid_at TIMESTAMP WITH TIME ZONE,
    line_items JSONB, -- Added based on Struct
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS usage_events (
    id VARCHAR(100) PRIMARY KEY,
    subscription_id VARCHAR(100) NOT NULL, 
    event_type VARCHAR(255),
    units BIGINT,
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_plans_active ON plans(is_active);
CREATE INDEX IF NOT EXISTS idx_subscriptions_org_id ON subscriptions(organization_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX IF NOT EXISTS idx_usage_events_sub_id ON usage_events(subscription_id);
