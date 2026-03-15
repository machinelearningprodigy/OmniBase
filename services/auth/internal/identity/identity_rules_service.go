package identity

import (
	"context"
)

type IdentityRule struct {
	ID        string `json:"id" db:"id"`
	RuleType  string `json:"rule_type" db:"rule_type"`
	Value     string `json:"value" db:"value"`
	IsActive  bool   `json:"is_active" db:"is_active"`
	CreatedAt string `json:"created_at" db:"created_at"`
}

func (s *Service) GetIdentityRules(ctx context.Context) ([]IdentityRule, error) {
	rows, err := s.db.Query(ctx, "SELECT id, rule_type, value, is_active, created_at::text FROM auth.identity_rules")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []IdentityRule
	for rows.Next() {
		var r IdentityRule
		if err := rows.Scan(&r.ID, &r.RuleType, &r.Value, &r.IsActive, &r.CreatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *Service) SaveIdentityRule(ctx context.Context, rule_type, value string, is_active bool) error {
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.identity_rules (rule_type, value, is_active)
		VALUES ($1, $2, $3)
	`, rule_type, value, is_active)
	return err
}

func (s *Service) DeleteIdentityRule(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.identity_rules WHERE id = $1", id)
	return err
}
