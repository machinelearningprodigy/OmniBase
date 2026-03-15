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

func (h *Handler) GetSMTPSettings(c *fiber.Ctx) error {
	projectID := c.Query("project_id", "default")
	settings, err := h.svc.GetSMTPSettings(c.Context(), projectID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(settings)
}

func (h *Handler) SaveSMTPSettings(c *fiber.Ctx) error {
	var body SMTPSettings
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	body.ProjectID = c.Query("project_id", "default")
	if err := h.svc.SaveSMTPSettings(c.Context(), body); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) SendTestEmail(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := h.svc.SendTestEmail(c.Context(), body.Email); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}
