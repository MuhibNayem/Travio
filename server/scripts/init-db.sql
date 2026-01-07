-- Travio Database Initialization Script
-- Creates schemas for each service in the shared database

-- Identity Service Schema
CREATE SCHEMA IF NOT EXISTS identity;

-- Create users table
CREATE TABLE IF NOT EXISTS identity.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    organization_id UUID,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create organizations table
CREATE TABLE IF NOT EXISTS identity.organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create refresh tokens table
CREATE TABLE IF NOT EXISTS identity.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES identity.users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    device_info VARCHAR(500),
    ip_address VARCHAR(50),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON identity.refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token_hash ON identity.refresh_tokens(token_hash);

-- Catalog Service Schema
CREATE SCHEMA IF NOT EXISTS catalog;

-- Create stations table
CREATE TABLE IF NOT EXISTS catalog.stations (
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

CREATE INDEX IF NOT EXISTS idx_stations_city ON catalog.stations(city);
CREATE INDEX IF NOT EXISTS idx_stations_org_id ON catalog.stations(organization_id);

-- Create routes table
CREATE TABLE IF NOT EXISTS catalog.routes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    origin_station_id UUID REFERENCES catalog.stations(id),
    destination_station_id UUID REFERENCES catalog.stations(id),
    intermediate_stops JSONB DEFAULT '[]',
    distance_km INTEGER,
    estimated_duration_minutes INTEGER,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_routes_origin ON catalog.routes(origin_station_id);
CREATE INDEX IF NOT EXISTS idx_routes_destination ON catalog.routes(destination_station_id);

-- Create trips table
CREATE TABLE IF NOT EXISTS catalog.trips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID,
    route_id UUID REFERENCES catalog.routes(id),
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

CREATE INDEX IF NOT EXISTS idx_trips_departure ON catalog.trips(departure_time);
CREATE INDEX IF NOT EXISTS idx_trips_route_id ON catalog.trips(route_id);
CREATE INDEX IF NOT EXISTS idx_trips_vehicle_type ON catalog.trips(vehicle_type);

-- Seed data for stations
INSERT INTO catalog.stations (id, code, name, city, country) VALUES
    ('11111111-1111-1111-1111-111111111111', 'DHA', 'Dhaka (Kamalapur)', 'Dhaka', 'Bangladesh'),
    ('22222222-2222-2222-2222-222222222222', 'CTG', 'Chittagong', 'Chittagong', 'Bangladesh'),
    ('33333333-3333-3333-3333-333333333333', 'SYL', 'Sylhet', 'Sylhet', 'Bangladesh'),
    ('44444444-4444-4444-4444-444444444444', 'CXB', 'Cox''s Bazar', 'Cox''s Bazar', 'Bangladesh'),
    ('55555555-5555-5555-5555-555555555555', 'KHL', 'Khulna', 'Khulna', 'Bangladesh'),
    ('66666666-6666-6666-6666-666666666666', 'RAJ', 'Rajshahi', 'Rajshahi', 'Bangladesh')
ON CONFLICT DO NOTHING;

-- Grant permissions (adjust as needed for your setup)
GRANT ALL PRIVILEGES ON SCHEMA identity TO postgres;
GRANT ALL PRIVILEGES ON SCHEMA catalog TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA identity TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA catalog TO postgres;

\echo 'Database initialization complete!'
