package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{DB: db}
}

func (r *UserRepository) Create(user *domain.User) error {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()

	query := `INSERT INTO users (id, email, password_hash, organization_id, role, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.DB.Exec(query, user.ID, user.Email, user.PasswordHash, user.OrganizationID, user.Role, user.CreatedAt)
	return err
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, organization_id, role, created_at FROM users WHERE email = $1`

	row := r.DB.QueryRow(query, email)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.OrganizationID, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, organization_id, role, created_at FROM users WHERE id = $1`

	row := r.DB.QueryRow(query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.OrganizationID, &user.Role, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	user.UpdatedAt = time.Now()
	query := `UPDATE users 
	          SET email = $1, password_hash = $2, organization_id = $3, role = $4, updated_at = $5 
			  WHERE id = $6`

	result, err := r.DB.Exec(query, user.Email, user.PasswordHash, user.OrganizationID, user.Role, user.UpdatedAt, user.ID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrUserNotFound
	}
	return nil
}

type OrgRepository struct {
	DB *sql.DB
}

func NewOrgRepository(db *sql.DB) *OrgRepository {
	return &OrgRepository{DB: db}
}

func (r *OrgRepository) Create(org *domain.Organization) error {
	org.ID = uuid.New().String()
	org.CreatedAt = time.Now()
	org.Status = "active" // Default status

	query := `INSERT INTO organizations (id, name, plan_id, status, created_at) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := r.DB.Exec(query, org.ID, org.Name, org.PlanID, org.Status, org.CreatedAt)
	return err
}

func (r *OrgRepository) FindByID(id string) (*domain.Organization, error) {
	query := `SELECT id, name, plan_id, status, created_at FROM organizations WHERE id = $1`
	row := r.DB.QueryRow(query, id)

	var org domain.Organization
	if err := row.Scan(&org.ID, &org.Name, &org.PlanID, &org.Status, &org.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return &org, nil
}
