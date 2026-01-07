package service

import (
	"context"
	"sync"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/queue/internal/repository"
)

// QueueService manages virtual waiting room
type QueueService struct {
	repo         *repository.QueueRepository
	tokenManager *domain.TokenManager
	workers      map[string]*AdmissionWorker
	workersMu    sync.Mutex
}

// NewQueueService creates a new queue service
func NewQueueService(repo *repository.QueueRepository, tokenSecret string) *QueueService {
	return &QueueService{
		repo:         repo,
		tokenManager: domain.NewTokenManager(tokenSecret),
		workers:      make(map[string]*AdmissionWorker),
	}
}

// JoinQueue adds a user to the virtual queue
func (s *QueueService) JoinQueue(ctx context.Context, eventID, userID, sessionID string) (*domain.QueueEntry, error) {
	// Add to queue via repository (atomic Lua script)
	return s.repo.Join(ctx, eventID, userID, sessionID)
}

// GetPosition returns user's current queue position
func (s *QueueService) GetPosition(ctx context.Context, eventID, userID string) (*domain.QueueEntry, error) {
	return s.repo.GetPosition(ctx, eventID, userID)
}

// LeaveQueue removes a user from the queue
func (s *QueueService) LeaveQueue(ctx context.Context, eventID, userID string) error {
	return s.repo.Leave(ctx, eventID, userID)
}

// ValidateAdmission checks if a user has a valid admission token
// DEPRECATED: Use ValidateToken for stateless validation
func (s *QueueService) ValidateAdmission(ctx context.Context, token string) (bool, string, string, error) {
	return s.ValidateToken(ctx, token)
}

// CompleteAdmission marks the token as used
// For stateless tokens, this is a no-op or could verify against a bloom filter
func (s *QueueService) CompleteAdmission(ctx context.Context, token string) error {
	// Stateless tokens don't need to be deleted
	return nil
}

// ProcessAdmission triggers the admission of the next batch of users.
// This is called by the AdmissionWorker on a configured interval.
func (s *QueueService) ProcessAdmission(ctx context.Context, eventID string) (int, error) {
	config, err := s.repo.GetConfig(ctx, eventID)
	if err != nil {
		return 0, err
	}

	if !config.QueueEnabled {
		return 0, nil
	}

	// Calculate how many to admit
	stats, err := s.repo.GetStats(ctx, eventID)
	if err != nil {
		return 0, err
	}

	if stats.TotalWaiting == 0 {
		return 0, nil
	}

	// Admit users
	userIDs, err := s.repo.AdmitNext(ctx, eventID, config.AdmissionBatchSize, config.TokenTTL)
	if err != nil {
		return 0, err
	}

	return len(userIDs), nil
}

// GenerateToken generates a stateless JWT for an admitted user
// This replaces the old stateful token validation
func (s *QueueService) GenerateToken(userID, eventID string, ttl time.Duration) (string, error) {
	return s.tokenManager.GenerateToken(userID, eventID, ttl)
}

// ValidateToken validates a stateless JWT token
func (s *QueueService) ValidateToken(ctx context.Context, token string) (bool, string, string, error) {
	claims, err := s.tokenManager.ValidateToken(token)
	if err != nil {
		return false, "", "", nil // Invalid token
	}
	return true, claims.UserID, claims.EventID, nil
}

// GetStats returns queue statistics
func (s *QueueService) GetStats(ctx context.Context, eventID string) (*domain.QueueStats, error) {
	return s.repo.GetStats(ctx, eventID)
}

// ConfigureQueue sets queue configuration for an event
func (s *QueueService) ConfigureQueue(ctx context.Context, config *domain.AdmissionConfig) error {
	if err := s.repo.SetConfig(ctx, config); err != nil {
		return err
	}

	// Restart worker with new config if exists
	s.workersMu.Lock()
	if worker, exists := s.workers[config.EventID]; exists {
		worker.Stop()
		delete(s.workers, config.EventID)
	}
	s.workersMu.Unlock()

	if config.QueueEnabled {
		s.ensureWorker(config.EventID)
	}

	return nil
}

// ensureWorker starts an admission worker for the event if not running
func (s *QueueService) ensureWorker(eventID string) {
	s.workersMu.Lock()
	defer s.workersMu.Unlock()

	if _, exists := s.workers[eventID]; exists {
		return
	}

	worker := NewAdmissionWorker(s.repo, eventID)
	s.workers[eventID] = worker
	go worker.Start()
}

// AdmissionWorker periodically admits users from the queue
type AdmissionWorker struct {
	repo    *repository.QueueRepository
	eventID string
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewAdmissionWorker creates a new admission worker
func NewAdmissionWorker(repo *repository.QueueRepository, eventID string) *AdmissionWorker {
	ctx, cancel := context.WithCancel(context.Background())
	return &AdmissionWorker{
		repo:    repo,
		eventID: eventID,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start begins the admission loop
func (w *AdmissionWorker) Start() {
	config, _ := w.repo.GetConfig(w.ctx, w.eventID)
	if config == nil {
		config = &domain.AdmissionConfig{
			AdmissionBatchSize: 10,
			AdmissionInterval:  time.Minute,
			TokenTTL:           10 * time.Minute,
		}
	}

	ticker := time.NewTicker(config.AdmissionInterval)
	defer ticker.Stop()

	logger.Info("admission worker started", "event_id", w.eventID)

	for {
		select {
		case <-w.ctx.Done():
			logger.Info("admission worker stopped", "event_id", w.eventID)
			return
		case <-ticker.C:
			w.admitBatch(config)
		}
	}
}

// admitBatch admits a batch of users
func (w *AdmissionWorker) admitBatch(config *domain.AdmissionConfig) {
	admitted, err := w.repo.AdmitNext(w.ctx, w.eventID, config.AdmissionBatchSize, config.TokenTTL)
	if err != nil {
		logger.Error("admission batch failed", "event_id", w.eventID, "error", err)
		return
	}

	if len(admitted) > 0 {
		logger.Info("users admitted from queue",
			"event_id", w.eventID,
			"count", len(admitted),
		)
	}
}

// Stop stops the admission worker
func (w *AdmissionWorker) Stop() {
	w.cancel()
}
