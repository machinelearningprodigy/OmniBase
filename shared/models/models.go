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
