package identity

import (
	"context"
)

func (s *Service) GetOAuthApps(ctx context.Context, projectID string) ([]IdentityProvider, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, project_id, type, name, client_id, client_secret, metadata_url, is_active 
		FROM auth.identity_providers 
		WHERE type = 'oauth' AND project_id = $1
	`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []IdentityProvider
	for rows.Next() {
		var p IdentityProvider
		var cid, cs, murl *string
		if err := rows.Scan(&p.ID, &p.ProjectID, &p.Type, &p.Name, &cid, &cs, &murl, &p.IsActive); err != nil {
			return nil, err
		}
		if cid != nil {
			p.ClientID = *cid
		}
		if cs != nil {
			p.ClientSecret = *cs
		}
		if murl != nil {
			p.MetadataURL = *murl
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func (s *Service) SaveOAuthApp(ctx context.Context, p IdentityProvider) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.identity_providers (project_id, type, name, client_id, client_secret, is_active)
		VALUES ($1, 'oauth', $2, NULLIF($3, ''), NULLIF($4, ''), $5)
		ON CONFLICT (name, project_id) DO UPDATE SET
		client_id = EXCLUDED.client_id, client_secret = EXCLUDED.client_secret, 
		is_active = EXCLUDED.is_active, updated_at = NOW()
	`, p.ProjectID, p.Name, p.ClientID, p.ClientSecret, p.IsActive)
	return err
}
