package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/crm/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrCouponNotFound = errors.New("coupon not found")
	ErrTicketNotFound = errors.New("ticket not found")
)

type CRMRepository struct {
	DB *sql.DB
}

func NewCRMRepository(db *sql.DB) *CRMRepository {
	return &CRMRepository{DB: db}
}

// --- Coupons ---

func (r *CRMRepository) CreateCoupon(c *domain.Coupon) error {
	c.ID = uuid.New().String()
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()

	query := `INSERT INTO coupons (id, organization_id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, usage_count, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	_, err := r.DB.Exec(query, c.ID, c.OrganizationID, c.Code, c.DiscountType, c.DiscountValue, c.MinPurchaseAmount, c.MaxDiscountAmount, c.StartDate, c.EndDate, c.UsageLimit, c.UsageCount, c.IsActive, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *CRMRepository) GetCouponByCode(code, orgID string) (*domain.Coupon, error) {
	query := `SELECT id, organization_id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, usage_count, is_active, created_at, updated_at FROM coupons WHERE code = $1 AND organization_id = $2`

	return r.scanCoupon(r.DB.QueryRow(query, code, orgID))
}

func (r *CRMRepository) GetCoupon(id string) (*domain.Coupon, error) {
	query := `SELECT id, organization_id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, usage_count, is_active, created_at, updated_at FROM coupons WHERE id = $1`

	return r.scanCoupon(r.DB.QueryRow(query, id))
}

func (r *CRMRepository) ListCoupons(orgID string) ([]*domain.Coupon, error) {
	query := `SELECT id, organization_id, code, discount_type, discount_value, min_purchase_amount, max_discount_amount, start_date, end_date, usage_limit, usage_count, is_active, created_at, updated_at FROM coupons WHERE organization_id = $1`

	rows, err := r.DB.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var coupons []*domain.Coupon
	for rows.Next() {
		c, err := r.scanCouponRow(rows)
		if err != nil {
			return nil, err
		}
		coupons = append(coupons, c)
	}
	return coupons, nil
}

func (r *CRMRepository) UpdateCouponUsage(id string) error {
	query := `UPDATE coupons SET usage_count = usage_count + 1, updated_at = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, time.Now(), id)
	return err
}

// --- Support Tickets ---

func (r *CRMRepository) CreateTicket(t *domain.SupportTicket) error {
	t.ID = uuid.New().String()
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	t.Status = "OPEN"

	query := `INSERT INTO support_tickets (id, organization_id, user_id, subject, status, priority, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.DB.Exec(query, t.ID, t.OrganizationID, t.UserID, t.Subject, t.Status, t.Priority, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *CRMRepository) CreateTicketMessage(msg *domain.TicketMessage) error {
	msg.ID = uuid.New().String()
	msg.CreatedAt = time.Now()

	query := `INSERT INTO ticket_messages (id, ticket_id, sender_id, message, created_at) 
              VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Exec(query, msg.ID, msg.TicketID, msg.SenderID, msg.Message, msg.CreatedAt)
	return err
}

func (r *CRMRepository) ListTickets(orgID string) ([]*domain.SupportTicket, error) {
	// Basic listing for admin/operator
	query := `SELECT id, organization_id, user_id, subject, status, priority, created_at, updated_at FROM support_tickets WHERE organization_id = $1 ORDER BY created_at DESC`
	// Note: Should probably support user_id filter too for customers

	rows, err := r.DB.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []*domain.SupportTicket
	for rows.Next() {
		var t domain.SupportTicket
		if err := rows.Scan(&t.ID, &t.OrganizationID, &t.UserID, &t.Subject, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, &t)
	}
	return tickets, nil
}

// --- Helpers ---

func (r *CRMRepository) scanCoupon(row *sql.Row) (*domain.Coupon, error) {
	var c domain.Coupon
	err := row.Scan(&c.ID, &c.OrganizationID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchaseAmount, &c.MaxDiscountAmount, &c.StartDate, &c.EndDate, &c.UsageLimit, &c.UsageCount, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCouponNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *CRMRepository) scanCouponRow(rows *sql.Rows) (*domain.Coupon, error) {
	var c domain.Coupon
	err := rows.Scan(&c.ID, &c.OrganizationID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchaseAmount, &c.MaxDiscountAmount, &c.StartDate, &c.EndDate, &c.UsageLimit, &c.UsageCount, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
