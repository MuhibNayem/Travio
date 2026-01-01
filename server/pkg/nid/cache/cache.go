package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/nid"
	"github.com/redis/go-redis/v9"
)

// CachingProvider wraps another provider with Redis caching
// Implements the Decorator pattern for transparent caching
type CachingProvider struct {
	wrapped    nid.Provider
	redis      *redis.Client
	defaultTTL time.Duration
	keyPrefix  string
}

// CacheConfig holds caching configuration
type CacheConfig struct {
	DefaultTTL time.Duration // Default: 24h
	KeyPrefix  string        // Default: "nid:cache:"
}

func NewCachingProvider(wrapped nid.Provider, redisClient *redis.Client, cfg CacheConfig) *CachingProvider {
	if cfg.DefaultTTL == 0 {
		cfg.DefaultTTL = 24 * time.Hour
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "nid:cache:"
	}
	return &CachingProvider{
		wrapped:    wrapped,
		redis:      redisClient,
		defaultTTL: cfg.DefaultTTL,
		keyPrefix:  cfg.KeyPrefix,
	}
}

func (p *CachingProvider) Name() string {
	return p.wrapped.Name() + "+cache"
}

func (p *CachingProvider) Country() string {
	return p.wrapped.Country()
}

func (p *CachingProvider) Verify(ctx context.Context, req *nid.VerifyRequest) (*nid.VerifyResponse, error) {
	// Build cache key
	cacheKey := p.buildCacheKey(req.NID, req.DateOfBirth)

	// Try cache first
	cached, err := p.getFromCache(ctx, cacheKey)
	if err == nil && cached != nil {
		// Cache hit - check if still valid
		if time.Now().Before(cached.ExpiresAt) {
			cached.ProviderName = p.Name() + " (cached)"
			return cached, nil
		}
		// Cache expired, continue to provider
	}

	// Cache miss or expired - call wrapped provider
	resp, err := p.wrapped.Verify(ctx, req)
	if err != nil {
		// On provider error, return stale cache if available
		if cached != nil {
			cached.ProviderName = p.Name() + " (stale)"
			return cached, nil
		}
		return nil, err
	}

	// Cache successful verifications
	if resp.IsValid {
		ttl := p.defaultTTL
		if !resp.ExpiresAt.IsZero() {
			ttl = time.Until(resp.ExpiresAt)
		}
		_ = p.setCache(ctx, cacheKey, resp, ttl)
	}

	return resp, nil
}

func (p *CachingProvider) HealthCheck(ctx context.Context) error {
	// Check both Redis and wrapped provider
	if err := p.redis.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("cache unavailable: %w", err)
	}
	return p.wrapped.HealthCheck(ctx)
}

func (p *CachingProvider) buildCacheKey(nidStr string, dob time.Time) string {
	dobStr := ""
	if !dob.IsZero() {
		dobStr = dob.Format("20060102")
	}
	return fmt.Sprintf("%s%s:%s:%s", p.keyPrefix, p.wrapped.Country(), nidStr, dobStr)
}

func (p *CachingProvider) getFromCache(ctx context.Context, key string) (*nid.VerifyResponse, error) {
	data, err := p.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var resp nid.VerifyResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (p *CachingProvider) setCache(ctx context.Context, key string, resp *nid.VerifyResponse, ttl time.Duration) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	return p.redis.Set(ctx, key, data, ttl).Err()
}

// InvalidateCache removes a cached entry
func (p *CachingProvider) InvalidateCache(ctx context.Context, nidStr string, dob time.Time) error {
	key := p.buildCacheKey(nidStr, dob)
	return p.redis.Del(ctx, key).Err()
}

// GetCacheStats returns cache statistics
type CacheStats struct {
	Hits    int64   `json:"hits"`
	Misses  int64   `json:"misses"`
	Size    int64   `json:"size"`
	HitRate float64 `json:"hit_rate"`
}

func (p *CachingProvider) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	// Get keys matching our prefix
	pattern := p.keyPrefix + "*"
	keys, err := p.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	return &CacheStats{
		Size: int64(len(keys)),
		// Note: For production, use Redis INFO for actual hit/miss stats
	}, nil
}
