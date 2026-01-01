-- ScyllaDB Schema for Inventory Service
-- Designed for 100K+ QPS during peak booking
-- Run with: cqlsh -f schema.cql

CREATE KEYSPACE IF NOT EXISTS travio_inventory
WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'datacenter1': 3  -- 3 replicas per datacenter
};

USE travio_inventory;

-- Segments table: stores trip segment metadata
-- Partition key: trip_id (all segments for a trip colocated)
CREATE TABLE IF NOT EXISTS segments (
    trip_id text,
    segment_index int,
    from_station_id text,
    to_station_id text,
    departure_time timestamp,
    arrival_time timestamp,
    PRIMARY KEY (trip_id, segment_index)
) WITH CLUSTERING ORDER BY (segment_index ASC)
  AND compaction = {'class': 'LeveledCompactionStrategy'}
  AND gc_grace_seconds = 86400;

-- Seat Inventory table: the HOT table for availability checks
-- Partition key: (trip_id, segment_index) for quick segment lookups
-- Each partition contains all seats for one segment of one trip
CREATE TABLE IF NOT EXISTS seat_inventory (
    trip_id text,
    segment_index int,
    seat_id text,
    seat_number text,
    seat_class text,          -- economy, business, ac
    seat_type text,           -- window, aisle, middle
    status text,              -- available, held, booked, blocked
    hold_id text,
    hold_user_id text,
    hold_expiry timestamp,
    booking_id text,
    price_paisa bigint,
    updated_at timestamp,
    PRIMARY KEY ((trip_id, segment_index), seat_id)
) WITH compaction = {'class': 'LeveledCompactionStrategy'}
  AND gc_grace_seconds = 86400
  AND default_time_to_live = 0;  -- No TTL on seats

-- Materialized view for finding seats by status (for admin dashboards)
CREATE MATERIALIZED VIEW IF NOT EXISTS seats_by_status AS
    SELECT trip_id, segment_index, status, seat_id, seat_number, seat_class
    FROM seat_inventory
    WHERE trip_id IS NOT NULL
      AND segment_index IS NOT NULL
      AND seat_id IS NOT NULL
      AND status IS NOT NULL
    PRIMARY KEY ((trip_id, status), segment_index, seat_id);

-- Counter table for quick availability counts (eventual consistency acceptable for search)
CREATE TABLE IF NOT EXISTS availability_counters (
    trip_id text,
    segment_index int,
    seat_class text,
    available_count counter,
    PRIMARY KEY ((trip_id, segment_index), seat_class)
);

-- Bookings table: stores confirmed bookings
CREATE TABLE IF NOT EXISTS bookings (
    booking_id text PRIMARY KEY,
    order_id text,
    trip_id text,
    user_id text,
    from_station_id text,
    to_station_id text,
    segment_range list<int>,
    seats list<frozen<map<text, text>>>,  -- [{seat_id: X, seat_number: Y, ticket_id: Z}]
    total_paisa bigint,
    status text,
    created_at timestamp,
    updated_at timestamp
);

-- Secondary index for user booking lookups
CREATE INDEX IF NOT EXISTS bookings_by_user ON bookings (user_id);
CREATE INDEX IF NOT EXISTS bookings_by_trip ON bookings (trip_id);
CREATE INDEX IF NOT EXISTS bookings_by_order ON bookings (order_id);

-- =========================================
-- Example Data Initialization
-- =========================================

-- For a trip DHA -> CTG with stops at DHA, COM, FEN, CTG:
-- Segments: 0 (DHA->COM), 1 (COM->FEN), 2 (FEN->CTG)
-- Journey DHA->CTG requires seats available on segments [0,1,2]
-- Journey COM->CTG requires seats available on segments [1,2]

-- Example insert for initializing a 40-seat bus
-- INSERT INTO seat_inventory (trip_id, segment_index, seat_id, seat_number, seat_class, seat_type, status, price_paisa, updated_at)
-- VALUES ('trip-001', 0, 'seat-a1', 'A1', 'ac', 'window', 'available', 80000, toTimestamp(now()));
