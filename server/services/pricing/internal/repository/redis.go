package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) CacheRules(ctx context.Context, orgID string, rules []*PricingRule, ttl time.Duration) error {
	key := fmt.Sprintf("pricing:rules:%s", orgID)
	data, err := json.Marshal(rules)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRepository) GetCachedRules(ctx context.Context, orgID string) ([]*PricingRule, error) {
	key := fmt.Sprintf("pricing:rules:%s", orgID)
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var rules []*PricingRule
	if err := json.Unmarshal([]byte(val), &rules); err != nil {
		return nil, err
	}
	return rules, nil
}

func (r *RedisRepository) InvalidateRules(ctx context.Context, orgID string) error {
	key := fmt.Sprintf("pricing:rules:%s", orgID)
	return r.client.Del(ctx, key).Err()
}
