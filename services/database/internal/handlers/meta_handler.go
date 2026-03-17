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
	env     string
}

func NewMetaHandler(service *services.MetaService, jwtManager *jwt.Manager, log *zap.Logger, env string) *MetaHandler {
	return &MetaHandler{
		service: service,
		jwt:     jwtManager,
		log:     log,
		env:     env,
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

func (h *MetaHandler) RequireAdmin(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwt.Verify(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}
	if claims.Role == "service_role" {
		return c.Next()
	}
	if h.env == "development" && claims.Role == "authenticated" {
		return c.Next()
	}

	var isSuperAdmin bool
	if err := h.service.DB().QueryRow(c.Context(), "SELECT is_super_admin FROM auth.users WHERE id = $1", claims.UserID).Scan(&isSuperAdmin); err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Admin access required"})
	}
	if !isSuperAdmin {
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

func (h *MetaHandler) ListTableColumns(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	table := strings.TrimSpace(c.Query("table"))
	if table == "" {
		return c.Status(400).JSON(fiber.Map{"error": "table is required"})
	}

	columns, err := h.service.GetTableColumns(c.Context(), schema, table)
	if err != nil {
		h.log.Error("failed to get table columns", zap.Error(err), zap.String("schema", schema), zap.String("table", table))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(columns)
}

func (h *MetaHandler) RunQuery(c *fiber.Ctx) error {
	var payload struct {
		Query         string `json:"query"`
		MigrationName string `json:"migration_name"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}

	if strings.TrimSpace(payload.Query) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "query cannot be empty"})
	}

	results, err := h.service.RunQuery(c.Context(), payload.Query, payload.MigrationName)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()}) // Return Postgres error directly to user
	}

	return c.JSON(results)
}

func (h *MetaHandler) ListMigrations(c *fiber.Ctx) error {
	migrations, err := h.service.ListMigrations(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(migrations)
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

func (h *MetaHandler) ToggleRLS(c *fiber.Ctx) error {
	var payload struct {
		Schema  string `json:"schema"`
		Table   string `json:"table"`
		Enabled bool   `json:"enabled"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if strings.TrimSpace(payload.Table) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "table is required"})
	}
	if strings.TrimSpace(payload.Schema) == "" {
		payload.Schema = "public"
	}

	if err := h.service.SetRLSEnabled(c.Context(), payload.Schema, payload.Table, payload.Enabled); err != nil {
		h.log.Error("failed to toggle rls", zap.Error(err), zap.String("schema", payload.Schema), zap.String("table", payload.Table), zap.Bool("enabled", payload.Enabled))
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "rls updated"})
}

func (h *MetaHandler) DeleteTable(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	table := strings.TrimSpace(c.Query("table"))
	if table == "" {
		return c.Status(400).JSON(fiber.Map{"error": "table is required"})
	}

	if err := h.service.DeleteTable(c.Context(), schema, table); err != nil {
		h.log.Error("failed to delete table", zap.Error(err), zap.String("schema", schema), zap.String("table", table))
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "table deleted"})
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
func (h *MetaHandler) ListTriggers(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	triggers, err := h.service.GetTriggers(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(triggers)
}

func (h *MetaHandler) CreateTrigger(c *fiber.Ctx) error {
	var req services.CreateTriggerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := h.service.CreateTrigger(c.Context(), req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "trigger created"})
}

func (h *MetaHandler) DeleteTrigger(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	table := c.Query("table")
	name := c.Query("name")
	if table == "" || name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "table and name are required"})
	}
	if err := h.service.DeleteTrigger(c.Context(), schema, table, name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "trigger deleted"})
}

func (h *MetaHandler) ListForeignKeys(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	fks, err := h.service.GetForeignKeys(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fks)
}

func (h *MetaHandler) GetFullSchema(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	fullSchema, err := h.service.GetFullSchema(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fullSchema)
}

// ─────────────────────────────────────────────────────
// ENUM HANDLERS
// ─────────────────────────────────────────────────────

func (h *MetaHandler) ListEnums(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	enums, err := h.service.GetEnums(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if enums == nil {
		enums = make([]services.EnumMeta, 0)
	}
	return c.JSON(enums)
}

func (h *MetaHandler) CreateEnum(c *fiber.Ctx) error {
	var req services.CreateEnumRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := h.service.CreateEnum(c.Context(), req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "enum created"})
}

func (h *MetaHandler) UpdateEnum(c *fiber.Ctx) error {
	var req services.UpdateEnumRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := h.service.UpdateEnum(c.Context(), req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "enum updated"})
}

func (h *MetaHandler) DeleteEnum(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if err := h.service.DeleteEnum(c.Context(), schema, name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "enum deleted"})
}

// ─────────────────────────────────────────────────────
// INDEX HANDLERS
// ─────────────────────────────────────────────────────

func (h *MetaHandler) ListIndexes(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	indexes, err := h.service.GetIndexes(c.Context(), schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if indexes == nil {
		indexes = make([]services.IndexMeta, 0)
	}
	return c.JSON(indexes)
}

func (h *MetaHandler) CreateIndex(c *fiber.Ctx) error {
	var req services.CreateIndexRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if err := h.service.CreateIndex(c.Context(), req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(fiber.Map{"message": "index created"})
}

func (h *MetaHandler) DeleteIndex(c *fiber.Ctx) error {
	schema := c.Query("schema", "public")
	name := strings.TrimSpace(c.Query("name"))
	table := strings.TrimSpace(c.Query("table"))
	if name == "" || table == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name and table required"})
	}
	if err := h.service.DeleteIndex(c.Context(), schema, table, name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "index deleted"})
}

// ─────────────────────────────────────────────────────
// EXTENSION HANDLERS
// ─────────────────────────────────────────────────────

func (h *MetaHandler) ListExtensions(c *fiber.Ctx) error {
	exts, err := h.service.GetExtensions(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(exts)
}

func (h *MetaHandler) EnableExtension(c *fiber.Ctx) error {
	var payload struct {
		Name   string `json:"name"`
		Schema string `json:"schema"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid payload"})
	}
	if strings.TrimSpace(payload.Name) == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if err := h.service.EnableExtension(c.Context(), payload.Name, payload.Schema); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "extension enabled"})
}

func (h *MetaHandler) DisableExtension(c *fiber.Ctx) error {
	name := strings.TrimSpace(c.Query("name"))
	if name == "" {
		return c.Status(400).JSON(fiber.Map{"error": "name is required"})
	}
	if err := h.service.DisableExtension(c.Context(), name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"message": "extension disabled"})
}
