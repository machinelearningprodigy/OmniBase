package identity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
	"go.uber.org/zap"
)

func (s *Service) GenerateWeb3Nonce(ctx context.Context, walletAddress string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(bytes)

	expiresAt := time.Now().UTC().Add(5 * time.Minute)

	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.web3_nonces (wallet_address, nonce, expires_at, created_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (wallet_address) DO UPDATE 
		SET nonce = EXCLUDED.nonce, expires_at = EXCLUDED.expires_at, created_at = NOW()
	`, walletAddress, nonce, expiresAt)

	if err != nil {
		return "", fmt.Errorf("failed to store nonce: %w", err)
	}

	return nonce, nil
}

func (s *Service) VerifyWeb3Signature(ctx context.Context, walletAddress, signature string) (*SignUpResponse, error) {
	var storedNonce string
	var expiresAt time.Time

	err := s.db.QueryRow(ctx, `
		SELECT nonce, expires_at FROM auth.web3_nonces 
		WHERE wallet_address = $1
	`, walletAddress).Scan(&storedNonce, &expiresAt)

	if err != nil {
		return nil, &AuthError{Code: "invalid_nonce", Message: "Nonce missing or expired. Request a new one."}
	}
	if time.Now().UTC().After(expiresAt) {
		return nil, &AuthError{Code: "expired_nonce", Message: "Nonce has expired"}
	}

	if len(signature) < 64 {
		return nil, &AuthError{Code: "invalid_signature", Message: "Invalid signature length"}
	}

	s.db.Exec(ctx, "DELETE FROM auth.web3_nonces WHERE wallet_address = $1", walletAddress)

	var userID string
	err = s.db.QueryRow(ctx, "SELECT id::text FROM auth.users WHERE raw_user_meta_data->>'wallet_address' = $1", walletAddress).Scan(&userID)
	if err != nil {
		now := time.Now().UTC()
		userID = generateUUID()

		_, err = s.db.Exec(ctx, `
			INSERT INTO auth.users (id, role, raw_user_meta_data, created_at, updated_at)
			VALUES ($1, 'authenticated', $2, $3, $3)
		`, userID, models.JSONB{"wallet_address": walletAddress}, now)
		if err != nil {
			return nil, err
		}
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	s.db.Exec(ctx, "UPDATE auth.users SET last_sign_in_at = NOW() WHERE id = $1", user.ID)

	accessToken, refreshToken, err := s.IssueTokenPair(ctx, user, "Web3", "")
	if err != nil {
		return nil, err
	}

	s.log.Info("web3 user signed in", zap.String("wallet", walletAddress))

	return &SignUpResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "bearer",
		ExpiresIn:    3600,
	}, nil
}

func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}
