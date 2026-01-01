package queue

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Virtual Queue System for high-demand sales
// Prevents thundering herd by issuing queue positions

var (
	ErrQueueFull       = errors.New("queue is full")
	ErrInvalidToken    = errors.New("invalid queue token")
	ErrTokenExpired    = errors.New("queue token has expired")
	ErrNotYourTurn     = errors.New("not your turn yet")
	ErrEventNotInQueue = errors.New("event not in queue mode")
)

// Config for virtual queue
type Config struct {
	// Maximum users allowed in queue per event
	MaxQueueSize int
	// How many users to let through per minute
	ThroughputPerMinute int
	// Queue token validity duration
	TokenTTL time.Duration
	// Secret for signing tokens
	SigningSecret []byte
}

// QueueToken represents a signed queue position token
type QueueToken struct {
	ID        string    `json:"id"`
	EventID   string    `json:"event_id"`
	UserID    string    `json:"user_id"`
	Position  int64     `json:"position"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Queue manages the virtual waiting room
type Queue struct {
	client *redis.Client
	config Config
}

// NewQueue creates a new virtual queue manager
func NewQueue(client *redis.Client, config Config) *Queue {
	if config.MaxQueueSize == 0 {
		config.MaxQueueSize = 100000
	}
	if config.ThroughputPerMinute == 0 {
		config.ThroughputPerMinute = 1000
	}
	if config.TokenTTL == 0 {
		config.TokenTTL = 15 * time.Minute
	}
	return &Queue{
		client: client,
		config: config,
	}
}

// EnableQueueMode activates queue for a specific event (e.g., during Eid sale)
func (q *Queue) EnableQueueMode(ctx context.Context, eventID string, startsAt time.Time) error {
	key := fmt.Sprintf("queue:enabled:%s", eventID)
	return q.client.Set(ctx, key, startsAt.Unix(), 24*time.Hour).Err()
}

// DisableQueueMode deactivates queue for an event
func (q *Queue) DisableQueueMode(ctx context.Context, eventID string) error {
	key := fmt.Sprintf("queue:enabled:%s", eventID)
	return q.client.Del(ctx, key).Err()
}

// IsQueueEnabled checks if queue mode is active for an event
func (q *Queue) IsQueueEnabled(ctx context.Context, eventID string) (bool, error) {
	key := fmt.Sprintf("queue:enabled:%s", eventID)
	exists, err := q.client.Exists(ctx, key).Result()
	return exists > 0, err
}

// JoinQueue puts a user in the queue and returns their position token
func (q *Queue) JoinQueue(ctx context.Context, eventID, userID string) (*QueueToken, error) {
	// Check if queue is enabled
	enabled, err := q.IsQueueEnabled(ctx, eventID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		return nil, ErrEventNotInQueue
	}

	queueKey := fmt.Sprintf("queue:positions:%s", eventID)

	// Check queue size
	size, err := q.client.ZCard(ctx, queueKey).Result()
	if err != nil {
		return nil, err
	}
	if size >= int64(q.config.MaxQueueSize) {
		return nil, ErrQueueFull
	}

	// Check if user already in queue
	existingScore, err := q.client.ZScore(ctx, queueKey, userID).Result()
	if err == nil && existingScore > 0 {
		// User already in queue, return existing position
		rank, _ := q.client.ZRank(ctx, queueKey, userID).Result()
		return q.generateToken(eventID, userID, rank+1)
	}

	// Add user to queue
	now := time.Now()
	_, err = q.client.ZAdd(ctx, queueKey, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: userID,
	}).Result()
	if err != nil {
		return nil, err
	}

	// Get user's position
	rank, err := q.client.ZRank(ctx, queueKey, userID).Result()
	if err != nil {
		return nil, err
	}

	// Set queue expiration (24 hours)
	q.client.Expire(ctx, queueKey, 24*time.Hour)

	return q.generateToken(eventID, userID, rank+1)
}

// ValidateToken checks if a queue token is valid and if it's the user's turn
func (q *Queue) ValidateToken(ctx context.Context, tokenString string) (*QueueToken, error) {
	token, err := q.parseAndVerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Check expiration
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	// Check if queue is still enabled
	enabled, err := q.IsQueueEnabled(ctx, token.EventID)
	if err != nil {
		return nil, err
	}
	if !enabled {
		// Queue disabled = everyone allowed
		return token, nil
	}

	// Calculate current serving position
	currentServing, err := q.GetCurrentServing(ctx, token.EventID)
	if err != nil {
		return nil, err
	}

	if token.Position > currentServing {
		return nil, ErrNotYourTurn
	}

	return token, nil
}

// GetCurrentServing returns which position is currently being served
func (q *Queue) GetCurrentServing(ctx context.Context, eventID string) (int64, error) {
	key := fmt.Sprintf("queue:serving:%s", eventID)
	val, err := q.client.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 1, nil // Start from position 1
	}
	return val, err
}

// AdvanceQueue moves the serving position forward (called by background job)
func (q *Queue) AdvanceQueue(ctx context.Context, eventID string, count int64) error {
	key := fmt.Sprintf("queue:serving:%s", eventID)
	return q.client.IncrBy(ctx, key, count).Err()
}

// GetQueueStatus returns current queue status for UI
type QueueStatus struct {
	Position       int64         `json:"position"`
	TotalInQueue   int64         `json:"total_in_queue"`
	EstimatedWait  time.Duration `json:"estimated_wait"`
	CurrentServing int64         `json:"current_serving"`
}

func (q *Queue) GetQueueStatus(ctx context.Context, eventID, userID string) (*QueueStatus, error) {
	queueKey := fmt.Sprintf("queue:positions:%s", eventID)

	// Get user's position
	rank, err := q.client.ZRank(ctx, queueKey, userID).Result()
	if err == redis.Nil {
		return nil, errors.New("user not in queue")
	}
	if err != nil {
		return nil, err
	}

	// Get total queue size
	total, err := q.client.ZCard(ctx, queueKey).Result()
	if err != nil {
		return nil, err
	}

	// Get current serving position
	currentServing, err := q.GetCurrentServing(ctx, eventID)
	if err != nil {
		return nil, err
	}

	position := rank + 1
	waitPosition := position - currentServing
	if waitPosition < 0 {
		waitPosition = 0
	}

	// Estimate wait time
	estimatedWait := time.Duration(waitPosition) * time.Minute / time.Duration(q.config.ThroughputPerMinute)

	return &QueueStatus{
		Position:       position,
		TotalInQueue:   total,
		EstimatedWait:  estimatedWait,
		CurrentServing: currentServing,
	}, nil
}

// --- Token Signing ---

func (q *Queue) generateToken(eventID, userID string, position int64) (*QueueToken, error) {
	now := time.Now()
	token := &QueueToken{
		ID:        uuid.New().String(),
		EventID:   eventID,
		UserID:    userID,
		Position:  position,
		IssuedAt:  now,
		ExpiresAt: now.Add(q.config.TokenTTL),
	}
	return token, nil
}

func (q *Queue) SignToken(token *QueueToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", err
	}

	// Create HMAC signature
	h := hmac.New(sha256.New, q.config.SigningSecret)
	h.Write(data)
	sig := h.Sum(nil)

	// Combine data and signature
	combined := append(data, sig...)
	return base64.URLEncoding.EncodeToString(combined), nil
}

func (q *Queue) parseAndVerifyToken(tokenString string) (*QueueToken, error) {
	combined, err := base64.URLEncoding.DecodeString(tokenString)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if len(combined) < sha256.Size {
		return nil, ErrInvalidToken
	}

	data := combined[:len(combined)-sha256.Size]
	sig := combined[len(combined)-sha256.Size:]

	// Verify HMAC
	h := hmac.New(sha256.New, q.config.SigningSecret)
	h.Write(data)
	expectedSig := h.Sum(nil)

	if !hmac.Equal(sig, expectedSig) {
		return nil, ErrInvalidToken
	}

	var token QueueToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, ErrInvalidToken
	}

	return &token, nil
}
