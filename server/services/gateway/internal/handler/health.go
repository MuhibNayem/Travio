package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

// Health handles health check requests
func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// Ready handles readiness check (depends on upstream services)
func Ready(w http.ResponseWriter, r *http.Request) {
	// In production, check connectivity to critical services
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ready",
	})
}
