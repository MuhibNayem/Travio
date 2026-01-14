-- Travio Master Database Initialization Script
-- ==============================================================================
-- This script runs on PostgreSQL container startup.
-- It populates the schemas for all services that require a relational database.
-- Prerequisites: Databases must be created by create-multiple-postgresql-databases.sh
-- ==============================================================================

-- ==============================================================================
-- 1. IDENTITY SERVICE (travio_identity)
-- ==============================================================================
\c travio_identity

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255), -- User Full Name
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID,
    role VARCHAR(50) DEFAULT 'user',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS role VARCHAR(50) DEFAULT 'user';

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan_id VARCHAR(100),
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(255),
    currency VARCHAR(3) DEFAULT 'BDT',
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    family_id UUID,
    token_hash VARCHAR(255) NOT NULL,
    device_info VARCHAR(500),
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    revoked BOOLEAN DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE IF EXISTS refresh_tokens
    ADD COLUMN IF NOT EXISTS family_id UUID,
    ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS user_agent VARCHAR(500),
    ADD COLUMN IF NOT EXISTS device_info VARCHAR(500),
    ADD COLUMN IF NOT EXISTS revoked BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ADD COLUMN IF NOT EXISTS token_hash VARCHAR(255) NOT NULL,
    ADD COLUMN IF NOT EXISTS ip_address VARCHAR(50);

CREATE TABLE IF NOT EXISTS organization_invites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    email VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    token VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, expired
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);

-- ==============================================================================
-- 2. CATALOG SERVICE (travio_catalog)
-- ==============================================================================
\c travio_catalog

CREATE TABLE IF NOT EXISTS stations (
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
    amenities JSONB DEFAULT '[]', -- JSON array of strings
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS routes (
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
);

CREATE TABLE IF NOT EXISTS trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    schedule_id UUID,
    service_date DATE,
    route_id UUID REFERENCES routes(id),
    vehicle_id VARCHAR(100),
    vehicle_type VARCHAR(50) NOT NULL, -- 'bus', 'train', 'launch'
    vehicle_class VARCHAR(50),
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE,
    total_seats INTEGER NOT NULL,
    available_seats INTEGER NOT NULL,
    pricing JSONB DEFAULT '{}',
    status VARCHAR(50) DEFAULT 'scheduled',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedule_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    route_id UUID NOT NULL REFERENCES routes(id),
    vehicle_id VARCHAR(100),
    vehicle_type VARCHAR(50) NOT NULL,
    vehicle_class VARCHAR(50),
    total_seats INTEGER NOT NULL DEFAULT 0,
    pricing JSONB DEFAULT '{}',
    departure_time TIME NOT NULL,
    arrival_offset_minutes INTEGER,
    timezone VARCHAR(50) DEFAULT 'Asia/Dhaka',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    days_of_week SMALLINT NOT NULL, -- bitmask: 0b0000001 = Mon ... 0b1000000 = Sun
    status VARCHAR(50) DEFAULT 'active',
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedule_exceptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedule_templates(id) ON DELETE CASCADE,
    service_date DATE NOT NULL,
    is_added BOOLEAN NOT NULL DEFAULT true,
    reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedule_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    schedule_id UUID NOT NULL REFERENCES schedule_templates(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    snapshot JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

ALTER TABLE trips
    ADD CONSTRAINT IF NOT EXISTS trips_schedule_fk
    FOREIGN KEY (schedule_id) REFERENCES schedule_templates(id);

CREATE TABLE IF NOT EXISTS trip_segments (
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL,
    from_station_id UUID REFERENCES stations(id),
    to_station_id UUID REFERENCES stations(id),
    departure_time TIMESTAMP WITH TIME ZONE NOT NULL,
    arrival_time TIMESTAMP WITH TIME ZONE NOT NULL,
    available_seats INTEGER NOT NULL,
    PRIMARY KEY (trip_id, segment_index)
);

CREATE INDEX IF NOT EXISTS idx_stations_city ON stations(city);
CREATE INDEX IF NOT EXISTS idx_stations_org_id ON stations(organization_id);
CREATE INDEX IF NOT EXISTS idx_routes_origin ON routes(origin_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_destination ON routes(destination_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_org_id ON routes(organization_id);
CREATE INDEX IF NOT EXISTS idx_trips_departure ON trips(departure_time);
CREATE INDEX IF NOT EXISTS idx_trips_route_id ON trips(route_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_type ON trips(vehicle_type);
CREATE INDEX IF NOT EXISTS idx_trips_schedule_id ON trips(schedule_id);
CREATE INDEX IF NOT EXISTS idx_trips_service_date ON trips(service_date);
CREATE INDEX IF NOT EXISTS idx_schedule_org_id ON schedule_templates(organization_id);
CREATE INDEX IF NOT EXISTS idx_schedule_route_id ON schedule_templates(route_id);
CREATE INDEX IF NOT EXISTS idx_schedule_active ON schedule_templates(status) WHERE status = 'active';

-- Seed data for stations
-- Legacy stations for backward compatibility
INSERT INTO stations (id, code, name, city, country) VALUES
    ('11111111-1111-1111-1111-111111111111', 'DHA', 'Dhaka (Kamalapur)', 'Dhaka', 'Bangladesh'),
    ('22222222-2222-2222-2222-222222222222', 'CTG', 'Chittagong', 'Chittagong', 'Bangladesh'),
    ('33333333-3333-3333-3333-333333333333', 'SLT', 'Sylhet', 'Sylhet', 'Bangladesh'),
    ('44444444-4444-4444-4444-444444444444', 'CXB', 'Cox''s Bazar', 'Cox''s Bazar', 'Bangladesh'),
    ('55555555-5555-5555-5555-555555555555', 'KHL', 'Khulna', 'Khulna', 'Bangladesh'),
    ('66666666-6666-6666-6666-666666666666', 'RAH', 'Rajshahi', 'Rajshahi', 'Bangladesh')
ON CONFLICT DO NOTHING;

-- Import comprehensive Bangladesh stations data
\i /docker-entrypoint-initdb.d/seed-bd-stations.sql

-- ============================================================================
-- STATIONS PRODUCTION IMPROVEMENTS
-- ============================================================================

-- Add constraints for data integrity
ALTER TABLE stations ADD CONSTRAINT IF NOT EXISTS stations_code_unique UNIQUE (code);
ALTER TABLE routes ADD CONSTRAINT IF NOT EXISTS routes_code_unique_per_org UNIQUE (organization_id, code);
ALTER TABLE trips ADD CONSTRAINT IF NOT EXISTS trips_unique_per_org_departure UNIQUE (organization_id, route_id, departure_time);
ALTER TABLE stations ADD CONSTRAINT IF NOT EXISTS stations_latitude_check CHECK (latitude >= -90 AND latitude <= 90);
ALTER TABLE stations ADD CONSTRAINT IF NOT EXISTS stations_longitude_check CHECK (longitude >= -180 AND longitude <= 180);
ALTER TABLE stations ADD CONSTRAINT IF NOT EXISTS stations_code_format_check CHECK (code ~ '^[A-Z]{3}$');
ALTER TABLE stations ADD CONSTRAINT IF NOT EXISTS stations_status_check CHECK (status IN ('active', 'inactive', 'under_construction', 'maintenance'));

-- Add performance indexes
CREATE INDEX IF NOT EXISTS idx_stations_state ON stations(state) WHERE state IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_stations_city_country ON stations(city, country);
CREATE INDEX IF NOT EXISTS idx_stations_active ON stations(status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_stations_name_search ON stations USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_stations_coordinates ON stations(latitude, longitude);

-- Add auto-update trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DROP TRIGGER IF EXISTS update_stations_updated_at ON stations;
CREATE TRIGGER update_stations_updated_at BEFORE UPDATE ON stations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create utility views
CREATE OR REPLACE VIEW active_stations AS
SELECT id, code, name, city, state, country, latitude, longitude, amenities, timezone, address, created_at, updated_at
FROM stations WHERE status = 'active';

CREATE OR REPLACE VIEW stations_by_division AS
SELECT state as division, COUNT(*) as station_count,
       json_agg(json_build_object('code', code, 'name', name, 'city', city, 'latitude', latitude, 'longitude', longitude) ORDER BY name) as stations
FROM stations WHERE state IS NOT NULL AND status = 'active'
GROUP BY state ORDER BY state;

-- Add table/column documentation
COMMENT ON TABLE stations IS 'Transportation stations (bus terminals, railway stations, ferry ghats) across Bangladesh';
COMMENT ON COLUMN stations.code IS 'IATA-style 3-letter station code (e.g., DHA, CTG, SYL)';
COMMENT ON COLUMN stations.state IS 'Division name (e.g., Dhaka, Chattogram, Sylhet)';

-- Optimize query planner
ANALYZE stations;

-- ==============================================================================
-- 3. ORDER SERVICE (travio_order)
-- ==============================================================================
\c travio_order

CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    user_id UUID,
    trip_id UUID,
    from_station_id UUID,
    to_station_id UUID,
    passengers JSONB DEFAULT '[]', -- Stores snapshot of passenger details
    seats JSONB DEFAULT '[]',      -- Stores booked seat details
    
    -- Pricing
    subtotal_paisa BIGINT DEFAULT 0,
    tax_paisa BIGINT DEFAULT 0,
    booking_fee_paisa BIGINT DEFAULT 0,
    discount_paisa BIGINT DEFAULT 0,
    total_paisa BIGINT DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'BDT',
    
    -- Payment
    payment_id UUID,
    payment_status VARCHAR(50) DEFAULT 'pending', -- pending, authorized, captured, failed, refunded
    payment_method VARCHAR(50),
    
    -- Booking Reference
    booking_id UUID,
    hold_id UUID,
    
    -- Order State
    status VARCHAR(50) DEFAULT 'pending', -- pending, confirmed, failed, cancelled, expired
    saga_id UUID,
    idempotency_key VARCHAR(255) UNIQUE,
    
    -- Contact
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Outbox pattern for reliable event publishing
CREATE TABLE IF NOT EXISTS event_outbox (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- pending, published, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    processed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_trip_id ON orders(trip_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);
CREATE INDEX IF NOT EXISTS idx_event_outbox_status ON event_outbox(status) WHERE status = 'pending';

-- ==============================================================================
-- 4. PAYMENT SERVICE (travio_payment)
-- ==============================================================================
\c travio_payment

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    order_id UUID NOT NULL,
    attempt INTEGER DEFAULT 1,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    gateway VARCHAR(50) NOT NULL, -- sslcommerz, bkash, nagad
    gateway_tx_id VARCHAR(255),   -- External ID from gateway
    status VARCHAR(20) NOT NULL,  -- PENDING, SUCCESS, FAILED
    idempotency_key VARCHAR(255) UNIQUE, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS payment_configs (
    organization_id UUID PRIMARY KEY,
    gateway VARCHAR(50) NOT NULL, -- 'sslcommerz', 'bkash', 'nagad'
    credentials JSONB NOT NULL,   -- Encrypted StoreID/Pass, AppKey/Secret
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_transactions_order_id ON transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_transactions_gateway_tx_id ON transactions(gateway_tx_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- ==============================================================================
-- 5. FULFILLMENT SERVICE (travio_fulfillment)
-- ==============================================================================
\c travio_fulfillment

CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID,
    order_id UUID,
    organization_id UUID,
    
    -- Trip Snapshot
    trip_id UUID,
    route_name VARCHAR(255),
    from_station VARCHAR(255),
    to_station VARCHAR(255),
    departure_time TIMESTAMP WITH TIME ZONE,
    arrival_time TIMESTAMP WITH TIME ZONE,
    
    -- Passenger Info
    passenger_nid VARCHAR(50),
    passenger_name VARCHAR(255),
    seat_number VARCHAR(10),
    seat_class VARCHAR(50),
    
    -- Pricing
    price_paisa BIGINT,
    currency VARCHAR(3),
    
    -- Digital Assets
    qr_code_data TEXT,
    qr_code_url TEXT,
    pdf_url TEXT,
    
    -- State
    status VARCHAR(50) DEFAULT 'active', -- active, used, cancelled, expired
    is_boarded BOOLEAN DEFAULT FALSE,
    boarded_at TIMESTAMP WITH TIME ZONE,
    boarded_by UUID,
    
    valid_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tickets_order_id ON tickets(order_id);
CREATE INDEX IF NOT EXISTS idx_tickets_booking_id ON tickets(booking_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_passenger_nid ON tickets(passenger_nid);

-- ==============================================================================
-- 6. PRICING SERVICE (travio_pricing)
-- ==============================================================================
\c travio_pricing

CREATE TABLE IF NOT EXISTS pricing_rules (
    id VARCHAR(36) PRIMARY KEY,
    organization_id UUID, -- Nullable for Global Rules
    name VARCHAR(255) NOT NULL,
    description TEXT,
    condition TEXT NOT NULL,
    multiplier DECIMAL(5,4) NOT NULL,
    adjustment_type VARCHAR(20) DEFAULT 'multiplier',
    adjustment_value DECIMAL(10,2) DEFAULT 0,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pricing_rules_active ON pricing_rules(is_active, priority);
CREATE INDEX IF NOT EXISTS idx_pricing_rules_org ON pricing_rules(organization_id);

-- Seed default pricing rules
INSERT INTO pricing_rules (id, organization_id, name, description, condition, multiplier, adjustment_type, adjustment_value, priority, is_active) VALUES
    (gen_random_uuid()::text, NULL, 'Weekend Surge', '20% increase on weekends', 'day_of_week == "Saturday" || day_of_week == "Friday"', 1.20, 'multiplier', 0, 10, true),
    (gen_random_uuid()::text, NULL, 'Early Bird', '15% off for 30+ days advance', 'days_until_departure > 30', 0.85, 'multiplier', 0, 20, true),
    (gen_random_uuid()::text, NULL, 'Last Minute', '10% increase for same-day booking', 'days_until_departure < 1', 1.10, 'multiplier', 0, 5, true),
    (gen_random_uuid()::text, NULL, 'Business Class', '40% premium', 'seat_class == "business"', 1.40, 'multiplier', 0, 1, true)
ON CONFLICT DO NOTHING;

-- ==============================================================================
-- 7. VENDOR SERVICE (travio_vendor)
-- ==============================================================================
\c travio_operator

CREATE TABLE IF NOT EXISTS vendors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    address TEXT,
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, suspended
    commission_rate DOUBLE PRECISION DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vendors_email ON vendors(contact_email);
CREATE INDEX IF NOT EXISTS idx_vendors_status ON vendors(status);


-- ==============================================================================
-- 8. SUBSCRIPTION SERVICE (travio_subscription)
-- ==============================================================================
\c travio_subscription

CREATE TABLE IF NOT EXISTS plans (
    id VARCHAR(100) PRIMARY KEY, -- Changed from UUID to support 'plan_free', 'plan_pro'
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_paisa BIGINT NOT NULL,
    interval VARCHAR(50) NOT NULL, -- 'month', 'year'
    features JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    -- Billing & Usage
    usage_price_paisa BIGINT DEFAULT 0, -- Per-unit usage fee
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
    line_items JSONB,
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

-- Seed default plans
INSERT INTO plans (id, name, description, price_paisa, interval, features, is_active) VALUES
    ('plan_starter', 'Basic', 'Essential features for small operators', 20000, 'month', '{"vehicles": "10", "users": "5", "support": "email"}', true),
    ('plan_growth', 'Pro', 'Advanced analytics and larger fleet', 50000, 'month', '{"vehicles": "50", "users": "20", "support": "priority"}', true),
    ('plan_enterprise', 'Enterprise', 'Unlimited scale and dedicated support', 250000, 'month', '{"vehicles": "unlimited", "users": "unlimited", "support": "dedicated"}', true)
ON CONFLICT DO NOTHING;

-- ==============================================================================
-- 9. FRAUD SERVICE (travio_fraud)
-- ==============================================================================
\c travio_fraud
CREATE EXTENSION IF NOT EXISTS vector;

-- ==============================================================================
-- 10. REPORTING SERVICE (travio_reporting)
-- ==============================================================================
SET allow_nullable_key = 1;
-- ==============================================================================
-- 11. EVENTS SERVICE (travio_events)
-- ==============================================================================
\c travio_events

CREATE TABLE IF NOT EXISTS venues (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    capacity INTEGER,
    type VARCHAR(50),
    sections JSONB DEFAULT '[]', -- Seating sections
    map_image_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    venue_id UUID REFERENCES venues(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    images TEXT[],
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'draft',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID REFERENCES events(id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_paisa BIGINT NOT NULL,
    total_quantity INTEGER NOT NULL,
    available_quantity INTEGER NOT NULL,
    max_per_user INTEGER DEFAULT 10,
    sales_start_time TIMESTAMP WITH TIME ZONE,
    sales_end_time TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_events_org ON events(organization_id);
CREATE INDEX IF NOT EXISTS idx_events_venue ON events(venue_id);
CREATE INDEX IF NOT EXISTS idx_events_start ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_ticket_types_event ON ticket_types(event_id);

-- ==============================================================================
-- 12. CRM SERVICE (travio_crm)
-- ==============================================================================
\c travio_crm

CREATE TABLE IF NOT EXISTS coupons (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    discount_type VARCHAR(50) NOT NULL, -- percentage, fixed
    discount_value DECIMAL(10,2) NOT NULL,
    min_purchase_amount BIGINT DEFAULT 0,
    max_discount_amount BIGINT DEFAULT 0,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    usage_limit INTEGER DEFAULT 0,
    usage_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS support_tickets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    user_id UUID,
    subject VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'open',
    priority VARCHAR(50) DEFAULT 'normal',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ticket_id UUID REFERENCES support_tickets(id),
    sender_id UUID,
    message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_coupons_code_org ON coupons(organization_id, code);
CREATE INDEX IF NOT EXISTS idx_tickets_org ON support_tickets(organization_id);
CREATE INDEX IF NOT EXISTS idx_tickets_user ON support_tickets(user_id);

-- ==============================================================================
-- 13. FLEET SERVICE (travio_fleet)
-- ==============================================================================
\c travio_fleet

CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL,
    name VARCHAR(255),
    license_plate VARCHAR(50),
    vin VARCHAR(100),
    make VARCHAR(100),
    model VARCHAR(255),
    year INTEGER,
    type VARCHAR(50) NOT NULL, -- bus, train, launch
    status VARCHAR(50) DEFAULT 'active',
    config JSONB DEFAULT '{}', -- Layout config
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS asset_locations (
    asset_id UUID PRIMARY KEY REFERENCES assets(id),
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    speed DOUBLE PRECISION,
    heading DOUBLE PRECISION,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- History table for tracking (optional, maybe uses TimeScaleDB in future)
CREATE TABLE IF NOT EXISTS location_history (
    time TIMESTAMP WITH TIME ZONE NOT NULL,
    asset_id UUID NOT NULL,
    latitude DOUBLE PRECISION,
    longitude DOUBLE PRECISION,
    speed DOUBLE PRECISION
);
-- SELECT create_hypertable('location_history', 'time', if_not_exists => TRUE); 

CREATE INDEX IF NOT EXISTS idx_assets_org ON assets(organization_id);
