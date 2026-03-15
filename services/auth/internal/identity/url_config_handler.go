package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) SaveSettings(c *fiber.Ctx) error {
	key := c.Params("key")
	var body map[string]any
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid config body"})
	}
	if err := h.svc.SaveSysSetting(c.Context(), key, body); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) GetSettings(c *fiber.Ctx) error {
	key := c.Params("key")
	val, err := h.svc.GetSysSetting(c.Context(), key)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Setting not found"})
	}
	return c.JSON(val)
}
