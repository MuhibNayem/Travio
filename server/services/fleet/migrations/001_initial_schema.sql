-- Initial Schema for Fleet Service
-- Creates assets and asset_locations tables

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

CREATE INDEX IF NOT EXISTS idx_assets_org ON assets(organization_id);
