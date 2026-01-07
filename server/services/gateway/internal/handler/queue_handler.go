package handler

import (
	"encoding/json"
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/gateway/internal/client"
)

// QueueHandler handles virtual waiting room requests via gRPC
type QueueHandler struct {
	client *client.QueueClient
}

// NewQueueHandler creates a new queue handler with gRPC client
func NewQueueHandler(queueClient *client.QueueClient) *QueueHandler {
	return &QueueHandler{client: queueClient}
}

// JoinQueueRequest represents the join queue request body
type JoinQueueRequest struct {
	EventID   string `json:"event_id"`
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

// JoinQueue adds user to the waiting queue via gRPC
func (h *QueueHandler) JoinQueue(w http.ResponseWriter, r *http.Request) {
	var req JoinQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	position, err := h.client.JoinQueue(r.Context(), req.EventID, req.UserID, req.SessionID)
	if err != nil {
		logger.Error("Failed to join queue", "error", err)
		http.Error(w, `{"error": "queue service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// GetQueuePosition returns current position in queue via gRPC
func (h *QueueHandler) GetQueuePosition(w http.ResponseWriter, r *http.Request) {
	eventID := r.URL.Query().Get("event_id")
	userID := r.URL.Query().Get("user_id")

	if eventID == "" || userID == "" {
		http.Error(w, `{"error": "event_id and user_id required"}`, http.StatusBadRequest)
		return
	}

	position, err := h.client.GetPosition(r.Context(), eventID, userID)
	if err != nil {
		logger.Error("Failed to get queue position", "error", err)
		http.Error(w, `{"error": "queue service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(position)
}

// VerifyTokenRequest represents the verify token request body
type VerifyTokenRequest struct {
	Token string `json:"token"`
}

// VerifyQueueToken verifies if user can proceed from queue via gRPC
func (h *QueueHandler) VerifyQueueToken(w http.ResponseWriter, r *http.Request) {
	var req VerifyTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.client.ValidateToken(r.Context(), req.Token)
	if err != nil {
		logger.Error("Failed to verify queue token", "error", err)
		http.Error(w, `{"error": "queue service unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
