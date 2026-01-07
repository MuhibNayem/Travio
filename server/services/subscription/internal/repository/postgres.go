package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Plan struct {
	ID              string
	Name            string
	Description     string
	PricePaisa      int64
	Interval        string
	Features        map[string]string
	IsActive        bool
	UsagePricePaisa int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type UsageEvent struct {
	ID             string
	SubscriptionID string
	EventType      string
	Units          int64
	IdempotencyKey string
	CreatedAt      time.Time
}

type LineItem struct {
	Description    string `json:"description"`
	AmountPaisa    int64  `json:"amount_paisa"`
	Quantity       int64  `json:"quantity"`
	UnitPricePaisa int64  `json:"unit_price_paisa"`
}

type Subscription struct {
	ID                 string
	OrganizationID     string
	PlanID             string
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Invoice struct {
	ID             string
	SubscriptionID string
	AmountPaisa    int64
	Status         string
	LineItems      []LineItem
	IssuedAt       time.Time
	DueDate        time.Time
	PaidAt         *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Repository interface {
	CreatePlan(ctx context.Context, plan *Plan) error
	ListPlans(ctx context.Context, includeInactive bool) ([]*Plan, error)
	GetPlan(ctx context.Context, id string) (*Plan, error)
	UpdatePlan(ctx context.Context, plan *Plan) error

	CreateSubscription(ctx context.Context, sub *Subscription) error
	GetSubscription(ctx context.Context, organizationID string) (*Subscription, error)
	ListSubscriptions(ctx context.Context, planID, status string) ([]*Subscription, error)
	CancelSubscription(ctx context.Context, organizationID string) error
	UpdateSubscriptionStatus(ctx context.Context, subID, status string) error

	CreateInvoice(ctx context.Context, invoice *Invoice) error
	ListInvoices(ctx context.Context, subscriptionID string) ([]*Invoice, error)

	// Usage Methods
	RecordUsage(ctx context.Context, event *UsageEvent) (string, error)
	GetUsageForPeriod(ctx context.Context, subID string, start, end time.Time) (int64, error)

	// Entitlement Check
	GetEntitlement(ctx context.Context, organizationID string) (*Entitlement, error)
}

// Entitlement represents the combined subscription and plan data for enforcement
type Entitlement struct {
	OrganizationID  string
	PlanID          string
	PlanName        string
	Status          string
	Features        map[string]string
	QuotaLimits     map[string]int64
	UsageThisPeriod map[string]int64
	PeriodStart     time.Time
	PeriodEnd       time.Time
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Plan Methods

func (r *PostgresRepository) CreatePlan(ctx context.Context, plan *Plan) error {
	if plan.ID == "" {
		plan.ID = uuid.New().String()
	}
	plan.CreatedAt = time.Now()
	plan.UpdatedAt = time.Now()

	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return err
	}

	query := `INSERT INTO plans (id, name, description, price_paisa, interval, features, is_active, usage_price_paisa, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.ExecContext(ctx, query, plan.ID, plan.Name, plan.Description, plan.PricePaisa, plan.Interval, featuresJSON, plan.IsActive, plan.UsagePricePaisa, plan.CreatedAt, plan.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListPlans(ctx context.Context, includeInactive bool) ([]*Plan, error) {
	query := `SELECT id, name, description, price_paisa, interval, features, is_active, usage_price_paisa, created_at, updated_at FROM plans`
	if !includeInactive {
		query += " WHERE is_active = true"
	}

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*Plan
	for rows.Next() {
		var p Plan
		var featuresBytes []byte
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PricePaisa, &p.Interval, &featuresBytes, &p.IsActive, &p.UsagePricePaisa, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if err := json.Unmarshal(featuresBytes, &p.Features); err != nil {
			return nil, err
		}
		plans = append(plans, &p)
	}
	return plans, nil
}

func (r *PostgresRepository) GetPlan(ctx context.Context, id string) (*Plan, error) {
	query := `SELECT id, name, description, price_paisa, interval, features, is_active, usage_price_paisa, created_at, updated_at FROM plans WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var p Plan
	var featuresBytes []byte
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.PricePaisa, &p.Interval, &featuresBytes, &p.IsActive, &p.UsagePricePaisa, &p.CreatedAt, &p.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(featuresBytes, &p.Features); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PostgresRepository) UpdatePlan(ctx context.Context, plan *Plan) error {
	plan.UpdatedAt = time.Now()
	featuresJSON, err := json.Marshal(plan.Features)
	if err != nil {
		return err
	}
	query := `UPDATE plans SET name=$1, description=$2, price_paisa=$3, features=$4, is_active=$5, usage_price_paisa=$6, updated_at=$7 WHERE id=$8`
	_, err = r.db.ExecContext(ctx, query, plan.Name, plan.Description, plan.PricePaisa, featuresJSON, plan.IsActive, plan.UsagePricePaisa, plan.UpdatedAt, plan.ID)
	return err
}

// Subscription Methods

func (r *PostgresRepository) CreateSubscription(ctx context.Context, sub *Subscription) error {
	if sub.ID == "" {
		sub.ID = uuid.New().String()
	}
	sub.CreatedAt = time.Now()
	sub.UpdatedAt = time.Now()

	query := `INSERT INTO subscriptions (id, organization_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query, sub.ID, sub.OrganizationID, sub.PlanID, sub.Status, sub.CurrentPeriodStart, sub.CurrentPeriodEnd, sub.CreatedAt, sub.UpdatedAt)
	return err
}

func (r *PostgresRepository) GetSubscription(ctx context.Context, organizationID string) (*Subscription, error) {
	// Gets the latest Active or Trialing subscription
	query := `SELECT id, organization_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at 
	          FROM subscriptions 
	          WHERE organization_id = $1 AND status IN ('active', 'trialing', 'past_due')
	          ORDER BY created_at DESC LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, organizationID)

	var s Subscription
	if err := row.Scan(&s.ID, &s.OrganizationID, &s.PlanID, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt, &s.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) ListSubscriptions(ctx context.Context, planID, status string) ([]*Subscription, error) {
	query := `SELECT id, organization_id, plan_id, status, current_period_start, current_period_end, created_at, updated_at FROM subscriptions WHERE 1=1`
	var args []interface{}
	argIdx := 1

	if planID != "" {
		query += fmt.Sprintf(" AND plan_id = $%d", argIdx)
		args = append(args, planID)
		argIdx++
	}
	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []*Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.ID, &s.OrganizationID, &s.PlanID, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, &s)
	}
	return subs, nil
}

func (r *PostgresRepository) CancelSubscription(ctx context.Context, organizationID string) error {
	query := `UPDATE subscriptions SET status = 'canceled', updated_at = NOW() 
	          WHERE organization_id = $1 AND status IN ('active', 'trialing')`

	result, err := r.db.ExecContext(ctx, query, organizationID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("no active subscription found to cancel")
	}
	return nil
}

func (r *PostgresRepository) UpdateSubscriptionStatus(ctx context.Context, subID, status string) error {
	query := `UPDATE subscriptions SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, subID)
	return err
}

// Invoice Methods

func (r *PostgresRepository) CreateInvoice(ctx context.Context, inv *Invoice) error {
	if inv.ID == "" {
		inv.ID = uuid.New().String()
	}
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = time.Now()

	lineItemsJSON, err := json.Marshal(inv.LineItems)
	if err != nil {
		return err
	}

	query := `INSERT INTO invoices (id, subscription_id, amount_paisa, status, issued_at, due_date, paid_at, line_items, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.ExecContext(ctx, query, inv.ID, inv.SubscriptionID, inv.AmountPaisa, inv.Status, inv.IssuedAt, inv.DueDate, inv.PaidAt, lineItemsJSON, inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListInvoices(ctx context.Context, subscriptionID string) ([]*Invoice, error) {
	query := `SELECT id, subscription_id, amount_paisa, status, issued_at, due_date, paid_at, line_items, created_at, updated_at FROM invoices WHERE subscription_id = $1 ORDER BY issued_at DESC`
	rows, err := r.db.QueryContext(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		var inv Invoice
		var lineItemsBytes []byte
		if err := rows.Scan(&inv.ID, &inv.SubscriptionID, &inv.AmountPaisa, &inv.Status, &inv.IssuedAt, &inv.DueDate, &inv.PaidAt, &lineItemsBytes, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		if lineItemsBytes != nil {
			if err := json.Unmarshal(lineItemsBytes, &inv.LineItems); err != nil {
				// Log error but probably don't fail entire fetch
			}
		}
		invoices = append(invoices, &inv)
	}
	return invoices, nil
}

// Usage Methods

func (r *PostgresRepository) RecordUsage(ctx context.Context, event *UsageEvent) (string, error) {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	query := `INSERT INTO usage_events (id, subscription_id, event_type, units, idempotency_key, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          ON CONFLICT (idempotency_key) DO NOTHING RETURNING id`

	// This assumes Postgres 9.5+
	var returnedID string
	err := r.db.QueryRowContext(ctx, query, event.ID, event.SubscriptionID, event.EventType, event.Units, event.IdempotencyKey, event.CreatedAt).Scan(&returnedID)
	if err != nil {
		if err == sql.ErrNoRows {
			// Duplicate (ON CONFLICT DO NOTHING returned no row)
			// Return ID of existing if we want? Or just empty id + no error
			// Let's query existing ID
			queryExisting := `SELECT id FROM usage_events WHERE idempotency_key = $1`
			err := r.db.QueryRowContext(ctx, queryExisting, event.IdempotencyKey).Scan(&returnedID)
			if err != nil {
				return "", err
			}
			return returnedID, nil // Return existing ID but maybe indicate duplicate? Service logic handles it.
		}
		return "", err
	}
	return returnedID, nil
}

func (r *PostgresRepository) GetUsageForPeriod(ctx context.Context, subID string, start, end time.Time) (int64, error) {
	query := `SELECT COALESCE(SUM(units), 0) FROM usage_events 
	          WHERE subscription_id = $1 AND created_at >= $2 AND created_at < $3`

	var total int64
	err := r.db.QueryRowContext(ctx, query, subID, start, end).Scan(&total)
	return total, err
}

// GetEntitlement fetches subscription + plan data in a single query for entitlement checks.
func (r *PostgresRepository) GetEntitlement(ctx context.Context, organizationID string) (*Entitlement, error) {
	query := `
		SELECT 
			s.organization_id,
			s.plan_id,
			p.name as plan_name,
			s.status,
			p.features,
			s.current_period_start,
			s.current_period_end
		FROM subscriptions s
		INNER JOIN plans p ON s.plan_id = p.id
		WHERE s.organization_id = $1 AND s.status != 'canceled'
		ORDER BY s.created_at DESC
		LIMIT 1
	`

	var ent Entitlement
	var featuresJSON []byte
	err := r.db.QueryRowContext(ctx, query, organizationID).Scan(
		&ent.OrganizationID,
		&ent.PlanID,
		&ent.PlanName,
		&ent.Status,
		&featuresJSON,
		&ent.PeriodStart,
		&ent.PeriodEnd,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No subscription found
		}
		return nil, fmt.Errorf("failed to get entitlement: %w", err)
	}

	// Parse features
	if len(featuresJSON) > 0 {
		if err := json.Unmarshal(featuresJSON, &ent.Features); err != nil {
			ent.Features = make(map[string]string)
		}
	} else {
		ent.Features = make(map[string]string)
	}

	// Build quota limits from numeric features
	ent.QuotaLimits = make(map[string]int64)
	for key, value := range ent.Features {
		if num, err := parseInt64(value); err == nil {
			ent.QuotaLimits[key] = num
		}
	}

	// Fetch usage for current period
	ent.UsageThisPeriod = make(map[string]int64)

	// Get subscription ID for usage query
	var subID string
	subQuery := `SELECT id FROM subscriptions WHERE organization_id = $1 AND status != 'canceled' ORDER BY created_at DESC LIMIT 1`
	if err := r.db.QueryRowContext(ctx, subQuery, organizationID).Scan(&subID); err == nil {
		// Get ticket_sale usage
		usageQuery := `SELECT event_type, COALESCE(SUM(units), 0) 
		               FROM usage_events 
		               WHERE subscription_id = $1 AND created_at >= $2 AND created_at < $3
		               GROUP BY event_type`
		rows, err := r.db.QueryContext(ctx, usageQuery, subID, ent.PeriodStart, ent.PeriodEnd)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var eventType string
				var units int64
				if err := rows.Scan(&eventType, &units); err == nil {
					ent.UsageThisPeriod[eventType] = units
				}
			}
		}
	}

	return &ent, nil
}

// parseInt64 safely parses a string to int64
func parseInt64(s string) (int64, error) {
	var n int64
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}
