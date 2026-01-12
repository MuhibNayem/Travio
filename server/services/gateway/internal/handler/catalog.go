package handler

import (
	"context"
	"encoding/json"
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

// CreateTripRequest is the HTTP request body for creating a trip
type CreateTripRequest struct {
	RouteID       string  `json:"route_id"`
	VehicleID     string  `json:"vehicle_id"`
	VehicleType   string  `json:"vehicle_type"`
	VehicleClass  string  `json:"vehicle_class"`
	DepartureTime string  `json:"departure_time"` // ISO 8601
	TotalSeats    int32   `json:"total_seats"`
	BasePrice     float64 `json:"base_price"`
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

// CreateTrip creates a new trip via gRPC
func (h *CatalogHandler) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var req CreateTripRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	deptTime, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		http.Error(w, `{"error": "invalid departure_time format (expected RFC3339)"}`, http.StatusBadRequest)
		return
	}

	// Get Org ID from context (set by middleware)
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error": "missing organization context"}`, http.StatusUnauthorized)
		return
	}

	grpcReq := &catalogpb.CreateTripRequest{
		OrganizationId: orgID,
		RouteId:        req.RouteID,
		VehicleId:      req.VehicleID,
		VehicleType:    req.VehicleType,
		VehicleClass:   req.VehicleClass,
		DepartureTime:  deptTime.Unix(),
		TotalSeats:     req.TotalSeats,
		Pricing: &catalogpb.TripPricing{
			BasePricePaisa: int64(req.BasePrice * 100), // Convert to Paisa
			Currency:       "BDT",
		},
	}

	resp, err := h.client.CreateTrip(r.Context(), grpcReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.ResourceExhausted {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPaymentRequired) // 402 Payload Required for Upgrade
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   "plan_limit_exceeded",
				"message": st.Message(),
				"upgrade": true,
			})
			return
		}
		logger.Error("Failed to create trip", "error", err)
		http.Error(w, `{"error": "failed to create trip"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tripToJSON(resp))
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

// GetTrip retrieves a trip by ID
func (h *CatalogHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tripID := chi.URLParam(r, "tripId")
	orgID := middleware.GetOrgID(r.Context())

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetTrip(ctx, &catalogpb.GetTripRequest{
			Id:             tripID,
			OrganizationId: orgID,
		})
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			http.Error(w, "Trip not found", http.StatusNotFound)
			return
		}
		logger.Error("Failed to get trip", "error", err)
		http.Error(w, "Failed to get trip", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.Trip)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tripToJSON(resp))
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

// ListTrips lists trips for an organization
func (h *CatalogHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Get organization ID from JWT token context (set by auth middleware)
	orgID := middleware.GetOrgID(r.Context())

	// Allow override from query parameter for admin/cross-org queries
	if queryOrgID := r.URL.Query().Get("organization_id"); queryOrgID != "" {
		orgID = queryOrgID
	}

	// If still empty, try header (legacy support)
	if orgID == "" {
		orgID = r.Header.Get("X-Organization-ID")
	}

	// Organization ID is required
	if orgID == "" {
		http.Error(w, `{"error": "organization_id is required. Please ensure your user account has an organization assigned and your JWT token contains 'oid' claim."}`, http.StatusBadRequest)
		return
	}

	routeID := r.URL.Query().Get("route_id")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListTrips(ctx, &catalogpb.ListTripsRequest{
			OrganizationId: orgID,
			RouteId:        routeID,
			PageSize:       100,
		})
	})
	if err != nil {
		logger.Error("Failed to list trips", "error", err)
		http.Error(w, "Failed to list trips", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListTripsResponse)

	trips := make([]map[string]interface{}, 0, len(resp.Trips))
	for _, t := range resp.Trips {
		trips = append(trips, tripToJSON(t))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"trips":           trips,
		"next_page_token": resp.NextPageToken,
		"total_count":     resp.TotalCount,
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
