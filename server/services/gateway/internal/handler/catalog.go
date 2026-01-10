package handler

import (
	"context"
	"encoding/json"
	"net/http"
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
		http.Error(w, "Failed to fetch stations", http.StatusInternalServerError)
		return
	}
	resp := result.(*catalogpb.ListStationsResponse)

	// Convert to JSON-friendly format
	stations := make([]map[string]interface{}, 0, len(resp.Stations))
	for _, s := range resp.Stations {
		stations = append(stations, map[string]interface{}{
			"id":        s.Id,
			"code":      s.Code,
			"name":      s.Name,
			"city":      s.City,
			"state":     s.State,
			"country":   s.Country,
			"latitude":  s.Latitude,
			"longitude": s.Longitude,
			"timezone":  s.Timezone,
			"amenities": s.Amenities,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stations": stations,
		"total":    len(stations),
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
	orgID := r.Header.Get("X-Organization-ID")
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
	json.NewEncoder(w).Encode(resp)
}

// GetTrip retrieves a trip by ID
func (h *CatalogHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tripID := chi.URLParam(r, "tripId")
	orgID := r.Header.Get("X-Organization-ID") // Optional, for admin/operator check?

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

	// Enrich/Map response if needed?
	// Provide standard JSON structure matching SearchTrips result item style?
	// Or just return protobuf JSON. Protobuf JSON has standard casing usually.

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ListTrips lists trips for an organization
func (h *CatalogHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		// Fallback to header if set by middleware for operators
		orgID = r.Header.Get("X-Organization-ID")
	}

	if orgID == "" {
		http.Error(w, "organization_id is required", http.StatusBadRequest)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetStation retrieves a station by ID
func (h *CatalogHandler) GetStation(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stationID := chi.URLParam(r, "stationId")
	orgID := r.Header.Get("X-Organization-ID")

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
	json.NewEncoder(w).Encode(resp)
}

// Close closes the gRPC connection
func (h *CatalogHandler) Close() error {
	return h.catalogConn.Close()
}
