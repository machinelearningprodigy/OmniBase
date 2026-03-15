package identity

import (
	"context"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
)

func (s *Service) SaveSysSetting(ctx context.Context, key string, value map[string]any) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.settings (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = NOW()
	`, key, models.JSONB(value))
	return err
}
