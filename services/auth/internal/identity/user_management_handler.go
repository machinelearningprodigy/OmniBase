package identity

import (
	"github.com/gofiber/fiber/v2"
)

func (h *Handler) AdminListUsers(c *fiber.Ctx) error {
	page := max(1, c.QueryInt("page", 1))
	perPage := min(100, max(1, c.QueryInt("per_page", 50)))
	search := c.Query("search")
	projectID := c.Query("project_id")
	if projectID == "" {
		projectID = c.Get("X-OmniBase-Project-ID")
	}

	users, total, err := h.svc.AdminListUsers(c.Context(), page, perPage, search, projectID)
	if err != nil {
		return h.HandleAuthError(c, err)
	}

	return c.JSON(fiber.Map{
		"users":    users,
		"total":    total,
		"page":     page,
		"per_page": perPage,
	})
}

func (h *Handler) AdminGetUser(c *fiber.Ctx) error {
	user, err := h.svc.GetUser(c.Context(), c.Params("id"))
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(user)
}

func (h *Handler) AdminUpdateUser(c *fiber.Ctx) error {
	var req AdminUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	user, err := h.svc.AdminUpdateUser(c.Context(), c.Params("id"), req)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(user)
}

func (h *Handler) AdminDeleteUser(c *fiber.Ctx) error {
	if err := h.svc.AdminDeleteUser(c.Context(), c.Params("id")); err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) AdminBanUser(c *fiber.Ctx) error {
	var body struct {
		Banned bool `json:"banned"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.svc.AdminBanUser(c.Context(), c.Params("id"), body.Banned); err != nil {
		return h.HandleAuthError(c, err)
	}

	return c.JSON(fiber.Map{"status": "success"})
}

func (h *Handler) AdminInviteUser(c *fiber.Ctx) error {
	var body struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	user, err := h.svc.AdminInviteUser(c.Context(), body.Email)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *Handler) AdminGetUserAuditLogs(c *fiber.Ctx) error {
	logs, err := h.svc.ListUserAuditLogs(c.Context(), c.Params("id"), 50)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(logs)
}
