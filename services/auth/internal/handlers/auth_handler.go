package handlers

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authSvc *services.AuthService
	log     *zap.Logger
}

func NewAuthHandler(authSvc *services.AuthService, log *zap.Logger) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, log: log}
}

func (h *AuthHandler) SignUp(c *fiber.Ctx) error {
	var req services.SignUpRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "invalid_body",
			"message": "Invalid request body",
		})
	}

	req.ProjectID = c.Get("X-OmniBase-Project-ID")
	resp, err := h.authSvc.SignUp(c.Context(), req)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *AuthHandler) SignIn(c *fiber.Ctx) error {
	var req services.SignInRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code":    "invalid_body",
			"message": "Invalid request body",
		})
	}

	resp, err := h.authSvc.SignIn(c.Context(), req, c.Get("User-Agent"), c.IP())
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(resp)
}

func (h *AuthHandler) SignOut(c *fiber.Ctx) error {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = c.BodyParser(&body)

	if body.RefreshToken != "" {
		if err := h.authSvc.RevokeRefreshToken(c.Context(), body.RefreshToken); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_token", "message": err.Error()})
		}
		return c.SendStatus(fiber.StatusNoContent)
	}

	userID, _, err := h.GetClaims(c)
	if err == nil && userID != "" {
		if err := h.authSvc.RevokeAllSessions(c.Context(), userID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
		}
	}
	return c.SendStatus(fiber.StatusNoContent)
}

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
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "unauthorized", "message": "Not authenticated",
		})
	}

	var req services.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"code": "invalid_body", "message": "Invalid request body",
		})
	}

	user, err := h.authSvc.UpdateUser(c.Context(), userID, req)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(user)
}

func (h *AuthHandler) RecoverPassword(c *fiber.Ctx) error {
	var body struct {
		Email      string `json:"email"`
		RedirectTo string `json:"redirect_to"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": "Invalid request body"})
	}
	if err := h.authSvc.CreateRecoveryToken(c.Context(), body.Email, body.RedirectTo); err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If this email exists, a reset link has been sent"})
}

func (h *AuthHandler) MagicLink(c *fiber.Ctx) error {
	var body struct {
		Email      string `json:"email"`
		RedirectTo string `json:"redirect_to"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": "Invalid request body"})
	}
	if body.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "email_required", "message": "Email is required"})
	}
	if err := h.authSvc.CreateMagicLink(c.Context(), body.Email, body.RedirectTo); err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "If this email exists, a sign-in link has been sent"})
}

func (h *AuthHandler) VerifyEmail(c *fiber.Ctx) error {
	token := c.Query("token")
	tokenType := services.ActionLinkType(c.Query("type", string(services.ActionLinkSignup)))
	redirectTo := c.Query("redirect_to", "/")
	if len(token) > 8 {
		h.log.Info("email verify attempt", zap.String("token_prefix", token[:8]))
	}

	// Magic link: consume token, issue session, redirect with tokens in hash
	if tokenType == services.ActionLinkMagicLogin {
		user, resolvedRedirect, err := h.authSvc.VerifyMagicLink(c.Context(), token)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_token", "message": err.Error()})
		}
		if resolvedRedirect != "" {
			redirectTo = resolvedRedirect
		}
		resp, err := h.authSvc.IssueTokenPairForUser(c.Context(), user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"code": "internal_error", "message": err.Error()})
		}
		redirectTo = redirectTo + "#access_token=" + urlQueryEscape(resp.AccessToken) +
			"&refresh_token=" + urlQueryEscape(resp.RefreshToken) +
			"&token_type=" + urlQueryEscape(resp.TokenType) +
			"&expires_in=" + fmt.Sprintf("%d", resp.ExpiresIn)
		return c.Redirect(redirectTo)
	}

	resolvedRedirect, err := h.authSvc.VerifyEmailToken(c.Context(), token, tokenType)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_token", "message": err.Error()})
	}
	if resolvedRedirect != "" {
		redirectTo = resolvedRedirect
	}
	return c.Redirect(redirectTo)
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_body", "message": "Invalid request body"})
	}
	if err := h.authSvc.ResetPassword(c.Context(), body.Token, body.Password); err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(fiber.Map{"message": "Password updated successfully"})
}

func (h *AuthHandler) OAuthAuthorize(c *fiber.Ctx) error {
	provider := c.Query("provider")
	redirectTo := c.Query("redirect_to")
	projectID := c.Query("project_id")
	if projectID == "" {
		projectID = c.Get("X-OmniBase-Project-ID")
	}

	authURL, err := h.authSvc.OAuthAuthorizeURL(c.Context(), provider, redirectTo, projectID)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Redirect(authURL)
}

func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	provider := c.Query("provider")
	code := c.Query("code")
	state := c.Query("state")
	if provider == "" || code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"code": "invalid_oauth_callback", "message": "provider, code, and state are required"})
	}

	resp, redirectTo, err := h.authSvc.CompleteOAuth(c.Context(), provider, code, state)
	if err != nil {
		return handleAuthError(c, err)
	}

	target := redirectTo
	if target == "" {
		target = "/"
	}
	target = target + "#access_token=" + urlQueryEscape(resp.AccessToken) +
		"&refresh_token=" + urlQueryEscape(resp.RefreshToken) +
		"&token_type=" + urlQueryEscape(resp.TokenType) +
		"&expires_in=" + fmt.Sprintf("%d", resp.ExpiresIn)
	return c.Redirect(target)
}



func (h *AuthHandler) AdminGenerateLink(c *fiber.Ctx) error {
	var body struct {
		Type       string `json:"type"`
		UserID     string `json:"user_id"`
		Email      string `json:"email"`
		RedirectTo string `json:"redirect_to"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	link, err := h.authSvc.GenerateActionLink(c.Context(), services.ActionLinkType(body.Type), body.UserID, body.Email, body.RedirectTo)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.JSON(fiber.Map{"action_link": link})
}

func (h *AuthHandler) ListProviders(c *fiber.Ctx) error {
	return c.JSON(h.authSvc.ListProviders())
}

func (h *AuthHandler) RequireAdmin(c *fiber.Ctx) error {
	userID, role, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"code": "unauthorized", "message": "Not authenticated",
		})
	}
	if role == "service_role" {
		return c.Next()
	}

	isAdmin, err := h.authSvc.IsAdmin(c.Context(), userID)
	if err != nil || !isAdmin {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"code": "forbidden", "message": "Admin access required",
		})
	}
	return c.Next()
}

func handleAuthError(c *fiber.Ctx, err error) error {
	if authErr, ok := err.(*services.AuthError); ok {
		status := fiber.StatusBadRequest
		if authErr.Code == "user_not_found" || authErr.Code == "invalid_credentials" || authErr.Code == "invalid_token" {
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
		if userID := c.Get("X-OmniBase-User-ID"); userID != "" {
			return userID
		}
	}
	return ""
}

func (h *AuthHandler) GetClaims(c *fiber.Ctx) (string, string, error) {
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

	claims, err := h.authSvc.VerifyToken(strings.TrimPrefix(auth, "Bearer "))
	if err != nil {
		return "", "", err
	}

	return claims.UserID, claims.Role, nil
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

var _ = jwt.AccessToken

func urlQueryEscape(value string) string {
	replacer := strings.NewReplacer(
		"%", "%25",
		" ", "%20",
		"#", "%23",
		"&", "%26",
		"+", "%2B",
		"=", "%3D",
		"?", "%3F",
	)
	return replacer.Replace(value)
}

