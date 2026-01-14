package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Promotion struct {
	ID                  string
	OrganizationID      *string
	Code                string
	Description         string
	DiscountType        string // "PERCENT", "FIXED"
	DiscountValue       float64
	MaxUsage            int64
	CurrentUsage        int64
	ValidFrom           *time.Time
	ValidUntil          *time.Time
	MinOrderAmountPaisa int64
	IsActive            bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// InitPromotionsSchema creates the promotions table
func (r *PostgresRepository) InitPromotionsSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS promotions (
			id VARCHAR(36) PRIMARY KEY,
			organization_id VARCHAR(36),
			code VARCHAR(50) NOT NULL,
			description TEXT,
			discount_type VARCHAR(20) NOT NULL,
			discount_value DECIMAL(10,2) NOT NULL,
			max_usage BIGINT DEFAULT 0,
			current_usage BIGINT DEFAULT 0,
			valid_from TIMESTAMP,
			valid_until TIMESTAMP,
			min_order_amount_paisa BIGINT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(organization_id, code)
		);
		CREATE INDEX IF NOT EXISTS idx_promotions_code ON promotions(code);
		CREATE INDEX IF NOT EXISTS idx_promotions_org ON promotions(organization_id);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *PostgresRepository) CreatePromotion(ctx context.Context, p *Promotion) error {
	query := `INSERT INTO promotions (
		id, organization_id, code, description, discount_type, discount_value, 
		max_usage, valid_from, valid_until, min_order_amount_paisa, is_active, created_at, updated_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`

	_, err := r.db.ExecContext(ctx, query,
		p.ID, p.OrganizationID, p.Code, p.Description, p.DiscountType, p.DiscountValue,
		p.MaxUsage, p.ValidFrom, p.ValidUntil, p.MinOrderAmountPaisa, p.IsActive, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetPromotionByCode(ctx context.Context, code, orgID string) (*Promotion, error) {
	query := `SELECT id, organization_id, code, description, discount_type, discount_value, 
		max_usage, current_usage, valid_from, valid_until, min_order_amount_paisa, is_active, created_at, updated_at
		FROM promotions WHERE code = $1 AND (organization_id = $2 OR organization_id IS NULL) AND is_active = true`

	row := r.db.QueryRowContext(ctx, query, code, orgID)

	var p Promotion
	err := row.Scan(
		&p.ID, &p.OrganizationID, &p.Code, &p.Description, &p.DiscountType, &p.DiscountValue,
		&p.MaxUsage, &p.CurrentUsage, &p.ValidFrom, &p.ValidUntil, &p.MinOrderAmountPaisa, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) IncrementPromotionUsage(ctx context.Context, id string) error {
	query := `UPDATE promotions SET current_usage = current_usage + 1 WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *PostgresRepository) GetPromotions(ctx context.Context, orgID string, activeOnly bool) ([]*Promotion, error) {
	query := `SELECT id, organization_id, code, description, discount_type, discount_value, 
		max_usage, current_usage, valid_from, valid_until, min_order_amount_paisa, is_active, created_at, updated_at
		FROM promotions WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if orgID != "" {
		query += fmt.Sprintf(" AND (organization_id = $%d OR organization_id IS NULL)", argIdx)
		args = append(args, orgID)
		argIdx++
	}

	if activeOnly {
		query += " AND is_active = true"
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var promotions []*Promotion
	for rows.Next() {
		var p Promotion
		if err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.Code, &p.Description, &p.DiscountType, &p.DiscountValue,
			&p.MaxUsage, &p.CurrentUsage, &p.ValidFrom, &p.ValidUntil, &p.MinOrderAmountPaisa, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		promotions = append(promotions, &p)
	}
	return promotions, nil
}
