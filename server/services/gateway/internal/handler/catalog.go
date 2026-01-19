package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	catalogpb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// CatalogHandler handles catalog-related REST endpoints
type CatalogHandler struct {
	catalogConn *grpc.ClientConn
	client      catalogpb.CatalogServiceClient
	cb          *middleware.CircuitBreaker
}

// NewCatalogHandler creates a catalog handler with gRPC connection
func NewCatalogHandler(catalogURL string, cb *middleware.CircuitBreaker) (*CatalogHandler, error) {
	conn, err := grpc.NewClient(catalogURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CatalogHandler{
		catalogConn: conn,
		client:      catalogpb.NewCatalogServiceClient(conn),
		cb:          cb,
	}, nil
}

// ListStations returns all stations
func (h *CatalogHandler) ListStations(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListStations(ctx, &catalogpb.ListStationsRequest{
			PageSize: 100,
		})
	})
	if err != nil {
		logger.Error("Failed to list stations", "error", err)
		http.Error(w, "Failed to fetch stations", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListStationsResponse)

	// Convert to JSON-friendly format
	stations := make([]map[string]interface{}, 0, len(resp.Stations))
	for _, s := range resp.Stations {
		stations = append(stations, stationToJSON(s))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stations":        stations,
		"total":           resp.TotalCount,
		"next_page_token": resp.NextPageToken,
	})
}

// SearchTrips searches for trips between stations
func (h *CatalogHandler) SearchTrips(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Parse query params
	originCity := r.URL.Query().Get("from")
	destCity := r.URL.Query().Get("to")
	dateStr := r.URL.Query().Get("date")
	vehicleType := r.URL.Query().Get("type")

	// Parse date
	var travelDate int64
	if dateStr != "" {
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			travelDate = t.Unix()
		}
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.SearchTrips(ctx, &catalogpb.SearchTripsRequest{
			OriginCity:      originCity,
			DestinationCity: destCity,
			TravelDate:      travelDate,
			VehicleType:     vehicleType,
			PageSize:        50,
		})
	})
	if err != nil {
		http.Error(w, "Failed to search trips", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.SearchTripsResponse)

	// Convert to JSON-friendly format
	results := make([]map[string]interface{}, 0, len(resp.Results))
	for _, r := range resp.Results {
		trip := r.Trip
		route := r.Route
		origin := r.OriginStation
		dest := r.DestinationStation

		// Get price from Pricing field
		var price int64
		if trip.Pricing != nil {
			price = trip.Pricing.BasePricePaisa / 100 // Convert to currency
		}

		results = append(results, map[string]interface{}{
			"id":              trip.Id,
			"route_id":        trip.RouteId,
			"type":            trip.VehicleType,
			"operator":        r.OperatorName,
			"vehicle_name":    trip.VehicleId,
			"departure_time":  time.Unix(trip.DepartureTime, 0).Format(time.RFC3339),
			"arrival_time":    time.Unix(trip.ArrivalTime, 0).Format(time.RFC3339),
			"price":           price,
			"class":           trip.VehicleClass,
			"available_seats": trip.TotalSeats,
			"total_seats":     trip.TotalSeats,
			"from":            origin.Name,
			"from_city":       origin.City,
			"to":              dest.Name,
			"to_city":         dest.City,
			"duration":        route.EstimatedDurationMinutes,
			"distance":        route.DistanceKm,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results":   results,
		"total":     resp.TotalCount,
		"next_page": resp.NextPageToken,
	})
}

type CreateRouteStopRequest struct {
	StationID              string `json:"station_id"`
	Sequence               int32  `json:"sequence"`
	ArrivalOffsetMinutes   int32  `json:"arrival_offset_minutes"`
	DepartureOffsetMinutes int32  `json:"departure_offset_minutes"`
	DistanceFromOriginKm   int32  `json:"distance_from_origin_km"`
}

type CreateRouteRequest struct {
	Code                     string                   `json:"code"`
	Name                     string                   `json:"name"`
	OriginStationID          string                   `json:"origin_station_id"`
	DestinationStationID     string                   `json:"destination_station_id"`
	DistanceKm               int32                    `json:"distance_km"`
	EstimatedDurationMinutes int32                    `json:"estimated_duration_minutes"`
	IntermediateStops        []CreateRouteStopRequest `json:"intermediate_stops"`
}

// CreateRoute creates a new route via gRPC
func (h *CatalogHandler) CreateRoute(w http.ResponseWriter, r *http.Request) {
	var req CreateRouteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Code == "" || req.Name == "" || req.OriginStationID == "" || req.DestinationStationID == "" {
		http.Error(w, `{"error": "missing required fields"}`, http.StatusBadRequest)
		return
	}

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		orgID = r.Header.Get("X-Organization-ID")
	}

	if orgID == "" {
		http.Error(w, `{"error": "missing organization context"}`, http.StatusUnauthorized)
		return
	}

	stops := make([]*catalogpb.RouteStop, 0, len(req.IntermediateStops))
	for _, stop := range req.IntermediateStops {
		stops = append(stops, &catalogpb.RouteStop{
			StationId:              stop.StationID,
			Sequence:               stop.Sequence,
			ArrivalOffsetMinutes:   stop.ArrivalOffsetMinutes,
			DepartureOffsetMinutes: stop.DepartureOffsetMinutes,
			DistanceFromOriginKm:   stop.DistanceFromOriginKm,
		})
	}

	grpcReq := &catalogpb.CreateRouteRequest{
		OrganizationId:           orgID,
		Code:                     req.Code,
		Name:                     req.Name,
		OriginStationId:          req.OriginStationID,
		DestinationStationId:     req.DestinationStationID,
		IntermediateStops:        stops,
		DistanceKm:               req.DistanceKm,
		EstimatedDurationMinutes: req.EstimatedDurationMinutes,
	}

	resp, err := h.client.CreateRoute(r.Context(), grpcReq)
	if err != nil {
		logger.Error("Failed to create route", "error", err)
		http.Error(w, `{"error": "failed to create route"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(routeToJSON(resp))
}

// ListRoutes returns all routes
func (h *CatalogHandler) ListRoutes(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		orgID = middleware.GetOrgID(r.Context())
	}
	originID := r.URL.Query().Get("origin_station_id")
	destID := r.URL.Query().Get("destination_station_id")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListRoutes(ctx, &catalogpb.ListRoutesRequest{
			OrganizationId:       orgID,
			OriginStationId:      originID,
			DestinationStationId: destID,
			PageSize:             100,
		})
	})
	if err != nil {
		logger.Error("Failed to list routes", "error", err)
		http.Error(w, "Failed to list routes", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListRoutesResponse)

	routes := make([]map[string]interface{}, 0, len(resp.Routes))
	for _, r := range resp.Routes {
		routes = append(routes, routeToJSON(r))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"routes":          routes,
		"next_page_token": resp.NextPageToken,
		"total_count":     resp.TotalCount,
	})
}

// GetRoute retrieves a route by ID
func (h *CatalogHandler) GetRoute(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	routeID := chi.URLParam(r, "routeId")
	orgID := middleware.GetOrgID(r.Context())

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetRoute(ctx, &catalogpb.GetRouteRequest{
			Id:             routeID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Route not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to get route", "error", err)
		http.Error(w, "Failed to get route", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.Route)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(routeToJSON(resp))
}

// ListTripInstances returns trip instances with route and station details
func (h *CatalogHandler) ListTripInstances(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	req := &catalogpb.ListTripInstancesRequest{
		OrganizationId: orgID,
		ScheduleId:     r.URL.Query().Get("schedule_id"),
		RouteId:        r.URL.Query().Get("route_id"),
		StartDate:      r.URL.Query().Get("start_date"),
		EndDate:        r.URL.Query().Get("end_date"),
		Status:         parseTripStatus(r.URL.Query().Get("status")),
		PageSize:       100,
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListTripInstances(ctx, req)
	})
	if err != nil {
		logger.Error("Failed to list trip instances", "error", err)
		http.Error(w, "Failed to list trip instances", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListTripInstancesResponse)

	results := make([]map[string]interface{}, 0, len(resp.Results))
	for _, r := range resp.Results {
		results = append(results, map[string]interface{}{
			"trip":                tripToJSON(r.Trip),
			"route":               routeToJSON(r.Route),
			"origin_station":      stationToJSON(r.OriginStation),
			"destination_station": stationToJSON(r.DestinationStation),
			"operator_name":       r.OperatorName,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results":         results,
		"next_page_token": resp.NextPageToken,
		"total_count":     resp.TotalCount,
	})
}

// GetTripInstance returns a single trip instance with route and station details
func (h *CatalogHandler) GetTripInstance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	tripID := chi.URLParam(r, "tripId")
	tripResult, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetTrip(ctx, &catalogpb.GetTripRequest{
			Id:             tripID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		logger.Error("Failed to get trip instance", "error", err)
		http.Error(w, "Failed to get trip instance", http.StatusInternalServerError)
		return
	}
	trip := tripResult.(*catalogpb.Trip)

	routeResult, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetRoute(ctx, &catalogpb.GetRouteRequest{
			Id:             trip.RouteId,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		logger.Error("Failed to get route", "error", err)
		http.Error(w, "Failed to get route", http.StatusInternalServerError)
		return
	}
	route := routeResult.(*catalogpb.Route)

	originResult, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetStation(ctx, &catalogpb.GetStationRequest{
			Id:             route.OriginStationId,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		logger.Error("Failed to get origin station", "error", err)
		http.Error(w, "Failed to get origin station", http.StatusInternalServerError)
		return
	}
	origin := originResult.(*catalogpb.Station)

	destResult, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetStation(ctx, &catalogpb.GetStationRequest{
			Id:             route.DestinationStationId,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		logger.Error("Failed to get destination station", "error", err)
		http.Error(w, "Failed to get destination station", http.StatusInternalServerError)
		return
	}
	dest := destResult.(*catalogpb.Station)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trip":                tripToJSON(trip),
		"route":               routeToJSON(route),
		"origin_station":      stationToJSON(origin),
		"destination_station": stationToJSON(dest),
	})
}

// Schedule request/response helpers
type ScheduleRequest struct {
	RouteID              string             `json:"route_id"`
	VehicleID            string             `json:"vehicle_id"`
	VehicleType          string             `json:"vehicle_type"`
	VehicleClass         string             `json:"vehicle_class"`
	TotalSeats           int32              `json:"total_seats"`
	Pricing              TripPricingRequest `json:"pricing"`
	DepartureMinutes     int32              `json:"departure_minutes"`
	ArrivalOffsetMinutes int32              `json:"arrival_offset_minutes"`
	Timezone             string             `json:"timezone"`
	StartDate            string             `json:"start_date"`
	EndDate              string             `json:"end_date"`
	DaysOfWeek           int32              `json:"days_of_week"`
	Status               string             `json:"status"`
}

type ScheduleExceptionRequest struct {
	ServiceDate string `json:"service_date"`
	IsAdded     bool   `json:"is_added"`
	Reason      string `json:"reason"`
}

type TripPricingRequest struct {
	BasePricePaisa     int64                   `json:"base_price_paisa"`
	TaxPaisa           int64                   `json:"tax_paisa"`
	BookingFeePaisa    int64                   `json:"booking_fee_paisa"`
	Currency           string                  `json:"currency"`
	ClassPrices        map[string]int64        `json:"class_prices"`
	SeatCategoryPrices map[string]int64        `json:"seat_category_prices"`
	SegmentPrices      []SegmentPricingRequest `json:"segment_prices"`
}

type SegmentPricingRequest struct {
	FromStationID      string           `json:"from_station_id"`
	ToStationID        string           `json:"to_station_id"`
	BasePricePaisa     int64            `json:"base_price_paisa"`
	ClassPrices        map[string]int64 `json:"class_prices"`
	SeatCategoryPrices map[string]int64 `json:"seat_category_prices"`
}

type BulkScheduleRequest struct {
	Schedules []ScheduleRequest `json:"schedules"`
}

func pricingRequestToProto(req TripPricingRequest) *catalogpb.TripPricing {
	var segmentPrices []*catalogpb.SegmentPricing
	for _, seg := range req.SegmentPrices {
		segmentPrices = append(segmentPrices, &catalogpb.SegmentPricing{
			FromStationId:      seg.FromStationID,
			ToStationId:        seg.ToStationID,
			BasePricePaisa:     seg.BasePricePaisa,
			ClassPrices:        seg.ClassPrices,
			SeatCategoryPrices: seg.SeatCategoryPrices,
		})
	}
	return &catalogpb.TripPricing{
		BasePricePaisa:     req.BasePricePaisa,
		TaxPaisa:           req.TaxPaisa,
		BookingFeePaisa:    req.BookingFeePaisa,
		Currency:           req.Currency,
		ClassPrices:        req.ClassPrices,
		SeatCategoryPrices: req.SeatCategoryPrices,
		SegmentPrices:      segmentPrices,
	}
}

// CreateSchedule creates a recurring schedule
func (h *CatalogHandler) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		log.Printf("CreateSchedule: organization_id is missing from context")
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("CreateSchedule: failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("CreateSchedule: orgID=%s, routeID=%s, vehicleID=%s, startDate=%s, endDate=%s",
		orgID, req.RouteID, req.VehicleID, req.StartDate, req.EndDate)

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.CreateSchedule(ctx, &catalogpb.CreateScheduleRequest{
			OrganizationId:       orgID,
			RouteId:              req.RouteID,
			VehicleId:            req.VehicleID,
			VehicleType:          req.VehicleType,
			VehicleClass:         req.VehicleClass,
			TotalSeats:           req.TotalSeats,
			Pricing:              pricingRequestToProto(req.Pricing),
			DepartureMinutes:     req.DepartureMinutes,
			ArrivalOffsetMinutes: req.ArrivalOffsetMinutes,
			Timezone:             req.Timezone,
			StartDate:            req.StartDate,
			EndDate:              req.EndDate,
			DaysOfWeek:           req.DaysOfWeek,
		})
	})
	if err != nil {
		log.Printf("CreateSchedule ERROR: %v", err)
		http.Error(w, fmt.Sprintf(`{"error": "Failed to create schedule: %v"}`, err), http.StatusInternalServerError)
		return
	}

	resp := result.(*catalogpb.Schedule)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scheduleToJSON(resp))
}

// BulkCreateSchedules creates multiple schedules in one request
func (h *CatalogHandler) BulkCreateSchedules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	var req BulkScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	schedules := make([]*catalogpb.ScheduleDefinition, 0, len(req.Schedules))
	for _, s := range req.Schedules {
		schedules = append(schedules, &catalogpb.ScheduleDefinition{
			RouteId:              s.RouteID,
			VehicleId:            s.VehicleID,
			VehicleType:          s.VehicleType,
			VehicleClass:         s.VehicleClass,
			TotalSeats:           s.TotalSeats,
			Pricing:              pricingRequestToProto(s.Pricing),
			DepartureMinutes:     s.DepartureMinutes,
			ArrivalOffsetMinutes: s.ArrivalOffsetMinutes,
			Timezone:             s.Timezone,
			StartDate:            s.StartDate,
			EndDate:              s.EndDate,
			DaysOfWeek:           s.DaysOfWeek,
		})
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.CreateSchedules(ctx, &catalogpb.BulkCreateSchedulesRequest{
			OrganizationId: orgID,
			Schedules:      schedules,
		})
	})
	if err != nil {
		http.Error(w, "Failed to create schedules", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.BulkCreateSchedulesResponse)

	created := make([]map[string]interface{}, 0, len(resp.Schedules))
	for _, s := range resp.Schedules {
		created = append(created, scheduleToJSON(s))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules":     created,
		"created_count": resp.CreatedCount,
	})
}

// ListSchedules lists schedules
func (h *CatalogHandler) ListSchedules(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		log.Printf("ListSchedules: organization_id missing")
		http.Error(w, `{"error": "organization_id is required"}`, http.StatusBadRequest)
		return
	}

	log.Printf("ListSchedules: orgID=%s, routeID=%s", orgID, r.URL.Query().Get("route_id"))

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListSchedules(ctx, &catalogpb.ListSchedulesRequest{
			OrganizationId: orgID,
			RouteId:        r.URL.Query().Get("route_id"),
			Status:         parseScheduleStatus(r.URL.Query().Get("status")),
			PageSize:       100,
		})
	})
	if err != nil {
		log.Printf("ListSchedules ERROR: %v", err)
		http.Error(w, fmt.Sprintf(`{"error": "Failed to list schedules: %v"}`, err), http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListSchedulesResponse)

	schedules := make([]map[string]interface{}, 0, len(resp.Schedules))
	for _, s := range resp.Schedules {
		schedules = append(schedules, scheduleToJSON(s))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"schedules":       schedules,
		"next_page_token": resp.NextPageToken,
		"total_count":     resp.TotalCount,
	})
}

// GetSchedule retrieves a schedule by ID
func (h *CatalogHandler) GetSchedule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetSchedule(ctx, &catalogpb.GetScheduleRequest{
			Id:             scheduleID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		http.Error(w, "Failed to get schedule", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.Schedule)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scheduleToJSON(resp))
}

// UpdateSchedule updates a schedule
func (h *CatalogHandler) UpdateSchedule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")

	var req ScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		status := parseScheduleStatus(req.Status)
		if status == catalogpb.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED {
			status = catalogpb.ScheduleStatus_SCHEDULE_STATUS_ACTIVE
		}
		return h.client.UpdateSchedule(ctx, &catalogpb.UpdateScheduleRequest{
			Id:                   scheduleID,
			OrganizationId:       orgID,
			TotalSeats:           req.TotalSeats,
			Pricing:              pricingRequestToProto(req.Pricing),
			DepartureMinutes:     req.DepartureMinutes,
			ArrivalOffsetMinutes: req.ArrivalOffsetMinutes,
			Timezone:             req.Timezone,
			StartDate:            req.StartDate,
			EndDate:              req.EndDate,
			DaysOfWeek:           req.DaysOfWeek,
			Status:               status,
		})
	})
	if err != nil {
		http.Error(w, "Failed to update schedule", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.Schedule)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scheduleToJSON(resp))
}

// DeleteSchedule removes a schedule
func (h *CatalogHandler) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")

	_, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.DeleteSchedule(ctx, &catalogpb.DeleteScheduleRequest{
			Id:             scheduleID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		http.Error(w, "Failed to delete schedule", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddScheduleException adds an exception for a schedule date
func (h *CatalogHandler) AddScheduleException(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")

	var req ScheduleExceptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.AddScheduleException(ctx, &catalogpb.AddScheduleExceptionRequest{
			ScheduleId:     scheduleID,
			OrganizationId: orgID,
			ServiceDate:    req.ServiceDate,
			IsAdded:        req.IsAdded,
			Reason:         req.Reason,
		})
	})
	if err != nil {
		http.Error(w, "Failed to add schedule exception", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ScheduleException)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(scheduleExceptionToJSON(resp))
}

// ListScheduleExceptions lists schedule exceptions
func (h *CatalogHandler) ListScheduleExceptions(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListScheduleExceptions(ctx, &catalogpb.ListScheduleExceptionsRequest{
			ScheduleId:     scheduleID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		http.Error(w, "Failed to list schedule exceptions", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListScheduleExceptionsResponse)

	exceptions := make([]map[string]interface{}, 0, len(resp.Exceptions))
	for _, ex := range resp.Exceptions {
		exceptions = append(exceptions, scheduleExceptionToJSON(ex))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exceptions": exceptions,
	})
}

// GenerateTripInstances generates trip instances for a schedule
func (h *CatalogHandler) GenerateTripInstances(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := middleware.GetOrgID(r.Context())
	scheduleID := chi.URLParam(r, "scheduleId")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	log.Printf("GenerateTripInstances: orgID=%s, scheduleID=%s, startDate=%s, endDate=%s",
		orgID, scheduleID, startDate, endDate)

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GenerateTripInstances(ctx, &catalogpb.GenerateTripInstancesRequest{
			ScheduleId:     scheduleID,
			OrganizationId: orgID,
			StartDate:      startDate,
			EndDate:        endDate,
		})
	})
	if err != nil {
		log.Printf("GenerateTripInstances ERROR: %v", err)
		http.Error(w, fmt.Sprintf(`{"error": "Failed to generate trip instances: %v"}`, err), http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.GenerateTripInstancesResponse)

	trips := make([]map[string]interface{}, 0, len(resp.Trips))
	for _, t := range resp.Trips {
		trips = append(trips, tripToJSON(t))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trips":         trips,
		"created_count": resp.CreatedCount,
	})
}

// GetStation retrieves a station by ID
func (h *CatalogHandler) GetStation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stationID := chi.URLParam(r, "stationId")
	orgID := middleware.GetOrgID(r.Context())

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetStation(ctx, &catalogpb.GetStationRequest{
			Id:             stationID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Station not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to get station", "error", err)
		http.Error(w, "Failed to get station", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.Station)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stationToJSON(resp))
}

// Close closes the gRPC connection
func (h *CatalogHandler) Close() error {
	return h.catalogConn.Close()
}

// Helper functions for JSON mapping

func formatEnum(s string, prefix string) string {
	return strings.ToLower(strings.TrimPrefix(s, prefix))
}

func tripToJSON(t *catalogpb.Trip) map[string]interface{} {
	if t == nil {
		return nil
	}
	return map[string]interface{}{
		"id":              t.Id,
		"organization_id": t.OrganizationId,
		"schedule_id":     t.ScheduleId,
		"service_date":    t.ServiceDate,
		"route_id":        t.RouteId,
		"vehicle_id":      t.VehicleId,
		"vehicle_type":    t.VehicleType,
		"vehicle_class":   t.VehicleClass,
		"departure_time":  t.DepartureTime,
		"arrival_time":    t.ArrivalTime,
		"total_seats":     t.TotalSeats,
		"available_seats": t.AvailableSeats,
		"pricing":         t.Pricing,
		"status":          formatEnum(t.Status.String(), "TRIP_STATUS_"),
		"segments":        t.Segments,
		"created_at":      t.CreatedAt,
		"updated_at":      t.UpdatedAt,
	}
}

func routeToJSON(r *catalogpb.Route) map[string]interface{} {
	if r == nil {
		return nil
	}
	return map[string]interface{}{
		"id":                         r.Id,
		"organization_id":            r.OrganizationId,
		"code":                       r.Code,
		"name":                       r.Name,
		"origin_station_id":          r.OriginStationId,
		"destination_station_id":     r.DestinationStationId,
		"intermediate_stops":         r.IntermediateStops,
		"distance_km":                r.DistanceKm,
		"estimated_duration_minutes": r.EstimatedDurationMinutes,
		"status":                     formatEnum(r.Status.String(), "ROUTE_STATUS_"),
		"created_at":                 r.CreatedAt,
		"updated_at":                 r.UpdatedAt,
		"origin_station":             stationToJSON(r.OriginStation),
		"destination_station":        stationToJSON(r.DestinationStation),
	}
}

func stationToJSON(s *catalogpb.Station) map[string]interface{} {
	if s == nil {
		return nil
	}
	return map[string]interface{}{
		"id":              s.Id,
		"organization_id": s.OrganizationId,
		"code":            s.Code,
		"name":            s.Name,
		"city":            s.City,
		"state":           s.State,
		"country":         s.Country,
		"latitude":        s.Latitude,
		"longitude":       s.Longitude,
		"timezone":        s.Timezone,
		"address":         s.Address,
		"amenities":       s.Amenities,
		"status":          formatEnum(s.Status.String(), "STATION_STATUS_"),
		"created_at":      s.CreatedAt,
		"updated_at":      s.UpdatedAt,
	}
}

func scheduleToJSON(s *catalogpb.Schedule) map[string]interface{} {
	if s == nil {
		return nil
	}
	var pricing map[string]interface{}
	if s.Pricing != nil {
		var segmentPrices []map[string]interface{}
		for _, seg := range s.Pricing.GetSegmentPrices() {
			if seg == nil {
				continue
			}
			segmentPrices = append(segmentPrices, map[string]interface{}{
				"from_station_id":      seg.GetFromStationId(),
				"to_station_id":        seg.GetToStationId(),
				"base_price_paisa":     seg.GetBasePricePaisa(),
				"class_prices":         seg.GetClassPrices(),
				"seat_category_prices": seg.GetSeatCategoryPrices(),
			})
		}
		pricing = map[string]interface{}{
			"base_price_paisa":     s.Pricing.GetBasePricePaisa(),
			"tax_paisa":            s.Pricing.GetTaxPaisa(),
			"booking_fee_paisa":    s.Pricing.GetBookingFeePaisa(),
			"currency":             s.Pricing.GetCurrency(),
			"class_prices":         s.Pricing.GetClassPrices(),
			"seat_category_prices": s.Pricing.GetSeatCategoryPrices(),
			"segment_prices":       segmentPrices,
		}
	}
	return map[string]interface{}{
		"id":                     s.Id,
		"organization_id":        s.OrganizationId,
		"route_id":               s.RouteId,
		"vehicle_id":             s.VehicleId,
		"vehicle_type":           s.VehicleType,
		"vehicle_class":          s.VehicleClass,
		"total_seats":            s.TotalSeats,
		"pricing":                pricing,
		"departure_minutes":      s.DepartureMinutes,
		"arrival_offset_minutes": s.ArrivalOffsetMinutes,
		"timezone":               s.Timezone,
		"start_date":             s.StartDate,
		"end_date":               s.EndDate,
		"days_of_week":           s.DaysOfWeek,
		"status":                 formatEnum(s.Status.String(), "SCHEDULE_STATUS_"),
		"created_at":             s.CreatedAt,
		"updated_at":             s.UpdatedAt,
		"version":                s.Version,
	}
}

func scheduleExceptionToJSON(e *catalogpb.ScheduleException) map[string]interface{} {
	if e == nil {
		return nil
	}
	return map[string]interface{}{
		"id":           e.Id,
		"schedule_id":  e.ScheduleId,
		"service_date": e.ServiceDate,
		"is_added":     e.IsAdded,
		"reason":       e.Reason,
		"created_at":   e.CreatedAt,
	}
}

func parseScheduleStatus(value string) catalogpb.ScheduleStatus {
	switch strings.ToLower(value) {
	case "active":
		return catalogpb.ScheduleStatus_SCHEDULE_STATUS_ACTIVE
	case "inactive":
		return catalogpb.ScheduleStatus_SCHEDULE_STATUS_INACTIVE
	default:
		return catalogpb.ScheduleStatus_SCHEDULE_STATUS_UNSPECIFIED
	}
}

func parseTripStatus(value string) catalogpb.TripStatus {
	switch strings.ToLower(value) {
	case "scheduled":
		return catalogpb.TripStatus_TRIP_STATUS_SCHEDULED
	case "boarding":
		return catalogpb.TripStatus_TRIP_STATUS_BOARDING
	case "departed":
		return catalogpb.TripStatus_TRIP_STATUS_DEPARTED
	case "in_transit":
		return catalogpb.TripStatus_TRIP_STATUS_IN_TRANSIT
	case "arrived":
		return catalogpb.TripStatus_TRIP_STATUS_ARRIVED
	case "cancelled":
		return catalogpb.TripStatus_TRIP_STATUS_CANCELLED
	case "delayed":
		return catalogpb.TripStatus_TRIP_STATUS_DELAYED
	default:
		return catalogpb.TripStatus_TRIP_STATUS_UNSPECIFIED
	}
}
