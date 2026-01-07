package repository

import (
	"database/sql"

	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
)

type InviteRepository struct {
	db *sql.DB
}

func NewInviteRepository(db *sql.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

func (r *InviteRepository) Create(invite *domain.Invite) error {
	query := `
		INSERT INTO organization_invites (organization_id, email, role, token, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return r.db.QueryRow(
		query,
		invite.OrganizationID,
		invite.Email,
		invite.Role,
		invite.Token,
		invite.Status,
		invite.ExpiresAt,
	).Scan(&invite.ID, &invite.CreatedAt)
}

func (r *InviteRepository) FindByToken(token string) (*domain.Invite, error) {
	query := `
		SELECT id, organization_id, email, role, token, status, expires_at, created_at
		FROM organization_invites
		WHERE token = $1`

	var i domain.Invite
	err := r.db.QueryRow(query, token).Scan(
		&i.ID, &i.OrganizationID, &i.Email, &i.Role, &i.Token, &i.Status, &i.ExpiresAt, &i.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (r *InviteRepository) FindByEmailAndOrg(email, orgID string) (*domain.Invite, error) {
	query := `
		SELECT id, organization_id, email, role, token, status, expires_at, created_at
		FROM organization_invites
		WHERE email = $1 AND organization_id = $2 AND status = 'pending'`

	var i domain.Invite
	err := r.db.QueryRow(query, email, orgID).Scan(
		&i.ID, &i.OrganizationID, &i.Email, &i.Role, &i.Token, &i.Status, &i.ExpiresAt, &i.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &i, nil
}

func (r *InviteRepository) UpdateStatus(id, status string) error {
	query := `UPDATE organization_invites SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

func (r *InviteRepository) ListByOrg(orgID string) ([]*domain.Invite, error) {
	query := `
		SELECT id, organization_id, email, role, token, status, expires_at, created_at
		FROM organization_invites
		WHERE organization_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invites []*domain.Invite
	for rows.Next() {
		var i domain.Invite
		if err := rows.Scan(
			&i.ID, &i.OrganizationID, &i.Email, &i.Role, &i.Token, &i.Status, &i.ExpiresAt, &i.CreatedAt,
		); err != nil {
			return nil, err
		}
		invites = append(invites, &i)
	}
	return invites, nil
}

// In a real generic user repo, but adding ListMembers here for SaaS refactor completeness
func (r *InviteRepository) ListMembers(orgID string, limit, offset int) ([]*domain.User, int, error) {
	// Count query
	var total int
	countQuery := `SELECT COUNT(*) FROM users WHERE organization_id = $1`
	err := r.db.QueryRow(countQuery, orgID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Data query
	query := `
		SELECT id, email, organization_id, role, status, created_at
		FROM users
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(
			&u.ID, &u.Email, &u.OrganizationID, &u.Role, &u.Status, &u.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, &u)
	}
	return users, total, nil
}

// Helper to remove member
func (r *InviteRepository) RemoveMember(userID, orgID string) error {
	// Set their OrgID to NULL (or soft delete logic depending on requirements)
	// For "Staff", we probably want to disable them or remove the link.
	// Assuming logic: Remove access = Set OrgID NULL

	// Check if correct org
	query := `UPDATE users SET organization_id = NULL, role = 'user' WHERE id = $1 AND organization_id = $2`
	_, err := r.db.Exec(query, userID, orgID)
	return err
}

// Helper to update role
func (r *InviteRepository) UpdateMemberRole(userID, orgID, role string) error {
	query := `UPDATE users SET role = $1 WHERE id = $2 AND organization_id = $3`
	_, err := r.db.Exec(query, role, userID, orgID)
	return err
}

func (r *InviteRepository) CountAdmins(orgID string) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE organization_id = $1 AND role = 'admin'`
	err := r.db.QueryRow(query, orgID).Scan(&count)
	return count, err
}
