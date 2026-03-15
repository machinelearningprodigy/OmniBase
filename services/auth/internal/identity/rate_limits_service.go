package identity

import "context"

type RateLimitConfig struct {
	ID      string `json:"id" db:"id"`
	Key     string `json:"key" db:"key"`
	Count   int    `json:"count" db:"count"`
	ResetAt string `json:"reset_at" db:"reset_at"`
}

func (s *Service) GetRateLimits(ctx context.Context) ([]RateLimitConfig, error) {
	rows, err := s.db.Query(ctx, "SELECT id, key, count, reset_at::text FROM auth.rate_limits")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var limits []RateLimitConfig
	for rows.Next() {
		var l RateLimitConfig
		if err := rows.Scan(&l.ID, &l.Key, &l.Count, &l.ResetAt); err != nil {
			return nil, err
		}
		limits = append(limits, l)
	}
	return limits, nil
}
