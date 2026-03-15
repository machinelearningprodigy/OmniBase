package identity

import (
	"context"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
)

// ListSessions returns all active sessions for a user
func (s *Service) ListSessions(ctx context.Context, userID string) ([]*models.Session, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, user_agent, ip, created_at, expires_at
		FROM auth.sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*models.Session
	for rows.Next() {
		var sess models.Session
		var ip *string
		if err := rows.Scan(&sess.ID, &sess.UserID, &sess.UserAgent, &ip, &sess.CreatedAt, &sess.ExpiresAt); err != nil {
			return nil, err
		}
		if ip != nil {
			sess.IPAddress = *ip
		}
		sessions = append(sessions, &sess)
	}
	return sessions, nil
}

// RevokeSession revokes a specific session by ID
func (s *Service) RevokeSession(ctx context.Context, userID, sessionID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.sessions WHERE id = $1 AND user_id = $2", sessionID, userID)
	return err
}
