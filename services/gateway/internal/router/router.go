package router

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/gateway/internal/apilogs"
	"github.com/machinelearningprodigy/OmniBase/gateway/internal/middleware"
	"github.com/machinelearningprodigy/OmniBase/gateway/internal/proxy"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"go.uber.org/zap"
)

// Register sets up all routes in the API Gateway.
// The gateway acts as a reverse proxy to each OmniBase service.
//
// Route structure (mirrors Supabase's routing):
//
//	/auth/v1/*        → Auth Service
//	/rest/v1/*        → PostgREST (auto-REST API)
//	/graphql/v1/*     → pg_graphql (auto-GraphQL)
//	/storage/v1/*     → Storage Service
//	/realtime/v1/*    → Realtime Service (WebSocket upgrade)
//	/functions/v1/*   → Functions Service (Phase 2)
//	/meta/*           → Database Meta Service (admin only)
//	/admin/v1/*       → Admin API (projects, logs, config)
func Register(app *fiber.App, cfg *config.Config, log *zap.Logger) {
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// Initialize DB pool for project management
	var db *pgxpool.Pool
	pool, err := pgxpool.New(context.Background(), cfg.DSN())
	if err != nil {
		log.Warn("Gateway DB connection failed (project API will have limited functionality)", zap.Error(err))
	} else {
		db = pool
		// Ensure omnibase schema & projects table exist
		ensureProjectsTable(context.Background(), db, log)
	}

	// Middleware factories
	authRequired := middleware.RequireAuth(jwtManager, log)

	// ─── Auth Routes (public — gateway passes through to auth service) ────────
	authProxy := proxy.NewProxy(cfg.AuthServiceURL, log)
	auth := app.Group("/auth/v1")
	auth.Use(middleware.InjectUserContext(jwtManager))
	auth.All("/*", authProxy.ForwardWithAuth)

	// ─── REST API Routes (requires anon key or JWT) ───────────────────────────
	// PostgREST serves at / ; strip /rest/v1 so e.g. /rest/v1/todos → /todos
	restProxy := proxy.NewProxy(cfg.PostgRESTURL, log)
	rest := app.Group("/rest/v1")
	rest.Use(middleware.InjectUserContext(jwtManager))
	rest.All("/*", func(c *fiber.Ctx) error {
		p := c.Path()
		if strings.HasPrefix(p, "/rest/v1") {
			c.Path(strings.TrimPrefix(p, "/rest/v1"))
			if c.Path() == "" {
				c.Path("/")
			}
		}
		return restProxy.ForwardWithAuth(c)
	})

	// ─── Storage Routes ───────────────────────────────────────────────────────
	storageProxy := proxy.NewProxy(cfg.StorageServiceURL, log)
	storage := app.Group("/storage/v1")
	storage.Use(middleware.InjectUserContext(jwtManager))
	storage.All("/*", storageProxy.ForwardWithAuth)

	// ─── Realtime Routes (WebSocket) ──────────────────────────────────────────
	// WebSocket upgrade is handled by the gateway, then proxied to realtime service
	realtimeProxy := proxy.NewProxy(cfg.RealtimeServiceURL, log)
	app.Get("/realtime/v1/websocket", authRequired, realtimeProxy.ForwardWebSocket)

	// ─── Functions Routes (Phase 2 placeholder) ───────────────────────────────
	functionsProxy := proxy.NewProxy(cfg.FunctionsServiceURL, log)
	functions := app.Group("/functions/v1")
	functions.All("/*", functionsProxy.Forward)

	// ─── Database Meta Routes (admin only) — Renamed to /pg to match diagram ──
	pgMetaProxy := proxy.NewProxy(cfg.DatabaseServiceURL, log)
	pg := app.Group("/pg")
	pg.Use(authRequired)
	pg.All("/*", pgMetaProxy.Forward)

	// ─── GraphQL API Routes (hits the DB's pg_graphql extension) ──────────────
	app.All("/graphql/v1", middleware.InjectUserContext(jwtManager), func(c *fiber.Ctx) error {
		c.Path("/api/graphql") // Rewrite to database service internal route
		return pgMetaProxy.ForwardWithAuth(c)
	})

	// ─── Admin API Routes (dashboard uses this) ────────────────────────────────
	admin := app.Group("/admin/v1")
	admin.Use(authRequired)

	// API Logs
	admin.Get("/logs", func(c *fiber.Ctx) error {
		return c.JSON(apilogs.GetEntries())
	})

	// ─── Project Management ────────────────────────────────────────────────────
	// List all projects for the authenticated user
	admin.Get("/projects", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		rows, err := db.Query(context.Background(), `
			SELECT id, name, slug, description, anon_key, service_key, region, plan, created_at, updated_at
			FROM omnibase.projects
			WHERE owner_id = $1
			ORDER BY created_at ASC
		`, fmt.Sprintf("%v", userID))
		if err != nil {
			log.Error("failed to list projects", zap.Error(err))
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to list projects"})
		}
		defer rows.Close()

		type ProjectRow struct {
			ID          string    `json:"id"`
			Name        string    `json:"name"`
			Slug        string    `json:"slug"`
			Description *string   `json:"description"`
			AnonKey     string    `json:"anon_key"`
			ServiceKey  string    `json:"service_key"`
			Region      string    `json:"region"`
			Plan        string    `json:"plan"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		}
		var projects []ProjectRow
		for rows.Next() {
			var p ProjectRow
			rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.AnonKey, &p.ServiceKey, &p.Region, &p.Plan, &p.CreatedAt, &p.UpdatedAt)
			projects = append(projects, p)
		}
		if projects == nil {
			projects = []ProjectRow{}
		}
		return c.JSON(projects)
	})

	// Create a new project (auto-generates anon/service keys)
	admin.Post("/projects", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}

		var body struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Region      string `json:"region"`
		}
		if err := c.BodyParser(&body); err != nil || body.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "project name is required"})
		}

		// Generate stable anon + service_role keys for this project
		projectID := uuid.New().String()
		longLivedManager := jwt.NewManager(cfg.JWTSecret, 100*365*24*time.Hour, 100*365*24*time.Hour)
		anonKey, err := longLivedManager.IssueAnonymousToken(projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate anon key"})
		}
		serviceKey, err := longLivedManager.IssueAccessToken("service-account-"+projectID, "", "service_role", projectID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to generate service key"})
		}

		slug := makeSlug(body.Name)
		region := body.Region
		if region == "" {
			region = "local"
		}

		_, err = db.Exec(context.Background(), `
			INSERT INTO omnibase.projects (id, name, slug, description, owner_id, anon_key, service_key, region, plan)
			VALUES ($1, $2, $3, $4, $5::uuid, $6, $7, $8, 'free')
		`, projectID, body.Name, slug, body.Description, fmt.Sprintf("%v", userID), anonKey, serviceKey, region)
		if err != nil {
			log.Error("failed to create project", zap.Error(err))
			// Project may already have a slug conflict — try with unique suffix
			slug = slug + "-" + projectID[:8]
			_, err = db.Exec(context.Background(), `
				INSERT INTO omnibase.projects (id, name, slug, description, owner_id, anon_key, service_key, region, plan)
				VALUES ($1, $2, $3, $4, $5::uuid, $6, $7, $8, 'free')
			`, projectID, body.Name, slug, body.Description, fmt.Sprintf("%v", userID), anonKey, serviceKey, region)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create project"})
			}
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"id":          projectID,
			"name":        body.Name,
			"slug":        slug,
			"description": body.Description,
			"anon_key":    anonKey,
			"service_key": serviceKey,
			"region":      region,
			"plan":        "free",
			"url":         cfg.APIExternalURL,
		})
	})

	// Get a specific project by ID
	admin.Get("/projects/:id", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}

		var result struct {
			ID          string    `json:"id"`
			Name        string    `json:"name"`
			Slug        string    `json:"slug"`
			AnonKey     string    `json:"anon_key"`
			ServiceKey  string    `json:"service_key"`
			Region      string    `json:"region"`
			Plan        string    `json:"plan"`
			CreatedAt   time.Time `json:"created_at"`
		}
		err := db.QueryRow(context.Background(), `
			SELECT id, name, slug, anon_key, service_key, region, plan, created_at
			FROM omnibase.projects
			WHERE id = $1 AND owner_id = $2::uuid
		`, c.Params("id"), fmt.Sprintf("%v", userID)).Scan(
			&result.ID, &result.Name, &result.Slug, &result.AnonKey,
			&result.ServiceKey, &result.Region, &result.Plan, &result.CreatedAt,
		)
		if err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
		}
		return c.JSON(result)
	})

	// Delete a project
	admin.Delete("/projects/:id", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}
		if db == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "database not available"})
		}
		result, err := db.Exec(context.Background(),
			"DELETE FROM omnibase.projects WHERE id = $1 AND owner_id = $2::uuid",
			c.Params("id"), fmt.Sprintf("%v", userID),
		)
		if err != nil || result.RowsAffected() == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "project not found"})
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	// Project Configuration (Tokens) — returns keys for CURRENT project (from query param or first project)
	admin.Get("/config", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		if userID == nil || userID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "not authenticated"})
		}

		projectID := c.Query("project_id")
		if db != nil && projectID != "" {
			var anonKey, serviceKey string
			err := db.QueryRow(context.Background(), `
				SELECT anon_key, service_key FROM omnibase.projects
				WHERE id = $1 AND owner_id = $2::uuid
			`, projectID, fmt.Sprintf("%v", userID)).Scan(&anonKey, &serviceKey)
			if err == nil {
				return c.JSON(fiber.Map{
					"anonKey":    anonKey,
					"serviceKey": serviceKey,
					"url":        cfg.APIExternalURL,
					"project_id": projectID,
				})
			}
		}

		// Fall back to first project or generate ephemeral keys
		if db != nil {
			var anonKey, serviceKey, pID string
			err := db.QueryRow(context.Background(), `
				SELECT id, anon_key, service_key FROM omnibase.projects
				WHERE owner_id = $1::uuid
				ORDER BY created_at ASC LIMIT 1
			`, fmt.Sprintf("%v", userID)).Scan(&pID, &anonKey, &serviceKey)
			if err == nil {
				return c.JSON(fiber.Map{
					"anonKey":    anonKey,
					"serviceKey": serviceKey,
					"url":        cfg.APIExternalURL,
					"project_id": pID,
				})
			}
		}

		// No projects yet — generate ephemeral keys
		anonKey, _ := jwtManager.IssueAnonymousToken("")
		longManager := jwt.NewManager(cfg.JWTSecret, time.Hour*24*365*10, time.Hour*24*365*10)
		serviceKey, _ := longManager.IssueAccessToken("dashboard-admin", "", "service_role", "")
		return c.JSON(fiber.Map{
			"anonKey":    anonKey,
			"serviceKey": serviceKey,
			"url":        cfg.APIExternalURL,
		})
	})

	log.Info("Routes registered",
		zap.Strings("groups", []string{
			"/auth/v1", "/rest/v1", "/graphql/v1", "/storage/v1",
			"/realtime/v1", "/functions/v1", "/pg", "/admin/v1",
		}),
	)
}

// ensureProjectsTable creates the projects table if it doesn't exist
func ensureProjectsTable(ctx context.Context, db *pgxpool.Pool, log *zap.Logger) {
	_, err := db.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = 'omnibase') THEN
				EXECUTE 'CREATE SCHEMA omnibase';
			END IF;
		END
		$$;
		CREATE TABLE IF NOT EXISTS omnibase.projects (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        TEXT NOT NULL,
			slug        TEXT UNIQUE,
			description TEXT,
			owner_id    UUID,
			anon_key    TEXT,
			service_key TEXT,
			region      TEXT DEFAULT 'local',
			plan        TEXT DEFAULT 'free',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Warn("Failed to ensure projects table (may already exist)", zap.Error(err))
	}
}

// makeSlug converts a name to a URL-safe slug
func makeSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	var result []byte
	for i := 0; i < len(slug); i++ {
		c := slug[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result = append(result, c)
		}
	}
	// Trim leading/trailing dashes
	s := strings.Trim(string(result), "-")
	if s == "" {
		s = "project"
	}
	return s
}
