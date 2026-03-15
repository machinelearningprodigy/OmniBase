package models

import (
	"time"
)

// ─── Auth Models ──────────────────────────────────────────────────────────────

// User represents an authenticated user in OmniBase
type User struct {
	ID               string     `json:"id" db:"id"`
	Email            string     `json:"email" db:"email"`
	EmailConfirmedAt *time.Time `json:"email_confirmed_at,omitempty" db:"email_confirmed_at"`
	Phone            string     `json:"phone,omitempty" db:"phone"`
	PhoneConfirmedAt *time.Time `json:"phone_confirmed_at,omitempty" db:"phone_confirmed_at"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	Role             string     `json:"role" db:"role"`
	RawUserMetaData  JSONB      `json:"user_metadata,omitempty" db:"raw_user_meta_data"`
	RawAppMetaData   JSONB      `json:"app_metadata,omitempty" db:"raw_app_meta_data"`
	IsSuperAdmin     bool       `json:"is_super_admin,omitempty" db:"is_super_admin"`
	IsBanned         bool       `json:"is_banned" db:"is_banned"`
	LastSignInAt     *time.Time `json:"last_sign_in_at,omitempty" db:"last_sign_in_at"`
	ProjectID        string     `json:"project_id,omitempty" db:"project_id"`
	
	// Helper fields for UI
	DisplayName      string     `json:"display_name,omitempty" db:"-"`
	AvatarURL        string     `json:"avatar_url,omitempty" db:"-"`
	Providers        []string   `json:"providers,omitempty" db:"-"`
	MFAEnabled       bool       `json:"mfa_enabled" db:"-"`
	EmailVerified    bool       `json:"email_verified" db:"-"`
	PhoneVerified    bool       `json:"phone_verified" db:"-"`
	PasskeyCount     int        `json:"passkey_count" db:"-"`

	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// Session represents an active user session
type Session struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	UserAgent    string    `json:"user_agent,omitempty" db:"user_agent"`
	IPAddress    string    `json:"ip,omitempty" db:"ip"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
}

// OAuthAccount links an external OAuth provider to a user
type OAuthAccount struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Provider     string    `json:"provider" db:"provider"` // "google" | "github" | ...
	ProviderID   string    `json:"provider_id" db:"provider_id"`
	AccessToken  string    `json:"-" db:"access_token"`
	RefreshToken string    `json:"-" db:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// AuditLog tracks authentication events
type AuditLog struct {
	ID        string    `json:"id" db:"id"`
	UserID    *string   `json:"user_id,omitempty" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	IPAddress *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string   `json:"user_agent,omitempty" db:"user_agent"`
	Details   JSONB     `json:"details,omitempty" db:"details"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// MFAFactor represents a user's 2FA mechanism
type MFAFactor struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"user_id" db:"user_id"`
	FactorType string    `json:"factor_type" db:"factor_type"` // e.g. "totp", "sms"
	Status     string    `json:"status" db:"status"`           // "verified", "unverified"
	Secret     string    `json:"-" db:"secret"`
	LastUsedAt *time.Time`json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// Passkey represents a WebAuthn credential
type Passkey struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	CredentialID string    `json:"credential_id" db:"credential_id"`
	PublicKey    []byte    `json:"-" db:"public_key"`
	AAGUID       string    `json:"aaguid,omitempty" db:"aaguid"`
	SignCount    int       `json:"sign_count" db:"sign_count"`
	LastUsedAt   *time.Time`json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// RateLimit represents an API rate limit status
type RateLimit struct {
	Key     string    `json:"key"`
	Limit   int       `json:"limit"`
	Warning int       `json:"warning,omitempty"`
	Count   int       `json:"count"`
	ResetSeconds int  `json:"reset_seconds"`
}

// ─── Storage Models ───────────────────────────────────────────────────────────

// Bucket represents a storage bucket
type Bucket struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Public         bool      `json:"public" db:"public"`
	AllowedMIMEs   []string  `json:"allowed_mime_types,omitempty" db:"allowed_mime_types"`
	FileSizeLimit  *int64    `json:"file_size_limit,omitempty" db:"file_size_limit"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}

// StorageObject represents a file stored in OmniBase storage
type StorageObject struct {
	ID           string    `json:"id" db:"id"`
	BucketID     string    `json:"bucket_id" db:"bucket_id"`
	Name         string    `json:"name" db:"name"`
	Owner        string    `json:"owner,omitempty" db:"owner"`
	ContentType  string    `json:"content_type" db:"content_type"`
	Size         int64     `json:"size" db:"size"`
	Metadata     JSONB     `json:"metadata,omitempty" db:"metadata"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// ─── Project Models ───────────────────────────────────────────────────────────

// Project represents an OmniBase project (for multi-project support in Phase 6)
type Project struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	AnonKey     string    `json:"anon_key" db:"anon_key"`
	ServiceKey  string    `json:"-" db:"service_key"`
	DatabaseURL string    `json:"-" db:"database_url"`
	Region      string    `json:"region" db:"region"`
	Status      string    `json:"status" db:"status"` // "active" | "paused" | "deleting"
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// IdentityProvider represents an external auth provider (Social, SAML, etc)
type IdentityProvider struct {
	ID           string `json:"id" db:"id"`
	ProjectID    string `json:"project_id" db:"project_id"`
	Type         string `json:"type" db:"type"` // "oauth" | "saml"
	Name         string `json:"name" db:"name"` // "google", "github", etc
	ClientID     string `json:"client_id,omitempty" db:"client_id"`
	ClientSecret string `json:"client_secret,omitempty" db:"client_secret"`
	MetadataURL  string `json:"metadata_url,omitempty" db:"metadata_url"`
	IsActive     bool   `json:"is_active" db:"is_active"`
}

// ─── API Response Models ──────────────────────────────────────────────────────

// APIError is the standard error response format across all services
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

// PaginatedResponse wraps list results with pagination metadata
type PaginatedResponse[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPages int   `json:"total_pages"`
}

// JSONB is a convenience type for Postgres JSONB columns
type JSONB map[string]any
