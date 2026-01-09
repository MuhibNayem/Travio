package handler

import (
	"encoding/json"
	"net/http"
	"time"

	fraudv1 "github.com/MuhibNayem/Travio/server/api/proto/fraud/v1"
	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// FraudHandler handles fraud-related HTTP requests.
type FraudHandler struct {
	client *client.FraudClient
}

// NewFraudHandler creates a new fraud handler.
func NewFraudHandler(c *client.FraudClient) *FraudHandler {
	return &FraudHandler{client: c}
}

// RegisterRoutes registers fraud routes.
func (h *FraudHandler) RegisterRoutes(r chi.Router) {
	r.Route("/fraud", func(r chi.Router) {
		r.Use(middleware.RequireAuth)
		r.Post("/analyze", h.AnalyzeBooking)
		r.Post("/verify-document", h.VerifyDocument)
		r.Get("/health", h.Health)
	})
}

// AnalyzeBooking analyzes a booking for fraud.
func (h *FraudHandler) AnalyzeBooking(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrderID             string   `json:"order_id"`
		UserID              string   `json:"user_id"`
		TripID              string   `json:"trip_id"`
		PassengerCount      int32    `json:"passenger_count"`
		PassengerNIDs       []string `json:"passenger_nids"`
		PassengerNames      []string `json:"passenger_names"`
		BookingTimestamp    string   `json:"booking_timestamp"`
		IPAddress           string   `json:"ip_address"`
		UserAgent           string   `json:"user_agent"`
		PaymentMethod       string   `json:"payment_method"`
		TotalAmountPaisa    int64    `json:"total_amount_paisa"`
		BookingsLast24Hours int32    `json:"bookings_last_24_hours"`
		BookingsLastWeek    int32    `json:"bookings_last_week"`
		PreviousFraudFlags  int32    `json:"previous_fraud_flags"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var bookingTime int64
	if req.BookingTimestamp != "" {
		if t, err := time.Parse(time.RFC3339, req.BookingTimestamp); err == nil {
			bookingTime = t.Unix()
		}
	}
	if bookingTime == 0 {
		bookingTime = time.Now().Unix()
	}

	resp, err := h.client.AnalyzeBooking(r.Context(), &fraudv1.AnalyzeBookingRequest{
		OrderId:              req.OrderID,
		UserId:               req.UserID,
		TripId:               req.TripID,
		PassengerCount:       req.PassengerCount,
		PassengerNids:        req.PassengerNIDs,
		PassengerNames:       req.PassengerNames,
		BookingTimestamp:     bookingTime,
		IpAddress:            req.IPAddress,
		UserAgent:            req.UserAgent,
		PaymentMethod:        req.PaymentMethod,
		TotalAmountPaisa:     req.TotalAmountPaisa,
		BookingsLast_24Hours: req.BookingsLast24Hours,
		BookingsLastWeek:     req.BookingsLastWeek,
		PreviousFraudFlags:   req.PreviousFraudFlags,
	})
	if err != nil {
		logger.Error("Fraud analysis failed", "error", err)
		http.Error(w, "Fraud analysis failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// VerifyDocument verifies a document.
func (h *FraudHandler) VerifyDocument(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Document file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imageData := make([]byte, header.Size)
	if _, err := file.Read(imageData); err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	resp, err := h.client.VerifyDocument(r.Context(), &fraudv1.VerifyDocumentRequest{
		DocumentType:  r.FormValue("document_type"),
		DocumentImage: imageData,
		ImageMimeType: header.Header.Get("Content-Type"),
		ExpectedNid:   r.FormValue("expected_nid"),
		ExpectedName:  r.FormValue("expected_name"),
	})
	if err != nil {
		logger.Error("Document verification failed", "error", err)
		http.Error(w, "Verification failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// Health checks fraud service health.
func (h *FraudHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp, err := h.client.Health(r.Context())
	if err != nil {
		http.Error(w, "Fraud service unhealthy", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": resp.Status})
}
