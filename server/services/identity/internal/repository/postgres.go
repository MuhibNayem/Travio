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
