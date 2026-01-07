package domain

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
	ErrSigningKey   = errors.New("signing key missing")
)

// TokenClaims defines the payload for queue admission tokens
type TokenClaims struct {
	UserID  string `json:"uid"`
	EventID string `json:"eid"`
	jwt.RegisteredClaims
}

// TokenManager handles JWT generation and validation
type TokenManager struct {
	signingKey []byte
	issuer     string
}

// NewTokenManager creates a new token manager
func NewTokenManager(signingKey string) *TokenManager {
	if signingKey == "" {
		signingKey = "default-dev-secret-do-not-use-in-prod"
	}
	return &TokenManager{
		signingKey: []byte(signingKey),
		issuer:     "travio-queue-service",
	}
}

// GenerateToken creates a signed JWT for an admitted user
func (tm *TokenManager) GenerateToken(userID, eventID string, ttl time.Duration) (string, error) {
	claims := TokenClaims{
		UserID:  userID,
		EventID: eventID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    tm.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(tm.signingKey)
}

// ValidateToken verifies a JWT and returns the claims
// This can be used by the Gateway locally
func (tm *TokenManager) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return tm.signingKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}
