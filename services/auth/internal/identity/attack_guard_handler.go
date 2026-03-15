package identity

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func (h *Handler) GetBannedIPs(c *fiber.Ctx) error {
	ips, err := h.svc.GetBannedIPs(c.Context())
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(ips)
}

func (h *Handler) BanIP(c *fiber.Ctx) error {
	var body struct {
		IP         string `json:"ip"`
		Reason     string `json:"reason"`
		Expiration int    `json:"expiration_minutes"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	duration := time.Duration(0)
	if body.Expiration > 0 {
		duration = time.Duration(body.Expiration) * time.Minute
	}

	if err := h.svc.BanIP(c.Context(), body.IP, body.Reason, duration); err != nil {
		return h.HandleAuthError(c, err)
	}
	h.log.Info("IP banned via attack guard", zap.String("ip", body.IP))
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) UnbanIP(c *fiber.Ctx) error {
	ip := c.Params("ip")
	if err := h.svc.UnbanIP(c.Context(), ip); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}
