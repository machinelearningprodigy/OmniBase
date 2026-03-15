package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListSessions(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	sessions, err := h.svc.ListSessions(c.Context(), userID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(sessions)
}

func (h *Handler) RevokeSession(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	sessionID := c.Params("id")
	if err := h.svc.RevokeSession(c.Context(), userID, sessionID); err != nil {
		return h.HandleAuthError(c, err)
	}
	h.LogAction(c, userID, "session_revoked", map[string]any{"session_id": sessionID})
	return c.SendStatus(fiber.StatusNoContent)
}
