package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/machinelearningprodigy/OmniBase/database/internal/services"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"go.uber.org/zap"
)

type MetaHandler struct {
	service *services.MetaService
	jwt     *jwt.Manager
	log     *zap.Logger
}

func NewMetaHandler(service *services.MetaService, jwtManager *jwt.Manager, log *zap.Logger) *MetaHandler {
	return &MetaHandler{
		service: service,
		jwt:     jwtManager,
		log:     log,
	}
}

// RequireServiceRole ensures only the Service Role (Admin) can access metadata endpoints.
func (h *MetaHandler) RequireServiceRole(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwt.Verify(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	if claims.Role != "service_role" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}

	return c.Next()
}

func (h *MetaHandler) ListTables(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	
	tables, err := h.service.GetTables(c.Context(), schema)
	if err != nil {
		h.log.Error("failed to get tables", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(tables)
}

func (h *MetaHandler) RunQuery(c *fiber.Ctx) error {
	var payload struct {
		Query string `json:"query"`
	}
	
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if strings.TrimSpace(payload.Query) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query cannot be empty"})
	}

	results, err := h.service.RunQuery(c.Context(), payload.Query)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()}) // Return Postgres error directly to user
	}

	return c.JSON(results)
}
