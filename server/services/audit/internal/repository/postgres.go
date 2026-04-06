package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/MuhibNayem/Travio/server/services/audit/internal/domain"
)

// AuditRepository handles audit log persistence
type AuditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateTable creates the audit_logs table (write-only, append-only)
func (r *AuditRepository) CreateTable(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS audit_logs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		aggregate_id VARCHAR(255) NOT NULL,
		event_type VARCHAR(100) NOT NULL,
		resource_type VARCHAR(100),
		action VARCHAR(100) NOT NULL,
		actor_id VARCHAR(255),
		actor_type VARCHAR(50) NOT NULL,
		organization_id VARCHAR(255),
		ip_address VARCHAR(45),
		user_agent TEXT,
		details JSONB,
		status VARCHAR(20) NOT NULL DEFAULT 'success',
		failure_reason TEXT,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		archived BOOLEAN DEFAULT FALSE,
		archived_at TIMESTAMP WITH TIME ZONE
	);
	
	CREATE INDEX IF NOT EXISTS idx_audit_logs_actor ON audit_logs(actor_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_aggregate ON audit_logs(aggregate_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_org ON audit_logs(organization_id);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_created ON audit_logs(created_at);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_event ON audit_logs(event_type);
	CREATE INDEX IF NOT EXISTS idx_audit_logs_archived ON audit_logs(archived) WHERE archived = FALSE;
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// Insert inserts a new audit log entry (append-only, no updates or deletes)
func (r *AuditRepository) Insert(ctx context.Context, log *domain.AuditLog) error {
	detailsJSON, _ := json.Marshal(log.Details)
	
	query := `
	INSERT INTO audit_logs (id, aggregate_id, event_type, resource_type, action, actor_id, actor_type,
		organization_id, ip_address, user_agent, details, status, failure_reason, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID,
		log.AggregateID,
		log.EventType,
		log.ResourceType,
		log.Action,
		log.ActorID,
		log.ActorType,
		log.OrganizationID,
		log.IPAddress,
		log.UserAgent,
		detailsJSON,
		log.Status,
		log.FailureReason,
		log.CreatedAt,
	)
	return err
}

// GetByActorID retrieves audit logs by actor ID with pagination
func (r *AuditRepository) GetByActorID(ctx context.Context, actorID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	query := `
	SELECT id, aggregate_id, event_type, resource_type, action, actor_id, actor_type,
		organization_id, ip_address, user_agent, details, status, failure_reason, created_at, archived
	FROM audit_logs
	WHERE actor_id = $1 AND archived = FALSE
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`
	return r.queryLogs(ctx, query, actorID, limit, offset)
}

// GetByAggregateID retrieves audit logs by aggregate ID
func (r *AuditRepository) GetByAggregateID(ctx context.Context, aggregateID string, limit, offset int) ([]*domain.AuditLog, int64, error) {
	query := `
	SELECT id, aggregate_id, event_type, resource_type, action, actor_id, actor_type,
		organization_id, ip_address, user_agent, details, status, failure_reason, created_at, archived
	FROM audit_logs
	WHERE aggregate_id = $1 AND archived = FALSE
	ORDER BY created_at DESC
	LIMIT $2 OFFSET $3
	`
	return r.queryLogs(ctx, query, aggregateID, limit, offset)
}

// GetByTimeRange retrieves audit logs within a time range
func (r *AuditRepository) GetByTimeRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*domain.AuditLog, int64, error) {
	query := `
	SELECT id, aggregate_id, event_type, resource_type, action, actor_id, actor_type,
		organization_id, ip_address, user_agent, details, status, failure_reason, created_at, archived
	FROM audit_logs
	WHERE created_at >= $1 AND created_at <= $2 AND archived = FALSE
	ORDER BY created_at DESC
	LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, from, to, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total int64
	countQuery := `SELECT COUNT(*) FROM audit_logs WHERE created_at >= $1 AND created_at <= $2 AND archived = FALSE`
	err = r.db.QueryRowContext(ctx, countQuery, from, to).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	logs, err := scanLogs(rows)
	return logs, total, err
}

// ArchiveOldLogs marks logs older than the given duration as archived
func (r *AuditRepository) ArchiveOldLogs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	query := `UPDATE audit_logs SET archived = TRUE, archived_at = NOW() WHERE created_at < $1 AND archived = FALSE`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}

// DeleteArchivedLogs deletes logs that have been archived for longer than the given duration
func (r *AuditRepository) DeleteArchivedLogs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	query := `DELETE FROM audit_logs WHERE archived = TRUE AND archived_at < $1`
	_, err := r.db.ExecContext(ctx, query, cutoff)
	return err
}

// CountTotal returns the total number of non-archived audit logs
func (r *AuditRepository) CountTotal(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_logs WHERE archived = FALSE`).Scan(&count)
	return count, err
}

// --- Internal Helpers ---

type querier interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func (r *AuditRepository) queryLogs(ctx context.Context, query string, filter interface{}, limit, offset int) ([]*domain.AuditLog, int64, error) {
	rows, err := r.db.QueryContext(ctx, query, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs, err := scanLogs(rows)
	return logs, 0, err
}

func scanLogs(rows *sql.Rows) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	for rows.Next() {
		var log domain.AuditLog
		var detailsJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.AggregateID,
			&log.EventType,
			&log.ResourceType,
			&log.Action,
			&log.ActorID,
			&log.ActorType,
			&log.OrganizationID,
			&log.IPAddress,
			&log.UserAgent,
			&detailsJSON,
			&log.Status,
			&log.FailureReason,
			&log.CreatedAt,
			&log.Archived,
		)
		if err != nil {
			return nil, err
		}

		if len(detailsJSON) > 0 {
			_ = json.Unmarshal(detailsJSON, &log.Details)
		}

		logs = append(logs, &log)
	}
	return logs, rows.Err()
}
