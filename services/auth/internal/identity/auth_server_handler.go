package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) ListOAuthClients(c *fiber.Ctx) error {
	if err := h.svc.EnsureOAuthClientsTable(c.Context()); err != nil {
		return h.HandleAuthError(c, err)
	}
	clients, err := h.svc.GetOAuthClients(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(clients)
}
