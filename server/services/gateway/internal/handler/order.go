package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	orderpb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderHandler handles order-related REST endpoints
type OrderHandler struct {
	conn   *grpc.ClientConn
	client orderpb.OrderServiceClient
	cb     *middleware.CircuitBreaker
}

// NewOrderHandler creates an order handler with gRPC connection
func NewOrderHandler(orderURL string, cb *middleware.CircuitBreaker) (*OrderHandler, error) {
	conn, err := grpc.NewClient(orderURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &OrderHandler{
		conn:   conn,
		client: orderpb.NewOrderServiceClient(conn),
		cb:     cb,
	}, nil
}

// CreateOrderRequest represents the order creation request
type CreateOrderRequest struct {
	TripID        string `json:"trip_id"`
	FromStationID string `json:"from_station_id"`
	ToStationID   string `json:"to_station_id"`
	HoldID        string `json:"hold_id"`
	Passengers    []struct {
		NID         string `json:"nid"`
		Name        string `json:"name"`
		SeatID      string `json:"seat_id"`
		DateOfBirth string `json:"date_of_birth"`
		Gender      string `json:"gender"`
		Age         int    `json:"age"`
	} `json:"passengers"`
	PaymentMethod struct {
		Type  string `json:"type"`
		Token string `json:"token,omitempty"`
	} `json:"payment_method"`
	ContactEmail   string `json:"contact_email"`
	ContactPhone   string `json:"contact_phone"`
	IdempotencyKey string `json:"idempotency_key"`
}

// CreateOrder creates a new order
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get idempotency key from header or body
	idempotencyKey := r.Header.Get("X-Idempotency-Key")
	if idempotencyKey == "" {
		idempotencyKey = req.IdempotencyKey
	}
	if idempotencyKey == "" {
		http.Error(w, "Idempotency key required", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r.Context())
	orgID := middleware.GetOrgID(r.Context())

	passengers := make([]*orderpb.PassengerRequest, 0, len(req.Passengers))
	for _, p := range req.Passengers {
		passengers = append(passengers, &orderpb.PassengerRequest{
			Nid:         p.NID,
			Name:        p.Name,
			SeatId:      p.SeatID,
			DateOfBirth: p.DateOfBirth,
			Gender:      p.Gender,
			Age:         int32(p.Age),
		})
	}

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
			OrganizationId: orgID,
			UserId:         userID,
			TripId:         req.TripID,
			FromStationId:  req.FromStationID,
			ToStationId:    req.ToStationID,
			HoldId:         req.HoldID,
			Passengers:     passengers,
			PaymentMethod: &orderpb.PaymentMethod{
				Type:  req.PaymentMethod.Type,
				Token: req.PaymentMethod.Token,
			},
			ContactEmail:   req.ContactEmail,
			ContactPhone:   req.ContactPhone,
			IdempotencyKey: idempotencyKey,
		})
	})
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}
	resp := result.(*orderpb.CreateOrderResponse)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(orderToJSON(resp.Order))
}

// GetOrder retrieves an order by ID
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orderID := chi.URLParam(r, "orderId")
	userID := middleware.GetUserID(r.Context())

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.GetOrder(ctx, &orderpb.GetOrderRequest{
			OrderId: orderID,
			UserId:  userID,
		})
	})
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}
	resp := result.(*orderpb.Order)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderToJSON(resp))
}

// ListOrders lists orders for a user
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := middleware.GetUserID(r.Context())
	pageToken := r.URL.Query().Get("page_token")

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.ListOrders(ctx, &orderpb.ListOrdersRequest{
			UserId:    userID,
			PageSize:  20,
			PageToken: pageToken,
		})
	})
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}
	resp := result.(*orderpb.ListOrdersResponse)

	orders := make([]map[string]interface{}, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, orderToJSON(o))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders":    orders,
		"next_page": resp.NextPageToken,
		"total":     resp.TotalCount,
	})
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	orderID := chi.URLParam(r, "orderId")
	userID := middleware.GetUserID(r.Context())

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	result, err := h.cb.Execute(func() (interface{}, error) {
		return h.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{
			OrderId: orderID,
			UserId:  userID,
			Reason:  req.Reason,
		})
	})
	if err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}
	resp := result.(*orderpb.CancelOrderResponse)

	response := map[string]interface{}{
		"success": resp.Success,
		"order":   orderToJSON(resp.Order),
	}

	if resp.Refund != nil {
		response["refund"] = map[string]interface{}{
			"refund_id":    resp.Refund.RefundId,
			"amount_paisa": resp.Refund.AmountPaisa,
			"status":       resp.Refund.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// orderToJSON converts a protobuf Order to a JSON-friendly map
func orderToJSON(o *orderpb.Order) map[string]interface{} {
	if o == nil {
		return nil
	}

	passengers := make([]map[string]interface{}, 0)
	for _, p := range o.Passengers {
		passengers = append(passengers, map[string]interface{}{
			"nid":         p.Nid,
			"name":        p.Name,
			"seat_id":     p.SeatId,
			"seat_number": p.SeatNumber,
			"seat_class":  p.SeatClass,
		})
	}

	return map[string]interface{}{
		"id":                o.Id,
		"trip_id":           o.TripId,
		"from_station_id":   o.FromStationId,
		"to_station_id":     o.ToStationId,
		"status":            o.Status.String(),
		"passengers":        passengers,
		"subtotal_paisa":    o.SubtotalPaisa,
		"tax_paisa":         o.TaxPaisa,
		"booking_fee_paisa": o.BookingFeePaisa,
		"discount_paisa":    o.DiscountPaisa,
		"total_paisa":       o.TotalPaisa,
		"currency":          o.Currency,
		"payment_status":    o.PaymentStatus.String(),
		"contact_email":     o.ContactEmail,
		"contact_phone":     o.ContactPhone,
		"created_at":        time.Unix(o.CreatedAt, 0).Format(time.RFC3339),
		"expires_at":        time.Unix(o.ExpiresAt, 0).Format(time.RFC3339),
	}
}

// Close closes the gRPC connection
func (h *OrderHandler) Close() error {
	return h.conn.Close()
}
