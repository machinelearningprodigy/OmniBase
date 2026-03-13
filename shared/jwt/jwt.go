package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType differentiates access and refresh tokens
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the OmniBase JWT payload.
// This is the universal claims structure used across all services.
type Claims struct {
	UserID    string    `json:"sub"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`  // "anon" | "authenticated" | "service_role"
	TokenType TokenType `json:"type"`
	ProjectID string    `json:"project_id,omitempty"`
	jwt.RegisteredClaims
}

// Manager handles JWT creation and validation
type Manager struct {
	secret          []byte
	accessExpiry    time.Duration
	refreshExpiry   time.Duration
}

// NewManager creates a JWT manager
func NewManager(secret string, accessExpiry, refreshExpiry time.Duration) *Manager {
	return &Manager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// IssueAccessToken creates a short-lived JWT access token
func (m *Manager) IssueAccessToken(userID, email, role, projectID string) (string, error) {
	return m.issue(userID, email, role, projectID, AccessToken, m.accessExpiry)
}

// IssueRefreshToken creates a long-lived JWT refresh token
func (m *Manager) IssueRefreshToken(userID, email, role, projectID string) (string, error) {
	return m.issue(userID, email, role, projectID, RefreshToken, m.refreshExpiry)
}

func (m *Manager) issue(userID, email, role, projectID string, tokenType TokenType, expiry time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		ProjectID: projectID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			Issuer:    "omnibase",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Verify validates a JWT and returns the claims
func (m *Manager) Verify(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// IssueAnonymousToken creates a JWT for unauthenticated (anon) requests
func (m *Manager) IssueAnonymousToken(projectID string) (string, error) {
	return m.issue("anon", "", "anon", projectID, AccessToken, m.accessExpiry)
}

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)
