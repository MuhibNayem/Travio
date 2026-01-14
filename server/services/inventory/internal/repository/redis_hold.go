package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/inventory/internal/domain"
	"github.com/redis/go-redis/v9"
)

// HoldRepository manages seat holds in Redis for fast TTL-based expiration
type HoldRepository struct {
	client *redis.Client
}

func NewHoldRepository(client *redis.Client) *HoldRepository {
	return &HoldRepository{client: client}
}

// CreateHold stores a hold record with automatic expiration
func (r *HoldRepository) CreateHold(ctx context.Context, orgID string, hold *domain.SeatHold) error {
	data, err := json.Marshal(hold)
	if err != nil {
		return err
	}

	// Key by hold_id
	holdKey := fmt.Sprintf("hold:%s:%s", orgID, hold.HoldID)
	ttl := time.Until(hold.ExpiresAt)

	pipe := r.client.Pipeline()

	// Store hold data
	pipe.Set(ctx, holdKey, data, ttl)

	// Index by user for lookup
	userHoldsKey := fmt.Sprintf("user_holds:%s:%s", orgID, hold.UserID)
	pipe.SAdd(ctx, userHoldsKey, hold.HoldID)
	pipe.Expire(ctx, userHoldsKey, 24*time.Hour)

	// Index by trip for admin lookup
	tripHoldsKey := fmt.Sprintf("trip_holds:%s:%s", orgID, hold.TripID)
	pipe.SAdd(ctx, tripHoldsKey, hold.HoldID)
	pipe.Expire(ctx, tripHoldsKey, 24*time.Hour)

	_, err = pipe.Exec(ctx)
	return err
}

// GetHold retrieves a hold by ID
func (r *HoldRepository) GetHold(ctx context.Context, orgID, holdID string) (*domain.SeatHold, error) {
	key := fmt.Sprintf("hold:%s:%s", orgID, holdID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrHoldNotFound
		}
		return nil, err
	}

	var hold domain.SeatHold
	if err := json.Unmarshal(data, &hold); err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(hold.ExpiresAt) {
		return nil, domain.ErrHoldExpired
	}

	return &hold, nil
}

// UpdateHoldStatus updates the status of a hold
func (r *HoldRepository) UpdateHoldStatus(ctx context.Context, orgID, holdID, status string) error {
	hold, err := r.GetHold(ctx, orgID, holdID)
	if err != nil {
		return err
	}

	hold.Status = status
	data, _ := json.Marshal(hold)

	key := fmt.Sprintf("hold:%s:%s", orgID, holdID)
	ttl := time.Until(hold.ExpiresAt)
	if ttl < 0 {
		ttl = time.Minute // Grace period for cleanup
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// DeleteHold removes a hold
func (r *HoldRepository) DeleteHold(ctx context.Context, orgID, holdID, userID, tripID string) error {
	pipe := r.client.Pipeline()

	holdKey := fmt.Sprintf("hold:%s:%s", orgID, holdID)
	userHoldsKey := fmt.Sprintf("user_holds:%s:%s", orgID, userID)
	tripHoldsKey := fmt.Sprintf("trip_holds:%s:%s", orgID, tripID)

	pipe.Del(ctx, holdKey)
	pipe.SRem(ctx, userHoldsKey, holdID)
	pipe.SRem(ctx, tripHoldsKey, holdID)

	_, err := pipe.Exec(ctx)
	return err
}

// GetUserHolds returns all active holds for a user
func (r *HoldRepository) GetUserHolds(ctx context.Context, orgID, userID string) ([]*domain.SeatHold, error) {
	key := fmt.Sprintf("user_holds:%s:%s", orgID, userID)
	holdIDs, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var holds []*domain.SeatHold
	for _, holdID := range holdIDs {
		hold, err := r.GetHold(ctx, orgID, holdID)
		if err != nil {
			continue // Skip expired/invalid holds
		}
		holds = append(holds, hold)
	}

	return holds, nil
}

// CountUserActiveHolds returns the number of active holds for a user
func (r *HoldRepository) CountUserActiveHolds(ctx context.Context, orgID, userID string) (int, error) {
	holds, err := r.GetUserHolds(ctx, orgID, userID)
	if err != nil {
		return 0, err
	}
	return len(holds), nil
}
