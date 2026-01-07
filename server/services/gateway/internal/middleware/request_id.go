package middleware

import (
	"net/http"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

// RequestID injects a unique request ID into each request
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Set on response header
		w.Header().Set(RequestIDHeader, requestID)

		// Add to request context (for logging)
		r.Header.Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}

// Logger logs each request with timing
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"request_id", r.Header.Get(RequestIDHeader),
		)
		next.ServeHTTP(w, r)
	})
}
