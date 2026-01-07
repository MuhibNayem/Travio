package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	StationCacheTTL = 24 * time.Hour
	RouteCacheTTL   = 24 * time.Hour
	TripCacheTTL    = 5 * time.Minute // Shorter TTL for dynamic availability
)

// CachedStationRepository decorates a StationRepository with Redis caching
type CachedStationRepository struct {
	next StationRepository
	rdb  *redis.Client
}

func NewCachedStationRepository(next StationRepository, rdb *redis.Client) *CachedStationRepository {
	return &CachedStationRepository{next: next, rdb: rdb}
}

func (r *CachedStationRepository) Create(ctx context.Context, station *domain.Station) error {
	// Write-through or invalidate? Invalidate is safer for consistency.
	// But let's just invalidate since Create isn't high frequency.
	return r.next.Create(ctx, station)
}

func (r *CachedStationRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Station, error) {
	key := fmt.Sprintf("station:%s:%s", orgID, id)

	// Read-Through
	val, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var station domain.Station
		if err := json.Unmarshal([]byte(val), &station); err == nil {
			return &station, nil
		}
	}

	// Fallback to DB
	station, err := r.next.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	// Cache it
	if data, err := json.Marshal(station); err == nil {
		r.rdb.Set(ctx, key, data, StationCacheTTL)
	}

	return station, nil
}

func (r *CachedStationRepository) List(ctx context.Context, orgID, city string, limit, offset int) ([]*domain.Station, int, error) {
	// Caching Lists is harder due to pagination and filters.
	// For FAANG scale, we might cache common queries (e.g., "all stations for org").
	// For now, let's pass through.
	return r.next.List(ctx, orgID, city, limit, offset)
}

func (r *CachedStationRepository) Update(ctx context.Context, station *domain.Station) error {
	err := r.next.Update(ctx, station)
	if err == nil {
		key := fmt.Sprintf("station:%s:%s", station.OrganizationID, station.ID)
		r.rdb.Del(ctx, key)
	}
	return err
}

func (r *CachedStationRepository) Delete(ctx context.Context, id, orgID string) error {
	err := r.next.Delete(ctx, id, orgID)
	if err == nil {
		key := fmt.Sprintf("station:%s:%s", orgID, id)
		r.rdb.Del(ctx, key)
	}
	return err
}

// CachedRouteRepository
type CachedRouteRepository struct {
	next RouteRepository
	rdb  *redis.Client
}

func NewCachedRouteRepository(next RouteRepository, rdb *redis.Client) *CachedRouteRepository {
	return &CachedRouteRepository{next: next, rdb: rdb}
}

func (r *CachedRouteRepository) Create(ctx context.Context, route *domain.Route) error {
	return r.next.Create(ctx, route)
}

func (r *CachedRouteRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Route, error) {
	key := fmt.Sprintf("route:%s:%s", orgID, id)

	val, err := r.rdb.Get(ctx, key).Result()
	if err == nil {
		var route domain.Route
		if err := json.Unmarshal([]byte(val), &route); err == nil {
			return &route, nil
		}
	}

	route, err := r.next.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(route); err == nil {
		r.rdb.Set(ctx, key, data, RouteCacheTTL)
	}

	return route, nil
}

func (r *CachedRouteRepository) List(ctx context.Context, orgID, originID, destID string, limit, offset int) ([]*domain.Route, int, error) {
	return r.next.List(ctx, orgID, originID, destID, limit, offset)
}
