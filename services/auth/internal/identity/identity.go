package identity

import (
	"context"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/models"
	"go.uber.org/zap"
)

type IdentityProvider = models.IdentityProvider

// Service encapsulates all identity features
type Service struct {
	db      *pgxpool.Pool
	log     *zap.Logger
	cfg     *config.Config
	authSvc *services.AuthService
}

func NewService(db *pgxpool.Pool, log *zap.Logger, cfg *config.Config, authSvc *services.AuthService) *Service {
	return &Service{db: db, log: log, cfg: cfg, authSvc: authSvc}
}

func (s *Service) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.authSvc.GetUser(ctx, id)
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	return s.authSvc.GetUserByEmail(ctx, email)
}

func (s *Service) SendTestEmail(ctx context.Context, email string) error {
	return s.authSvc.SendTestEmail(ctx, email)
}

func (s *Service) SignUp(ctx context.Context, req services.SignUpRequest) (*services.SignUpResponse, error) {
	return s.authSvc.SignUp(ctx, req)
}

func (s *Service) IssueTokenPair(ctx context.Context, user *models.User, source string, ip string) (string, string, error) {
	return s.authSvc.IssueTokenPair(ctx, user, source, ip)
}

func (s *Service) Config() *config.Config {
	return s.authSvc.Config()
}

func (s *Service) DB() *pgxpool.Pool {
	return s.authSvc.DB()
}

func (s *Service) GetProjectKeys() (string, string, error) {
	return s.authSvc.GetProjectKeys()
}

// Handler maps identity services to fiber endpoints
type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// Re-using common structures
type AuthError = services.AuthError
type SignUpResponse = services.SignUpResponse
type SignUpRequest = services.SignUpRequest

func handleAuthError(c *fiber.Ctx, err error) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
}

func (h *Handler) HandleAuthError(c *fiber.Ctx, err error) error {
	return handleAuthError(c, err)
}

func (h *Handler) LogAction(c *fiber.Ctx, userID, action string, details map[string]any) {
	ip := c.IP()
	ua := c.Get("User-Agent")
	go h.svc.LogAudit(context.Background(), &userID, action, ip, ua, details)
}

func (h *Handler) GetClaims(c *fiber.Ctx) (string, string, error) {
	if userID := c.Get("X-OmniBase-User-ID"); userID != "" {
		role := c.Get("X-OmniBase-Role")
		if role == "" {
			role = "authenticated"
		}
		return userID, role, nil
	}

	auth := c.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return "", "", fiber.ErrUnauthorized
	}

	claims, err := h.svc.authSvc.VerifyToken(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return "", "", err
	}

	return claims.UserID, claims.Role, nil
}

func (h *Handler) GetProjectKeys(c *fiber.Ctx) error {
	anon, service, err := h.svc.GetProjectKeys()
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(fiber.Map{
		"anon_key":    anon,
		"service_key": service,
	})
}
