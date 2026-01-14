package repository

import (
	"context"
	"fmt"
)

// InitSchema ensures the ScyllaDB schema is up to date
func (r *LocationRepository) InitSchema(ctx context.Context, keyspace string) error {
	queries := []string{
		fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}", keyspace),

		`CREATE TABLE IF NOT EXISTS asset_latest_locations (
			asset_id text PRIMARY KEY,
			organization_id text,
			latitude double,
			longitude double,
			speed double,
			heading double,
			timestamp timestamp
		) WITH compaction = {'class': 'LeveledCompactionStrategy'}`,

		`CREATE TABLE IF NOT EXISTS location_history (
			asset_id text,
			bucket text,
			timestamp timestamp,
			latitude double,
			longitude double,
			speed double,
			heading double,
			PRIMARY KEY ((asset_id, bucket), timestamp)
		) WITH CLUSTERING ORDER BY (timestamp DESC)
		  AND compaction = {'class': 'TimeWindowCompactionStrategy', 'compaction_window_unit': 'DAYS', 'compaction_window_size': 1}`,
	}

	for _, query := range queries {
		if err := r.Session.Query(query).WithContext(ctx).Exec(); err != nil {
			return fmt.Errorf("scylla schema init failed: %w query: %s", err, query)
		}
	}
	return nil
}
