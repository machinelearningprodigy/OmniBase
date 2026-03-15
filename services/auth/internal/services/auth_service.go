package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"github.com/machinelearningprodigy/OmniBase/shared/models"
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
	client     *http.Client
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

	service := &AuthService{
		db:         pool,
		jwtManager: jwtManager,
		cfg:        cfg,
		log:        log,
		client:     &http.Client{Timeout: 20 * time.Second},
	}
	if err := service.ensureSystemTables(context.Background()); err != nil {
		pool.Close()
		return nil, err
	}
	return service, nil
}

func (s *AuthService) Close() {
	s.db.Close()
}

func (s *AuthService) VerifyToken(token string) (*jwt.Claims, error) {
	return s.jwtManager.Verify(token)
}

// ─── Core Auth Operations ──────────────────────────────────────────────────────

// SignUpRequest represents the signup payload
type SignUpRequest struct {
	Email    string         `json:"email"`
	Password string         `json:"password"`
	Phone     string         `json:"phone,omitempty"`
	ProjectID string         `json:"project_id,omitempty"`
	Options   *SignUpOptions `json:"options,omitempty"`
}

type SignUpOptions struct {
	Data       map[string]any `json:"data"`       // User metadata
	RedirectTo string         `json:"redirectTo"` // Post-verification redirect
}

type ActionLinkType string

const (
	ActionLinkSignup      ActionLinkType = "signup"
	ActionLinkInvite      ActionLinkType = "invite"
	ActionLinkRecovery    ActionLinkType = "recovery"
	ActionLinkMagicLogin  ActionLinkType = "magic_link"
)

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
	var userCount int64
	if err := s.db.QueryRow(ctx, "SELECT COUNT(*) FROM auth.users").Scan(&userCount); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	user := &models.User{
		ID:           userID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         "authenticated",
		IsSuperAdmin: userCount == 0,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if s.cfg.Env == "development" {
		user.EmailConfirmedAt = &now
	}

	if req.Options != nil && req.Options.Data != nil {
		user.RawUserMetaData = models.JSONB(req.Options.Data)
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth.users (
			id, email, password_hash, role, raw_user_meta_data, is_super_admin, email_confirmed_at, project_id, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, user.ID, user.Email, user.PasswordHash, user.Role, user.RawUserMetaData, user.IsSuperAdmin, user.EmailConfirmedAt, req.ProjectID, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	s.log.Info("user signed up", zap.String("user_id", user.ID), zap.String("email", user.Email))

	// Send confirmation email (async)
	go s.sendEmailConfirmation(user, req.Options)

	// Issue tokens only if email auto-confirm is enabled
	if s.cfg.Env == "development" {
		// In development, auto-confirm and issue tokens
		accessToken, refreshToken, err := s.IssueTokenPair(ctx, user, "", "")
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
	GrantType    string `json:"grant_type"` // "password" | "refresh_token"
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// SignIn authenticates a user and returns tokens
func (s *AuthService) SignIn(ctx context.Context, req SignInRequest, userAgent, ip string) (*SignUpResponse, error) {
	switch req.GrantType {
	case "password":
		return s.signInWithPassword(ctx, req, userAgent, ip)
	case "refresh_token":
		return s.refreshTokenGrant(ctx, req.RefreshToken)
	default:
		return nil, &AuthError{Code: "unsupported_grant_type", Message: "Supported grant types: password, refresh_token"}
	}
}

func (s *AuthService) signInWithPassword(ctx context.Context, req SignInRequest, userAgent, ip string) (*SignUpResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, &AuthError{Code: "invalid_credentials", Message: "Email and password are required"}
	}

	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, password_hash, role, is_super_admin, is_banned, email_confirmed_at, last_sign_in_at, created_at, updated_at
		FROM auth.users WHERE email = $1
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.IsSuperAdmin,
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
	if s.cfg.Env != "development" && user.EmailConfirmedAt == nil {
		return nil, &AuthError{Code: "email_not_confirmed", Message: "Email address has not been confirmed"}
	}

	if !verifyPassword(req.Password, user.PasswordHash) {
		return nil, &AuthError{Code: "invalid_credentials", Message: "Invalid email or password"}
	}

	// Update last sign in
	s.db.Exec(ctx, "UPDATE auth.users SET last_sign_in_at = NOW() WHERE id = $1", user.ID)

	accessToken, refreshToken, err := s.IssueTokenPair(ctx, &user, userAgent, ip)
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
		SELECT id, email, role, is_super_admin, is_banned FROM auth.users WHERE id = $1
	`, claims.UserID).Scan(&user.ID, &user.Email, &user.Role, &user.IsSuperAdmin, &user.IsBanned)
	if err != nil || user.IsBanned {
		return nil, &AuthError{Code: "invalid_refresh_token", Message: "User not found or banned"}
	}

	var sessionExists bool
	if err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM auth.sessions
			WHERE user_id = $1 AND refresh_token = $2 AND expires_at > NOW()
		)
	`, claims.UserID, refreshToken).Scan(&sessionExists); err != nil || !sessionExists {
		return nil, &AuthError{Code: "invalid_refresh_token", Message: "Refresh session not found or expired"}
	}

	if err := s.revokeRefreshToken(ctx, refreshToken); err != nil {
		s.log.Warn("failed to revoke previous refresh token", zap.Error(err), zap.String("user_id", claims.UserID))
	}

	accessToken, newRefreshToken, err := s.IssueTokenPair(ctx, &user, "", "")
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
func (s *AuthService) DB() *pgxpool.Pool      { return s.db }
func (s *AuthService) Config() *config.Config { return s.cfg }

func (s *AuthService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, role, is_banned, email_confirmed_at, last_sign_in_at, 
		       raw_user_meta_data, raw_app_meta_data, is_super_admin, created_at, updated_at
		FROM auth.users WHERE email = $1
	`, email).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsBanned,
		&user.EmailConfirmedAt, &user.LastSignInAt,
		&user.RawUserMetaData, &user.RawAppMetaData, &user.IsSuperAdmin,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, &AuthError{Code: "user_not_found", Message: "User not found"}
	}
	return &user, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, role, is_banned, email_confirmed_at, last_sign_in_at, 
		       raw_user_meta_data, raw_app_meta_data, is_super_admin, created_at, updated_at
		FROM auth.users WHERE id = $1
	`, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.IsBanned,
		&user.EmailConfirmedAt, &user.LastSignInAt,
		&user.RawUserMetaData, &user.RawAppMetaData, &user.IsSuperAdmin,
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
		SELECT id, email, role, is_super_admin, is_banned, email_confirmed_at, last_sign_in_at, created_at, updated_at
		FROM auth.users ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsSuperAdmin, &u.IsBanned, &u.EmailConfirmedAt, &u.LastSignInAt, &u.CreatedAt, &u.UpdatedAt)
		users = append(users, &u)
	}

	return users, total, nil
}

// AdminDeleteUser permanently deletes a user
func (s *AuthService) AdminDeleteUser(ctx context.Context, userID string) error {
	result, err := s.db.Exec(ctx, "DELETE FROM auth.users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return &AuthError{Code: "user_not_found", Message: "User not found"}
	}
	s.log.Info("user deleted by admin", zap.String("user_id", userID))
	return nil
}

// AdminBanUser bans or unbans a user
func (s *AuthService) AdminBanUser(ctx context.Context, userID string, banned bool) error {
	result, err := s.db.Exec(ctx, "UPDATE auth.users SET is_banned = $1, updated_at = NOW() WHERE id = $2", banned, userID)
	if err != nil {
		return fmt.Errorf("failed to update user ban status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return &AuthError{Code: "user_not_found", Message: "User not found"}
	}
	action := "banned"
	if !banned {
		action = "unbanned"
	}
	s.log.Info("user "+action+" by admin", zap.String("user_id", userID))
	return nil
}

// AdminUpdateUser updates user metadata
func (s *AuthService) AdminUpdateUser(ctx context.Context, userID string, updates AdminUpdateRequest) (*models.User, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if updates.Role != "" {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argNum))
		args = append(args, updates.Role)
		argNum++
	}
	if updates.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argNum))
		args = append(args, updates.Email)
		argNum++
	}

	if len(setClauses) == 0 {
		return s.GetUser(ctx, userID)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, userID)

	sql := fmt.Sprintf("UPDATE auth.users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argNum)
	_, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// AdminInviteUser creates a user and sends them an invite email
func (s *AuthService) AdminInviteUser(ctx context.Context, email string) (*models.User, error) {
	if email == "" {
		return nil, &AuthError{Code: "email_required", Message: "Email is required"}
	}

	var exists bool
	s.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM auth.users WHERE email = $1)", email).Scan(&exists)
	if exists {
		return nil, &AuthError{Code: "user_already_exists", Message: "A user with this email already exists"}
	}

	// Generate a random temporary password
	tempPwd, _ := GenerateSecureToken(16)
	passwordHash, err := hashPassword(tempPwd)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New().String()
	now := time.Now().UTC()

	user := &models.User{
		ID:           userID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         "authenticated",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth.users (id, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, user.ID, user.Email, user.PasswordHash, user.Role, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create invited user: %w", err)
	}

	s.log.Info("user invited by admin", zap.String("user_id", user.ID), zap.String("email", email))
	if link, linkErr := s.GenerateActionLink(ctx, ActionLinkInvite, user.ID, user.Email, s.cfg.SiteURL); linkErr == nil {
		s.log.Info("invite link generated", zap.String("user_id", user.ID), zap.String("invite_url", link))
	}
	return user, nil
}

// AdminUpdateRequest for admin user updates
type AdminUpdateRequest struct {
	Email        string         `json:"email"`
	Role         string         `json:"role"`
	IsSuperAdmin *bool          `json:"is_super_admin,omitempty"`
	UserMetadata map[string]any `json:"user_metadata,omitempty"`
}

type UpdateUserRequest struct {
	Email        string         `json:"email,omitempty"`
	Password     string         `json:"password,omitempty"`
	UserMetadata map[string]any `json:"data,omitempty"`
}

func (s *AuthService) UpdateUser(ctx context.Context, userID string, updates UpdateUserRequest) (*models.User, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if updates.Email != "" {
		setClauses = append(setClauses, fmt.Sprintf("email = $%d", argNum))
		args = append(args, updates.Email)
		argNum++
	}
	if updates.Password != "" {
		if len(updates.Password) < 8 {
			return nil, &AuthError{Code: "weak_password", Message: "Password must be at least 8 characters"}
		}
		passwordHash, err := hashPassword(updates.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		setClauses = append(setClauses, fmt.Sprintf("password_hash = $%d", argNum))
		args = append(args, passwordHash)
		argNum++
	}
	if updates.UserMetadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("raw_user_meta_data = $%d", argNum))
		args = append(args, models.JSONB(updates.UserMetadata))
		argNum++
	}

	if len(setClauses) == 0 {
		return s.GetUser(ctx, userID)
	}

	setClauses = append(setClauses, "updated_at = NOW()")
	args = append(args, userID)

	sql := fmt.Sprintf("UPDATE auth.users SET %s WHERE id = $%d", strings.Join(setClauses, ", "), argNum)
	result, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return nil, &AuthError{Code: "user_not_found", Message: "User not found"}
	}

	return s.GetUser(ctx, userID)
}

func (s *AuthService) IsAdmin(ctx context.Context, userID string) (bool, error) {
	var isSuperAdmin bool
	err := s.db.QueryRow(ctx, "SELECT is_super_admin FROM auth.users WHERE id = $1", userID).Scan(&isSuperAdmin)
	if err != nil {
		return false, err
	}
	return isSuperAdmin, nil
}

func (s *AuthService) CreateRecoveryToken(ctx context.Context, email, redirectTo string) error {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, is_banned
		FROM auth.users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.IsBanned)
	if err != nil || user.IsBanned {
		return nil
	}
	if redirectTo == "" {
		redirectTo = s.cfg.SiteURL + "/auth/reset-password"
	}
	token, err := GenerateSecureToken(32)
	if err != nil {
		return err
	}
	tokenHash := hashToken(token)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	if _, err := s.db.Exec(ctx, `
		INSERT INTO omnibase.auth_flow_tokens (token_hash, user_id, email, token_type, redirect_to, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, tokenHash, user.ID, user.Email, string(ActionLinkRecovery), redirectTo, expiresAt); err != nil {
		return err
	}
	// Link goes directly to client with token so user can submit new password (token consumed on reset)
	clientLink := redirectTo
	if strings.Contains(redirectTo, "?") {
		clientLink = redirectTo + "&token=" + token
	} else {
		clientLink = redirectTo + "?token=" + token
	}
	if s.mailer != nil {
		return s.mailer.SendPasswordReset(user.Email, "", clientLink)
	}
	s.log.Info("password recovery link generated", zap.String("email", user.Email), zap.String("reset_url", clientLink))
	return nil
}

// CreateMagicLink sends a passwordless sign-in link to the given email (if user exists).
func (s *AuthService) CreateMagicLink(ctx context.Context, email, redirectTo string) error {
	var user models.User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, is_banned
		FROM auth.users
		WHERE email = $1
	`, email).Scan(&user.ID, &user.Email, &user.IsBanned)
	if err != nil || user.IsBanned {
		return nil // Do not reveal whether user exists
	}

	link, err := s.GenerateActionLink(ctx, ActionLinkMagicLogin, user.ID, user.Email, redirectTo)
	if err != nil {
		return err
	}

	if s.mailer != nil {
		return s.mailer.SendMagicLink(user.Email, link)
	}
	s.log.Info("magic link generated", zap.String("email", user.Email), zap.String("link", link))
	return nil
}

// VerifyMagicLink consumes a magic_link token and returns the user for issuing a session.
func (s *AuthService) VerifyMagicLink(ctx context.Context, token string) (*models.User, string, error) {
	record, err := s.consumeFlowToken(ctx, token, ActionLinkMagicLogin)
	if err != nil {
		return nil, "", err
	}
	var user models.User
	err = s.db.QueryRow(ctx, `
		SELECT id, email, role, is_super_admin, is_banned, created_at, updated_at
		FROM auth.users WHERE id = $1
	`, record.UserID).Scan(&user.ID, &user.Email, &user.Role, &user.IsSuperAdmin, &user.IsBanned, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, "", err
	}
	if user.IsBanned {
		return nil, "", &AuthError{Code: "user_banned", Message: "Account is suspended"}
	}
	return &user, record.RedirectTo, nil
}

func (s *AuthService) VerifyEmailToken(ctx context.Context, token string, tokenType ActionLinkType) (string, error) {
	record, err := s.consumeFlowToken(ctx, token, tokenType)
	if err != nil {
		return "", err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE auth.users
		SET email_confirmed_at = COALESCE(email_confirmed_at, NOW()), updated_at = NOW()
		WHERE id = $1
	`, record.UserID); err != nil {
		return "", err
	}

	return record.RedirectTo, nil
}

func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if len(newPassword) < 8 {
		return &AuthError{Code: "weak_password", Message: "Password must be at least 8 characters"}
	}

	record, err := s.consumeFlowToken(ctx, token, ActionLinkRecovery)
	if err != nil {
		return err
	}

	passwordHash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	if _, err := s.db.Exec(ctx, `
		UPDATE auth.users
		SET password_hash = $1, updated_at = NOW()
		WHERE id = $2
	`, passwordHash, record.UserID); err != nil {
		return err
	}

	return s.RevokeAllSessions(ctx, record.UserID)
}

func (s *AuthService) RevokeRefreshToken(ctx context.Context, refreshToken string) error {
	return s.revokeRefreshToken(ctx, refreshToken)
}

func (s *AuthService) RevokeAllSessions(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.sessions WHERE user_id = $1", userID)
	return err
}

func (s *AuthService) GenerateActionLink(ctx context.Context, tokenType ActionLinkType, userID, email, redirectTo string) (string, error) {
	if redirectTo == "" {
		redirectTo = s.cfg.SiteURL
	}

	token, err := GenerateSecureToken(32)
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	if tokenType == ActionLinkRecovery || tokenType == ActionLinkMagicLogin {
		expiresAt = time.Now().UTC().Add(1 * time.Hour)
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO omnibase.auth_flow_tokens (token_hash, user_id, email, token_type, redirect_to, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, hashToken(token), userID, email, string(tokenType), redirectTo, expiresAt); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/auth/v1/verify?token=%s&type=%s&redirect_to=%s", s.cfg.APIExternalURL, token, tokenType, redirectTo), nil
}

func (s *AuthService) ListProviders() []OAuthProvider {
	providers := []OAuthProvider{
		{
			Name:         "google",
			Enabled:      s.cfg.GoogleClientID != "" && s.cfg.GoogleClientSecret != "",
			ClientID:     s.cfg.GoogleClientID,
			AuthorizeURL: fmt.Sprintf("%s/auth/v1/authorize?provider=google", s.cfg.APIExternalURL),
		},
		{
			Name:         "github",
			Enabled:      s.cfg.GitHubClientID != "" && s.cfg.GitHubClientSecret != "",
			ClientID:     s.cfg.GitHubClientID,
			AuthorizeURL: fmt.Sprintf("%s/auth/v1/authorize?provider=github", s.cfg.APIExternalURL),
		},
	}
	return providers
}

func (s *AuthService) OAuthAuthorizeURL(ctx context.Context, provider, redirectTo, projectID string) (string, error) {
	cfg, err := s.providerConfig(ctx, provider, projectID)
	if err != nil {
		return "", err
	}

	state, err := GenerateSecureToken(24)
	if err != nil {
		return "", err
	}
	if redirectTo == "" {
		redirectTo = s.cfg.SiteURL
	}

	if _, err := s.db.Exec(ctx, `
		INSERT INTO omnibase.oauth_states (state_hash, provider, redirect_to, project_id, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`, hashToken(state), provider, redirectTo, projectID, time.Now().UTC().Add(10*time.Minute)); err != nil {
		return "", err
	}

	values := url.Values{}
	values.Set("client_id", cfg.ClientID)
	values.Set("redirect_uri", s.oauthCallbackURL(provider))
	values.Set("response_type", "code")
	values.Set("state", state)
	values.Set("scope", cfg.Scope)
	if provider == "google" {
		values.Set("access_type", "offline")
		values.Set("prompt", "consent")
	}

	return cfg.AuthURL + "?" + values.Encode(), nil
}

func (s *AuthService) CompleteOAuth(ctx context.Context, provider, code, state string) (*SignUpResponse, string, error) {
	stateRecord, err := s.consumeOAuthState(ctx, provider, state)
	if err != nil {
		return nil, "", err
	}

	cfg, err := s.providerConfig(ctx, provider, stateRecord.ProjectID)
	if err != nil {
		return nil, "", err
	}

	tokenResp, err := s.exchangeOAuthCode(ctx, cfg, provider, code)
	if err != nil {
		return nil, "", err
	}

	email, providerID, err := s.fetchOAuthIdentity(ctx, cfg, provider, tokenResp.AccessToken)
	if err != nil {
		return nil, "", err
	}

	user, err := s.findOrCreateOAuthUser(ctx, provider, providerID, email, tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn)
	if err != nil {
		return nil, "", err
	}

	accessToken, refreshToken, err := s.IssueTokenPair(ctx, user, "", "")
	if err != nil {
		return nil, "", err
	}

	return &SignUpResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    int(s.cfg.JWTAccessExpiry.Seconds()),
	}, stateRecord.RedirectTo, nil
}

// IssueTokenPairForUser issues access and refresh tokens for a user (e.g. after magic link verify).
func (s *AuthService) IssueTokenPairForUser(ctx context.Context, user *models.User) (*SignUpResponse, error) {
	accessToken, refreshToken, err := s.IssueTokenPair(ctx, user, "", "")
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

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *AuthService) IssueTokenPair(ctx context.Context, user *models.User, userAgent, ip string) (accessToken, refreshToken string, err error) {
	accessToken, err = s.jwtManager.IssueAccessToken(user.ID, user.Email, user.Role, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to issue access token: %w", err)
	}
	refreshToken, err = s.jwtManager.IssueRefreshToken(user.ID, user.Email, user.Role, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to issue refresh token: %w", err)
	}
	if err := s.persistSession(ctx, user.ID, refreshToken, userAgent, ip); err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (s *AuthService) sendEmailConfirmation(user *models.User, opts *SignUpOptions) {
	redirectTo := s.cfg.SiteURL
	if opts != nil && opts.RedirectTo != "" {
		redirectTo = opts.RedirectTo
	}
	confirmURL, err := s.GenerateActionLink(context.Background(), ActionLinkSignup, user.ID, user.Email, redirectTo)
	if err != nil {
		s.log.Error("failed to generate confirmation link", zap.Error(err), zap.String("user_id", user.ID))
		return
	}

	if s.mailer != nil {
		if err := s.mailer.SendConfirmation(user.Email, "", confirmURL); err != nil {
			s.log.Error("failed to send confirmation email", zap.Error(err), zap.String("user_id", user.ID))
		}
	} else {
		s.log.Info("email confirmation (no mailer configured)", zap.String("confirm_url", confirmURL))
	}
}

// hashPassword uses Argon2id (winner of the Password Hashing Competition)
// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt_hex>$<hash_hex>
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=1,p=4$%x$%x", salt, hash), nil
}

// verifyPassword checks an Argon2id hash
// Format: $argon2id$v=19$m=65536,t=1,p=4$<salt_hex>$<hash_hex>
func verifyPassword(password, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	// Expected: ["", "argon2id", "v=19", "m=65536,t=1,p=4", "<salt_hex>", "<hash_hex>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}

	saltBytes, err := hex.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expectedHashBytes, err := hex.DecodeString(parts[5])
	if err != nil {
		return false
	}

	// Re-hash with same parameters
	computedHash := argon2.IDKey([]byte(password), saltBytes, 1, 64*1024, 4, 32)

	// Constant-time comparison
	return errors.Is(nil, nil) && safeCompare(computedHash, expectedHashBytes)
}

// safeCompare performs a constant-time byte slice comparison
func safeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

func GenerateSecureToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type flowTokenRecord struct {
	UserID     string
	RedirectTo string
}

type OAuthProvider struct {
	Name         string `json:"name"`
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"client_id,omitempty"`
	AuthorizeURL string `json:"authorize_url,omitempty"`
}

type oauthStateRecord struct {
	Provider   string
	RedirectTo string
	ProjectID  string
}

type oauthProviderConfig struct {
	ClientID     string
	ClientSecret string
	AuthURL      string
	TokenURL     string
	UserURL      string
	Scope        string
}

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func (s *AuthService) ensureSystemTables(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'auth') THEN
				EXECUTE 'CREATE SCHEMA auth';
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'omnibase') THEN
				EXECUTE 'CREATE SCHEMA omnibase';
			END IF;
		END
		$$;
		CREATE TABLE IF NOT EXISTS auth.users (
			id UUID PRIMARY KEY,
			email TEXT UNIQUE,
			phone TEXT UNIQUE,
			password_hash TEXT,
			role TEXT NOT NULL DEFAULT 'authenticated',
			raw_user_meta_data JSONB DEFAULT '{}'::jsonb,
			raw_app_meta_data JSONB DEFAULT '{}'::jsonb,
			is_super_admin BOOLEAN DEFAULT FALSE,
			is_banned BOOLEAN DEFAULT FALSE,
			email_confirmed_at TIMESTAMPTZ,
			phone_confirmed_at TIMESTAMPTZ,
			last_sign_in_at TIMESTAMPTZ,
			banned_until TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS auth.sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
			refresh_token TEXT NOT NULL UNIQUE,
			user_agent TEXT,
			ip INET,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE TABLE IF NOT EXISTS auth.oauth_accounts (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
			provider TEXT NOT NULL,
			provider_id TEXT NOT NULL,
			access_token TEXT,
			refresh_token TEXT,
			expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(provider, provider_id)
		);
		CREATE INDEX IF NOT EXISTS auth_users_email_idx ON auth.users(email);
		CREATE INDEX IF NOT EXISTS auth_users_created_at_idx ON auth.users(created_at DESC);
		CREATE INDEX IF NOT EXISTS auth_sessions_user_id_idx ON auth.sessions(user_id);
		CREATE INDEX IF NOT EXISTS auth_sessions_refresh_token_idx ON auth.sessions(refresh_token);
		CREATE TABLE IF NOT EXISTS omnibase.auth_flow_tokens (
			token_hash TEXT PRIMARY KEY,
			user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
			email TEXT,
			token_type TEXT NOT NULL,
			redirect_to TEXT,
			expires_at TIMESTAMPTZ NOT NULL,
			consumed_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS auth_flow_tokens_user_id_idx ON omnibase.auth_flow_tokens(user_id);
		CREATE INDEX IF NOT EXISTS auth_flow_tokens_type_idx ON omnibase.auth_flow_tokens(token_type);
		CREATE TABLE IF NOT EXISTS omnibase.oauth_states (
			state_hash TEXT PRIMARY KEY,
			provider TEXT NOT NULL,
			redirect_to TEXT,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS oauth_states_provider_idx ON omnibase.oauth_states(provider);

		CREATE TABLE IF NOT EXISTS auth.audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID REFERENCES auth.users(id) ON DELETE SET NULL,
			action TEXT NOT NULL,
			ip_address INET,
			user_agent TEXT,
			details JSONB,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS audit_logs_user_id_idx ON auth.audit_logs(user_id);

		CREATE TABLE IF NOT EXISTS auth.mfa_factors (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
			factor_type TEXT NOT NULL,
			status TEXT NOT NULL,
			secret TEXT,
			last_used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS mfa_factors_user_id_idx ON auth.mfa_factors(user_id);

		CREATE TABLE IF NOT EXISTS auth.passkeys (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
			credential_id TEXT NOT NULL UNIQUE,
			public_key BYTEA NOT NULL,
			aaguid TEXT,
			sign_count INTEGER DEFAULT 0,
			last_used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS passkeys_user_id_idx ON auth.passkeys(user_id);

		CREATE TABLE IF NOT EXISTS auth.rate_limits (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			key TEXT NOT NULL UNIQUE,
			count INTEGER DEFAULT 1,
			reset_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS auth.settings (
			key TEXT PRIMARY KEY,
			value JSONB NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth.email_templates (
			type TEXT PRIMARY KEY,
			subject TEXT NOT NULL,
			body_html TEXT NOT NULL,
			body_text TEXT,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth.hooks (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			event TEXT NOT NULL, -- e.g., 'before_signup', 'after_login'
			endpoint_url TEXT NOT NULL,
			secret TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth.identity_providers (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			type TEXT NOT NULL, -- 'oauth2', 'saml', 'oidc'
			name TEXT NOT NULL UNIQUE,
			client_id TEXT,
			client_secret TEXT,
			metadata_url TEXT,
			is_active BOOLEAN DEFAULT TRUE,
			config JSONB DEFAULT '{}'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth.banned_ips (
			ip INET PRIMARY KEY,
			reason TEXT,
			expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS auth.web3_nonces (
			wallet_address TEXT PRIMARY KEY,
			nonce TEXT NOT NULL,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *AuthService) persistSession(ctx context.Context, userID, refreshToken, userAgent, ip string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.sessions (user_id, refresh_token, user_agent, ip, expires_at)
		VALUES ($1, $2, $3, NULLIF($4, '')::inet, $5)
	`, userID, refreshToken, userAgent, ip, time.Now().UTC().Add(s.cfg.JWTRefreshExpiry))
	return err
}

func (s *AuthService) revokeRefreshToken(ctx context.Context, refreshToken string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.sessions WHERE refresh_token = $1", refreshToken)
	return err
}

func (s *AuthService) consumeFlowToken(ctx context.Context, token string, tokenType ActionLinkType) (*flowTokenRecord, error) {
	var record flowTokenRecord
	err := s.db.QueryRow(ctx, `
		UPDATE omnibase.auth_flow_tokens
		SET consumed_at = NOW()
		WHERE token_hash = $1
		  AND token_type = $2
		  AND consumed_at IS NULL
		  AND expires_at > NOW()
		RETURNING user_id::text, COALESCE(redirect_to, '')
	`, hashToken(token), string(tokenType)).Scan(&record.UserID, &record.RedirectTo)
	if err != nil {
		return nil, &AuthError{Code: "invalid_token", Message: "Invalid or expired token"}
	}
	return &record, nil
}

func (s *AuthService) consumeOAuthState(ctx context.Context, provider, state string) (*oauthStateRecord, error) {
	var record oauthStateRecord
	err := s.db.QueryRow(ctx, `
		DELETE FROM omnibase.oauth_states
		WHERE state_hash = $1
		  AND provider = $2
		  AND expires_at > NOW()
		RETURNING provider, COALESCE(redirect_to, ''), COALESCE(project_id::text, '')
	`, hashToken(state), provider).Scan(&record.Provider, &record.RedirectTo, &record.ProjectID)
	if err != nil {
		return nil, &AuthError{Code: "invalid_oauth_state", Message: "Invalid or expired OAuth state"}
	}
	return &record, nil
}

func (s *AuthService) providerConfig(ctx context.Context, provider, projectID string) (*oauthProviderConfig, error) {
	// 1. Try DB first
	if projectID != "" {
		var cid, cs string
		var isActive bool
		err := s.db.QueryRow(ctx, `
			SELECT client_id, client_secret, is_active 
			FROM auth.identity_providers 
			WHERE name = $1 AND project_id = $2 AND type = 'oauth'
		`, provider, projectID).Scan(&cid, &cs, &isActive)
		
		if err == nil && isActive && cid != "" && cs != "" {
			conf := s.getStaticProviderConfig(provider)
			if conf != nil {
				conf.ClientID = cid
				conf.ClientSecret = cs
				return conf, nil
			}
		}
	}

	// 2. Fallback to Env vars (Platform level / Default)
	switch provider {
	case "google":
		if s.cfg.GoogleClientID == "" || s.cfg.GoogleClientSecret == "" {
			return nil, &AuthError{Code: "provider_not_configured", Message: "Google OAuth is not configured"}
		}
		return &oauthProviderConfig{
			ClientID:     s.cfg.GoogleClientID,
			ClientSecret: s.cfg.GoogleClientSecret,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			UserURL:      "https://openidconnect.googleapis.com/v1/userinfo",
			Scope:        "openid email profile",
		}, nil
	case "github":
		if s.cfg.GitHubClientID == "" || s.cfg.GitHubClientSecret == "" {
			return nil, &AuthError{Code: "provider_not_configured", Message: "GitHub OAuth is not configured"}
		}
		return &oauthProviderConfig{
			ClientID:     s.cfg.GitHubClientID,
			ClientSecret: s.cfg.GitHubClientSecret,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			UserURL:      "https://api.github.com/user",
			Scope:        "read:user user:email",
		}, nil
	default:
		// Try static config even if not in DB, in case we add more env support later
		static := s.getStaticProviderConfig(provider)
		if static != nil && static.ClientID != "" {
			return static, nil
		}
		return nil, &AuthError{Code: "unsupported_provider", Message: "Unsupported OAuth provider or not configured"}
	}
}

func (s *AuthService) getStaticProviderConfig(provider string) *oauthProviderConfig {
	switch provider {
	case "google":
		return &oauthProviderConfig{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			UserURL:  "https://openidconnect.googleapis.com/v1/userinfo",
			Scope:    "openid email profile",
		}
	case "github":
		return &oauthProviderConfig{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
			UserURL:  "https://api.github.com/user",
			Scope:    "read:user user:email",
		}
	case "line":
		return &oauthProviderConfig{
			AuthURL:  "https://access.line.me/oauth2/v2.1/authorize",
			TokenURL: "https://api.line.me/oauth2/v2.1/token",
			UserURL:  "https://api.line.me/v2/profile",
			Scope:    "profile openid email",
		}
	case "paypal":
		return &oauthProviderConfig{
			AuthURL:  "https://www.paypal.com/signin/authorize",
			TokenURL: "https://api-m.paypal.com/v1/oauth2/token",
			UserURL:  "https://api-m.paypal.com/v1/identity/openidconnect/userinfo",
			Scope:    "openid email profile",
		}
	case "amazon":
		return &oauthProviderConfig{
			AuthURL:  "https://www.amazon.com/ap/oa",
			TokenURL: "https://api.amazon.com/auth/o2/token",
			UserURL:  "https://api.amazon.com/user/profile",
			Scope:    "profile profile:user_id",
		}
	case "tiktok":
		return &oauthProviderConfig{
			AuthURL:  "https://www.tiktok.com/v2/auth/authorize/",
			TokenURL: "https://open.tiktokapis.com/v2/oauth/token/",
			UserURL:  "https://open.tiktokapis.com/v2/user/info/",
			Scope:    "user.info.basic",
		}
	case "pinterest":
		return &oauthProviderConfig{
			AuthURL:  "https://www.pinterest.com/oauth/",
			TokenURL: "https://api.pinterest.com/v5/oauth/token",
			UserURL:  "https://api.pinterest.com/v5/user_account",
			Scope:    "user_accounts:read",
		}
	case "snapchat":
		return &oauthProviderConfig{
			AuthURL:  "https://accounts.snapchat.com/login/oauth2/authorize",
			TokenURL: "https://accounts.snapchat.com/login/oauth2/access_token",
			UserURL:  "https://kit.snapchat.com/v1/me",
			Scope:    "https://auth.snapchat.com/oauth2/api/user.display_name",
		}
	case "yahoo":
		return &oauthProviderConfig{
			AuthURL:  "https://api.login.yahoo.com/oauth2/request_auth",
			TokenURL: "https://api.login.yahoo.com/oauth2/get_token",
			UserURL:  "https://api.login.yahoo.com/openid/v1/userinfo",
			Scope:    "openid profile email",
		}
	case "okta":
		return &oauthProviderConfig{
			AuthURL:  "https://okta.com/oauth2/v1/authorize", // Placeholder, usually tenant-specific
			TokenURL: "https://okta.com/oauth2/v1/token",     // Placeholder, usually tenant-specific
			UserURL:  "https://okta.com/oauth2/v1/userinfo",  // Placeholder, usually tenant-specific
			Scope:    "openid profile email",
		}
	case "yandex":
		return &oauthProviderConfig{
			AuthURL:  "https://oauth.yandex.com/authorize",
			TokenURL: "https://oauth.yandex.com/token",
			UserURL:  "https://login.yandex.ru/info",
			Scope:    "login:email login:info",
		}
	case "wordpress":
		return &oauthProviderConfig{
			AuthURL:  "https://public-api.wordpress.com/oauth2/authorize",
			TokenURL: "https://public-api.wordpress.com/oauth2/token",
			UserURL:  "https://public-api.wordpress.com/rest/v1/me",
			Scope:    "auth",
		}
	case "vk":
		return &oauthProviderConfig{
			AuthURL:  "https://oauth.vk.com/authorize",
			TokenURL: "https://oauth.vk.com/access_token",
			UserURL:  "https://api.vk.com/method/users.get",
			Scope:    "email",
		}
	case "facebook":
		return &oauthProviderConfig{
			AuthURL:  "https://www.facebook.com/v12.0/dialog/oauth",
			TokenURL: "https://graph.facebook.com/v12.0/oauth/access_token",
			UserURL:  "https://graph.facebook.com/me?fields=id,name,email,picture",
			Scope:    "email public_profile",
		}
	case "microsoft":
		return &oauthProviderConfig{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			UserURL:  "https://graph.microsoft.com/v1.0/me",
			Scope:    "openid email profile User.Read",
		}
	case "apple":
		return &oauthProviderConfig{
			AuthURL:  "https://appleid.apple.com/auth/authorize",
			TokenURL: "https://appleid.apple.com/auth/token",
			UserURL:  "", // Apple uses ID tokens
			Scope:    "name email",
		}
	}
	return nil
}

func (s *AuthService) oauthCallbackURL(provider string) string {
	return fmt.Sprintf("%s/auth/v1/callback?provider=%s", s.cfg.APIExternalURL, provider)
}

func (s *AuthService) exchangeOAuthCode(ctx context.Context, cfg *oauthProviderConfig, provider, code string) (*oauthTokenResponse, error) {
	form := url.Values{}
	form.Set("client_id", cfg.ClientID)
	form.Set("client_secret", cfg.ClientSecret)
	form.Set("code", code)
	form.Set("redirect_uri", s.oauthCallbackURL(provider))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, &AuthError{Code: "oauth_exchange_failed", Message: string(body)}
	}

	var tokenResp oauthTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}
	if tokenResp.AccessToken == "" {
		return nil, &AuthError{Code: "oauth_exchange_failed", Message: "OAuth provider did not return an access token"}
	}
	return &tokenResp, nil
}

func (s *AuthService) fetchOIDCUser(ctx context.Context, userURL, accessToken string) (email, providerID string, err error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, userURL, nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := s.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var payload struct {
		Sub           string `json:"sub"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}
	if payload.Email == "" || !payload.EmailVerified {
		// Fallback if email is not directly available or verified, use sub as a placeholder
		if payload.Sub != "" {
			return payload.Sub + "@oidc.invalid", payload.Sub, nil
		}
		return "", "", &AuthError{Code: "oauth_identity_invalid", Message: "OIDC provider did not return a verified email or sub"}
	}
	return payload.Email, payload.Sub, nil
}

func (s *AuthService) fetchOAuthIdentity(ctx context.Context, cfg *oauthProviderConfig, provider, accessToken string) (email, providerID string, err error) {
	switch provider {
	case "google":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var payload struct {
			Sub           string `json:"sub"`
			Email         string `json:"email"`
			EmailVerified bool   `json:"email_verified"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return "", "", err
		}
		if payload.Email == "" || !payload.EmailVerified {
			return "", "", &AuthError{Code: "oauth_identity_invalid", Message: "Google account does not have a verified email"}
		}
		return payload.Email, payload.Sub, nil
	case "github":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var user struct {
			ID    int64  `json:"id"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", "", err
		}
		email = user.Email
		if email == "" {
			email, err = s.fetchGitHubPrimaryEmail(ctx, accessToken)
			if err != nil {
				return "", "", err
			}
		}
		if email == "" {
			return "", "", &AuthError{Code: "oauth_identity_invalid", Message: "GitHub account does not expose a verified email"}
		}
		return email, fmt.Sprintf("%d", user.ID), nil
	case "discord":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var user struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", "", err
		}
		return user.Email, user.ID, nil
	case "facebook":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL+"&access_token="+accessToken, nil)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var user struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", "", err
		}
		return user.Email, user.ID, nil
	case "microsoft":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var user struct {
			ID    string `json:"id"`
			Email string `json:"mail"` // MS Graph uses 'mail'
		}
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			return "", "", err
		}
		return user.Email, user.ID, nil
	case "line":
		// Line's user info endpoint returns a profile object
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var lineUser struct {
			UserID      string `json:"userId"`
			DisplayName string `json:"displayName"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&lineUser); err != nil {
			return "", "", err
		}
		// Line API does not directly provide email from /profile. It's usually in the ID token or requires specific scopes.
		// For simplicity, we'll use a placeholder email.
		return lineUser.UserID + "@line.invalid", lineUser.UserID, nil
	case "paypal":
		return s.fetchOIDCUser(ctx, cfg.UserURL, accessToken)
	case "amazon":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var amz struct {
			UserID string `json:"user_id"`
			Email  string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&amz); err != nil {
			return "", "", err
		}
		return amz.Email, amz.UserID, nil
	case "tiktok":
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cfg.UserURL, strings.NewReader(`{"fields":["open_id","email"]}`))
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var tt struct {
			Data struct {
				User struct {
					OpenID string `json:"open_id"`
					Email  string `json:"email"`
				} `json:"user"`
			} `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&tt); err != nil {
			return "", "", err
		}
		if tt.Data.User.Email == "" {
			tt.Data.User.Email = tt.Data.User.OpenID + "@tiktok.invalid"
		}
		return tt.Data.User.Email, tt.Data.User.OpenID, nil
	case "pinterest":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var pin struct {
			Username string `json:"username"`
			ID       string `json:"id"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&pin); err != nil {
			return "", "", err
		}
		// Pinterest API v5 does not directly provide email in user_account endpoint.
		// It's usually tied to specific scopes or not available.
		return pin.Username + "@pinterest.invalid", pin.ID, nil
	case "snapchat":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var snap struct {
			Me struct {
				DisplayName string `json:"displayName"`
				ExternalID  string `json:"externalId"` // This is the stable ID
			} `json:"me"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
			return "", "", err
		}
		// SnapKit does not reliably return email without particular scopes.
		// Using externalId as providerID and a placeholder email.
		return snap.Me.ExternalID + "@snapchat.invalid", snap.Me.ExternalID, nil
	case "yahoo":
		return s.fetchOIDCUser(ctx, cfg.UserURL, accessToken)
	case "okta":
		return s.fetchOIDCUser(ctx, cfg.UserURL, accessToken)
	case "yandex":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL+"?format=json", nil)
		req.Header.Set("Authorization", "OAuth "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var yx struct {
			ID           string `json:"id"`
			DefaultEmail string `json:"default_email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&yx); err != nil {
			return "", "", err
		}
		return yx.DefaultEmail, yx.ID, nil
	case "wordpress":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var wp struct {
			ID    int    `json:"ID"`
			Email string `json:"email"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&wp); err != nil {
			return "", "", err
		}
		return wp.Email, fmt.Sprintf("%d", wp.ID), nil
	case "vk":
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.UserURL+"?v=5.131&access_token="+accessToken, nil)
		resp, err := s.client.Do(req)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()
		var vk struct {
			Response []struct {
				ID int `json:"id"`
			} `json:"response"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&vk); err != nil {
			return "", "", err
		}
		if len(vk.Response) == 0 {
			return "", "", &AuthError{Code: "oauth_identity_invalid", Message: "VK did not return user"}
		}
		// Note: VK returns email in token response, not user response usually.
		// For simplicity, we'll use a placeholder email.
		return fmt.Sprintf("%d@vk.invalid", vk.Response[0].ID), fmt.Sprintf("%d", vk.Response[0].ID), nil
	default:
		return "", "", &AuthError{Code: "unsupported_provider", Message: "Unsupported OAuth provider"}
	}
}

func (s *AuthService) fetchGitHubPrimaryEmail(ctx context.Context, accessToken string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}
	for _, item := range emails {
		if item.Primary && item.Verified {
			return item.Email, nil
		}
	}
	for _, item := range emails {
		if item.Verified {
			return item.Email, nil
		}
	}
	return "", nil
}

func (s *AuthService) findOrCreateOAuthUser(ctx context.Context, provider, providerID, email, accessToken, refreshToken string, expiresIn int) (*models.User, error) {
	var userID string
	err := s.db.QueryRow(ctx, `
		SELECT user_id::text
		FROM auth.oauth_accounts
		WHERE provider = $1 AND provider_id = $2
	`, provider, providerID).Scan(&userID)
	if err == nil {
		user, getErr := s.GetUser(ctx, userID)
		if getErr != nil {
			return nil, getErr
		}
		_, _ = s.db.Exec(ctx, `
			UPDATE auth.oauth_accounts
			SET access_token = $1, refresh_token = $2, expires_at = $3
			WHERE provider = $4 AND provider_id = $5
		`, accessToken, refreshToken, nullableExpiry(expiresIn), provider, providerID)
		return user, nil
	}

	var existingUserID string
	err = s.db.QueryRow(ctx, `SELECT id::text FROM auth.users WHERE email = $1`, email).Scan(&existingUserID)
	if err != nil {
		now := time.Now().UTC()
		existingUserID = uuid.New().String()
		_, err = s.db.Exec(ctx, `
			INSERT INTO auth.users (id, email, role, email_confirmed_at, created_at, updated_at)
			VALUES ($1, $2, 'authenticated', $3, $4, $5)
		`, existingUserID, email, now, now, now)
		if err != nil {
			return nil, err
		}
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO auth.oauth_accounts (user_id, provider, provider_id, access_token, refresh_token, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (provider, provider_id)
		DO UPDATE SET
			user_id = EXCLUDED.user_id,
			access_token = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			expires_at = EXCLUDED.expires_at
	`, existingUserID, provider, providerID, accessToken, refreshToken, nullableExpiry(expiresIn))
	if err != nil {
		return nil, err
	}

	return s.GetUser(ctx, existingUserID)
}

func nullableExpiry(expiresIn int) interface{} {
	if expiresIn <= 0 {
		return nil
	}
	return time.Now().UTC().Add(time.Duration(expiresIn) * time.Second)
}

// AuthError is the standard error type for auth operations
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

func (s *AuthService) GetProjectKeys() (string, string, error) {
	anon, err := s.jwtManager.IssueAnonymousToken("default")
	if err != nil {
		return "", "", err
	}
	service, err := s.jwtManager.IssueServiceRoleToken("default")
	if err != nil {
		return "", "", err
	}
	return anon, service, nil
}
