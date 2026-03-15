package identity

import (
	"context"
	"encoding/json"
	"time"

	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AuthHook struct {
	ID          string `json:"id" db:"id"`
	Event       string `json:"event" db:"event"`
	EndpointURL string `json:"endpoint_url" db:"endpoint_url"`
	Secret      string `json:"-" db:"secret"`
	IsActive    bool   `json:"is_active" db:"is_active"`
}

var restyClient = resty.New().SetTimeout(5 * time.Second).SetRetryCount(2)

// ExecuteHooks triggers all active webhooks for a specific event
func (s *Service) ExecuteHooks(ctx context.Context, event string, payload any) {
	hooks, err := s.getActiveHooksByEvent(ctx, event)
	if err != nil || len(hooks) == 0 {
		return
	}

	payloadBytes, _ := json.Marshal(payload)

	for _, hook := range hooks {
		go func(h AuthHook, data []byte) {
			resp, err := restyClient.R().
				SetHeader("Content-Type", "application/json").
				SetHeader("X-OmniBase-Hook-Secret", h.Secret).
				SetBody(data).
				Post(h.EndpointURL)

			if err != nil || resp.IsError() {
				s.log.Error("Webhook execution failed", zap.String("event", h.Event), zap.String("url", h.EndpointURL), zap.Error(err))
			} else {
				s.log.Info("Webhook executed successfully", zap.String("event", h.Event), zap.String("url", h.EndpointURL))
			}
		}(hook, payloadBytes)
	}
}

func (s *Service) getActiveHooksByEvent(ctx context.Context, event string) ([]AuthHook, error) {
	rows, err := s.db.Query(ctx, "SELECT id, event, endpoint_url, secret, is_active FROM auth.hooks WHERE is_active = TRUE AND event = $1", event)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []AuthHook
	for rows.Next() {
		var h AuthHook
		if err := rows.Scan(&h.ID, &h.Event, &h.EndpointURL, &h.Secret, &h.IsActive); err == nil {
			hooks = append(hooks, h)
		}
	}
	return hooks, nil
}

func (s *Service) GetHooks(ctx context.Context) ([]AuthHook, error) {
	rows, err := s.db.Query(ctx, "SELECT id, event, endpoint_url, is_active FROM auth.hooks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hooks []AuthHook
	for rows.Next() {
		var h AuthHook
		if err := rows.Scan(&h.ID, &h.Event, &h.EndpointURL, &h.IsActive); err != nil {
			return nil, err
		}
		hooks = append(hooks, h)
	}
	return hooks, nil
}

func (s *Service) SaveHook(ctx context.Context, hook AuthHook) error {
	secret, _ := services.GenerateSecureToken(16)
	_, err := s.db.Exec(ctx, `
		INSERT INTO auth.hooks (event, endpoint_url, secret, is_active)
		VALUES ($1, $2, $3, $4)
	`, hook.Event, hook.EndpointURL, secret, hook.IsActive)
	return err
}

func (s *Service) DeleteHook(ctx context.Context, id string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.hooks WHERE id = $1", id)
	return err
}
