package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
)

// CachedTripRepository wraps a TripRepository with multi-level caching.
type CachedTripRepository struct {
	repo  repository.TripRepository
	cache *MultiLevelCache
}

// NewCachedTripRepository creates a new cached trip repository.
func NewCachedTripRepository(repo repository.TripRepository, cache *MultiLevelCache) *CachedTripRepository {
	return &CachedTripRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *CachedTripRepository) cacheKey(id, orgID string) string {
	return fmt.Sprintf("trip:%s:%s", orgID, id)
}

// GetByID retrieves a trip, checking cache first.
func (r *CachedTripRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	key := r.cacheKey(id, orgID)

	// Check cache
	var trip domain.Trip
	if err := r.cache.GetAs(ctx, key, &trip); err == nil && trip.ID != "" {
		return &trip, nil
	}

	// Cache miss - fetch from repository
	result, err := r.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	// Populate cache
	_ = r.cache.SetAs(ctx, key, result)
	return result, nil
}

// Create creates a trip and caches it.
func (r *CachedTripRepository) Create(ctx context.Context, trip *domain.Trip) error {
	if err := r.repo.Create(ctx, trip); err != nil {
		return err
	}
	// Cache the new trip
	_ = r.cache.SetAs(ctx, r.cacheKey(trip.ID, trip.OrganizationID), trip)
	return nil
}

// Search searches for trips (no caching for search results - too dynamic).
func (r *CachedTripRepository) Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error) {
	return r.repo.Search(ctx, orgID, originCity, destCity, travelDate, limit, offset)
}

// UpdateStatus updates trip status and invalidates cache.
func (r *CachedTripRepository) UpdateStatus(ctx context.Context, id, orgID, status string) error {
	if err := r.repo.UpdateStatus(ctx, id, orgID, status); err != nil {
		return err
	}
	// Invalidate cache
	_ = r.cache.Delete(ctx, r.cacheKey(id, orgID))
	return nil
}

// DecrementSeats decrements available seats and invalidates cache by trip ID pattern.
func (r *CachedTripRepository) DecrementSeats(ctx context.Context, id string, count int) error {
	if err := r.repo.DecrementSeats(ctx, id, count); err != nil {
		return err
	}
	// Invalidate all caches matching this trip ID (any org)
	// Pattern: trip:*:<id>
	return r.cache.InvalidateByPattern(ctx, fmt.Sprintf("trip:*:%s", id))
}

// CachedStationRepository wraps a StationRepository with multi-level caching.
type CachedStationRepository struct {
	repo  repository.StationRepository
	cache *MultiLevelCache
}

// NewCachedStationRepository creates a new cached station repository.
func NewCachedStationRepository(repo repository.StationRepository, cache *MultiLevelCache) *CachedStationRepository {
	return &CachedStationRepository{
		repo:  repo,
		cache: cache,
	}
}

func (r *CachedStationRepository) cacheKey(id, orgID string) string {
	return fmt.Sprintf("station:%s:%s", orgID, id)
}

// GetByID retrieves a station, checking cache first.
func (r *CachedStationRepository) GetByID(ctx context.Context, id, orgID string) (*domain.Station, error) {
	key := r.cacheKey(id, orgID)

	var station domain.Station
	if err := r.cache.GetAs(ctx, key, &station); err == nil && station.ID != "" {
		return &station, nil
	}

	result, err := r.repo.GetByID(ctx, id, orgID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, nil
	}

	_ = r.cache.SetAs(ctx, key, result)
	return result, nil
}

func (r *CachedStationRepository) Create(ctx context.Context, station *domain.Station) error {
	if err := r.repo.Create(ctx, station); err != nil {
		return err
	}
	_ = r.cache.SetAs(ctx, r.cacheKey(station.ID, station.OrganizationID), station)
	return nil
}

func (r *CachedStationRepository) Update(ctx context.Context, station *domain.Station) error {
	if err := r.repo.Update(ctx, station); err != nil {
		return err
	}
	_ = r.cache.SetAs(ctx, r.cacheKey(station.ID, station.OrganizationID), station)
	return nil
}

func (r *CachedStationRepository) Delete(ctx context.Context, id, orgID string) error {
	if err := r.repo.Delete(ctx, id, orgID); err != nil {
		return err
	}
	_ = r.cache.Delete(ctx, r.cacheKey(id, orgID))
	return nil
}

func (r *CachedStationRepository) List(ctx context.Context, orgID, city string, limit, offset int) ([]*domain.Station, int, error) {
	return r.repo.List(ctx, orgID, city, limit, offset)
}
