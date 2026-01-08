package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
	"github.com/redis/go-redis/v9"
)

const (
	tokenBlacklistChannel = "identity:token:blacklist"
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

// --- Token Blacklist (Revocation) with Pub/Sub ---

// BlacklistToken adds a token to the blacklist with TTL synced to token expiry.
// Also publishes to a channel for distributed invalidation.
func (r *RedisRepository) BlacklistToken(ctx context.Context, jti string, ttl time.Duration) error {
	key := fmt.Sprintf("identity:revoked:%s", jti)

	// Use pipeline for atomic operation
	pipe := r.Client.Pipeline()
	pipe.Set(ctx, key, "1", ttl)
	pipe.Publish(ctx, tokenBlacklistChannel, jti)
	_, err := pipe.Exec(ctx)
	return err
}

// BlacklistTokenWithExpiry blacklists a token using the actual token expiry time.
func (r *RedisRepository) BlacklistTokenWithExpiry(ctx context.Context, jti string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}
	return r.BlacklistToken(ctx, jti, ttl)
}

// IsBlacklisted checks if a single token is blacklisted.
func (r *RedisRepository) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("identity:revoked:%s", jti)
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// AreBlacklisted performs batch blacklist check for multiple tokens.
// Returns a map of jti -> isBlacklisted.
func (r *RedisRepository) AreBlacklisted(ctx context.Context, jtis []string) (map[string]bool, error) {
	if len(jtis) == 0 {
		return map[string]bool{}, nil
	}

	keys := make([]string, len(jtis))
	for i, jti := range jtis {
		keys[i] = fmt.Sprintf("identity:revoked:%s", jti)
	}

	// Use MGET for batch check
	results, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	blacklisted := make(map[string]bool, len(jtis))
	for i, jti := range jtis {
		blacklisted[jti] = results[i] != nil
	}
	return blacklisted, nil
}

// BlacklistUserTokens revokes all tokens for a user by pattern.
func (r *RedisRepository) BlacklistUserTokens(ctx context.Context, userID string, ttl time.Duration) error {
	pattern := fmt.Sprintf("identity:token:%s:*", userID)

	// Scan for all tokens belonging to user
	iter := r.Client.Scan(ctx, 0, pattern, 100).Iterator()
	pipe := r.Client.Pipeline()
	count := 0

	for iter.Next(ctx) {
		key := iter.Val()
		pipe.Set(ctx, fmt.Sprintf("identity:revoked:%s", key), "1", ttl)
		pipe.Del(ctx, key)
		count++

		// Execute in batches of 100
		if count >= 100 {
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
			pipe = r.Client.Pipeline()
			count = 0
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if count > 0 {
		_, err := pipe.Exec(ctx)
		return err
	}

	return nil
}

// SubscribeToBlacklist returns a channel that receives blacklisted token JTIs.
// Used for distributed cache invalidation.
func (r *RedisRepository) SubscribeToBlacklist(ctx context.Context) <-chan string {
	pubsub := r.Client.Subscribe(ctx, tokenBlacklistChannel)
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-pubsub.Channel():
				if msg != nil {
					ch <- msg.Payload
				}
			}
		}
	}()

	return ch
}
