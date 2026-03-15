package identity

import (
	"context"

	"github.com/gofiber/fiber/v2"
)

// logAction pushes an audit log to the database asynchronously
func (h *Handler) logAction(c *fiber.Ctx, userID, action string, details map[string]any) {
	ip := c.IP()
	ua := c.Get("User-Agent")
	go h.svc.LogAudit(context.Background(), &userID, action, ip, ua, details)
}

// AdminListAuditLogs Returns global audit logs (requires admin)
func (h *Handler) AdminListAuditLogs(c *fiber.Ctx) error {
	page := max(1, c.QueryInt("page", 1))
	perPage := min(100, max(1, c.QueryInt("per_page", 50)))

	logs, total, err := h.svc.ListAuditLogs(c.Context(), page, perPage)
	if err != nil {
		return h.HandleAuthError(c, err)
	}
	return c.JSON(fiber.Map{
		"data":  logs,
		"total": total,
		"page":  page,
	})
}
