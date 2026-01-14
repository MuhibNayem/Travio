package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/pricing/internal/engine"
	"github.com/google/uuid"
)

// PricingRule represents a pricing rule in the database
type PricingRule struct {
	ID              string
	OrganizationID  *string // Nullable
	Name            string
	Description     string
	Condition       string
	Multiplier      float64
	AdjustmentType  string
	AdjustmentValue float64
	Priority        int
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Repository interface for pricing rules
type Repository interface {
	GetActiveRules(ctx context.Context) ([]*PricingRule, error)
	GetAllRules(ctx context.Context, includeInactive bool) ([]*PricingRule, error)
	CreateRule(ctx context.Context, rule *PricingRule) error
	UpdateRule(ctx context.Context, rule *PricingRule) error
	DeleteRule(ctx context.Context, id string) error
}

// PostgresRepository implements Repository using PostgreSQL
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// InitSchema creates the pricing_rules table if not exists
func (r *PostgresRepository) InitSchema(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS pricing_rules (
			id VARCHAR(36) PRIMARY KEY,
			organization_id VARCHAR(36),
			name VARCHAR(255) NOT NULL,
			description TEXT,
			condition TEXT NOT NULL,
			multiplier DECIMAL(5,4) NOT NULL,
			adjustment_type VARCHAR(20) DEFAULT 'multiplier',
			adjustment_value DECIMAL(10,2) DEFAULT 0,
			priority INT DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_pricing_rules_active ON pricing_rules(is_active, priority);
		CREATE INDEX IF NOT EXISTS idx_pricing_rules_org ON pricing_rules(organization_id);
	`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

// SeedDefaultRules inserts sample rules if none exist
func (r *PostgresRepository) SeedDefaultRules(ctx context.Context) error {
	var count int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pricing_rules").Scan(&count)
	if count > 0 {
		return nil
	}

	defaultRules := []*PricingRule{
		{
			ID:             uuid.New().String(),
			OrganizationID: nil,
			Name:           "Weekend Surge",
			Description:    "20% increase on weekends",
			Condition:      `day_of_week == "Saturday" || day_of_week == "Sunday"`,
			Multiplier:     1.20,
			AdjustmentType: "multiplier",
			Priority:       10,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: nil,
			Name:           "Early Bird Discount",
			Description:    "15% off for bookings 30+ days in advance",
			Condition:      `days_until_departure > 30`,
			Multiplier:     0.85,
			AdjustmentType: "multiplier",
			Priority:       20,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: nil,
			Name:           "Last Minute Surge",
			Description:    "50% increase for bookings within 3 days",
			Condition:      `days_until_departure < 3`,
			Multiplier:     1.50,
			AdjustmentType: "multiplier",
			Priority:       5,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: nil,
			Name:           "High Demand Surge",
			Description:    "25% increase when occupancy > 80%",
			Condition:      `occupancy_rate > 0.8`,
			Multiplier:     1.25,
			AdjustmentType: "multiplier",
			Priority:       15,
			IsActive:       true,
		},
		{
			ID:             uuid.New().String(),
			OrganizationID: nil,
			Name:           "Business Class Premium",
			Description:    "40% premium for business class",
			Condition:      `seat_class == "business"`,
			Multiplier:     1.40,
			AdjustmentType: "multiplier",
			Priority:       1,
			IsActive:       true,
		},
	}

	for _, rule := range defaultRules {
		if err := r.CreateRule(ctx, rule); err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) GetActiveRules(ctx context.Context) ([]*PricingRule, error) {
	// Get ALL rules (inactive=false, orgID="") so we can cache everything
	return r.GetAllRules(ctx, false, "")
}

func (r *PostgresRepository) GetAllRules(ctx context.Context, includeInactive bool, organizationID string) ([]*PricingRule, error) {
	query := `SELECT id, organization_id, name, description, condition, multiplier, adjustment_type, adjustment_value, priority, is_active, created_at, updated_at 
	          FROM pricing_rules WHERE 1=1`

	args := []interface{}{}
	argIdx := 1

	if !includeInactive {
		query += " AND is_active = true"
	}

	if organizationID != "" {
		query += fmt.Sprintf(" AND organization_id = $%d", argIdx)
		args = append(args, organizationID)
		argIdx++
	}

	query += " ORDER BY priority ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*PricingRule
	for rows.Next() {
		var rule PricingRule
		if err := rows.Scan(&rule.ID, &rule.OrganizationID, &rule.Name, &rule.Description, &rule.Condition, &rule.Multiplier, &rule.AdjustmentType, &rule.AdjustmentValue, &rule.Priority, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, nil
}

func (r *PostgresRepository) CreateRule(ctx context.Context, rule *PricingRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	query := `INSERT INTO pricing_rules (id, organization_id, name, description, condition, multiplier, adjustment_type, adjustment_value, priority, is_active, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	_, err := r.db.ExecContext(ctx, query, rule.ID, rule.OrganizationID, rule.Name, rule.Description, rule.Condition, rule.Multiplier, rule.AdjustmentType, rule.AdjustmentValue, rule.Priority, rule.IsActive, rule.CreatedAt, rule.UpdatedAt)
	return err
}

func (r *PostgresRepository) UpdateRule(ctx context.Context, rule *PricingRule) error {
	rule.UpdatedAt = time.Now()
	query := `UPDATE pricing_rules SET name=$2, description=$3, condition=$4, multiplier=$5, adjustment_type=$6, adjustment_value=$7, priority=$8, is_active=$9, updated_at=$10 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, query, rule.ID, rule.Name, rule.Description, rule.Condition, rule.Multiplier, rule.AdjustmentType, rule.AdjustmentValue, rule.Priority, rule.IsActive, rule.UpdatedAt)
	return err
}

func (r *PostgresRepository) DeleteRule(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM pricing_rules WHERE id=$1", id)
	return err
}

// ToEngineRules converts repository rules to engine rules
func ToEngineRules(rules []*PricingRule) []*engine.Rule {
	var engineRules []*engine.Rule
	for _, r := range rules {
		engineRules = append(engineRules, &engine.Rule{
			ID:              r.ID,
			Name:            r.Name,
			Condition:       r.Condition,
			Multiplier:      r.Multiplier,
			AdjustmentType:  r.AdjustmentType,
			AdjustmentValue: r.AdjustmentValue,
			Priority:        r.Priority,
		})
	}
	return engineRules
}
