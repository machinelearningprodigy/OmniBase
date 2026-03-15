package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetEmailTemplates(c *fiber.Ctx) error {
	templates, err := h.svc.GetEmailTemplates(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(templates)
}

func (h *Handler) SaveEmailTemplate(c *fiber.Ctx) error {
	var body EmailTemplate
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := h.svc.SaveEmailTemplate(c.Context(), body); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
