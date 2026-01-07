package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/MuhibNayem/Travio/server/services/operator/internal/domain"

	"github.com/google/uuid"
)

type VendorRepository interface {
	Create(ctx context.Context, vendor *domain.Vendor) error
	GetByID(ctx context.Context, id string) (*domain.Vendor, error)
	Update(ctx context.Context, vendor *domain.Vendor) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]*domain.Vendor, int64, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) VendorRepository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Create(ctx context.Context, vendor *domain.Vendor) error {
	if vendor.ID == "" {
		vendor.ID = uuid.New().String()
	}
	vendor.CreatedAt = time.Now()
	vendor.UpdatedAt = time.Now()

	query := `
		INSERT INTO vendors (id, name, contact_email, contact_phone, address, status, commission_rate, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.ExecContext(ctx, query,
		vendor.ID, vendor.Name, vendor.ContactEmail, vendor.ContactPhone,
		vendor.Address, vendor.Status, vendor.CommissionRate,
		vendor.CreatedAt, vendor.UpdatedAt,
	)
	return err
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*domain.Vendor, error) {
	query := `
		SELECT id, name, contact_email, contact_phone, address, status, commission_rate, created_at, updated_at
		FROM vendors WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var v domain.Vendor
	err := row.Scan(
		&v.ID, &v.Name, &v.ContactEmail, &v.ContactPhone,
		&v.Address, &v.Status, &v.CommissionRate,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("vendor not found")
		}
		return nil, err
	}
	return &v, nil
}

func (r *postgresRepository) Update(ctx context.Context, vendor *domain.Vendor) error {
	vendor.UpdatedAt = time.Now()
	query := `
		UPDATE vendors 
		SET name=$2, contact_email=$3, contact_phone=$4, address=$5, status=$6, commission_rate=$7, updated_at=$8
		WHERE id=$1
	`
	res, err := r.db.ExecContext(ctx, query,
		vendor.ID, vendor.Name, vendor.ContactEmail, vendor.ContactPhone,
		vendor.Address, vendor.Status, vendor.CommissionRate, vendor.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("vendor not found")
	}
	return nil
}

func (r *postgresRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM vendors WHERE id=$1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("vendor not found")
	}
	return nil
}

func (r *postgresRepository) List(ctx context.Context, page, limit int) ([]*domain.Vendor, int64, error) {
	offset := (page - 1) * limit

	// Get total count
	var total int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vendors").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, name, contact_email, contact_phone, address, status, commission_rate, created_at, updated_at
		FROM vendors
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var vendors []*domain.Vendor
	for rows.Next() {
		var v domain.Vendor
		if err := rows.Scan(
			&v.ID, &v.Name, &v.ContactEmail, &v.ContactPhone,
			&v.Address, &v.Status, &v.CommissionRate,
			&v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		vendors = append(vendors, &v)
	}

	return vendors, total, nil
}
