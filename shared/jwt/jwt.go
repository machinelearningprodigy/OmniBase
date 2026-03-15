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

// IssueAnonymousToken creates a long-lived JWT for public (anon) requests
func (m *Manager) IssueAnonymousToken(projectID string) (string, error) {
	// API keys for self-hosted instances usually don't expire soon (10 years)
	return m.issue("00000000-0000-0000-0000-000000000000", "", "anon", projectID, AccessToken, 10*365*24*time.Hour)
}

// IssueServiceRoleToken creates a long-lived JWT with master access
func (m *Manager) IssueServiceRoleToken(projectID string) (string, error) {
	return m.issue("00000000-0000-0000-0000-000000000000", "", "service_role", projectID, AccessToken, 10*365*24*time.Hour)
}

var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)

// IssueStorageToken generates a custom token for signed storage URLs
func (m *Manager) IssueStorageToken(bucket, path string, expiresIn int) (string, error) {
	now := time.Now().UTC()
	claims := jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("storage:%s:%s", bucket, path),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiresIn) * time.Second)),
		Issuer:    "omnibase-storage",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// VerifyStorageToken validates a storage signed URL token
func (m *Manager) VerifyStorageToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid storage token")
	}

	return claims, nil
}
