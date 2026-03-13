package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/omnibase/omnibase/auth/internal/services"
	"github.com/omnibase/omnibase/shared/jwt"
	"go.uber.org/zap"
)

// AuthHandler handles HTTP requests for the auth service
type AuthHandler struct {
	authSvc *services.AuthService
	log     *zap.Logger
}

func NewAuthHandler(authSvc *services.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, log: log}
}

// SignUp godoc
//
//	@Summary		Register a new user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		services.SignUpRequest	true	"Signup payload"
//	@Success		200		{object}	services.SignUpResponse
//	@Failure		400		{object}	fiber.Map
//	@Router			/auth/v1/signup [post]
func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var req services.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "invalid_body",
			"message": "Invalid request body",
		})
	}

	resp, err := h.authSvc.SignUp(c.Context(), req)
	if err != nil {
		return handleAuthError(c, err)
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

// SignIn godoc
//
//	@Summary		Sign in a user (password or refresh_token grant)
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			body	body		services.SignInRequest	true	"Sign in payload"
//	@Success		200		{object}	services.SignUpResponse
//	@Failure		400		{object}	fiber.Map
//	@Router			/auth/v1/token [post]
func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
	var req services.SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "invalid_body",
			"message": "Invalid request body",
		})
	}

	resp, err := h.authSvc.SignIn(c.Context(), req)
	if err != nil {
		return handleAuthError(c, err)
	}

	return c.JSON(resp)
}

// SignOut godoc
//
//	@Summary		Sign out a user (invalidate session)
//	@Tags			auth
//	@Success		204
//	@Router			/auth/v1/logout [post]
func (h *AuthHandler) SignOut(c *fiber.Ctx) error {
	// TODO: Phase 1 — Invalidate refresh token in DB/Valkey
	return c.SendStatus(fiber.StatusNoContent)
}

// GetUser godoc
//
//	@Summary		Get the currently authenticated user
//	@Tags			auth
//	@Produce		json
//	@Success		200	{object}	models.User
//	@Router			/auth/v1/user [get]
func (h *AuthHandler) GetUser(c *fiber.Ctx) error {
	userID := extractUserIDFromToken(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "unauthorized", "message": "Not authenticated",
		})
	}

	user, err := h.authSvc.GetUser(c.Context(), userID)
	if err != nil {
		return handleAuthError(c, err)
	}

	return c.JSON(user)
}

func (h *AuthHandler) UpdateUser(c *fiber.Ctx) error {
	// TODO: Email change, password change, metadata update
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Coming soon"})
}

func (h *AuthHandler) RecoverPassword(c *fiber.Ctx) error {
	// TODO: Send password reset email
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If this email exists, a reset link has been sent"})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	// TODO: Verify email confirmation token
	token := c.Query("token")
	redirectTo := c.Query("redirect_to", "/")
	h.log.Info("email verify attempt", zap.String("token_prefix", token[:min(8, len(token))]))
	return c.Redirect(redirectTo)
}

func (h *AuthHandler) OAuthAuthorize(c *fiber.Ctx) error {
	// TODO: Phase 1 — Google and GitHub OAuth2
	provider := c.Query("provider")
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"code":    "not_implemented",
		"message": "OAuth coming soon for provider: " + provider,
	})
}

func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"code": "not_implemented"})
}

// ─── Admin Handlers ────────────────────────────────────────────────────────────

func (h *AuthHandler) AdminListUsers(c *fiber.Ctx) error {
	page := max(1, c.QueryInt("page", 1))
	perPage := min(100, max(1, c.QueryInt("per_page", 50)))

	users, total, err := h.authSvc.AdminListUsers(c.Context(), page, perPage)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"code": "internal_error", "message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"users":    users,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *AuthHandler) AdminGetUser(c *fiber.Ctx) error {
	user, err := h.authSvc.GetUser(c.Context(), c.Params("id"))
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(user)
}

func (h *AuthHandler) AdminUpdateUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Coming soon"})
}

func (h *AuthHandler) AdminDeleteUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Coming soon"})
}

func (h *AuthHandler) AdminBanUser(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Coming soon"})
}

func (h *AuthHandler) AdminGenerateLink(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{"message": "Coming soon"})
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func handleAuthError(c *fiber.Ctx, err error) error {
	if authErr, ok := err.(*services.AuthError); ok {
		status := fiber.StatusBadRequest
		if authErr.Code == "user_not_found" || authErr.Code == "invalid_credentials" {
			status = fiber.StatusUnauthorized
		}
		return c.Status(status).JSON(fiber.Map{
			"code":    authErr.Code,
			"message": authErr.Message,
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"code": "internal_error", "message": "An unexpected error occurred",
	})
}

func extractUserIDFromToken(c *fiber.Ctx) string {
	auth := c.Get("Authorization")
	if len(auth) > 7 && auth[:7] == "Bearer " {
		// In production, use the JWT manager to verify and extract
		// For now, we rely on the gateway to inject X-OmniBase-User-ID
		if userID := c.Get("X-OmniBase-User-ID"); userID != "" {
			return userID
		}
	}
	return ""
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var _ = strconv.Itoa // suppress unused import
var _ = jwt.AccessToken
