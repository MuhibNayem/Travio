package ratelimit

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// TieredRateLimiter provides tiered rate limiting based on user type
type TieredRateLimiter struct {
	client *redis.Client
	tiers  map[string]TierConfig
}

// TierConfig defines rate limits for a tier
type TierConfig struct {
	Name           string
	RequestsPerMin int
	BurstSize      int
	CostMultiplier float64 // For weighted endpoints
}

// DefaultTiers provides standard tier configurations
var DefaultTiers = map[string]TierConfig{
	"anonymous": {Name: "anonymous", RequestsPerMin: 30, BurstSize: 10, CostMultiplier: 1.0},
	"free":      {Name: "free", RequestsPerMin: 60, BurstSize: 20, CostMultiplier: 1.0},
	"premium":   {Name: "premium", RequestsPerMin: 300, BurstSize: 50, CostMultiplier: 1.0},
	"business":  {Name: "business", RequestsPerMin: 1000, BurstSize: 100, CostMultiplier: 1.0},
}

// EndpointCosts defines relative costs for different endpoints
var EndpointCosts = map[string]int{
	"/v1/trips/search":       1,
	"/v1/holds":              5,  // More expensive - involves inventory
	"/v1/orders":             10, // Most expensive - creates order
	"/v1/trips/{id}/seatmap": 2,
}

// NewTieredRateLimiter creates a new tiered rate limiter
func NewTieredRateLimiter(redisAddr string) *TieredRateLimiter {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &TieredRateLimiter{
		client: client,
		tiers:  DefaultTiers,
	}
}

// Check checks if a request should be allowed
func (r *TieredRateLimiter) Check(ctx context.Context, identifier, tier, endpoint string) (*RateLimitResult, error) {
	tierConfig, ok := r.tiers[tier]
	if !ok {
		tierConfig = r.tiers["anonymous"]
	}

	// Get endpoint cost
	cost := 1
	if c, exists := EndpointCosts[endpoint]; exists {
		cost = c
	}

	key := fmt.Sprintf("ratelimit:%s:%s", tier, identifier)

	now := time.Now().Unix()
	windowStart := now - 60 // 1 minute window

	// Use Redis transaction for atomic operations
	pipe := r.client.Pipeline()

	// Remove old entries outside window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request with timestamp as score
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: fmt.Sprintf("%d:%d", now, cost)})

	// Set expiry
	pipe.Expire(ctx, key, 2*time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	currentCount := countCmd.Val()
	limit := int64(tierConfig.RequestsPerMin)
	remaining := limit - currentCount - int64(cost)
	if remaining < 0 {
		remaining = 0
	}

	allowed := currentCount+int64(cost) <= limit

	return &RateLimitResult{
		Allowed:    allowed,
		Limit:      int(limit),
		Remaining:  int(remaining),
		ResetAt:    time.Now().Add(60 * time.Second),
		RetryAfter: 60 - int(now%60),
		Tier:       tierConfig.Name,
	}, nil
}

// RateLimitResult contains rate limit check result
type RateLimitResult struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter int // seconds
	Tier       string
}

// Middleware returns an HTTP middleware for rate limiting
func (r *TieredRateLimiter) Middleware(getTier func(r *http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Get identifier (IP or user ID)
			identifier := req.RemoteAddr
			if userID := req.Header.Get("X-User-ID"); userID != "" {
				identifier = userID
			}

			// Get tier
			tier := "anonymous"
			if getTier != nil {
				tier = getTier(req)
			}

			result, err := r.Check(req.Context(), identifier, tier, req.URL.Path)
			if err != nil {
				// Fail open on error
				next.ServeHTTP(w, req)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", result.Limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", result.Remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
			w.Header().Set("X-RateLimit-Tier", result.Tier)

			if !result.Allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", result.RetryAfter))
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// Close closes the Redis connection
func (r *TieredRateLimiter) Close() error {
	return r.client.Close()
}
