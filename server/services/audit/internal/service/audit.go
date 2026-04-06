package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/domain"
	"github.com/MuhibNayem/Travio/server/services/audit/internal/repository"
	"github.com/google/uuid"
)

// AuditLogInput represents audit log input from HTTP requests
type AuditLogInput struct {
	AggregateID    string
	EventType      string
	ResourceType   string
	Action         string
	ActorID        string
	ActorType      string
	OrganizationID string
	IPAddress      string
	UserAgent      string
	Details        map[string]interface{}
	Status         string
	FailureReason  string
}

// AuditService handles audit log operations
type AuditService struct {
	repo *repository.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log creates a new audit log entry
func (s *AuditService) Log(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}
	if log.ActorType == "" {
		log.ActorType = domain.ActorSystem
	}
	if log.Status == "" {
		log.Status = domain.StatusSuccess
	}

	return s.repo.Insert(ctx, log)
}

// GetByActorID retrieves audit logs by actor ID
func (s *AuditService) GetByActorID(ctx context.Context, actorID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	return s.repo.GetByActorID(ctx, actorID, limit, offset)
}

// GetByResourceID retrieves audit logs by aggregate/resource ID
func (s *AuditService) GetByResourceID(ctx context.Context, resourceID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	return s.repo.GetByAggregateID(ctx, resourceID, limit, offset)
}

// GetByTimeRange retrieves audit logs within a time range
func (s *AuditService) GetByTimeRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*domain.AuditLog, int64, error) {
	return s.repo.GetByTimeRange(ctx, from, to, limit, offset)
}

// ArchiveOldLogs archives logs older than the retention period
func (s *AuditService) ArchiveOldLogs(ctx context.Context, retentionPeriod time.Duration) error {
	return s.repo.ArchiveOldLogs(ctx, retentionPeriod)
}

// DeleteArchivedLogs deletes logs that have been archived for too long
func (s *AuditService) DeleteArchivedLogs(ctx context.Context, olderThan time.Duration) error {
	return s.repo.DeleteArchivedLogs(ctx, olderThan)
}

// CountTotal returns the total number of non-archived audit logs
func (s *AuditService) CountTotal(ctx context.Context) (int64, error) {
	return s.repo.CountTotal(ctx)
}

// LogFromHTTP creates a new audit log entry from HTTP input
func (s *AuditService) LogFromHTTP(ctx context.Context, input *AuditLogInput) (string, error) {
	log := &domain.AuditLog{
		ID:             uuid.New().String(),
		AggregateID:    input.AggregateID,
		EventType:      input.EventType,
		ResourceType:   input.ResourceType,
		Action:         input.Action,
		ActorID:        input.ActorID,
		ActorType:      input.ActorType,
		OrganizationID: input.OrganizationID,
		IPAddress:      input.IPAddress,
		UserAgent:      input.UserAgent,
		Status:         input.Status,
		FailureReason:  input.FailureReason,
		CreatedAt:      time.Now(),
	}

	if log.ActorType == "" {
		log.ActorType = domain.ActorSystem
	}
	if log.Status == "" {
		log.Status = domain.StatusSuccess
	}

	// Marshal details
	if input.Details != nil {
		detailsJSON, err := json.Marshal(input.Details)
		if err != nil {
			return "", err
		}
		log.Details = detailsJSON
	}

	if err := s.repo.Insert(ctx, log); err != nil {
		return "", err
	}

	return log.ID, nil
}

// RetentionService handles automatic cleanup of old audit logs
type RetentionService struct {
	repo            *repository.AuditRepository
	retentionPeriod time.Duration
	stopCh          chan struct{}
}

// NewRetentionService creates a new retention service
func NewRetentionService(repo *repository.AuditRepository, retentionPeriod time.Duration) *RetentionService {
	return &RetentionService{
		repo:            repo,
		retentionPeriod: retentionPeriod,
		stopCh:          make(chan struct{}),
	}
}

// StartCleanup runs periodic cleanup of old audit logs
func (s *RetentionService) StartCleanup(ctx context.Context, interval time.Duration) {
	logger.Info("Starting audit log retention cleanup", "interval", interval, "retention_period", s.retentionPeriod)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping audit log retention cleanup")
			return
		case <-s.stopCh:
			logger.Info("Audit log retention cleanup stopped")
			return
		case <-ticker.C:
			if err := s.runCleanup(ctx); err != nil {
				logger.Error("Audit log cleanup failed", "error", err)
			} else {
				logger.Info("Audit log cleanup completed successfully")
			}
		}
	}
}

// Stop stops the cleanup service
func (s *RetentionService) Stop() {
	close(s.stopCh)
}

func (s *RetentionService) runCleanup(ctx context.Context) error {
	// Archive old logs
	if err := s.repo.ArchiveOldLogs(ctx, s.retentionPeriod); err != nil {
		return err
	}

	// Delete very old archived logs (keep for additional 30 days after archival)
	if err := s.repo.DeleteArchivedLogs(ctx, 30*24*time.Hour); err != nil {
		return err
	}

	return nil
}
