package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// --- Configuration (Load from ENV in prod) ---
var (
	AccessTokenSecret  = []byte("access-secret-key-change-me-in-prod")
	RefreshTokenSecret = []byte("refresh-secret-key-change-me-in-prod")
	AccessTokenTTL     = 15 * time.Minute
	RefreshTokenTTL    = 7 * 24 * time.Hour
)

// --- Errors ---
var (
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrRefreshTokenReused = errors.New("refresh token reuse detected, family revoked")
)

// --- Claims ---
type AccessTokenClaims struct {
	UserID         string `json:"uid"`
	OrganizationID string `json:"oid"`
	Role           string `json:"role"`
	jwt.RegisteredClaims
}

type RefreshTokenClaims struct {
	UserID   string `json:"uid"`
	FamilyID string `json:"fid"` // Token Family for rotation tracking
	jwt.RegisteredClaims
}

// --- Token Generation ---

// GenerateAccessToken creates a short-lived Access Token (15 min)
func GenerateAccessToken(userID, orgID, role string) (string, error) {
	jti := uuid.New().String()
	claims := AccessTokenClaims{
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "travio-identity",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(AccessTokenSecret)
}

// GenerateRefreshToken creates a long-lived Refresh Token (7 days)
// Returns: signed token string, raw familyID, raw tokenID (JTI)
func GenerateRefreshToken(userID, familyID string) (signedToken string, jti string, err error) {
	jti = uuid.New().String()
	if familyID == "" {
		familyID = uuid.New().String() // New family on first login
	}

	claims := RefreshTokenClaims{
		UserID:   userID,
		FamilyID: familyID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "travio-identity",
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err = token.SignedString(RefreshTokenSecret)
	return signedToken, jti, err
}

// --- Token Validation ---

func ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return AccessTokenSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

func ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return RefreshTokenSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*RefreshTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, ErrInvalidToken
}

// --- Utilities ---

// GenerateOpaqueToken creates a cryptographically secure random string (for refresh tokens if not using JWT)
func GenerateOpaqueToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashToken creates a SHA256 hash of a token (for secure storage)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(hash[:])
}
