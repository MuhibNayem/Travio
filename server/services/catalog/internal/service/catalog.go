package service

import (
	"context"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
)

// CatalogService handles business logic for catalog operations
type CatalogService struct {
	stationRepo repository.StationRepository
	routeRepo   repository.RouteRepository
	tripRepo    repository.TripRepository
	checker     entitlement.EntitlementChecker
}

func NewCatalogService(
	stationRepo repository.StationRepository,
	routeRepo repository.RouteRepository,
	tripRepo repository.TripRepository,
	checker entitlement.EntitlementChecker,
) *CatalogService {
	return &CatalogService{
		stationRepo: stationRepo,
		routeRepo:   routeRepo,
		tripRepo:    tripRepo,
		checker:     checker,
	}
}

// --- Station Operations ---

func (s *CatalogService) CreateStation(ctx context.Context, station *domain.Station) (*domain.Station, error) {
	if err := s.stationRepo.Create(ctx, station); err != nil {
		return nil, err
	}
	return station, nil
}

func (s *CatalogService) GetStation(ctx context.Context, id, orgID string) (*domain.Station, error) {
	return s.stationRepo.GetByID(ctx, id, orgID)
}

func (s *CatalogService) ListStations(ctx context.Context, orgID, city string, pageSize int, pageToken string) ([]*domain.Station, int, string, error) {
	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	stations, total, err := s.stationRepo.List(ctx, orgID, city, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return stations, total, nextToken, nil
}

func (s *CatalogService) UpdateStation(ctx context.Context, station *domain.Station) (*domain.Station, error) {
	if err := s.stationRepo.Update(ctx, station); err != nil {
		return nil, err
	}
	return s.stationRepo.GetByID(ctx, station.ID, station.OrganizationID)
}

func (s *CatalogService) DeleteStation(ctx context.Context, id, orgID string) error {
	return s.stationRepo.Delete(ctx, id, orgID)
}

// --- Route Operations ---

func (s *CatalogService) CreateRoute(ctx context.Context, route *domain.Route) (*domain.Route, error) {
	// Validate origin and destination stations exist
	_, err := s.stationRepo.GetByID(ctx, route.OriginStationID, route.OrganizationID)
	if err != nil {
		return nil, err
	}
	_, err = s.stationRepo.GetByID(ctx, route.DestinationStationID, route.OrganizationID)
	if err != nil {
		return nil, err
	}

	if err := s.routeRepo.Create(ctx, route); err != nil {
		return nil, err
	}
	return route, nil
}

func (s *CatalogService) GetRoute(ctx context.Context, id, orgID string) (*domain.Route, error) {
	return s.routeRepo.GetByID(ctx, id, orgID)
}

func (s *CatalogService) ListRoutes(ctx context.Context, orgID, originID, destID string, pageSize int, pageToken string) ([]*domain.Route, int, string, error) {
	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}

	routes, total, err := s.routeRepo.List(ctx, orgID, originID, destID, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return routes, total, nextToken, nil
}

// --- Trip Operations ---

func (s *CatalogService) CreateTrip(ctx context.Context, trip *domain.Trip, planID string) (*domain.Trip, error) {
	// FAANG: Enforce SaaS plan limits
	// Check Trip Quota (Max Trips Per Month)
	allowed, remaining, err := s.checker.CheckQuota(ctx, trip.OrganizationID, "max_trips_per_month")
	if err != nil {
		return nil, fmt.Errorf("failed to check entitlement: %w", err)
	}
	if !allowed {
		return nil, &PlanLimitError{
			Limit:   "max_trips_per_month",
			Current: int(remaining), // Typically 0 or negative if blocked
			Message: fmt.Sprintf("Trip creation blocked. You have reached your monthly limit. Please upgrade your plan."),
		}
	}

	// Calculate arrival time based on route duration
	route, err := s.routeRepo.GetByID(ctx, trip.RouteID, trip.OrganizationID)
	if err != nil {
		return nil, err
	}
	trip.ArrivalTime = trip.DepartureTime.Add(time.Duration(route.EstimatedDurationMin) * time.Minute)

	// Check Schedule Horizon (Max Schedule Days)
	// We need the full entitlement object to get the limit value (not just usage)
	entitlements, err := s.checker.CheckEntitlement(ctx, trip.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to check entitlement: %w", err)
	}

	maxDays := 30 // Default safe fallback
	if entitlements != nil {
		limit := entitlements.GetQuotaLimit("max_schedule_days")
		if limit > 0 {
			maxDays = int(limit)
		}
	}

	maxScheduleDate := time.Now().AddDate(0, 0, maxDays)
	if trip.DepartureTime.After(maxScheduleDate) {
		return nil, &PlanLimitError{
			Limit:   "max_schedule_days",
			Current: maxDays,
			Message: fmt.Sprintf("Trip departure exceeds your plan's scheduling horizon (%d days). Please upgrade your plan.", maxDays),
		}
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}
	return trip, nil
}

func (s *CatalogService) GetTrip(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	return s.tripRepo.GetByID(ctx, id, orgID)
}

func (s *CatalogService) SearchTrips(ctx context.Context, originCity, destCity string, travelDate time.Time, pageSize int, pageToken string) ([]*TripSearchResult, int, string, error) {
	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}

	trips, total, err := s.tripRepo.Search(ctx, "", originCity, destCity, travelDate, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	// Enrich results with route and station info
	var results []*TripSearchResult
	for _, trip := range trips {
		route, _ := s.routeRepo.GetByID(ctx, trip.RouteID, trip.OrganizationID)
		origin, _ := s.stationRepo.GetByID(ctx, route.OriginStationID, trip.OrganizationID)
		dest, _ := s.stationRepo.GetByID(ctx, route.DestinationStationID, trip.OrganizationID)

		results = append(results, &TripSearchResult{
			Trip:               trip,
			Route:              route,
			OriginStation:      origin,
			DestinationStation: dest,
		})
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return results, total, nextToken, nil
}

func (s *CatalogService) CancelTrip(ctx context.Context, id, orgID, reason string) (*domain.Trip, error) {
	if err := s.tripRepo.UpdateStatus(ctx, id, orgID, domain.TripStatusCancelled); err != nil {
		return nil, err
	}
	return s.tripRepo.GetByID(ctx, id, orgID)
}

// --- DTOs ---

type TripSearchResult struct {
	Trip               *domain.Trip
	Route              *domain.Route
	OriginStation      *domain.Station
	DestinationStation *domain.Station
	OperatorName       string
}

// --- Errors ---

type PlanLimitError struct {
	Limit   string
	Current int
	Message string
}

func (e *PlanLimitError) Error() string {
	return e.Message
}

// --- Helpers ---

func parsePageToken(token string) int {
	if token == "" {
		return 0
	}
	var offset int
	fmt.Sscanf(token, "%d", &offset)
	return offset
}

func generatePageToken(offset int) string {
	return fmt.Sprintf("%d", offset)
}
