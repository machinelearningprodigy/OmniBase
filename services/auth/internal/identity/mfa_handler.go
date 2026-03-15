package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) EnrollMFA(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		FactorType string `json:"factor_type"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if req.FactorType == "" {
		req.FactorType = "totp"
	}
	factor, err := h.svc.EnrollMFA(c.Context(), userID, req.FactorType)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	h.LogAction(c, userID, "mfa_enroll_initiated", map[string]any{"factor_id": factor.ID})
	return c.JSON(factor)
}

func (h *Handler) VerifyMFA(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	var req struct {
		FactorID string `json:"factor_id"`
		Code     string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := h.svc.VerifyMFA(c.Context(), userID, req.FactorID, req.Code); err != nil {
		return h.HandleAuthError(c, err)
	}
	h.LogAction(c, userID, "mfa_verified", map[string]any{"factor_id": req.FactorID})
	return c.SendStatus(fiber.StatusOK)
}

func (h *Handler) DeleteMFA(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	if err := h.svc.DeleteMFAFactor(c.Context(), userID, c.Params("id")); err != nil {
		return h.HandleAuthError(c, err)
	}
	h.LogAction(c, userID, "mfa_deleted", map[string]any{"factor_id": c.Params("id")})
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) ListMFA(c *fiber.Ctx) error {
	userID, _, err := h.GetClaims(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "unauthorized"})
	}
	factors, err := h.svc.ListMFAFactors(c.Context(), userID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(factors)
}
