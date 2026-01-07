package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	inventorypb "github.com/MuhibNayem/Travio/server/api/proto/inventory/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InventoryHandler handles inventory-related REST endpoints
type InventoryHandler struct {
	conn   *grpc.ClientConn
	client inventorypb.InventoryServiceClient
	cb     *middleware.CircuitBreaker
}

// NewInventoryHandler creates an inventory handler with gRPC connection
func NewInventoryHandler(inventoryURL string, cb *middleware.CircuitBreaker) (*InventoryHandler, error) {
	conn, err := grpc.NewClient(inventoryURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &InventoryHandler{
		conn:   conn,
		client: inventorypb.NewInventoryServiceClient(conn),
		cb:     cb,
	}, nil
}

// CheckAvailability checks seat availability for a trip segment
func (h *InventoryHandler) CheckAvailability(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tripID := chi.URLParam(r, "tripId")
	fromStation := r.URL.Query().Get("from")
	toStation := r.URL.Query().Get("to")
	passengers, _ := strconv.Atoi(r.URL.Query().Get("passengers"))
	if passengers == 0 {
		passengers = 1
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.CheckAvailability(ctx, &inventorypb.CheckAvailabilityRequest{
			TripId:        tripID,
			FromStationId: fromStation,
			ToStationId:   toStation,
			Passengers:    int32(passengers),
		})
	})
	if err != nil {
		http.Error(w, "Failed to check availability", http.StatusInternalServerError)
		return
	}
	resp := result.(*inventorypb.CheckAvailabilityResponse)

	seats := make([]map[string]interface{}, 0)
	for _, s := range resp.Seats {
		seats = append(seats, map[string]interface{}{
			"seatId":     s.SeatId,
			"seatNumber": s.SeatNumber,
			"seatClass":  s.SeatClass,
			"seatType":   s.SeatType,
			"status":     s.Status.String(),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"isAvailable":    resp.IsAvailable,
		"availableSeats": resp.AvailableSeats,
		"pricePaisa":     resp.PricePaisa,
		"seats":          seats,
	})
}

// GetSeatMap returns the seat map for a trip
func (h *InventoryHandler) GetSeatMap(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tripID := chi.URLParam(r, "tripId")
	fromStation := r.URL.Query().Get("from")
	toStation := r.URL.Query().Get("to")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetSeatMap(ctx, &inventorypb.GetSeatMapRequest{
			TripId:        tripID,
			FromStationId: fromStation,
			ToStationId:   toStation,
		})
	})
	if err != nil {
		http.Error(w, "Failed to get seat map", http.StatusInternalServerError)
		return
	}
	resp := result.(*inventorypb.GetSeatMapResponse)

	rows := make([]map[string]interface{}, 0)
	for _, row := range resp.Rows {
		seats := make([]map[string]interface{}, 0)
		for _, s := range row.Seats {
			seats = append(seats, map[string]interface{}{
				"seatId":     s.SeatId,
				"seatNumber": s.SeatNumber,
				"column":     s.Column,
				"seatType":   s.SeatType,
				"seatClass":  s.SeatClass,
				"status":     s.Status.String(),
				"pricePaisa": s.PricePaisa,
			})
		}
		rows = append(rows, map[string]interface{}{
			"rowNumber": row.RowNumber,
			"seats":     seats,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rows":   rows,
		"legend": resp.Legend.StatusColors,
	})
}

// HoldSeatsRequest represents the request to hold seats
type HoldSeatsRequest struct {
	TripID        string   `json:"tripId"`
	FromStationID string   `json:"fromStationId"`
	ToStationID   string   `json:"toStationId"`
	SeatIDs       []string `json:"seatIds"`
	SessionID     string   `json:"sessionId"`
}

// HoldSeats creates a hold on selected seats
func (h *InventoryHandler) HoldSeats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var req HoldSeatsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user ID from auth header (simplified - in production use JWT claims)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.HoldSeats(ctx, &inventorypb.HoldSeatsRequest{
			TripId:              req.TripID,
			FromStationId:       req.FromStationID,
			ToStationId:         req.ToStationID,
			SeatIds:             req.SeatIDs,
			UserId:              userID,
			SessionId:           req.SessionID,
			HoldDurationSeconds: 600, // 10 minutes
		})
	})
	if err != nil {
		http.Error(w, "Failed to hold seats", http.StatusInternalServerError)
		return
	}
	resp := result.(*inventorypb.HoldSeatsResponse)

	w.Header().Set("Content-Type", "application/json")
	if resp.Success {
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusConflict)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"holdId":        resp.HoldId,
		"success":       resp.Success,
		"heldSeatIds":   resp.HeldSeatIds,
		"failedSeatIds": resp.FailedSeatIds,
		"expiresAt":     time.Unix(resp.ExpiresAt, 0).Format(time.RFC3339),
		"failureReason": resp.FailureReason,
	})
}

// ReleaseHold releases a seat hold
func (h *InventoryHandler) ReleaseHold(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	holdID := chi.URLParam(r, "holdId")
	userID := r.Header.Get("X-User-ID")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ReleaseSeats(ctx, &inventorypb.ReleaseSeatsRequest{
			HoldId: holdID,
			UserId: userID,
		})
	})
	if err != nil {
		http.Error(w, "Failed to release hold", http.StatusInternalServerError)
		return
	}
	resp := result.(*inventorypb.ReleaseSeatsResponse)

	if resp.Success {
		w.WriteHeader(http.StatusNoContent)
	} else {
		http.Error(w, "Failed to release hold", http.StatusBadRequest)
	}
}

// Close closes the gRPC connection
func (h *InventoryHandler) Close() error {
	return h.conn.Close()
}
