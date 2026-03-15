package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListIdentityRules(c *fiber.Ctx) error {
	rules, err := h.svc.GetIdentityRules(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(rules)
}

func (h *Handler) SaveIdentityRule(c *fiber.Ctx) error {
	var body struct {
		RuleType string `json:"rule_type"`
		Value    string `json:"value"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := h.svc.SaveIdentityRule(c.Context(), body.RuleType, body.Value, body.IsActive); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) DeleteIdentityRule(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteIdentityRule(c.Context(), id); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
