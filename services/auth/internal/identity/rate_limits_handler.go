package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListRateLimits(c *fiber.Ctx) error {
	limits, err := h.svc.GetRateLimits(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(limits)
}
