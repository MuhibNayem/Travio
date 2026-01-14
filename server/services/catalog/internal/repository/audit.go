package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID         string
	EntityType string
	EntityID   string
	Action     string
	ActorID    string
	Changes    map[string]interface{}
	CreatedAt  time.Time
}

type PostgresAuditRepository struct {
	DB *sql.DB
}

func NewAuditRepository(db *sql.DB) *PostgresAuditRepository {
	return &PostgresAuditRepository{DB: db}
}

func (r *PostgresAuditRepository) Log(ctx context.Context, log AuditLog) error {
	log.ID = uuid.New().String()
	log.CreatedAt = time.Now()

	changesJSON, _ := json.Marshal(log.Changes)

	query := `INSERT INTO audit_logs (id, entity_type, entity_id, action, actor_id, changes, created_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.DB.ExecContext(ctx, query,
		log.ID, log.EntityType, log.EntityID, log.Action, log.ActorID, changesJSON, log.CreatedAt)
	return err
}
