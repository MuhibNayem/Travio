package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	Client *redis.Client
}

func NewRedisRepository(addr string) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisRepository{Client: client}, nil
}

func (r *RedisRepository) Close() error {
	return r.Client.Close()
}

// --- User Profile Caching ---

func (r *RedisRepository) CacheUser(ctx context.Context, user *domain.User, ttl time.Duration) error {
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}
	key := fmt.Sprintf("identity:user:%s", user.ID)
	return r.Client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisRepository) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	key := fmt.Sprintf("identity:user:%s", userID)
	data, err := r.Client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, err
	}

	var user domain.User
	if err := json.Unmarshal(data, &user); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *RedisRepository) InvalidateUser(ctx context.Context, userID string) error {
	key := fmt.Sprintf("identity:user:%s", userID)
	return r.Client.Del(ctx, key).Err()
}

// --- Token Blacklist (Revocation) ---

func (r *RedisRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("identity:revoked:%s", jti)
	// We just need existence, so value "1" is fine
	return r.Client.Set(ctx, key, "1", ttl).Err()
}

func (r *RedisRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("identity:revoked:%s", jti)
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
