package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type IdempotencyMiddleware struct {
	redisClient *redis.Client
}

func NewIdempotencyMiddleware(redisClient *redis.Client) *IdempotencyMiddleware {
	return &IdempotencyMiddleware{redisClient: redisClient}
}

// Middleware checks for Idempotency-Key header and returns cached response if found
func (m *IdempotencyMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extract Key
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		redisKey := fmt.Sprintf("idempotency:%s", key)

		// 2. Check Redis
		// State: "PROCESSING" | Body
		val, err := m.redisClient.Get(ctx, redisKey).Result()
		if err == nil {
			if val == "PROCESSING" {
				http.Error(w, "Request currently being processed", http.StatusConflict)
				return
			}
			// Return cached response
			w.Header().Set("X-Idempotency-Hit", "true")
			w.Header().Set("Content-Type", "application/json") // Assumption
			w.Write([]byte(val))
			return
		} else if err != redis.Nil {
			// Redis error
			fmt.Printf("Redis Error in Idempotency: %v\n", err) // Debug log
			http.Error(w, "Idempotency check failed", http.StatusInternalServerError)
			return
		}

		// 3. Mark as Processing
		// Set NX with short TTL (e.g. 1 min) to prevent infinite lock on crash
		ok, err := m.redisClient.SetNX(ctx, redisKey, "PROCESSING", 60*time.Second).Result()
		if err != nil {
			http.Error(w, "Failed to acquire lock", http.StatusInternalServerError)
			return
		}
		if !ok {
			// Race condition: another request just took it
			http.Error(w, "Request currently being processed", http.StatusConflict)
			return
		}

		// 4. Capture Response
		// We need to wrap the writer to capture valid JSON responses
		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next.ServeHTTP(rec, r)

		// 5. Store Result or Delete Lock
		if rec.statusCode >= 200 && rec.statusCode < 300 {
			// Success: Store body with 24h TTL
			m.redisClient.Set(ctx, redisKey, rec.body.String(), 24*time.Hour)
		} else {
			// Failure: Delete key so user can retry
			m.redisClient.Del(ctx, redisKey)
		}
	})
}

// responseRecorder wraps http.ResponseWriter to capture status and body
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
