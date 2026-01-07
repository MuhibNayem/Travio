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
	repo      *repository.QueueRepository
	workers   map[string]*AdmissionWorker
	workersMu sync.Mutex
}

// NewQueueService creates a new queue service
func NewQueueService(repo *repository.QueueRepository) *QueueService {
	return &QueueService{
		repo:    repo,
		workers: make(map[string]*AdmissionWorker),
	}
}

// JoinQueue adds a user to the virtual queue
func (s *QueueService) JoinQueue(ctx context.Context, eventID, userID, sessionID string) (*domain.QueueEntry, error) {
	// Check if queue is enabled for this event
	config, err := s.repo.GetConfig(ctx, eventID)
	if err != nil {
		return nil, err
	}

	if !config.QueueEnabled {
		// Queue disabled - immediately admit
		return &domain.QueueEntry{
			UserID:   userID,
			EventID:  eventID,
			Status:   domain.QueueStatusReady,
			Position: 0,
		}, nil
	}

	entry, err := s.repo.Join(ctx, eventID, userID, sessionID)
	if err != nil {
		return nil, err
	}

	// Ensure admission worker is running
	s.ensureWorker(eventID)

	logger.Info("user joined queue",
		"event_id", eventID,
		"user_id", userID,
		"position", entry.Position,
	)

	return entry, nil
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
func (s *QueueService) ValidateAdmission(ctx context.Context, token string) (bool, string, string, error) {
	userID, eventID, err := s.repo.ValidateToken(ctx, token)
	if err == repository.ErrTokenInvalid {
		return false, "", "", nil
	}
	if err != nil {
		return false, "", "", err
	}
	return true, userID, eventID, nil
}

// CompleteAdmission marks the token as used (user started purchase)
func (s *QueueService) CompleteAdmission(ctx context.Context, token string) error {
	return s.repo.ConsumeToken(ctx, token)
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
		// TODO: Send WebSocket notifications to admitted users
	}
}

// Stop stops the admission worker
func (w *AdmissionWorker) Stop() {
	w.cancel()
}
