// Package cache provides a multi-level caching solution for the Catalog service.
// L1: In-memory LRU cache for ultra-low latency (sub-millisecond).
// L2: Redis cache for distributed consistency.
package cache

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/redis/go-redis/v9"
)

// CacheEntry holds a cached value with its expiration time.
type CacheEntry struct {
	Value     []byte
	ExpiresAt time.Time
}

// IsExpired returns true if the entry has expired.
func (e *CacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpiresAt)
}

// MultiLevelCache implements L1 (in-memory) + L2 (Redis) caching.
type MultiLevelCache struct {
	l1     *lru.Cache[string, *CacheEntry]
	l2     *redis.Client
	l1TTL  time.Duration
	l2TTL  time.Duration
	mu     sync.RWMutex
	prefix string
}

// Config holds configuration for the multi-level cache.
type Config struct {
	// L1MaxItems is the maximum number of items in the L1 cache.
	L1MaxItems int
	// L1TTL is the time-to-live for L1 cache entries.
	L1TTL time.Duration
	// L2TTL is the time-to-live for L2 (Redis) cache entries.
	L2TTL time.Duration
	// Prefix is the key prefix for Redis.
	Prefix string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		L1MaxItems: 10000,
		L1TTL:      1 * time.Minute,
		L2TTL:      5 * time.Minute,
		Prefix:     "catalog:",
	}
}

// NewMultiLevelCache creates a new multi-level cache.
func NewMultiLevelCache(redisClient *redis.Client, cfg Config) (*MultiLevelCache, error) {
	l1Cache, err := lru.New[string, *CacheEntry](cfg.L1MaxItems)
	if err != nil {
		return nil, err
	}

	return &MultiLevelCache{
		l1:     l1Cache,
		l2:     redisClient,
		l1TTL:  cfg.L1TTL,
		l2TTL:  cfg.L2TTL,
		prefix: cfg.Prefix,
	}, nil
}

// Get retrieves a value from the cache, checking L1 first, then L2.
// Returns nil if not found or expired.
func (c *MultiLevelCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := c.prefix + key

	// Check L1 first (in-memory)
	c.mu.RLock()
	if entry, ok := c.l1.Get(fullKey); ok && !entry.IsExpired() {
		c.mu.RUnlock()
		return entry.Value, nil
	}
	c.mu.RUnlock()

	// Check L2 (Redis)
	data, err := c.l2.Get(ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	// Populate L1 from L2
	c.mu.Lock()
	c.l1.Add(fullKey, &CacheEntry{
		Value:     data,
		ExpiresAt: time.Now().Add(c.l1TTL),
	})
	c.mu.Unlock()

	return data, nil
}

// GetAs retrieves and unmarshals a value from the cache.
func (c *MultiLevelCache) GetAs(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key)
	if err != nil {
		return err
	}
	if data == nil {
		return nil // Cache miss
	}
	return json.Unmarshal(data, dest)
}

// Set stores a value in both L1 and L2 caches.
func (c *MultiLevelCache) Set(ctx context.Context, key string, value []byte) error {
	fullKey := c.prefix + key

	// Set in L1
	c.mu.Lock()
	c.l1.Add(fullKey, &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.l1TTL),
	})
	c.mu.Unlock()

	// Set in L2 (Redis)
	return c.l2.Set(ctx, fullKey, value, c.l2TTL).Err()
}

// SetAs marshals and stores a value in both caches.
func (c *MultiLevelCache) SetAs(ctx context.Context, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Set(ctx, key, data)
}

// Delete removes a value from both L1 and L2 caches.
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
	fullKey := c.prefix + key

	// Remove from L1
	c.mu.Lock()
	c.l1.Remove(fullKey)
	c.mu.Unlock()

	// Remove from L2
	return c.l2.Del(ctx, fullKey).Err()
}

// InvalidateByPrefix removes all entries with the given prefix from both caches.
func (c *MultiLevelCache) InvalidateByPrefix(ctx context.Context, prefix string) error {
	fullPrefix := c.prefix + prefix

	// Clear matching L1 entries
	c.mu.Lock()
	keys := c.l1.Keys()
	for _, k := range keys {
		if len(k) >= len(fullPrefix) && k[:len(fullPrefix)] == fullPrefix {
			c.l1.Remove(k)
		}
	}
	c.mu.Unlock()

	// Clear matching L2 entries using SCAN
	iter := c.l2.Scan(ctx, 0, fullPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		c.l2.Del(ctx, iter.Val())
	}

	return iter.Err()
}

// InvalidateByPattern removes all entries matching a glob pattern from both caches.
// Pattern supports * as wildcard (e.g., "trip:*:123" matches any org with trip ID 123).
func (c *MultiLevelCache) InvalidateByPattern(ctx context.Context, pattern string) error {
	fullPattern := c.prefix + pattern

	// Clear matching L1 entries using glob matching
	c.mu.Lock()
	keys := c.l1.Keys()
	for _, k := range keys {
		if matchGlob(fullPattern, k) {
			c.l1.Remove(k)
		}
	}
	c.mu.Unlock()

	// Clear matching L2 entries using Redis SCAN with pattern
	iter := c.l2.Scan(ctx, 0, fullPattern, 100).Iterator()
	for iter.Next(ctx) {
		c.l2.Del(ctx, iter.Val())
	}

	return iter.Err()
}

// matchGlob performs simple glob pattern matching with * as wildcard.
func matchGlob(pattern, s string) bool {
	for len(pattern) > 0 {
		switch pattern[0] {
		case '*':
			if len(pattern) == 1 {
				return true
			}
			for i := 0; i <= len(s); i++ {
				if matchGlob(pattern[1:], s[i:]) {
					return true
				}
			}
			return false
		default:
			if len(s) == 0 || s[0] != pattern[0] {
				return false
			}
			pattern = pattern[1:]
			s = s[1:]
		}
	}
	return len(s) == 0
}

// Stats returns cache statistics.
type Stats struct {
	L1Len    int
	L1MaxLen int
}

// Stats returns current cache statistics.
func (c *MultiLevelCache) Stats() Stats {
	return Stats{
		L1Len:    c.l1.Len(),
		L1MaxLen: 10000, // Use configured max instead of Cap()
	}
}
