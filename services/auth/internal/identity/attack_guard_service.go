package identity

import (
	"context"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type BannedIP struct {
	IP        string `json:"ip" db:"ip"`
	Reason    string `json:"reason" db:"reason"`
	ExpiresAt string `json:"expires_at,omitempty" db:"expires_at"`
}

type AttackGuard struct {
	mu     sync.RWMutex
	banned map[string]bool
	svc    *Service
}

// Memory-synced layer for 0ms rejection middleware.
var globalGuard = &AttackGuard{
	banned: make(map[string]bool),
}

func (s *Service) InitAttackGuardCache() {
	globalGuard.svc = s
	ips, _ := s.GetBannedIPs(context.Background())
	globalGuard.mu.Lock()
	defer globalGuard.mu.Unlock()
	for _, b := range ips {
		globalGuard.banned[b.IP] = true
	}
}

// AttackGuardMiddleware completely blocks incoming IPs if they exist in the fast map cache.
func (s *Service) AttackGuardMiddleware() fiber.Handler {
	s.InitAttackGuardCache()
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		globalGuard.mu.RLock()
		isBanned := globalGuard.banned[ip]
		globalGuard.mu.RUnlock()

		if isBanned {
			s.log.Warn("Attack guard blocked request", zap.String("ip", ip), zap.String("path", c.Path()))
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Your IP has been flagged for malicious activity and is blocked."})
		}
		return c.Next()
	}
}

func (s *Service) GetBannedIPs(ctx context.Context) ([]BannedIP, error) {
	rows, err := s.db.Query(ctx, "SELECT ip::text, COALESCE(reason, ''), CASE WHEN expires_at IS NULL THEN '' ELSE expires_at::text END FROM auth.banned_ips WHERE expires_at IS NULL OR expires_at > NOW()")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []BannedIP
	for rows.Next() {
		var b BannedIP
		if err := rows.Scan(&b.IP, &b.Reason, &b.ExpiresAt); err != nil {
			return nil, err
		}
		ips = append(ips, b)
	}
	return ips, nil
}

func (s *Service) BanIP(ctx context.Context, ip, reason string, expiration time.Duration) error {
	var expiresAt *time.Time
	if expiration > 0 {
		t := time.Now().UTC().Add(expiration)
		expiresAt = &t
	}

	_, err := s.db.Exec(ctx, "INSERT INTO auth.banned_ips (ip, reason, expires_at) VALUES ($1::inet, $2, $3) ON CONFLICT (ip) DO UPDATE SET reason = EXCLUDED.reason, expires_at = EXCLUDED.expires_at", ip, reason, expiresAt)
	if err == nil {
		globalGuard.mu.Lock()
		globalGuard.banned[ip] = true
		globalGuard.mu.Unlock()
	}
	return err
}

func (s *Service) UnbanIP(ctx context.Context, ip string) error {
	_, err := s.db.Exec(ctx, "DELETE FROM auth.banned_ips WHERE ip = $1::inet", ip)
	if err == nil {
		globalGuard.mu.Lock()
		delete(globalGuard.banned, ip)
		globalGuard.mu.Unlock()
	}
	return err
}
