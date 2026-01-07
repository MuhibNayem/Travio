package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MuhibNayem/Travio/server/services/queue/internal/service"
)

// HTTPHandler handles REST requests for queue service
type HTTPHandler struct {
	svc *service.QueueService
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(svc *service.QueueService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

// JoinQueueRequest for REST API
type JoinQueueRequest struct {
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// JoinQueue handles POST /v1/queue/join
func (h *HTTPHandler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	entry, err := h.svc.JoinQueue(r.Context(), req.EventID, req.UserID, req.SessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position":       entry.Position,
		"estimated_wait": entry.EstimatedWait.Seconds(),
		"token":          entry.Token,
		"status":         entry.Status,
	})
}

// GetPosition handles GET /v1/queue/position?event_id=X&user_id=Y
func (h *HTTPHandler) GetPosition(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID := r.URL.Query().Get("event_id")
	userID := r.URL.Query().Get("user_id")

	entry, err := h.svc.GetPosition(r.Context(), eventID, userID)
	if err != nil {
		http.Error(w, "Not in queue", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"position":       entry.Position,
		"estimated_wait": entry.EstimatedWait.Seconds(),
		"token":          entry.Token,
		"status":         entry.Status,
	})
}

// ValidateToken handles POST /v1/queue/validate
func (h *HTTPHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	valid, userID, eventID, err := h.svc.ValidateAdmission(r.Context(), req.Token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":    valid,
		"user_id":  userID,
		"event_id": eventID,
	})
}

// GetStats handles GET /v1/queue/stats?event_id=X
func (h *HTTPHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	eventID := r.URL.Query().Get("event_id")

	stats, err := h.svc.GetStats(r.Context(), eventID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Health handles GET /health
func (h *HTTPHandler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
