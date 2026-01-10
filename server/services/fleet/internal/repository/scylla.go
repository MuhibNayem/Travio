package repository

import (
	"time"

	"github.com/MuhibNayem/Travio/server/services/fleet/internal/domain"
	"github.com/gocql/gocql"
)

type LocationRepository struct {
	Session *gocql.Session
}

func NewLocationRepository(hosts ...string) (*LocationRepository, error) {
	cluster := gocql.NewCluster(hosts...)
	cluster.Keyspace = "travio_fleet"
	cluster.Consistency = gocql.One // High write throughput
	cluster.Timeout = 5 * time.Second
	cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}

	return &LocationRepository{Session: session}, nil
}

func (r *LocationRepository) Close() {
	r.Session.Close()
}

func (r *LocationRepository) UpdateLocation(loc *domain.AssetLocation) error {
	// 1. Update Latest Location (Overwrite)
	stmtLatest := `INSERT INTO asset_latest_locations (asset_id, organization_id, latitude, longitude, speed, heading, timestamp) 
                   VALUES (?, ?, ?, ?, ?, ?, ?)`

	if err := r.Session.Query(stmtLatest, loc.AssetID, loc.OrganizationID, loc.Latitude, loc.Longitude, loc.Speed, loc.Heading, loc.Timestamp).Exec(); err != nil {
		return err
	}

	// 2. Append to History (Time Series)
	// Partition by day
	bucket := loc.Timestamp.Format("2006-01-02")
	stmtHistory := `INSERT INTO location_history (asset_id, bucket, timestamp, latitude, longitude, speed, heading) 
                    VALUES (?, ?, ?, ?, ?, ?, ?)`

	return r.Session.Query(stmtHistory, loc.AssetID, bucket, loc.Timestamp, loc.Latitude, loc.Longitude, loc.Speed, loc.Heading).Exec()
}

func (r *LocationRepository) GetLatestLocation(assetID string) (*domain.AssetLocation, error) {
	query := `SELECT asset_id, organization_id, latitude, longitude, speed, heading, timestamp FROM asset_latest_locations WHERE asset_id = ?`

	var loc domain.AssetLocation
	if err := r.Session.Query(query, assetID).Scan(&loc.AssetID, &loc.OrganizationID, &loc.Latitude, &loc.Longitude, &loc.Speed, &loc.Heading, &loc.Timestamp); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (r *LocationRepository) GetLocationHistory(assetID string, start, end time.Time) ([]*domain.AssetLocation, error) {
	// Note: Cross-bucket queries (multiple days) need multiple queries in Cassandra/Scylla.
	// For simplicity, we'll assume the query is within 24h or handle basic single-day query for now.
	// Ideally, we'd iterate through days between start and end.

	bucket := start.Format("2006-01-02") // Simplified: single day query

	query := `SELECT latitude, longitude, speed, heading, timestamp FROM location_history 
              WHERE asset_id = ? AND bucket = ? AND timestamp >= ? AND timestamp <= ?`

	iter := r.Session.Query(query, assetID, bucket, start, end).Iter()

	var locations []*domain.AssetLocation
	var loc domain.AssetLocation
	// Note: We need separate scan vars to avoid rewriting the same struct pointer if we were passing &loc
	// But scan copies values, so reusing struct is fine if we append copies or new pointers.
	for iter.Scan(&loc.Latitude, &loc.Longitude, &loc.Speed, &loc.Heading, &loc.Timestamp) {
		l := loc // copy
		l.AssetID = assetID
		locations = append(locations, &l)
	}

	if err := iter.Close(); err != nil {
		return nil, err
	}
	return locations, nil
}
