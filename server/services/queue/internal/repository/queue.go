package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/MuhibNayem/Travio/server/services/queue/internal/domain"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotInQueue   = errors.New("user not in queue")
	ErrTokenInvalid = errors.New("token invalid or expired")
)

// QueueRepository manages queue state in Redis
type QueueRepository struct {
	client *redis.Client
}

// NewQueueRepository creates a new queue repository
func NewQueueRepository(redisAddr string) *QueueRepository {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return &QueueRepository{client: client}
}

// Keys
func queueKey(eventID string) string { return fmt.Sprintf("queue:%s", eventID) }
func entryKey(eventID, userID string) string {
	return fmt.Sprintf("queue:%s:entry:%s", eventID, userID)
}
func tokenKey(token string) string    { return fmt.Sprintf("queue:token:%s", token) }
func statsKey(eventID string) string  { return fmt.Sprintf("queue:%s:stats", eventID) }
func configKey(eventID string) string { return fmt.Sprintf("queue:%s:config", eventID) }

// Join adds a user to the queue
func (r *QueueRepository) Join(ctx context.Context, eventID, userID, sessionID string) (*domain.QueueEntry, error) {
	token := uuid.New().String()
	score := float64(time.Now().UnixNano())

	// Add to sorted set (score = join time for FIFO ordering)
	if err := r.client.ZAdd(ctx, queueKey(eventID), redis.Z{
		Score:  score,
		Member: userID,
	}).Err(); err != nil {
		return nil, err
	}

	// Get position
	rank, _ := r.client.ZRank(ctx, queueKey(eventID), userID).Result()

	entry := &domain.QueueEntry{
		ID:        uuid.New().String(),
		UserID:    userID,
		SessionID: sessionID,
		EventID:   eventID,
		Position:  int(rank) + 1,
		Token:     token,
		JoinedAt:  time.Now(),
		Status:    domain.QueueStatusWaiting,
		ExpiresAt: time.Now().Add(2 * time.Hour), // Queue position expires in 2 hours
	}

	// Store entry details
	entryJSON, _ := json.Marshal(entry)
	r.client.Set(ctx, entryKey(eventID, userID), entryJSON, 2*time.Hour)

	return entry, nil
}

// GetPosition returns user's current position in queue
func (r *QueueRepository) GetPosition(ctx context.Context, eventID, userID string) (*domain.QueueEntry, error) {
	// Get entry details
	entryJSON, err := r.client.Get(ctx, entryKey(eventID, userID)).Result()
	if err == redis.Nil {
		return nil, ErrNotInQueue
	}
	if err != nil {
		return nil, err
	}

	var entry domain.QueueEntry
	json.Unmarshal([]byte(entryJSON), &entry)

	// Get updated position
	rank, err := r.client.ZRank(ctx, queueKey(eventID), userID).Result()
	if err == redis.Nil {
		entry.Status = domain.QueueStatusExpired
		return &entry, nil
	}
	entry.Position = int(rank) + 1

	// Estimate wait time (assume 10 users/minute admission rate)
	entry.EstimatedWait = time.Duration(entry.Position) * 6 * time.Second

	return &entry, nil
}

// Leave removes a user from the queue
func (r *QueueRepository) Leave(ctx context.Context, eventID, userID string) error {
	r.client.ZRem(ctx, queueKey(eventID), userID)
	r.client.Del(ctx, entryKey(eventID, userID))
	return nil
}

// AdmitNext admits the next batch of users
func (r *QueueRepository) AdmitNext(ctx context.Context, eventID string, count int, tokenTTL time.Duration) ([]string, error) {
	// Get next N users from front of queue
	users, err := r.client.ZRange(ctx, queueKey(eventID), 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}

	admitted := make([]string, 0, len(users))
	for _, userID := range users {
		// Generate admission token
		token := uuid.New().String()
		tokenData := map[string]string{
			"user_id":  userID,
			"event_id": eventID,
		}
		tokenJSON, _ := json.Marshal(tokenData)
		r.client.Set(ctx, tokenKey(token), tokenJSON, tokenTTL)

		// Update entry status
		entryJSON, _ := r.client.Get(ctx, entryKey(eventID, userID)).Result()
		var entry domain.QueueEntry
		json.Unmarshal([]byte(entryJSON), &entry)
		entry.Status = domain.QueueStatusReady
		entry.Token = token
		entry.ExpiresAt = time.Now().Add(tokenTTL)
		updatedJSON, _ := json.Marshal(entry)
		r.client.Set(ctx, entryKey(eventID, userID), updatedJSON, tokenTTL)

		// Remove from waiting queue
		r.client.ZRem(ctx, queueKey(eventID), userID)

		admitted = append(admitted, userID)
	}

	// Update stats
	r.client.HIncrBy(ctx, statsKey(eventID), "admitted", int64(len(admitted)))

	return admitted, nil
}

// ValidateToken checks if an admission token is valid
func (r *QueueRepository) ValidateToken(ctx context.Context, token string) (string, string, error) {
	tokenJSON, err := r.client.Get(ctx, tokenKey(token)).Result()
	if err == redis.Nil {
		return "", "", ErrTokenInvalid
	}
	if err != nil {
		return "", "", err
	}

	var data map[string]string
	json.Unmarshal([]byte(tokenJSON), &data)

	return data["user_id"], data["event_id"], nil
}

// ConsumeToken marks a token as used
func (r *QueueRepository) ConsumeToken(ctx context.Context, token string) error {
	return r.client.Del(ctx, tokenKey(token)).Err()
}

// GetStats returns queue statistics
func (r *QueueRepository) GetStats(ctx context.Context, eventID string) (*domain.QueueStats, error) {
	waiting, _ := r.client.ZCard(ctx, queueKey(eventID)).Result()
	admitted, _ := r.client.HGet(ctx, statsKey(eventID), "admitted").Result()
	admittedCount, _ := strconv.Atoi(admitted)

	return &domain.QueueStats{
		EventID:       eventID,
		TotalWaiting:  int(waiting),
		TotalAdmitted: admittedCount,
		AvgWaitTime:   time.Duration(waiting) * 6 * time.Second / 10, // Estimate
		AdmissionRate: 10,                                            // Users per minute
		EstimatedWait: time.Duration(waiting) * 6 * time.Second,
	}, nil
}

// SetConfig stores queue configuration
func (r *QueueRepository) SetConfig(ctx context.Context, config *domain.AdmissionConfig) error {
	configJSON, _ := json.Marshal(config)
	return r.client.Set(ctx, configKey(config.EventID), configJSON, 0).Err()
}

// GetConfig retrieves queue configuration
func (r *QueueRepository) GetConfig(ctx context.Context, eventID string) (*domain.AdmissionConfig, error) {
	configJSON, err := r.client.Get(ctx, configKey(eventID)).Result()
	if err == redis.Nil {
		// Return default config
		return &domain.AdmissionConfig{
			EventID:            eventID,
			MaxConcurrent:      100,
			AdmissionBatchSize: 10,
			AdmissionInterval:  time.Minute,
			TokenTTL:           10 * time.Minute,
			QueueEnabled:       true,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var config domain.AdmissionConfig
	json.Unmarshal([]byte(configJSON), &config)
	return &config, nil
}

// Close closes the Redis connection
func (r *QueueRepository) Close() error {
	return r.client.Close()
}
