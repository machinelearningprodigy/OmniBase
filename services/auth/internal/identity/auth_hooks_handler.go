package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) GetAuthHooks(c *fiber.Ctx) error {
	hooks, err := h.svc.GetHooks(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(hooks)
}

func (h *Handler) SaveAuthHook(c *fiber.Ctx) error {
	var param AuthHook
	if err := c.BodyParser(&param); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid body"})
	}
	if err := h.svc.SaveHook(c.Context(), param); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) DeleteAuthHook(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.svc.DeleteHook(c.Context(), id); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
