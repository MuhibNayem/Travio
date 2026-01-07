package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter provides Redis-backed rate limiting
type RateLimiter struct {
	client     *redis.Client
	maxReqs    int
	windowSecs int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redisURL string, maxReqs, windowSecs int) *RateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})
	return &RateLimiter{
		client:     client,
		maxReqs:    maxReqs,
		windowSecs: windowSecs,
	}
}

// Middleware returns the rate limiting middleware
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use IP as key; in production add user ID for authenticated requests
		key := fmt.Sprintf("rate:%s", r.RemoteAddr)

		ctx := context.Background()

		// Sliding window counter using Redis
		count, err := rl.client.Incr(ctx, key).Result()
		if err != nil {
			// If Redis fails, allow the request (fail open)
			next.ServeHTTP(w, r)
			return
		}

		// Set expiry on first request
		if count == 1 {
			rl.client.Expire(ctx, key, time.Duration(rl.windowSecs)*time.Second)
		}

		// Check limit
		if count > int64(rl.maxReqs) {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", rl.windowSecs))
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
			w.Header().Set("X-RateLimit-Remaining", "0")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Set rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.maxReqs))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.maxReqs-int(count)))

		next.ServeHTTP(w, r)
	})
}

// Close closes the Redis connection
func (rl *RateLimiter) Close() error {
	return rl.client.Close()
}
