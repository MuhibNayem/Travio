package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	orderpb "github.com/MuhibNayem/Travio/server/api/proto/order/v1"
	"github.com/go-chi/chi/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OrderHandler handles order-related REST endpoints
type OrderHandler struct {
	conn   *grpc.ClientConn
	client orderpb.OrderServiceClient
}

// NewOrderHandler creates an order handler with gRPC connection
func NewOrderHandler(orderURL string) (*OrderHandler, error) {
	conn, err := grpc.NewClient(orderURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &OrderHandler{
		conn:   conn,
		client: orderpb.NewOrderServiceClient(conn),
	}, nil
}

// CreateOrderRequest represents the order creation request
type CreateOrderRequest struct {
	TripID        string `json:"tripId"`
	FromStationID string `json:"fromStationId"`
	ToStationID   string `json:"toStationId"`
	HoldID        string `json:"holdId"`
	Passengers    []struct {
		NID         string `json:"nid"`
		Name        string `json:"name"`
		SeatID      string `json:"seatId"`
		DateOfBirth string `json:"dateOfBirth"`
		Gender      string `json:"gender"`
		Age         int    `json:"age"`
	} `json:"passengers"`
	PaymentMethod struct {
		Type  string `json:"type"`
		Token string `json:"token,omitempty"`
	} `json:"paymentMethod"`
	ContactEmail   string `json:"contactEmail"`
	ContactPhone   string `json:"contactPhone"`
	IdempotencyKey string `json:"idempotencyKey"`
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

	userID := r.Header.Get("X-User-ID")
	orgID := r.Header.Get("X-Organization-ID")

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

	resp, err := h.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{
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
	if err != nil {
		http.Error(w, "Failed to create order", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(orderToJSON(resp.Order))
}

// GetOrder retrieves an order by ID
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	orderID := chi.URLParam(r, "orderId")
	userID := r.Header.Get("X-User-ID")

	resp, err := h.client.GetOrder(ctx, &orderpb.GetOrderRequest{
		OrderId: orderID,
		UserId:  userID,
	})
	if err != nil {
		http.Error(w, "Order not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orderToJSON(resp))
}

// ListOrders lists orders for a user
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	userID := r.Header.Get("X-User-ID")
	pageToken := r.URL.Query().Get("pageToken")

	resp, err := h.client.ListOrders(ctx, &orderpb.ListOrdersRequest{
		UserId:    userID,
		PageSize:  20,
		PageToken: pageToken,
	})
	if err != nil {
		http.Error(w, "Failed to list orders", http.StatusInternalServerError)
		return
	}

	orders := make([]map[string]interface{}, 0, len(resp.Orders))
	for _, o := range resp.Orders {
		orders = append(orders, orderToJSON(o))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"orders":   orders,
		"nextPage": resp.NextPageToken,
		"total":    resp.TotalCount,
	})
}

// CancelOrder cancels an order
func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	orderID := chi.URLParam(r, "orderId")
	userID := r.Header.Get("X-User-ID")

	var req struct {
		Reason string `json:"reason"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	resp, err := h.client.CancelOrder(ctx, &orderpb.CancelOrderRequest{
		OrderId: orderID,
		UserId:  userID,
		Reason:  req.Reason,
	})
	if err != nil {
		http.Error(w, "Failed to cancel order", http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"success": resp.Success,
		"order":   orderToJSON(resp.Order),
	}

	if resp.Refund != nil {
		result["refund"] = map[string]interface{}{
			"refundId":    resp.Refund.RefundId,
			"amountPaisa": resp.Refund.AmountPaisa,
			"status":      resp.Refund.Status,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// orderToJSON converts a protobuf Order to a JSON-friendly map
func orderToJSON(o *orderpb.Order) map[string]interface{} {
	if o == nil {
		return nil
	}

	passengers := make([]map[string]interface{}, 0)
	for _, p := range o.Passengers {
		passengers = append(passengers, map[string]interface{}{
			"nid":        p.Nid,
			"name":       p.Name,
			"seatId":     p.SeatId,
			"seatNumber": p.SeatNumber,
			"seatClass":  p.SeatClass,
		})
	}

	return map[string]interface{}{
		"id":              o.Id,
		"tripId":          o.TripId,
		"fromStationId":   o.FromStationId,
		"toStationId":     o.ToStationId,
		"status":          o.Status.String(),
		"passengers":      passengers,
		"subtotalPaisa":   o.SubtotalPaisa,
		"taxPaisa":        o.TaxPaisa,
		"bookingFeePaisa": o.BookingFeePaisa,
		"discountPaisa":   o.DiscountPaisa,
		"totalPaisa":      o.TotalPaisa,
		"currency":        o.Currency,
		"paymentStatus":   o.PaymentStatus.String(),
		"contactEmail":    o.ContactEmail,
		"contactPhone":    o.ContactPhone,
		"createdAt":       time.Unix(o.CreatedAt, 0).Format(time.RFC3339),
		"expiresAt":       time.Unix(o.ExpiresAt, 0).Format(time.RFC3339),
	}
}

// Close closes the gRPC connection
func (h *OrderHandler) Close() error {
	return h.conn.Close()
}
