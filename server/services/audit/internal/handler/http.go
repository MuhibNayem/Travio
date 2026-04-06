package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/service"
)

// HTTPHandler handles HTTP requests for audit operations
type HTTPHandler struct {
	auditService *service.AuditService
}

// NewHTTPHandler creates a new audit HTTP handler
func NewHTTPHandler(auditService *service.AuditService) *HTTPHandler {
	return &HTTPHandler{
		auditService: auditService,
	}
}

// Log handles POST /v1/audit/log
func (h *HTTPHandler) Log(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	var req struct {
		AggregateID    string                 `json:"aggregate_id"`
		EventType      string                 `json:"event_type"`
		ResourceType   string                 `json:"resource_type"`
		Action         string                 `json:"action"`
		ActorID        string                 `json:"actor_id"`
		ActorType      string                 `json:"actor_type"`
		OrganizationID string                 `json:"organization_id"`
		IPAddress      string                 `json:"ip_address"`
		UserAgent      string                 `json:"user_agent"`
		Details        map[string]interface{} `json:"details"`
		Status         string                 `json:"status"`
		FailureReason  string                 `json:"failure_reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	log := &service.AuditLogInput{
		AggregateID:    req.AggregateID,
		EventType:      req.EventType,
		ResourceType:   req.ResourceType,
		Action:         req.Action,
		ActorID:        req.ActorID,
		ActorType:      req.ActorType,
		OrganizationID: req.OrganizationID,
		IPAddress:      req.IPAddress,
		UserAgent:      req.UserAgent,
		Details:        req.Details,
		Status:         req.Status,
		FailureReason:  req.FailureReason,
	}

	id, err := h.auditService.LogFromHTTP(r.Context(), log)
	if err != nil {
		logger.Error("Failed to create audit log", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to create audit log"})
		return
	}

	respondJSON(w, http.StatusCreated, map[string]string{"id": id, "status": "created"})
}

// GetByActorID handles GET /v1/audit/actor/{id}
func (h *HTTPHandler) GetByActorID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	actorID := r.URL.Query().Get("actor_id")
	if actorID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "actor_id parameter required"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}

	logs, total, err := h.auditService.GetByActorID(r.Context(), actorID, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by actor", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get audit logs"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

// GetByResourceID handles GET /v1/audit/resource/{id}
func (h *HTTPHandler) GetByResourceID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	resourceID := r.URL.Query().Get("resource_id")
	if resourceID == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "resource_id parameter required"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}

	logs, total, err := h.auditService.GetByResourceID(r.Context(), resourceID, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by resource", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get audit logs"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

// GetByTimeRange handles GET /v1/audit/timerange
func (h *HTTPHandler) GetByTimeRange(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "from and to parameters required (RFC3339)"})
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid from format, use RFC3339"})
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		respondJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid to format, use RFC3339"})
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}

	logs, total, err := h.auditService.GetByTimeRange(r.Context(), from, to, limit, offset)
	if err != nil {
		logger.Error("Failed to get audit logs by time range", "error", err)
		respondJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get audit logs"})
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"total": total,
	})
}

// RegisterRoutes registers audit HTTP routes
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/v1/audit/log", h.Log)
	mux.HandleFunc("/v1/audit/actor", h.GetByActorID)
	mux.HandleFunc("/v1/audit/resource", h.GetByResourceID)
	mux.HandleFunc("/v1/audit/timerange", h.GetByTimeRange)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode JSON response", "error", err)
	}
}
