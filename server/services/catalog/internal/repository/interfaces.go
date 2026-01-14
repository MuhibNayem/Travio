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
	List(ctx context.Context, orgID, routeID, scheduleID, serviceDateFrom, serviceDateTo string, limit, offset int) ([]*domain.Trip, int, error)
	Search(ctx context.Context, orgID, originCity, destCity string, travelDate time.Time, limit, offset int) ([]*domain.Trip, int, error)
	UpdateStatus(ctx context.Context, id, orgID, status string) error
	DecrementSeats(ctx context.Context, id string, count int) error
	CreateSegments(ctx context.Context, tripID string, segments []domain.TripSegment) error
	GetSegments(ctx context.Context, tripID string) ([]domain.TripSegment, error)
}

type ScheduleRepository interface {
	Create(ctx context.Context, schedule *domain.ScheduleTemplate) error
	GetByID(ctx context.Context, id, orgID string) (*domain.ScheduleTemplate, error)
	List(ctx context.Context, orgID, routeID, status string, limit, offset int) ([]*domain.ScheduleTemplate, int, error)
	Update(ctx context.Context, schedule *domain.ScheduleTemplate) error
	Delete(ctx context.Context, id, orgID string) error
	AddException(ctx context.Context, exception *domain.ScheduleException) error
	ListExceptions(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleException, error)
	HasVehicleConflict(ctx context.Context, schedule *domain.ScheduleTemplate, excludeID string) (bool, error)
	GetHistory(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleVersion, error)
}
