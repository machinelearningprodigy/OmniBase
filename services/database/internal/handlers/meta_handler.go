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

func (h *MetaHandler) CreateTable(c *fiber.Ctx) error {
	var req services.CreateTableRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if err := h.service.CreateTable(c.Context(), req); err != nil {
		h.log.Error("failed to create table", zap.Error(err), zap.String("table", req.Name))
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "table created"})
}

func (h *MetaHandler) CreateProject(c *fiber.Ctx) error {
	var payload struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if strings.TrimSpace(payload.Name) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "project name required"})
	}

	if err := h.service.CreateProject(c.Context(), payload.Name); err != nil {
		h.log.Error("failed to create project", zap.Error(err), zap.String("name", payload.Name))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{"message": "project created"})
}

func (h *MetaHandler) ListFunctions(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	fns, err := h.service.GetFunctions(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fns)
}

func (h *MetaHandler) ListPolicies(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	policies, err := h.service.GetPolicies(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(policies)
}

func (h *MetaHandler) ListSchemas(c *fiber.Ctx) error {
	schemas, err := h.service.GetSchemas(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(schemas)
}

func (h *MetaHandler) ReloadSchema(c *fiber.Ctx) error {
	if err := h.service.ReloadSchemaCache(c.Context()); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "schema reload triggered"})
}

func (h *MetaHandler) ResolveGraphQL(c *fiber.Ctx) error {
	var payload struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := c.BodyParser(&payload); err != nil {
		h.log.Error("failed to parse graphql payload", zap.Error(err), zap.ByteString("body", c.Body()))
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	userID := c.Get("X-OmniBase-User-ID")
	role := c.Get("X-OmniBase-Role")

	// If no role provided, default to anon
	if role == "" {
		role = "anon"
	}

	result, err := h.service.ResolveGraphQL(c.Context(), payload.Query, payload.Variables, userID, role)
	if err != nil {
		h.log.Error("graphql resolution failed", zap.Error(err), zap.String("query", payload.Query))
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(result)
}


