package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/MuhibNayem/Travio/server/services/queue/internal/domain"
	"github.com/redis/go-redis/v9"
)

var (
	ErrNotInQueue   = errors.New("user not in queue")
	ErrTokenInvalid = errors.New("token invalid or expired")
)

// QueueRepository manages queue state in Redis using atomic Lua scripts
type QueueRepository struct {
	client     *redis.Client
	scripts    map[string]string
	scriptSHAs map[string]string
}

// NewQueueRepository creates a new queue repository
func NewQueueRepository(redisAddr string) (*QueueRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	repo := &QueueRepository{
		client:     client,
		scripts:    make(map[string]string),
		scriptSHAs: make(map[string]string),
	}

	if err := repo.loadScripts(); err != nil {
		return nil, err
	}

	return repo, nil
}

func (r *QueueRepository) loadScripts() error {
	scripts := map[string]string{
		"enqueue": "internal/scripts/enqueue.lua",
		"admit":   "internal/scripts/admit.lua",
	}

	for name, path := range scripts {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read script %s: %w", name, err)
		}
		r.scripts[name] = string(content)
		sha, err := r.client.ScriptLoad(context.Background(), string(content)).Result()
		if err != nil {
			return fmt.Errorf("failed to load script %s into Redis: %w", name, err)
		}
		r.scriptSHAs[name] = sha
	}
	return nil
}

// Keys
func queueKey(eventID string) string { return fmt.Sprintf("queue:%s", eventID) }
func entryKey(eventID, userID string) string {
	return fmt.Sprintf("queue:%s:entry:%s", eventID, userID)
}
func statsKey(eventID string) string  { return fmt.Sprintf("queue:%s:stats", eventID) }
func configKey(eventID string) string { return fmt.Sprintf("queue:%s:config", eventID) }

// Join adds a user to the queue atomically
func (r *QueueRepository) Join(ctx context.Context, eventID, userID, sessionID string) (*domain.QueueEntry, error) {
	entry := &domain.QueueEntry{
		ID:        userID, // Simplified ID
		UserID:    userID,
		SessionID: sessionID,
		EventID:   eventID,
		JoinedAt:  time.Now(),
		Status:    domain.QueueStatusWaiting,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}

	entryJSON, _ := json.Marshal(entry)
	score := float64(time.Now().UnixNano())

	// Execute atomic enqueue script
	res, err := r.client.EvalSha(ctx, r.scriptSHAs["enqueue"],
		[]string{queueKey(eventID), entryKey(eventID, userID)}, // KEYS
		userID, score, 0, entryJSON, int(2*time.Hour.Seconds()), // ARGV
	).Result()

	if err != nil {
		return nil, fmt.Errorf("enqueue script execution failed: %w", err)
	}

	// Result: [status, rank], status 0=new, 1=existing
	resSlice := res.([]interface{})
	rank := int(resSlice[1].(int64))

	entry.Position = rank
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
		// If not in ZSet but entry exists, check status
		if entry.Status == domain.QueueStatusReady {
			return &entry, nil
		}
		entry.Status = domain.QueueStatusExpired
		return &entry, nil
	}
	entry.Position = int(rank) + 1

	// Estimate wait time (assume 10 users/minute admission rate - simplistic)
	entry.EstimatedWait = time.Duration(entry.Position) * 6 * time.Second

	return &entry, nil
}

// AdmitNext atomically admits the next batch of users and removes them from the waiting queue
func (r *QueueRepository) AdmitNext(ctx context.Context, eventID string, count int, tokenTTL time.Duration) ([]string, error) {
	// Execute atomic admission script
	res, err := r.client.EvalSha(ctx, r.scriptSHAs["admit"],
		[]string{queueKey(eventID), statsKey(eventID)}, // KEYS
		count, int(tokenTTL.Seconds()), eventID, // ARGV
	).Result()

	if err != nil {
		return nil, fmt.Errorf("admit script execution failed: %w", err)
	}

	// Result should be list of userIDs
	usersSlice := res.([]interface{})
	admitted := make([]string, len(usersSlice))
	for i, v := range usersSlice {
		admitted[i] = v.(string)
	}

	return admitted, nil
}

// Legacy cleanup
func (r *QueueRepository) Leave(ctx context.Context, eventID, userID string) error {
	r.client.ZRem(ctx, queueKey(eventID), userID)
	r.client.Del(ctx, entryKey(eventID, userID))
	return nil
}

func (r *QueueRepository) GetStats(ctx context.Context, eventID string) (*domain.QueueStats, error) {
	waiting, _ := r.client.ZCard(ctx, queueKey(eventID)).Result()
	admitted, _ := r.client.HGet(ctx, statsKey(eventID), "admitted").Result()
	// Safely handle nil/empty admitted count
	admittedCount := 0
	if admitted != "" {
		fmt.Sscanf(admitted, "%d", &admittedCount)
	}

	return &domain.QueueStats{
		EventID:       eventID,
		TotalWaiting:  int(waiting),
		TotalAdmitted: admittedCount,
		AvgWaitTime:   time.Duration(waiting) * 6 * time.Second / 10,
		AdmissionRate: 10,
		EstimatedWait: time.Duration(waiting) * 6 * time.Second,
	}, nil
}

func (r *QueueRepository) SetConfig(ctx context.Context, config *domain.AdmissionConfig) error {
	configJSON, _ := json.Marshal(config)
	return r.client.Set(ctx, configKey(config.EventID), configJSON, 0).Err()
}

func (r *QueueRepository) GetConfig(ctx context.Context, eventID string) (*domain.AdmissionConfig, error) {
	configJSON, err := r.client.Get(ctx, configKey(eventID)).Result()
	if err == redis.Nil {
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

func (r *QueueRepository) Close() error {
	return r.client.Close()
}
