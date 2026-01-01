package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/MuhibNayem/Travio/server/services/identity/internal/domain"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrFamilyRevoked        = errors.New("token family has been revoked")
)

type RefreshTokenRepository struct {
	DB *sql.DB
}

func NewRefreshTokenRepository(db *sql.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{DB: db}
}

// Create stores a new refresh token
func (r *RefreshTokenRepository) Create(token *domain.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, family_id, token_hash, revoked, expires_at, created_at, last_used_at, user_agent, ip_address) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.DB.Exec(query,
		token.ID,
		token.UserID,
		token.FamilyID,
		token.TokenHash,
		token.Revoked,
		token.ExpiresAt,
		token.CreatedAt,
		token.LastUsedAt,
		token.UserAgent,
		token.IPAddress,
	)
	return err
}

// FindByID finds a refresh token by its JTI
func (r *RefreshTokenRepository) FindByID(id string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, family_id, token_hash, revoked, expires_at, created_at, last_used_at, user_agent, ip_address 
			  FROM refresh_tokens WHERE id = $1`

	var token domain.RefreshToken
	err := r.DB.QueryRow(query, id).Scan(
		&token.ID,
		&token.UserID,
		&token.FamilyID,
		&token.TokenHash,
		&token.Revoked,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.LastUsedAt,
		&token.UserAgent,
		&token.IPAddress,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}
	return &token, nil
}

// MarkUsed updates the last_used_at timestamp
func (r *RefreshTokenRepository) MarkUsed(id string) error {
	query := `UPDATE refresh_tokens SET last_used_at = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, time.Now(), id)
	return err
}

// Revoke marks a specific token as revoked
func (r *RefreshTokenRepository) Revoke(id string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}

// RevokeFamily revokes ALL tokens in a family (reuse detection / logout-all)
func (r *RefreshTokenRepository) RevokeFamily(familyID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE family_id = $1`
	_, err := r.DB.Exec(query, familyID)
	return err
}

// RevokeAllForUser revokes all refresh tokens for a user (password change, security event)
func (r *RefreshTokenRepository) RevokeAllForUser(userID string) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}

// GetActiveSessionsForUser lists all non-revoked tokens for session management UI
func (r *RefreshTokenRepository) GetActiveSessionsForUser(userID string) ([]*domain.RefreshToken, error) {
	query := `SELECT id, user_id, family_id, token_hash, revoked, expires_at, created_at, last_used_at, user_agent, ip_address 
			  FROM refresh_tokens 
			  WHERE user_id = $1 AND revoked = false AND expires_at > NOW()
			  ORDER BY last_used_at DESC`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []*domain.RefreshToken
	for rows.Next() {
		var t domain.RefreshToken
		if err := rows.Scan(
			&t.ID, &t.UserID, &t.FamilyID, &t.TokenHash, &t.Revoked, &t.ExpiresAt, &t.CreatedAt, &t.LastUsedAt, &t.UserAgent, &t.IPAddress,
		); err != nil {
			return nil, err
		}
		tokens = append(tokens, &t)
	}
	return tokens, nil
}

// DeleteExpired cleans up old expired tokens (run via cron job)
func (r *RefreshTokenRepository) DeleteExpired() (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`
	result, err := r.DB.Exec(query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
