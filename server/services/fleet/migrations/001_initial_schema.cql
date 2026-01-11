-- ScyllaDB Schema for Fleet Service
-- Optimized for high-write location tracking

CREATE KEYSPACE IF NOT EXISTS travio_fleet
WITH replication = {
    'class': 'NetworkTopologyStrategy',
    'datacenter1': 3
};

USE travio_fleet;

-- Asset Last Location (Latest State)
-- Fast lookups for "Where is bus X right now?"
CREATE TABLE IF NOT EXISTS asset_latest_locations (
    asset_id text PRIMARY KEY,
    organization_id text,
    latitude double,
    longitude double,
    speed double,
    heading double,
    timestamp timestamp
) WITH compaction = {'class': 'LeveledCompactionStrategy'};

-- Location History (Time Series)
-- "Where was bus X yesterday?"
-- Partitioned by asset_id and day (bucket) to avoid huge partitions
CREATE TABLE IF NOT EXISTS location_history (
    asset_id text,
    bucket text, -- YYYY-MM-DD
    timestamp timestamp,
    latitude double,
    longitude double,
    speed double,
    heading double,
    PRIMARY KEY ((asset_id, bucket), timestamp)
) WITH CLUSTERING ORDER BY (timestamp DESC)
  AND compaction = {
      'class': 'TimeWindowCompactionStrategy',
      'compaction_window_unit': 'DAYS',
      'compaction_window_size': 1
  }
  AND default_time_to_live = 3888000; -- 45 Days Retention (stays within 50 window limit)

-- Secondary index for finding assets by organization (for dashboard list)
CREATE INDEX IF NOT EXISTS asset_latest_locations_org_idx ON asset_latest_locations (organization_id);
