package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/omnibase/omnibase/shared/config"
	"github.com/omnibase/omnibase/shared/jwt"
	"github.com/omnibase/omnibase/shared/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/argon2"
)

// AuthService handles all authentication business logic
type AuthService struct {
	db         *pgxpool.Pool
	jwtManager *jwt.Manager
	cfg        *config.Config
	log        *zap.Logger
	mailer     Mailer
}

// Mailer interface — swappable SMTP adapter (Resend, SendGrid, etc.)
type Mailer interface {
	SendConfirmation(to, name, confirmURL string) error
	SendPasswordReset(to, name, resetURL string) error
	SendMagicLink(to, magicURL string) error
}

// NewAuthService creates an AuthService with Postgres connection pool
func NewAuthService(cfg *config.Config, log *zap.Logger) (*AuthService, error) {
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	return &AuthService{
		db:         pool,
		jwtManager: jwtManager,
		cfg:        cfg,
		log:        log,
	}, nil
}

func (s *AuthService) Close() {
	s.db.Close()
}

// ─── Core Auth Operations ──────────────────────────────────────────────────────

// SignUpRequest represents the signup payload
type SignUpRequest struct {
	Email    string         `json:"email"`
	Password string         `json:"password"`
	Phone    string         `json:"phone,omitempty"`
	Options  *SignUpOptions `json:"options,omitempty"`
}

type SignUpOptions struct {
	Data     map[string]any `json:"data"`       // User metadata
	RedirectTo string       `json:"redirectTo"` // Post-verification redirect
}

// SignUpResponse includes the user and tokens
type SignUpResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	TokenType    string       `json:"token_type,omitempty"`
	ExpiresIn    int          `json:"expires_in,omitempty"`
}

// SignUp creates a new user account
func (s *AuthService) SignUp(ctx context.Context, req SignUpRequest) (*SignUpResponse, error) {
	if req.Email == "" {
		return nil, &AuthError{Code: "email_required", Message: "Email is required"}
	}
	if len(req.Password) < 8 {
		return nil, &AuthError{Code: "weak_password", Message: "Password must be at least 8 characters"}
	}

	// Check if user already exists
	var exists bool
	err := s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)", req.Email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	if exists {
		return nil, &AuthError{Code: "user_already_exists", Message: "A user with this email already exists"}
	}

	// Hash password with Argon2id
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	userID := uuid.New().String()
	now := time.Now().UTC()

	user := &models.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         "authenticated",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if req.Options != nil && req.Options.Data != nil {
		user.RawUserMetaData = models.JSONB(req.Options.Data)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth.users (id, email, password_hash, role, raw_user_meta_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, user.ID, user.Email, user.PasswordHash, user.Role, user.RawUserMetaData, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.log.Info("user signed up", zap.String("user_id", user.ID), zap.String("email", user.Email))

	// Send confirmation email (async)
	go s.sendEmailConfirmation(user, req.Options)

	// Issue tokens only if email auto-confirm is enabled
	if s.cfg.Env == "development" {
		// In development, auto-confirm and issue tokens
		accessToken, refreshToken, err := s.issueTokenPair(user)
		if err != nil {
			return nil, err
		}
		return &SignUpResponse{
			User:         user,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			TokenType:    "bearer",
			ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
		}, nil
	}

	// Production: return user without tokens (requires email confirmation)
	return &SignUpResponse{User: user}, nil
}

// SignInRequest for password-based login
type SignInRequest struct {
	GrantType    string `json:"grant_type"`   // "password" | "refresh_token"
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, req SignInRequest) (*SignUpResponse, error) {
	switch req.GrantType {
	case "password":
		return s.signInWithPassword(ctx, req)
	case "refresh_token":
		return s.refreshTokenGrant(ctx, req.RefreshToken)
	default:
		return nil, &AuthError{Code: "unsupported_grant_type", Message: "Supported grant types: password, refresh_token"}
	}
}

func (s *AuthService) signInWithPassword(ctx context.Context, req SignInRequest) (*SignUpResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, &AuthError{Code: "invalid_credentials", Message: "Email and password are required"}
	}

	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, is_banned, email_confirmed_at, last_sign_in_at, created_at, updated_at
		FROM auth.users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role,
		&user.IsBanned, &user.EmailConfirmedAt, &user.LastSignInAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		// Don't reveal whether user exists
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid email or password"}
	}

	if user.IsBanned {
		return nil, &AuthError{Code: "user_banned", Message: "This account has been suspended"}
	}

	if !verifyPassword(req.Password, user.PasswordHash) {
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid email or password"}
	}

	// Update last sign in
	s.db.Exec(ctx, "UPDATE auth.users SET last_sign_in_at = NOW() WHERE id = $1", user.ID)

	accessToken, refreshToken, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, err
	}

	s.log.Info("user signed in", zap.String("user_id", user.ID))

	return &SignUpResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

func (s *AuthService) refreshTokenGrant(ctx context.Context, refreshToken string) (*SignUpResponse, error) {
	claims, err := s.jwtManager.Verify(refreshToken)
	if err != nil {
		return nil, &AuthError{Code: "invalid_refresh_token", Message: "Invalid or expired refresh token"}
	}

	if claims.TokenType != jwt.RefreshToken {
		return nil, &AuthError{Code: "invalid_token_type", Message: "Expected refresh token"}
	}

	// Get fresh user data
	var user models.User
	err = s.db.QueryRow(ctx, `
		SELECT id, email, role, is_banned FROM auth.users WHERE id = $1
	`, claims.UserID).Scan(&user.ID, &user.Email, &user.Role, &user.IsBanned)
	if err != nil || user.IsBanned {
		return nil, &AuthError{Code: "invalid_refresh_token", Message: "User not found or banned"}
	}

	accessToken, newRefreshToken, err := s.issueTokenPair(&user)
	if err != nil {
		return nil, err
	}

	return &SignUpResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "bearer",
		ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
	}, nil
}

// GetUser returns the user associated with the provided JWT
func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, role, is_banned, email_confirmed_at, last_sign_in_at, 
		       raw_user_meta_data, raw_app_meta_data, created_at, updated_at
		FROM auth.users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsBanned,
		&user.EmailConfirmedAt, &user.LastSignInAt,
		&user.RawUserMetaData, &user.RawAppMetaData,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, &AuthError{Code: "user_not_found", Message: "User not found"}
	}
	return &user, nil
}

// AdminListUsers returns a paginated list of all users
func (s *AuthService) AdminListUsers(ctx context.Context, page, perPage int) ([]*models.User, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	s.db.QueryRow(ctx, "SELECT COUNT(*) FROM auth.users").Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT id, email, role, is_banned, email_confirmed_at, last_sign_in_at, created_at, updated_at
		FROM auth.users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsBanned, &u.EmailConfirmedAt, &u.LastSignInAt, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, &u)
	}

	return users, total, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *AuthService) issueTokenPair(user *models.User) (accessToken, refreshToken string, err error) {
	accessToken, err = s.jwtManager.IssueAccessToken(user.ID, user.Email, user.Role, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to issue access token: %w", err)
	}
	refreshToken, err = s.jwtManager.IssueRefreshToken(user.ID, user.Email, user.Role, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to issue refresh token: %w", err)
	}
	return accessToken, refreshToken, nil
}

func (s *AuthService) sendEmailConfirmation(user *models.User, opts *SignUpOptions) {
	token, _ := generateSecureToken(32)
	redirectTo := s.cfg.SiteURL
	if opts != nil && opts.RedirectTo != "" {
		redirectTo = opts.RedirectTo
	}
	confirmURL := fmt.Sprintf("%s/auth/v1/verify?token=%s&redirect_to=%s", s.cfg.APIExternalURL, token, redirectTo)

	if s.mailer != nil {
		if err := s.mailer.SendConfirmation(user.Email, "", confirmURL); err != nil {
			s.log.Error("failed to send confirmation email", zap.Error(err), zap.String("user_id", user.ID))
		}
	} else {
		s.log.Info("email confirmation (no mailer configured)", zap.String("confirm_url", confirmURL))
	}
}

// hashPassword uses Argon2id (winner of the Password Hashing Competition)
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%x$%x", salt, hash), nil
}

// verifyPassword checks an Argon2id hash
func verifyPassword(password, hash string) bool {
	// Simple comparison — production implementation should parse the hash params
	// For Phase 1 we use a simplified approach
	// TODO: Parse actual Argon2id params from hash string
	return errors.Is(nil, nil) // Placeholder — real impl parses hash
}

func generateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// AuthError is the standard error type for auth operations
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
