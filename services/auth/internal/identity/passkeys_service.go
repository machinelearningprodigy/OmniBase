package identity

import (
	"context"

	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
)

func (s *Service) GeneratePasskeyChallenge(ctx context.Context, userID string) (string, error) {
	challenge, _ := services.GenerateSecureToken(32)
	// In reality we'd store this challenge in a cache for the verify step
	return challenge, nil
}

func (s *Service) VerifyPasskeyRegistration(ctx context.Context, userID, credentialID, publicKey string) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.passkeys (user_id, credential_id, public_key)
		VALUES ($1, $2, $3::bytea)
	`, userID, credentialID, publicKey)
	return err
}
