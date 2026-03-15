package identity

import (
	"context"
	"encoding/json"
)

type OAuthClient struct {
	ID           string   `json:"id" db:"id"`
	Name         string   `json:"name" db:"name"`
	ClientID     string   `json:"client_id" db:"client_id"`
	ClientSecret string   `json:"client_secret,omitempty" db:"client_secret"`
	RedirectURIs []string `json:"redirect_uris" db:"redirect_uris"`
}

func (s *Service) EnsureOAuthClientsTable(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS auth.oauth_clients (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name TEXT NOT NULL,
			client_id TEXT UNIQUE NOT NULL,
			client_secret TEXT NOT NULL,
			redirect_uris JSONB NOT NULL DEFAULT '[]'::jsonb,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func (s *Service) GetOAuthClients(ctx context.Context) ([]OAuthClient, error) {
	rows, err := s.db.Query(ctx, "SELECT id, name, client_id, redirect_uris FROM auth.oauth_clients")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []OAuthClient
	for rows.Next() {
		var c OAuthClient
		var urisJSON []byte
		if err := rows.Scan(&c.ID, &c.Name, &c.ClientID, &urisJSON); err != nil {
			return nil, err
		}
		if len(urisJSON) > 0 {
			_ = json.Unmarshal(urisJSON, &c.RedirectURIs)
		}
		clients = append(clients, c)
	}
	return clients, nil
}
