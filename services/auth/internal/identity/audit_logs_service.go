package identity

import (
	"context"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
)

// LogAudit records an audit event
func (s *Service) LogAudit(ctx context.Context, userID *string, action, ip, userAgent string, details map[string]any) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.audit_logs (user_id, action, ip_address, user_agent, details)
		VALUES ($1, $2, NULLIF($3, '')::inet, $4, $5)
	`, userID, action, ip, userAgent, models.JSONB(details))
	return err
}

// ListAuditLogs returns paginated audit logs
func (s *Service) ListAuditLogs(ctx context.Context, page, perPage int) ([]*models.AuditLog, int64, error) {
	offset := (page - 1) * perPage

	var total int64
	s.db.QueryRow(ctx, "SELECT COUNT(*) FROM auth.audit_logs").Scan(&total)

	rows, err := s.db.Query(ctx, `
		SELECT id, CASE WHEN user_id IS NULL THEN '' ELSE user_id::text END as uid, action, CASE WHEN ip_address IS NULL THEN '' ELSE ip_address::text END as ip, user_agent, details, created_at
		FROM auth.audit_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, perPage, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		var uid string
		var ip string
		if err := rows.Scan(&l.ID, &uid, &l.Action, &ip, &l.UserAgent, &l.Details, &l.CreatedAt); err != nil {
			return nil, 0, err
		}
		if uid != "" {
			l.UserID = &uid
		}
		if ip != "" {
			l.IPAddress = &ip
		}
		logs = append(logs, &l)
	}
	return logs, total, nil
}

// ListUserAuditLogs returns audit logs for a specific user.
func (s *Service) ListUserAuditLogs(ctx context.Context, userID string, limit int) ([]*models.AuditLog, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id::text, action, CASE WHEN ip_address IS NULL THEN '' ELSE ip_address::text END as ip, user_agent, details, created_at
		FROM auth.audit_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var l models.AuditLog
		var uid string
		var ip string
		if err := rows.Scan(&l.ID, &uid, &l.Action, &ip, &l.UserAgent, &l.Details, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.UserID = &uid
		if ip != "" {
			l.IPAddress = &ip
		}
		logs = append(logs, &l)
	}
	return logs, nil
}
