package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	catalogpb "github.com/MuhibNayem/Travio/server/api/proto/catalog/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
			"id":             trip.Id,
			"routeId":        trip.RouteId,
			"type":           trip.VehicleType,
			"operator":       r.OperatorName,
			"vehicleName":    trip.VehicleId,
			"departureTime":  time.Unix(trip.DepartureTime, 0).Format(time.RFC3339),
			"arrivalTime":    time.Unix(trip.ArrivalTime, 0).Format(time.RFC3339),
			"price":          price,
			"class":          trip.VehicleClass,
			"availableSeats": trip.TotalSeats,
			"totalSeats":     trip.TotalSeats,
			"from":           origin.Name,
			"fromCity":       origin.City,
			"to":             dest.Name,
			"toCity":         dest.City,
			"duration":       route.EstimatedDurationMinutes,
			"distance":       route.DistanceKm,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results":  results,
		"total":    resp.TotalCount,
		"nextPage": resp.NextPageToken,
	})
}

// Close closes the gRPC connection
func (h *CatalogHandler) Close() error {
	return h.catalogConn.Close()
}
