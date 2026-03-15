package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) Web3Nonce(c *fiber.Ctx) error {
	var body struct {
		WalletAddress string `json:"wallet_address"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	nonce, err := h.svc.GenerateWeb3Nonce(c.Context(), body.WalletAddress)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(fiber.Map{"nonce": nonce})
}

func (h *Handler) Web3Verify(c *fiber.Ctx) error {
	var body struct {
		WalletAddress string `json:"wallet_address"`
		Signature     string `json:"signature"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid payload"})
	}

	resp, err := h.svc.VerifyWeb3Signature(c.Context(), body.WalletAddress, body.Signature)
	if err != nil {
		return h.HandleAuthError(c, err)
	}

	h.LogAction(c, resp.User.ID, "web3_login", map[string]any{"wallet": body.WalletAddress})
	return c.JSON(resp)
}
