package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GeneratePasskeyChallenge(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	challenge, err := h.svc.GeneratePasskeyChallenge(c.Context(), userID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(fiber.Map{"challenge": challenge})
}

func (h *Handler) VerifyPasskey(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}

	var req struct {
		CredentialID string `json:"credential_id"`
		PublicKey    string `json:"public_key"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	if err := h.svc.VerifyPasskeyRegistration(c.Context(), userID, req.CredentialID, req.PublicKey); err != nil {
		return h.HandleAuthError(c, err)
	}

	h.LogAction(c, userID, "passkey_registered", map[string]any{"credential_id": req.CredentialID})
	return c.SendStatus(fiber.StatusCreated)
}
