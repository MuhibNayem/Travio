package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
)

// QueueClient interface for queue service
type QueueClient interface {
	ValidateToken(ctx context.Context, token string) (bool, string, string, error)
}

// QueueMiddleware checks if high-demand endpoints require queue token
type QueueMiddleware struct {
	queueClient QueueClient
	// Endpoints that require queue token during high demand
	protectedEndpoints map[string]bool
}

// NewQueueMiddleware creates a new queue middleware
func NewQueueMiddleware(client QueueClient) *QueueMiddleware {
	return &QueueMiddleware{
		queueClient: client,
		protectedEndpoints: map[string]bool{
			"/v1/holds":  true,
			"/v1/orders": true,
		},
	}
}

// Middleware returns the queue validation middleware
func (m *QueueMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if endpoint is protected
		if !m.protectedEndpoints[r.URL.Path] {
			next.ServeHTTP(w, r)
			return
		}

		// Check for queue token
		queueToken := r.Header.Get("X-Queue-Token")
		if queueToken == "" {
			// Check query param
			queueToken = r.URL.Query().Get("queue_token")
		}

		if queueToken == "" {
			// No token - check if queue is active for this event
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusPreconditionRequired)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":     "queue_required",
				"message":   "This event requires joining the queue first",
				"queue_url": "/v1/queue/join",
			})
			return
		}

		// Validate token
		if m.queueClient != nil {
			valid, userID, eventID, err := m.queueClient.ValidateToken(r.Context(), queueToken)
			if err != nil {
				logger.Error("queue token validation failed", "error", err)
				http.Error(w, "Queue validation failed", http.StatusInternalServerError)
				return
			}

			if !valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":   "token_invalid",
					"message": "Queue token is invalid or expired",
				})
				return
			}

			// Add queue context to request
			r.Header.Set("X-Queue-User-ID", userID)
			r.Header.Set("X-Queue-Event-ID", eventID)
		}

		next.ServeHTTP(w, r)
	})
}
