-- ============================================================================
-- Stations Database Production Improvements Migration
-- ============================================================================
-- Version: V003
-- Description: Add constraints, indexes, and validations for production-ready stations
-- Author: Travio Team
-- Date: 2026-01-11

\c travio_catalog

-- ============================================================================
-- 1. ADD CONSTRAINTS
-- ============================================================================

-- Add unique constraint on station code (industry standard)
ALTER TABLE stations 
ADD CONSTRAINT stations_code_unique UNIQUE (code);

-- Add check constraint for valid coordinate ranges
ALTER TABLE stations
ADD CONSTRAINT stations_latitude_check 
CHECK (latitude >= -90 AND latitude <= 90);

ALTER TABLE stations
ADD CONSTRAINT stations_longitude_check 
CHECK (longitude >= -180 AND longitude <= 180);

-- Add check constraint for station code format (3 uppercase letters)
ALTER TABLE stations
ADD CONSTRAINT stations_code_format_check 
CHECK (code ~ '^[A-Z]{3}$');

-- Add check constraint for valid status values
ALTER TABLE stations
ADD CONSTRAINT stations_status_check 
CHECK (status IN ('active', 'inactive', 'under_construction', 'maintenance'));

-- ============================================================================
-- 2. ADD PERFORMANCE INDEXES
-- ============================================================================

-- Index on state/division for filtering by region
-- (idx_stations_city already exists from init-db.sql)
CREATE INDEX IF NOT EXISTS idx_stations_state ON stations(state) 
WHERE state IS NOT NULL;

-- Composite index for common query: city + country
CREATE INDEX IF NOT EXISTS idx_stations_city_country ON stations(city, country);

-- Index for status filtering (partial index for active stations only)
CREATE INDEX IF NOT EXISTS idx_stations_active ON stations(status) 
WHERE status = 'active';

-- Full-text search index on station name
CREATE INDEX IF NOT EXISTS idx_stations_name_search ON stations 
USING gin(to_tsvector('english', name));

-- Spatial index for coordinate-based queries (if needed for nearby stations)
-- Note: For true spatial queries, consider upgrading to PostGIS in future
CREATE INDEX IF NOT EXISTS idx_stations_coordinates ON stations(latitude, longitude);

-- ============================================================================
-- 3. ADD PERFORMANCE OPTIMIZATIONS
-- ============================================================================

-- Update updated_at trigger function if not exists
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Add trigger to automatically update updated_at
DROP TRIGGER IF EXISTS update_stations_updated_at ON stations;
CREATE TRIGGER update_stations_updated_at
    BEFORE UPDATE ON stations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- 4. ADD HELPFUL VIEWS
-- ============================================================================

-- Create view for active stations only
CREATE OR REPLACE VIEW active_stations AS
SELECT id, code, name, city, state, country, latitude, longitude, 
       amenities, timezone, address, created_at, updated_at
FROM stations
WHERE status = 'active';

-- Create view for stations grouped by division
CREATE OR REPLACE VIEW stations_by_division AS
SELECT 
    state as division,
    COUNT(*) as station_count,
    json_agg(
        json_build_object(
            'code', code,
            'name', name,
            'city', city,
            'latitude', latitude,
            'longitude', longitude
        ) ORDER BY name
    ) as stations
FROM stations
WHERE state IS NOT NULL AND status = 'active'
GROUP BY state
ORDER BY state;

-- ============================================================================
-- 5. ADD DATA QUALITY FUNCTIONS
-- ============================================================================

-- Function to find nearby stations (simple distance calculation)
CREATE OR REPLACE FUNCTION find_nearby_stations(
    p_latitude DOUBLE PRECISION,
    p_longitude DOUBLE PRECISION,
    p_radius_km DOUBLE PRECISION DEFAULT 50
)
RETURNS TABLE (
    station_code VARCHAR,
    station_name VARCHAR,
    distance_km DOUBLE PRECISION
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        s.code,
        s.name,
        -- Haversine formula for distance (approximate)
        (6371 * acos(
            cos(radians(p_latitude)) * 
            cos(radians(s.latitude)) * 
            cos(radians(s.longitude) - radians(p_longitude)) + 
            sin(radians(p_latitude)) * 
            sin(radians(s.latitude))
        )) as distance
    FROM stations s
    WHERE s.status = 'active'
        AND s.latitude IS NOT NULL
        AND s.longitude IS NOT NULL
    HAVING (6371 * acos(
        cos(radians(p_latitude)) * 
        cos(radians(s.latitude)) * 
        cos(radians(s.longitude) - radians(p_longitude)) + 
        sin(radians(p_latitude)) * 
        sin(radians(s.latitude))
    )) <= p_radius_km
    ORDER BY distance;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- 6. ADD COMMENTS FOR DOCUMENTATION
-- ============================================================================

COMMENT ON TABLE stations IS 'Transportation stations (bus terminals, railway stations, ferry ghats) across Bangladesh';
COMMENT ON COLUMN stations.code IS 'IATA-style 3-letter station code (e.g., DHA, CTG, SYL)';
COMMENT ON COLUMN stations.state IS 'Division name (e.g., Dhaka, Chattogram, Sylhet)';
COMMENT ON COLUMN stations.latitude IS 'Latitude in decimal degrees (-90 to 90)';
COMMENT ON COLUMN stations.longitude IS 'Longitude in decimal degrees (-180 to 180)';
COMMENT ON COLUMN stations.amenities IS 'Array of available amenities (wifi, parking, restroom, etc.)';

-- ============================================================================
-- 7. ANALYZE FOR QUERY OPTIMIZATION
-- ============================================================================

ANALYZE stations;

-- ============================================================================
-- VERIFICATION QUERIES
-- ============================================================================

-- Verify constraints are in place
-- SELECT conname, contype FROM pg_constraint WHERE conrelid = 'stations'::regclass;

-- Verify indexes are created
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'stations';

-- Test full-text search
-- SELECT code, name FROM stations WHERE to_tsvector('english', name) @@ to_tsquery('dhaka');

-- Test nearby stations function
-- SELECT * FROM find_nearby_stations(23.8103, 90.4125, 50);
