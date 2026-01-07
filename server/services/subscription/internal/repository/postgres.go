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
	ID          string
	Name        string
	Description string
	PricePaisa  int64
	Interval    string
	Features    map[string]string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

	query := `INSERT INTO plans (id, name, description, price_paisa, interval, features, is_active, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err = r.db.ExecContext(ctx, query, plan.ID, plan.Name, plan.Description, plan.PricePaisa, plan.Interval, featuresJSON, plan.IsActive, plan.CreatedAt, plan.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListPlans(ctx context.Context, includeInactive bool) ([]*Plan, error) {
	query := `SELECT id, name, description, price_paisa, interval, features, is_active, created_at, updated_at FROM plans`
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
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.PricePaisa, &p.Interval, &featuresBytes, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
	query := `SELECT id, name, description, price_paisa, interval, features, is_active, created_at, updated_at FROM plans WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var p Plan
	var featuresBytes []byte
	if err := row.Scan(&p.ID, &p.Name, &p.Description, &p.PricePaisa, &p.Interval, &featuresBytes, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
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
	query := `UPDATE plans SET name=$1, description=$2, price_paisa=$3, features=$4, is_active=$5, updated_at=$6 WHERE id=$7`
	_, err = r.db.ExecContext(ctx, query, plan.Name, plan.Description, plan.PricePaisa, featuresJSON, plan.IsActive, plan.UpdatedAt, plan.ID)
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

	query := `INSERT INTO invoices (id, subscription_id, amount_paisa, status, issued_at, due_date, paid_at, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query, inv.ID, inv.SubscriptionID, inv.AmountPaisa, inv.Status, inv.IssuedAt, inv.DueDate, inv.PaidAt, inv.CreatedAt, inv.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListInvoices(ctx context.Context, subscriptionID string) ([]*Invoice, error) {
	query := `SELECT id, subscription_id, amount_paisa, status, issued_at, due_date, paid_at, created_at, updated_at FROM invoices WHERE subscription_id = $1 ORDER BY issued_at DESC`
	rows, err := r.db.QueryContext(ctx, query, subscriptionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []*Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.SubscriptionID, &inv.AmountPaisa, &inv.Status, &inv.IssuedAt, &inv.DueDate, &inv.PaidAt, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, &inv)
	}
	return invoices, nil
}
