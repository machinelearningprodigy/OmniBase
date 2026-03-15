package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListOAuthApps(c *fiber.Ctx) error {
	projectID := c.Get("X-OmniBase-Project-ID")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing project ID"})
	}

	providers, err := h.svc.GetOAuthApps(c.Context(), projectID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(providers)
}

func (h *Handler) SaveOAuthApp(c *fiber.Ctx) error {
	projectID := c.Get("X-OmniBase-Project-ID")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing project ID"})
	}

	var p IdentityProvider
	if err := c.BodyParser(&p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	p.ProjectID = projectID
	if err := h.svc.SaveOAuthApp(c.Context(), p); err != nil {
		return h.HandleAuthError(c, err)
	}

	return c.JSON(fiber.Map{"message": "OAuth app saved successfully"})
}
