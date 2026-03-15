package identity

import (
	"context"
	"fmt"

	"github.com/machinelearningprodigy/OmniBase/shared/models"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

type MFAEnrollResponse struct {
	ID         string `json:"id"`
	FactorType string `json:"factor_type"`
	Secret     string `json:"secret"`
	QRCodeURL  string `json:"qr_code_url,omitempty"`
}

// EnrollMFA initiates enrollment for an MFA factor using robust TOTP algorithms
func (s *Service) EnrollMFA(ctx context.Context, userID, factorType string) (*MFAEnrollResponse, error) {
	if factorType != "totp" {
		return nil, &AuthError{Code: "unsupported_factor", Message: "Only TOTP is currently supported natively."}
	}

	user, err := s.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "OmniBase",
		AccountName: user.Email,
	})
	if err != nil {
		s.log.Error("failed to generate totp secret", zap.Error(err))
		return nil, &AuthError{Code: "internal_error", Message: "Failed to generate MFA secret."}
	}
	secret := key.Secret()
	qrCodeURL := key.URL()

	var factorID string
	err = s.db.QueryRow(ctx, `
		INSERT INTO auth.mfa_factors (user_id, factor_type, status, secret)
		VALUES ($1, $2, 'unverified', $3)
		RETURNING id
	`, userID, factorType, secret).Scan(&factorID)

	if err != nil {
		return nil, fmt.Errorf("failed to save factor: %w", err)
	}

	return &MFAEnrollResponse{
		ID:         factorID,
		FactorType: factorType,
		Secret:     secret,
		QRCodeURL:  qrCodeURL,
	}, nil
}

// VerifyMFA validates the 6-digit TOTP code and activates the MFA factor
func (s *Service) VerifyMFA(ctx context.Context, userID, factorID, code string) error {
	var secret string
	var status string

	err := s.db.QueryRow(ctx, "SELECT secret, status FROM auth.mfa_factors WHERE id = $1 AND user_id = $2", factorID, userID).Scan(&secret, &status)
	if err != nil {
		return &AuthError{Code: "invalid_factor", Message: "Factor not found or does not belong to user."}
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return &AuthError{Code: "invalid_code", Message: "The TOTP code provided is incorrect."}
	}

	if status == "unverified" {
		_, err = s.db.Exec(ctx, "UPDATE auth.mfa_factors SET status = 'verified', updated_at = NOW() WHERE id = $1", factorID)
		if err != nil {
			return fmt.Errorf("failed to activate factor: %w", err)
		}
	}

	_, _ = s.db.Exec(ctx, "UPDATE auth.mfa_factors SET last_used_at = NOW() WHERE id = $1", factorID)

	return nil
}

func (s *Service) DeleteMFAFactor(ctx context.Context, userID, factorID string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.mfa_factors WHERE id = $1 AND user_id = $2", factorID, userID)
	return err
}

func (s *Service) ListMFAFactors(ctx context.Context, userID string) ([]*models.MFAFactor, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, user_id, factor_type, status, last_used_at, created_at, updated_at 
		FROM auth.mfa_factors WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var factors []*models.MFAFactor
	for rows.Next() {
		var f models.MFAFactor
		if err := rows.Scan(&f.ID, &f.UserID, &f.FactorType, &f.Status, &f.LastUsedAt, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		factors = append(factors, &f)
	}
	return factors, nil
}
