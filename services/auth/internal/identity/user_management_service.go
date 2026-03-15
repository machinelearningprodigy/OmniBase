package identity

import (
	"context"
	"fmt"
	"strings"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
	"go.uber.org/zap"
)

// AdminListUsers returns a list of users with optional filtering.
func (s *Service) AdminListUsers(ctx context.Context, page, perPage int, search string, projectID string) ([]*models.User, int64, error) {
	offset := (page - 1) * perPage

	var total int64

	var conditions []string
	var args []any

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(email ILIKE $%d OR id::text ILIKE $%d)", len(args)+1, len(args)+1))
		args = append(args, "%"+search+"%")
	}

	if projectID != "" {
		conditions = append(conditions, fmt.Sprintf("project_id = $%d", len(args)+1))
		args = append(args, projectID)
	} else {
		conditions = append(conditions, "project_id IS NULL")
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM auth.users" + where
	err := s.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	sql := fmt.Sprintf(`
		SELECT 
			u.id, u.email, u.role, u.is_super_admin, u.is_banned, u.email_confirmed_at, 
			u.last_sign_in_at, u.created_at, u.updated_at, u.raw_user_meta_data, 
			u.raw_app_meta_data, u.project_id,
			COALESCE(u.raw_user_meta_data->>'full_name', u.raw_user_meta_data->>'name', '') as display_name,
			COALESCE(u.raw_user_meta_data->>'avatar_url', '') as avatar_url,
			EXISTS(SELECT 1 FROM auth.mfa_factors WHERE user_id = u.id AND status = 'verified') as mfa_enabled,
			(u.email_confirmed_at IS NOT NULL) as email_verified,
			(u.phone_confirmed_at IS NOT NULL) as phone_verified,
			(SELECT COUNT(*) FROM auth.passkeys WHERE user_id = u.id) as passkey_count,
			(
				SELECT COALESCE(array_agg(DISTINCT p), '{}') FROM (
					SELECT provider FROM auth.oauth_accounts WHERE user_id = u.id
					UNION
					SELECT 'email' WHERE u.password_hash IS NOT NULL
				) t(p)
			) as providers
		FROM auth.users u %s
		ORDER BY created_at DESC LIMIT $%d OFFSET $%d
	`, where, len(args)+1, len(args)+2)

	args = append(args, perPage, offset)

	rows, err := s.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var u models.User
		var providers []string
		err := rows.Scan(
			&u.ID, &u.Email, &u.Role, &u.IsSuperAdmin, &u.IsBanned,
			&u.EmailConfirmedAt, &u.LastSignInAt, &u.CreatedAt, &u.UpdatedAt,
			&u.RawUserMetaData, &u.RawAppMetaData, &u.ProjectID,
			&u.DisplayName, &u.AvatarURL, &u.MFAEnabled, 
			&u.EmailVerified, &u.PhoneVerified, &u.PasskeyCount, &providers,
		)
		if err != nil {
			return nil, 0, err
		}
		u.Providers = providers
		users = append(users, &u)
	}

	return users, total, nil
}

// AdminDeleteUser permanently deletes a user.
func (s *Service) AdminDeleteUser(ctx context.Context, userID string) error {
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

// AdminBanUser bans or unbans a user.
func (s *Service) AdminBanUser(ctx context.Context, userID string, banned bool) error {
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

// AdminUpdateUser updates user data by admin.
func (s *Service) AdminUpdateUser(ctx context.Context, userID string, updates AdminUpdateRequest) (*models.User, error) {
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
	if updates.IsSuperAdmin != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_super_admin = $%d", argNum))
		args = append(args, *updates.IsSuperAdmin)
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
	_, err := s.db.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.GetUser(ctx, userID)
}

// AdminInviteUser creates a user and sends them an invite email.
func (s *Service) AdminInviteUser(ctx context.Context, email string) (*models.User, error) {
	return s.authSvc.AdminInviteUser(ctx, email)
}

type AdminUpdateRequest struct {
	Email        string         `json:"email"`
	Role         string         `json:"role"`
	IsSuperAdmin *bool          `json:"is_super_admin,omitempty"`
	UserMetadata map[string]any `json:"user_metadata,omitempty"`
}
