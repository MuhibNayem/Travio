package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/inventory/internal/domain"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// AcquireSeatLock attempts to acquire a distributed lock for a specific seat on a segment
// Returns true if lock acquired, false if already locked
func (r *RedisRepository) AcquireSeatLock(ctx context.Context, tripID, seatID string, segmentIndex int, userID string, ttl time.Duration) (bool, error) {
	key := fmt.Sprintf("inventory:lock:%s:%d:%s", tripID, segmentIndex, seatID)
	// SET key value NX EX ttl
	success, err := r.client.SetNX(ctx, key, userID, ttl).Result()
	if err != nil {
		return false, err
	}
	return success, nil
}

// ReleaseSeatLock releases the lock only if it belongs to the user
func (r *RedisRepository) ReleaseSeatLock(ctx context.Context, tripID, seatID string, segmentIndex int, userID string) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`
	key := fmt.Sprintf("inventory:lock:%s:%d:%s", tripID, segmentIndex, seatID)
	return r.client.Eval(ctx, script, []string{key}, userID).Err()
}

// CacheSeatMap caches the entire seat availability for a trip (short TTL)
// Maps: segment_index -> []SeatInventory
func (r *RedisRepository) CacheSeatMap(ctx context.Context, tripID string, seats []domain.SeatInventory, ttl time.Duration) error {
	key := fmt.Sprintf("inventory:cache:seatmap:%s", tripID)
	data, err := json.Marshal(seats)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

// GetCachedSeatMap retrieves the cached seat map
func (r *RedisRepository) GetCachedSeatMap(ctx context.Context, tripID string) ([]domain.SeatInventory, error) {
	key := fmt.Sprintf("inventory:cache:seatmap:%s", tripID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, err
	}

	var seats []domain.SeatInventory
	if err := json.Unmarshal(data, &seats); err != nil {
		return nil, err
	}
	return seats, nil
}

// InvalidateSeatMap explicitly deletes the cache (e.g., after booking)
func (r *RedisRepository) InvalidateSeatMap(ctx context.Context, tripID string) error {
	key := fmt.Sprintf("inventory:cache:seatmap:%s", tripID)
	return r.client.Del(ctx, key).Err()
}
