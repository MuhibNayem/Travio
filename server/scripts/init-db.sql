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
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_info VARCHAR(500),
    ip_address VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
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
    amenities TEXT[], -- PostgreSQL array
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
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
    organization_id UUID,
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

CREATE INDEX IF NOT EXISTS idx_stations_city ON stations(city);
CREATE INDEX IF NOT EXISTS idx_stations_org_id ON stations(organization_id);
CREATE INDEX IF NOT EXISTS idx_routes_origin ON routes(origin_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_destination ON routes(destination_station_id);
CREATE INDEX IF NOT EXISTS idx_trips_departure ON trips(departure_time);
CREATE INDEX IF NOT EXISTS idx_trips_route_id ON trips(route_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_type ON trips(vehicle_type);

-- Seed data for stations
INSERT INTO stations (id, code, name, city, country) VALUES
    ('11111111-1111-1111-1111-111111111111', 'DHA', 'Dhaka (Kamalapur)', 'Dhaka', 'Bangladesh'),
    ('22222222-2222-2222-2222-222222222222', 'CTG', 'Chittagong', 'Chittagong', 'Bangladesh'),
    ('33333333-3333-3333-3333-333333333333', 'SYL', 'Sylhet', 'Sylhet', 'Bangladesh'),
    ('44444444-4444-4444-4444-444444444444', 'CXB', 'Cox''s Bazar', 'Cox''s Bazar', 'Bangladesh'),
    ('55555555-5555-5555-5555-555555555555', 'KHL', 'Khulna', 'Khulna', 'Bangladesh'),
    ('66666666-6666-6666-6666-666666666666', 'RAJ', 'Rajshahi', 'Rajshahi', 'Bangladesh')
ON CONFLICT DO NOTHING;

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
    order_id UUID NOT NULL,
    attempt INTEGER DEFAULT 1,
    amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL,
    gateway VARCHAR(50) NOT NULL, -- sslcommerz, bkash, nagad
    gateway_tx_id VARCHAR(255),   -- External ID from gateway
    status VARCHAR(20) NOT NULL,  -- PENDING, SUCCESS, FAILED
    idempotency_key UUID UNIQUE, 
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
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
    name VARCHAR(255) NOT NULL,
    description TEXT,
    condition TEXT NOT NULL,
    multiplier DECIMAL(5,4) NOT NULL,
    priority INT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_pricing_rules_active ON pricing_rules(is_active, priority);

-- Seed default pricing rules
INSERT INTO pricing_rules (id, name, description, condition, multiplier, priority, is_active) VALUES
    (gen_random_uuid()::text, 'Weekend Surge', '20% increase on weekends', 'day_of_week == "Saturday" || day_of_week == "Friday"', 1.20, 10, true),
    (gen_random_uuid()::text, 'Early Bird', '15% off for 30+ days advance', 'days_until_departure > 30', 0.85, 20, true),
    (gen_random_uuid()::text, 'Last Minute', '10% increase for < 24h', 'hours_until_departure < 24', 1.10, 5, true),
    (gen_random_uuid()::text, 'Business Class', '40% premium', 'seat_class == "business"', 1.40, 1, true)
ON CONFLICT DO NOTHING;

\echo 'Master database initialization complete!'
