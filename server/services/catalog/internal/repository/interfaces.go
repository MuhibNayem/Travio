package repository

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
)

type StationRepository interface {
	Create(ctx context.Context, station *domain.Station) error
	GetByID(ctx context.Context, id, orgID string) (*domain.Station, error)
	List(ctx context.Context, orgID, city string, limit, offset int) ([]*domain.Station, int, error)
	Update(ctx context.Context, station *domain.Station) error
	Delete(ctx context.Context, id, orgID string) error
}

type RouteRepository interface {
	Create(ctx context.Context, route *domain.Route) error
	GetByID(ctx context.Context, id, orgID string) (*domain.Route, error)
	List(ctx context.Context, orgID, originID, destID string, limit, offset int) ([]*domain.Route, int, error)
}

type TripRepository interface {
	Create(ctx context.Context, trip *domain.Trip) error
	GetByID(ctx context.Context, id, orgID string) (*domain.Trip, error)
	Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error)
	UpdateStatus(ctx context.Context, id, orgID, status string) error
	DecrementSeats(ctx context.Context, id string, count int) error
}
