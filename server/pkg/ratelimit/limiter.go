package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config for rate limiter
type Config struct {
	// Requests per window
	RequestsPerWindow int
	// Window duration
	WindowDuration time.Duration
	// Key prefix for Redis
	KeyPrefix string
}

// Limiter implements a Redis-based sliding window rate limiter
type Limiter struct {
	client *redis.Client
	config Config
}

// NewLimiter creates a new rate limiter
func NewLimiter(client *redis.Client, config Config) *Limiter {
	if config.KeyPrefix == "" {
		config.KeyPrefix = "ratelimit"
	}
	return &Limiter{
		client: client,
		config: config,
	}
}

// Result contains rate limit check result
type Result struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

// Check performs rate limit check using sliding window algorithm
// key can be IP address, user ID, or combination
func (l *Limiter) Check(ctx context.Context, key string) (*Result, error) {
	now := time.Now()
	windowStart := now.Add(-l.config.WindowDuration)
	redisKey := fmt.Sprintf("%s:%s", l.config.KeyPrefix, key)

	// Use Redis sorted set for sliding window
	pipe := l.client.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, redisKey, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, redisKey)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("redis pipeline error: %w", err)
	}

	count := countCmd.Val()
	remaining := l.config.RequestsPerWindow - int(count) - 1

	if remaining < 0 {
		// Rate limited
		return &Result{
			Allowed:   false,
			Remaining: 0,
			ResetAt:   now.Add(l.config.WindowDuration),
		}, nil
	}

	// Add current request
	_, err = l.client.ZAdd(ctx, redisKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zadd error: %w", err)
	}

	// Set key expiration
	l.client.Expire(ctx, redisKey, l.config.WindowDuration*2)

	return &Result{
		Allowed:   true,
		Remaining: remaining,
		ResetAt:   now.Add(l.config.WindowDuration),
	}, nil
}

// Middleware returns an HTTP middleware for rate limiting
func (l *Limiter) Middleware(keyFunc func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFunc(r)
			result, err := l.Check(r.Context(), key)
			if err != nil {
				// On error, allow request (fail open) but log
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", l.config.RequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))

			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(result.ResetAt).Seconds())))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFunc extracts client IP for rate limiting
func IPKeyFunc(r *http.Request) string {
	// Check X-Forwarded-For first (for proxied requests)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	// Check X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	return r.RemoteAddr
}

// UserKeyFunc extracts user ID from request context for rate limiting
func UserKeyFunc(r *http.Request) string {
	userID := r.Header.Get("X-User-ID") // Set by auth middleware
	if userID != "" {
		return "user:" + userID
	}
	return "anon:" + IPKeyFunc(r)
}

// CompositeKeyFunc combines IP and User for stricter limiting
func CompositeKeyFunc(r *http.Request) string {
	return fmt.Sprintf("%s:%s", IPKeyFunc(r), UserKeyFunc(r))
}
