package entitlement

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	cacheKeyPrefix    = "entitlement:"
	invalidateChannel = "entitlement:invalidate"
)

// CachedChecker implements EntitlementChecker with Redis caching.
type CachedChecker struct {
	redis   *redis.Client
	fetcher EntitlementFetcher
	config  Config
}

// EntitlementFetcher is the interface for fetching entitlements from the source of truth.
type EntitlementFetcher interface {
	FetchEntitlement(ctx context.Context, orgID string) (*Entitlements, error)
}

// NewCachedChecker creates a new CachedChecker with the given configuration.
func NewCachedChecker(redisClient *redis.Client, fetcher EntitlementFetcher, cfg Config) *CachedChecker {
	return &CachedChecker{
		redis:   redisClient,
		fetcher: fetcher,
		config:  cfg,
	}
}

// cacheKey returns the Redis key for an organization's entitlements.
func cacheKey(orgID string) string {
	return cacheKeyPrefix + orgID
}

// CheckEntitlement retrieves entitlements from cache or fetches from source.
func (c *CachedChecker) CheckEntitlement(ctx context.Context, orgID string) (*Entitlements, error) {
	if !c.config.Enabled {
		// Enforcement disabled, return permissive entitlements
		return &Entitlements{
			OrganizationID: orgID,
			Status:         "active",
			Features:       map[string]string{},
			QuotaLimits:    map[string]int64{},
		}, nil
	}

	// Try cache first
	cached, err := c.getFromCache(ctx, orgID)
	if err == nil && cached != nil {
		return cached, nil
	}

	// Cache miss - fetch from source
	entitlements, err := c.fetcher.FetchEntitlement(ctx, orgID)
	if err != nil {
		if c.config.FailOpen {
			// Return permissive entitlements on failure
			return &Entitlements{
				OrganizationID: orgID,
				Status:         "active",
				Features:       map[string]string{},
				QuotaLimits:    map[string]int64{},
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch entitlement: %w", err)
	}

	// Cache the result
	if entitlements != nil {
		entitlements.CachedAt = time.Now()
		_ = c.setInCache(ctx, orgID, entitlements)
	}

	return entitlements, nil
}

// HasFeature checks if the organization has access to a specific feature.
func (c *CachedChecker) HasFeature(ctx context.Context, orgID, featureKey string) (bool, error) {
	ent, err := c.CheckEntitlement(ctx, orgID)
	if err != nil {
		return false, err
	}
	if ent == nil || !ent.IsActive() {
		return false, nil
	}

	value := ent.GetFeature(featureKey)
	if value == "" {
		return false, nil
	}

	// Parse boolean features
	if value == "true" || value == "1" {
		return true, nil
	}
	// Numeric features are considered "has feature" if > 0
	if num, err := strconv.ParseInt(value, 10, 64); err == nil && num > 0 {
		return true, nil
	}

	return false, nil
}

// CheckQuota verifies if the organization is within quota for a given metric.
func (c *CachedChecker) CheckQuota(ctx context.Context, orgID, quotaKey string) (bool, int64, error) {
	ent, err := c.CheckEntitlement(ctx, orgID)
	if err != nil {
		return false, 0, err
	}
	if ent == nil || !ent.IsActive() {
		return false, 0, nil
	}

	allowed, remaining := ent.IsWithinQuota(quotaKey)
	return allowed, remaining, nil
}

// InvalidateCache removes the cached entitlement for an organization.
func (c *CachedChecker) InvalidateCache(ctx context.Context, orgID string) error {
	err := c.redis.Del(ctx, cacheKey(orgID)).Err()
	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	// Publish invalidation event for distributed cache invalidation
	_ = c.redis.Publish(ctx, invalidateChannel, orgID).Err()

	return nil
}

// getFromCache retrieves entitlements from Redis cache.
func (c *CachedChecker) getFromCache(ctx context.Context, orgID string) (*Entitlements, error) {
	data, err := c.redis.Get(ctx, cacheKey(orgID)).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var ent Entitlements
	if err := json.Unmarshal(data, &ent); err != nil {
		return nil, err
	}

	// Check if cache is stale
	if time.Since(ent.CachedAt) > c.config.CacheTTL {
		return nil, nil // Treat as cache miss
	}

	return &ent, nil
}

// setInCache stores entitlements in Redis cache.
func (c *CachedChecker) setInCache(ctx context.Context, orgID string, ent *Entitlements) error {
	data, err := json.Marshal(ent)
	if err != nil {
		return err
	}

	return c.redis.Set(ctx, cacheKey(orgID), data, c.config.CacheTTL).Err()
}

// StartInvalidationListener starts a goroutine that listens for cache invalidation events.
func (c *CachedChecker) StartInvalidationListener(ctx context.Context) {
	go func() {
		pubsub := c.redis.Subscribe(ctx, invalidateChannel)
		defer pubsub.Close()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-ch:
				if msg != nil {
					// Invalidate local cache (if any in-memory caching is added later)
					_ = c.redis.Del(ctx, cacheKey(msg.Payload)).Err()
				}
			}
		}
	}()
}
