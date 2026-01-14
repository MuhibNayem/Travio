package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	fleetpb "github.com/MuhibNayem/Travio/server/api/proto/fleet/v1"
	inventorypb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"github.com/MuhibNayem/Travio/server/pkg/entitlement"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/clients"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/catalog/internal/repository"
	"github.com/google/uuid"
)

// CatalogService handles business logic for catalog operations
type CatalogService struct {
	stationRepo     repository.StationRepository
	routeRepo       repository.RouteRepository
	tripRepo        repository.TripRepository
	scheduleRepo    repository.ScheduleRepository
	checker         entitlement.EntitlementChecker
	fleetClient     *clients.FleetClient
	inventoryClient *clients.InventoryClient
	auditRepo       *repository.PostgresAuditRepository
}

func NewCatalogService(
	stationRepo repository.StationRepository,
	routeRepo repository.RouteRepository,
	tripRepo repository.TripRepository,
	scheduleRepo repository.ScheduleRepository,
	checker entitlement.EntitlementChecker,
	fleetClient *clients.FleetClient,
	inventoryClient *clients.InventoryClient,
	auditRepo *repository.PostgresAuditRepository,
) *CatalogService {
	return &CatalogService{
		stationRepo:     stationRepo,
		routeRepo:       routeRepo,
		tripRepo:        tripRepo,
		scheduleRepo:    scheduleRepo,
		checker:         checker,
		fleetClient:     fleetClient,
		inventoryClient: inventoryClient,
		auditRepo:       auditRepo,
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
	if orgID == "" {
		return nil, 0, "", fmt.Errorf("organization_id is required")
	}

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
	trip.ServiceDate = deriveServiceDate(trip.DepartureTime, route)

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

	// 4. Operational Constraints: Vehicle Availability & Status
	// Check Database Overlaps
	isBusy, err := s.tripRepo.CheckVehicleAvailability(ctx, trip.VehicleID, trip.DepartureTime, trip.ArrivalTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check vehicle availability: %w", err)
	}
	if isBusy {
		return nil, fmt.Errorf("vehicle is already booked for this time slot")
	}

	// Fetch Asset to check Status (and reuse for Inventory)
	var asset *fleetpb.Asset
	if s.fleetClient != nil {
		asset, err = s.fleetClient.GetAsset(ctx, trip.VehicleID, trip.OrganizationID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch asset: %w", err)
		}
		if asset.Status == fleetpb.AssetStatus_ASSET_STATUS_MAINTENANCE {
			return nil, fmt.Errorf("vehicle is currently in MAINTENANCE status")
		}
	}

	if err := s.tripRepo.Create(ctx, trip); err != nil {
		return nil, err
	}

	segments := buildTripSegments(trip, route)
	if err := s.tripRepo.CreateSegments(ctx, trip.ID, segments); err != nil {
		return nil, err
	}

	// Initialize Inventory with Vehicle Layout
	if s.inventoryClient != nil && asset != nil {
		// asset is already fetched above
		seatConfig := mapAssetConfigToSeatConfig(asset, trip.Pricing)

		var pbSegments []*inventorypb.SegmentDefinition
		for _, seg := range segments {
			pbSegments = append(pbSegments, &inventorypb.SegmentDefinition{
				SegmentIndex:  int32(seg.SegmentIndex),
				FromStationId: seg.FromStationID,
				ToStationId:   seg.ToStationID,
				DepartureTime: seg.DepartureTime.Unix(),
				ArrivalTime:   seg.ArrivalTime.Unix(),
			})
		}

		_, err = s.inventoryClient.InitializeTripInventory(ctx, &inventorypb.InitializeTripInventoryRequest{
			TripId:         trip.ID,
			OrganizationId: trip.OrganizationID,
			VehicleId:      trip.VehicleID,
			Segments:       pbSegments,
			SeatConfig:     seatConfig,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize inventory: %w", err)
		}
	}

	return trip, nil
}

func (s *CatalogService) GetTrip(ctx context.Context, id, orgID string) (*domain.Trip, error) {
	return s.tripRepo.GetByID(ctx, id, orgID)
}

func (s *CatalogService) ListTrips(ctx context.Context, orgID, routeID, scheduleID, serviceDate string, pageSize int, pageToken string) ([]*domain.Trip, int, string, error) {
	if orgID == "" {
		return nil, 0, "", fmt.Errorf("organization_id is required")
	}

	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	trips, total, err := s.tripRepo.List(ctx, orgID, routeID, scheduleID, serviceDate, serviceDate, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return trips, total, nextToken, nil
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

// --- Schedule Operations ---

func (s *CatalogService) CreateSchedule(ctx context.Context, schedule *domain.ScheduleTemplate) (*domain.ScheduleTemplate, error) {
	if schedule.OrganizationID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}
	// Validate route exists
	route, err := s.routeRepo.GetByID(ctx, schedule.RouteID, schedule.OrganizationID)
	if err != nil {
		return nil, err
	}
	if err := s.validateSchedule(ctx, schedule, route, ""); err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *CatalogService) CreateSchedules(ctx context.Context, orgID string, schedules []domain.ScheduleTemplate) ([]*domain.ScheduleTemplate, int, error) {
	if orgID == "" {
		return nil, 0, fmt.Errorf("organization_id is required")
	}

	var created []*domain.ScheduleTemplate
	for i := range schedules {
		schedules[i].OrganizationID = orgID
		route, err := s.routeRepo.GetByID(ctx, schedules[i].RouteID, orgID)
		if err != nil {
			return nil, 0, err
		}
		if err := s.validateSchedule(ctx, &schedules[i], route, ""); err != nil {
			return nil, 0, err
		}
		if err := s.scheduleRepo.Create(ctx, &schedules[i]); err != nil {
			return nil, 0, err
		}
		created = append(created, &schedules[i])
	}

	return created, len(created), nil
}

func (s *CatalogService) GetSchedule(ctx context.Context, id, orgID string) (*domain.ScheduleTemplate, error) {
	return s.scheduleRepo.GetByID(ctx, id, orgID)
}

func (s *CatalogService) ListSchedules(ctx context.Context, orgID, routeID, status string, pageSize int, pageToken string) ([]*domain.ScheduleTemplate, int, string, error) {
	if orgID == "" {
		return nil, 0, "", fmt.Errorf("organization_id is required")
	}

	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	schedules, total, err := s.scheduleRepo.List(ctx, orgID, routeID, status, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	nextToken := ""
	if offset+pageSize < total {
		nextToken = generatePageToken(offset + pageSize)
	}

	return schedules, total, nextToken, nil
}

func (s *CatalogService) UpdateSchedule(ctx context.Context, schedule *domain.ScheduleTemplate) (*domain.ScheduleTemplate, error) {
	existing, err := s.scheduleRepo.GetByID(ctx, schedule.ID, schedule.OrganizationID)
	if err != nil {
		return nil, err
	}
	route, err := s.routeRepo.GetByID(ctx, existing.RouteID, schedule.OrganizationID)
	if err != nil {
		return nil, err
	}

	merged := *existing
	merged.TotalSeats = schedule.TotalSeats
	merged.Pricing = schedule.Pricing
	merged.DepartureMinutes = schedule.DepartureMinutes
	merged.ArrivalOffsetMinutes = schedule.ArrivalOffsetMinutes
	merged.Timezone = schedule.Timezone
	merged.StartDate = schedule.StartDate
	merged.EndDate = schedule.EndDate
	merged.DaysOfWeek = schedule.DaysOfWeek
	merged.Status = schedule.Status

	if err := s.validateSchedule(ctx, &merged, route, schedule.ID); err != nil {
		return nil, err
	}

	if err := s.scheduleRepo.Update(ctx, &merged); err != nil {
		return nil, err
	}

	// Audit Log
	if s.auditRepo != nil {
		// Attempt to get actor from metadata (assuming generic key for now)
		actorID := "unknown" // TODO: Extract from context properly
		// Simple change log
		changes := map[string]interface{}{
			"version_from": existing.Version,
			"version_to":   merged.Version,
		}
		s.auditRepo.Log(ctx, repository.AuditLog{
			EntityType: "schedule",
			EntityID:   schedule.ID,
			Action:     "update",
			ActorID:    actorID,
			Changes:    changes,
		})
	}

	return s.scheduleRepo.GetByID(ctx, schedule.ID, schedule.OrganizationID)
}

func (s *CatalogService) DeleteSchedule(ctx context.Context, id, orgID string) error {
	return s.scheduleRepo.Delete(ctx, id, orgID)
}

func (s *CatalogService) AddScheduleException(ctx context.Context, exception *domain.ScheduleException, orgID string) (*domain.ScheduleException, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}
	if _, err := s.scheduleRepo.GetByID(ctx, exception.ScheduleID, orgID); err != nil {
		return nil, err
	}
	if err := s.scheduleRepo.AddException(ctx, exception); err != nil {
		return nil, err
	}
	return exception, nil
}

func (s *CatalogService) ListScheduleExceptions(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleException, error) {
	return s.scheduleRepo.ListExceptions(ctx, scheduleID, orgID)
}

func (s *CatalogService) GetScheduleHistory(ctx context.Context, scheduleID, orgID string) ([]*domain.ScheduleVersion, error) {
	if orgID == "" {
		return nil, fmt.Errorf("organization_id is required")
	}
	// Check schedule existence first
	if _, err := s.scheduleRepo.GetByID(ctx, scheduleID, orgID); err != nil {
		return nil, err
	}
	return s.scheduleRepo.GetHistory(ctx, scheduleID, orgID)
}

func (s *CatalogService) GenerateTripInstances(ctx context.Context, scheduleID, orgID, startDate, endDate string) ([]*domain.Trip, int, error) {
	if orgID == "" {
		return nil, 0, fmt.Errorf("organization_id is required")
	}
	schedule, err := s.scheduleRepo.GetByID(ctx, scheduleID, orgID)
	if err != nil {
		return nil, 0, err
	}
	if schedule.TotalSeats <= 0 {
		return nil, 0, fmt.Errorf("schedule total_seats must be greater than zero")
	}

	route, err := s.routeRepo.GetByID(ctx, schedule.RouteID, schedule.OrganizationID)
	if err != nil {
		return nil, 0, err
	}
	if err := s.validateSchedule(ctx, schedule, route, schedule.ID); err != nil {
		return nil, 0, err
	}

	exceptions, err := s.scheduleRepo.ListExceptions(ctx, scheduleID, orgID)
	if err != nil {
		return nil, 0, err
	}
	exceptionMap := make(map[string]bool)
	for _, ex := range exceptions {
		exceptionMap[ex.ServiceDate] = ex.IsAdded
	}

	dates, err := expandScheduleDates(schedule, startDate, endDate, exceptionMap)
	if err != nil {
		return nil, 0, err
	}

	// Fetch Asset once for efficiency
	var asset *fleetpb.Asset
	if s.fleetClient != nil && schedule.VehicleID != "" {
		asset, err = s.fleetClient.GetAsset(ctx, schedule.VehicleID, schedule.OrganizationID)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to fetch asset: %w", err)
		}
	}

	var tripsToCreate []*domain.Trip
	for _, serviceDate := range dates {
		departureTime, err := buildDepartureTime(serviceDate, schedule)
		if err != nil {
			return nil, 0, err
		}

		trip := &domain.Trip{
			ID:             uuid.New().String(),
			OrganizationID: schedule.OrganizationID,
			ScheduleID:     schedule.ID,
			ServiceDate:    serviceDate,
			RouteID:        schedule.RouteID,
			VehicleID:      schedule.VehicleID,
			VehicleType:    schedule.VehicleType,
			VehicleClass:   schedule.VehicleClass,
			DepartureTime:  departureTime,
			TotalSeats:     schedule.TotalSeats,
			AvailableSeats: schedule.TotalSeats,
			Pricing:        schedule.Pricing,
			Status:         domain.TripStatusScheduled,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if route.EstimatedDurationMin > 0 {
			trip.ArrivalTime = departureTime.Add(time.Duration(route.EstimatedDurationMin) * time.Minute)
		} else if schedule.ArrivalOffsetMinutes > 0 {
			trip.ArrivalTime = departureTime.Add(time.Duration(schedule.ArrivalOffsetMinutes) * time.Minute)
		}

		// Generate segments
		trip.Segments = buildTripSegments(trip, route)
		tripsToCreate = append(tripsToCreate, trip)
	}

	// Batch DB Insert
	if err := s.tripRepo.BatchCreate(ctx, tripsToCreate); err != nil {
		return nil, 0, err
	}

	// Initialize Inventory for each trip
	if s.inventoryClient != nil && asset != nil {
		seatConfig := mapAssetConfigToSeatConfig(asset, schedule.Pricing)

		for _, trip := range tripsToCreate {
			var pbSegments []*inventorypb.SegmentDefinition
			for _, seg := range trip.Segments {
				pbSegments = append(pbSegments, &inventorypb.SegmentDefinition{
					SegmentIndex:  int32(seg.SegmentIndex),
					FromStationId: seg.FromStationID,
					ToStationId:   seg.ToStationID,
					DepartureTime: seg.DepartureTime.Unix(),
					ArrivalTime:   seg.ArrivalTime.Unix(),
				})
			}

			// We launch these in parallel or sequence? Sequence for safety for now.
			// Ideally worker pool.
			_, err := s.inventoryClient.InitializeTripInventory(ctx, &inventorypb.InitializeTripInventoryRequest{
				TripId:         trip.ID,
				OrganizationId: trip.OrganizationID,
				VehicleId:      trip.VehicleID,
				Segments:       pbSegments,
				SeatConfig:     seatConfig,
			})
			if err != nil {
				// Log error but don't fail entire batch?
				// Or return error (partial failure).
				// For now return error, but transactions are already committed.
				// This is a distributed transaction issue.
				// SAGA or Log.
				// Proceeding, but logging would be better.
				// Since I don't have logger in struct (or do I? auditRepo is there), I'll just return error.
				return nil, len(tripsToCreate), fmt.Errorf("failed to init inventory for trip %s: %w", trip.ID, err)
			}
		}
	}

	return tripsToCreate, len(tripsToCreate), nil
}

func (s *CatalogService) ListTripInstances(ctx context.Context, orgID, scheduleID, routeID, startDate, endDate string, status string, pageSize int, pageToken string) ([]*TripSearchResult, int, string, error) {
	if orgID == "" {
		return nil, 0, "", fmt.Errorf("organization_id is required")
	}

	offset := parsePageToken(pageToken)
	if pageSize <= 0 {
		pageSize = 20
	}

	trips, total, err := s.tripRepo.List(ctx, orgID, routeID, scheduleID, startDate, endDate, pageSize, offset)
	if err != nil {
		return nil, 0, "", err
	}

	var results []*TripSearchResult
	for _, trip := range trips {
		if status != "" && trip.Status != status {
			continue
		}
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

func (s *CatalogService) validateSchedule(ctx context.Context, schedule *domain.ScheduleTemplate, route *domain.Route, excludeID string) error {
	if schedule.OrganizationID == "" {
		return fmt.Errorf("organization_id is required")
	}
	if schedule.RouteID == "" {
		return fmt.Errorf("route_id is required")
	}
	if schedule.TotalSeats <= 0 {
		return fmt.Errorf("total_seats must be greater than zero")
	}
	if schedule.DepartureMinutes < 0 || schedule.DepartureMinutes >= 1440 {
		return fmt.Errorf("departure_minutes must be between 0 and 1439")
	}
	if schedule.DaysOfWeek <= 0 {
		return fmt.Errorf("days_of_week must include at least one day")
	}
	start, err := time.Parse("2006-01-02", schedule.StartDate)
	if err != nil {
		return fmt.Errorf("invalid start_date")
	}
	end, err := time.Parse("2006-01-02", schedule.EndDate)
	if err != nil {
		return fmt.Errorf("invalid end_date")
	}
	if end.Before(start) {
		return fmt.Errorf("end_date must be after start_date")
	}

	if schedule.Timezone == "" {
		schedule.Timezone = "Asia/Dhaka"
	}
	if _, err := time.LoadLocation(schedule.Timezone); err != nil {
		return fmt.Errorf("invalid timezone")
	}

	if schedule.ArrivalOffsetMinutes <= 0 && route != nil && route.EstimatedDurationMin > 0 {
		schedule.ArrivalOffsetMinutes = route.EstimatedDurationMin
	}
	if schedule.ArrivalOffsetMinutes <= 0 {
		return fmt.Errorf("arrival_offset_minutes must be greater than zero")
	}
	if schedule.DepartureMinutes+schedule.ArrivalOffsetMinutes > 1440 {
		return fmt.Errorf("arrival_offset_minutes exceeds same-day limit; overnight schedules require explicit handling")
	}

	if schedule.Pricing.Currency == "" {
		schedule.Pricing.Currency = "BDT"
	}
	schedule.Pricing.Currency = strings.ToUpper(schedule.Pricing.Currency)
	if len(schedule.Pricing.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter ISO code")
	}
	if schedule.Pricing.BasePricePaisa <= 0 {
		return fmt.Errorf("base_price_paisa must be greater than zero")
	}
	for className, price := range schedule.Pricing.ClassPrices {
		if strings.TrimSpace(className) == "" || price <= 0 {
			return fmt.Errorf("class_prices must have positive values")
		}
	}
	for category, price := range schedule.Pricing.SeatCategoryPrices {
		if strings.TrimSpace(category) == "" || price <= 0 {
			return fmt.Errorf("seat_category_prices must have positive values")
		}
	}

	if route != nil && len(schedule.Pricing.SegmentPrices) > 0 {
		segmentPairs := make(map[string]struct{})
		seenSegments := make(map[string]struct{})
		stops := []string{route.OriginStationID}
		stops = append(stops, sortedIntermediateStops(route.IntermediateStops)...)
		stops = append(stops, route.DestinationStationID)
		for i := 0; i < len(stops)-1; i++ {
			key := stops[i] + "->" + stops[i+1]
			segmentPairs[key] = struct{}{}
		}
		for _, seg := range schedule.Pricing.SegmentPrices {
			key := seg.FromStationID + "->" + seg.ToStationID
			if _, ok := segmentPairs[key]; !ok {
				return fmt.Errorf("segment pricing does not match route stops")
			}
			if _, ok := seenSegments[key]; ok {
				return fmt.Errorf("duplicate segment pricing entries detected")
			}
			seenSegments[key] = struct{}{}
			if seg.BasePricePaisa <= 0 {
				return fmt.Errorf("segment base price must be greater than zero")
			}
			for className, price := range seg.ClassPrices {
				if strings.TrimSpace(className) == "" || price <= 0 {
					return fmt.Errorf("segment class_prices must have positive values")
				}
			}
			for category, price := range seg.SeatCategoryPrices {
				if strings.TrimSpace(category) == "" || price <= 0 {
					return fmt.Errorf("segment seat_category_prices must have positive values")
				}
			}
		}
	}

	if schedule.Status == "" {
		schedule.Status = domain.ScheduleStatusActive
	}
	if schedule.VehicleID != "" && schedule.Status == domain.ScheduleStatusActive {
		conflict, err := s.scheduleRepo.HasVehicleConflict(ctx, schedule, excludeID)
		if err != nil {
			return err
		}
		if conflict {
			return fmt.Errorf("schedule conflicts with existing vehicle assignment")
		}
	}

	return nil
}

func mapAssetConfigToSeatConfig(asset *fleetpb.Asset, pricing domain.TripPricing) *inventorypb.SeatConfiguration {
	var seats []*inventorypb.SeatDefinition
	totalSeats := 0

	// Helper to calculate price
	getPrice := func(seatClass, seatCategory string) int64 {
		price := pricing.BasePricePaisa
		if p, ok := pricing.ClassPrices[seatClass]; ok {
			price = p
		}
		if p, ok := pricing.SeatCategoryPrices[seatCategory]; ok {
			price = p
		}
		return price
	}

	if asset.Config.GetBus() != nil {
		bus := asset.Config.GetBus()
		totalSeats = int(bus.Rows * bus.SeatsPerRow) // Approximation
		// Generate simple seat map for bus
		// A1, A2, aisle, A3, A4...
		// This is a naive generation. Ideally Fleet should return exact seat list.
		// Detailed layout parsing should happen in Fleet service or shared lib.
		// For now, we assume simple grid.
		for r := 1; r <= int(bus.Rows); r++ {
			char := string(rune('A' + r - 1))
			for c := 1; c <= int(bus.SeatsPerRow); c++ {
				seatNum := fmt.Sprintf("%s%d", char, c)
				seatType := "window"
				if c > 1 && c < int(bus.SeatsPerRow) {
					seatType = "aisle"
				}

				seatClass := "economy" // default
				// Check categories
				for _, cat := range bus.Categories {
					// Logic to map specific seats to categories would be complex here
					// For simplicity using default class/category
					seatClass = strings.ToLower(cat.Name)
				}

				seats = append(seats, &inventorypb.SeatDefinition{
					SeatId:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
					SeatNumber: seatNum,
					Row:        int32(r),
					Column:     int32(c),
					SeatType:   seatType,
					SeatClass:  seatClass,
					PricePaisa: getPrice(seatClass, ""),
				})
			}
		}
	} else if asset.Config.GetTrain() != nil {
		// Train logic
		train := asset.Config.GetTrain()
		for _, coach := range train.Coaches {
			for r := 1; r <= int(coach.Rows); r++ {
				for s := 1; s <= int(coach.SeatsPerRow); s++ {
					seatNum := fmt.Sprintf("%s-%d-%d", coach.Id, r, s)
					seatClass := strings.ToLower(coach.Name) // e.g. "shovan"
					seats = append(seats, &inventorypb.SeatDefinition{
						SeatId:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
						SeatNumber: seatNum,
						Row:        int32(r),
						Column:     int32(s),
						SeatClass:  seatClass,
						PricePaisa: getPrice(seatClass, ""),
					})
					totalSeats++
				}
			}
		}
	} else if asset.Config.GetLaunch() != nil {
		// Launch logic
		launch := asset.Config.GetLaunch()
		for _, deck := range launch.Decks {
			// Deck seating
			for r := 1; r <= int(deck.Rows); r++ {
				for c := 1; c <= int(deck.Cols); c++ {
					seatNum := fmt.Sprintf("%s-%d-%d", deck.Id, r, c)
					seatClass := strings.ToLower(deck.Name)
					seats = append(seats, &inventorypb.SeatDefinition{
						SeatId:     fmt.Sprintf("%s-%s", asset.Id, seatNum),
						SeatNumber: seatNum,
						Row:        int32(r),
						Column:     int32(c),
						SeatClass:  seatClass,
						PricePaisa: int64(deck.SeatPricePaisa), // Launch has explicit price in config
					})
					totalSeats++
				}
			}
			// Cabins
			for _, cabin := range deck.Cabins {
				seatNum := cabin.Name
				seatClass := "cabin"
				if cabin.IsSuite {
					seatClass = "suite"
				}
				seats = append(seats, &inventorypb.SeatDefinition{
					SeatId:     fmt.Sprintf("%s-%s", asset.Id, cabin.Id),
					SeatNumber: seatNum,
					SeatClass:  seatClass,
					PricePaisa: int64(cabin.PricePaisa),
				})
				totalSeats++
			}
		}
	}

	return &inventorypb.SeatConfiguration{
		TotalSeats: int32(totalSeats),
		Seats:      seats,
	}
}

func sortedIntermediateStops(stops []domain.RouteStop) []string {
	if len(stops) == 0 {
		return nil
	}
	stopsCopy := make([]domain.RouteStop, len(stops))
	copy(stopsCopy, stops)
	sort.Slice(stopsCopy, func(i, j int) bool {
		return stopsCopy[i].Sequence < stopsCopy[j].Sequence
	})
	result := make([]string, 0, len(stopsCopy))
	for _, stop := range stopsCopy {
		result = append(result, stop.StationID)
	}
	return result
}

func deriveServiceDate(departure time.Time, route *domain.Route) string {
	return departure.In(departure.Location()).Format("2006-01-02")
}

type routeStopWithOffsets struct {
	StationID              string
	ArrivalOffsetMinutes   int
	DepartureOffsetMinutes int
}

func buildTripSegments(trip *domain.Trip, route *domain.Route) []domain.TripSegment {
	if trip == nil || route == nil {
		return nil
	}

	stopsCopy := make([]domain.RouteStop, len(route.IntermediateStops))
	copy(stopsCopy, route.IntermediateStops)
	sort.Slice(stopsCopy, func(i, j int) bool {
		return stopsCopy[i].Sequence < stopsCopy[j].Sequence
	})

	stops := make([]routeStopWithOffsets, 0, len(route.IntermediateStops)+2)
	stops = append(stops, routeStopWithOffsets{
		StationID:              route.OriginStationID,
		ArrivalOffsetMinutes:   0,
		DepartureOffsetMinutes: 0,
	})

	for _, stop := range stopsCopy {
		stops = append(stops, routeStopWithOffsets{
			StationID:              stop.StationID,
			ArrivalOffsetMinutes:   stop.ArrivalOffsetMinutes,
			DepartureOffsetMinutes: stop.DepartureOffsetMinutes,
		})
	}

	estimatedArrival := route.EstimatedDurationMin
	stops = append(stops, routeStopWithOffsets{
		StationID:              route.DestinationStationID,
		ArrivalOffsetMinutes:   estimatedArrival,
		DepartureOffsetMinutes: estimatedArrival,
	})

	segments := make([]domain.TripSegment, 0, len(stops)-1)
	for i := 0; i < len(stops)-1; i++ {
		departure := trip.DepartureTime.Add(time.Duration(stops[i].DepartureOffsetMinutes) * time.Minute)
		arrival := trip.DepartureTime.Add(time.Duration(stops[i+1].ArrivalOffsetMinutes) * time.Minute)
		segments = append(segments, domain.TripSegment{
			SegmentIndex:   i,
			FromStationID:  stops[i].StationID,
			ToStationID:    stops[i+1].StationID,
			DepartureTime:  departure,
			ArrivalTime:    arrival,
			AvailableSeats: trip.TotalSeats,
		})
	}

	return segments
}

func expandScheduleDates(schedule *domain.ScheduleTemplate, startDate, endDate string, exceptions map[string]bool) ([]string, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}
	scheduleStart, err := time.Parse("2006-01-02", schedule.StartDate)
	if err != nil {
		return nil, err
	}
	scheduleEnd, err := time.Parse("2006-01-02", schedule.EndDate)
	if err != nil {
		return nil, err
	}

	if start.Before(scheduleStart) {
		start = scheduleStart
	}
	if end.After(scheduleEnd) {
		end = scheduleEnd
	}

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if override, ok := exceptions[dateStr]; ok {
			if override {
				dates = append(dates, dateStr)
			}
			continue
		}

		if isScheduledDay(schedule.DaysOfWeek, d.Weekday()) {
			dates = append(dates, dateStr)
		}
	}
	return dates, nil
}

func buildDepartureTime(serviceDate string, schedule *domain.ScheduleTemplate) (time.Time, error) {
	loc, err := time.LoadLocation(schedule.Timezone)
	if err != nil {
		loc = time.UTC
	}
	date, err := time.ParseInLocation("2006-01-02", serviceDate, loc)
	if err != nil {
		return time.Time{}, err
	}
	hours := schedule.DepartureMinutes / 60
	mins := schedule.DepartureMinutes % 60
	return time.Date(date.Year(), date.Month(), date.Day(), hours, mins, 0, 0, loc), nil
}

func isScheduledDay(mask int, weekday time.Weekday) bool {
	var bit int
	switch weekday {
	case time.Monday:
		bit = 1
	case time.Tuesday:
		bit = 2
	case time.Wednesday:
		bit = 4
	case time.Thursday:
		bit = 8
	case time.Friday:
		bit = 16
	case time.Saturday:
		bit = 32
	case time.Sunday:
		bit = 64
	default:
		return false
	}
	return mask&bit != 0
}
