package router

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/omnibase/omnibase/shared/config"
	"github.com/omnibase/omnibase/shared/jwt"
	"github.com/omnibase/omnibase/gateway/internal/middleware"
	"github.com/omnibase/omnibase/gateway/internal/proxy"
	"go.uber.org/zap"
)

// Register sets up all routes in the API Gateway.
// The gateway acts as a reverse proxy to each OmniBase service.
//
// Route structure (mirrors Supabase's routing):
//   /auth/v1/*        → Auth Service
//   /rest/v1/*        → PostgREST (auto-REST API)
//   /graphql/v1/*     → pg_graphql (auto-GraphQL)
//   /storage/v1/*     → Storage Service
//   /realtime/v1/*    → Realtime Service (WebSocket upgrade)
//   /functions/v1/*   → Functions Service (Phase 2)
//   /meta/*           → Database Meta Service (admin only)
//   /dashboard/*      → Admin Dashboard (Phase 1)
func Register(app *fiber.App, cfg *config.Config, log *zap.Logger) {
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

	// Middleware factories
	authRequired := middleware.RequireAuth(jwtManager, log)
	serviceRoleRequired := middleware.RequireServiceRole(jwtManager, log)

	// ─── Auth Routes (public — gateway passes through to auth service) ────────
	authProxy := proxy.NewProxy(cfg.AuthServiceURL, log)
	auth := app.Group("/auth/v1")
	auth.All("/*", authProxy.Forward)

	// ─── REST API Routes (requires anon key or JWT) ───────────────────────────
	// PostgREST handles the actual query; gateway validates auth and forwards
	restProxy := proxy.NewProxy(cfg.PostgRESTURL, log)
	rest := app.Group("/rest/v1")
	rest.Use(middleware.InjectUserContext(jwtManager))
	rest.All("/*", restProxy.ForwardWithAuth)

	// ─── GraphQL API Routes ───────────────────────────────────────────────────
	// pg_graphql is embedded in PostgREST in newer versions; or runs separately
	app.All("/graphql/v1/*", middleware.InjectUserContext(jwtManager), restProxy.ForwardWithAuth)

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
	app.All("/functions/v1/*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
			"code":    "not_implemented",
			"message": "Serverless Functions are coming in Phase 2. Subscribe to github.com/omnibase/omnibase for updates.",
		})
	})

	// ─── Database Meta Routes (admin only) ────────────────────────────────────
	metaProxy := proxy.NewProxy(fmt.Sprintf("http://localhost:%d", cfg.DatabaseMetaPort), log)
	meta := app.Group("/meta")
	meta.Use(serviceRoleRequired)
	meta.All("/*", metaProxy.Forward)

	// ─── Admin API Routes (dashboard uses this) ────────────────────────────────
	admin := app.Group("/admin/v1")
	admin.Use(serviceRoleRequired)

	// Project management
	admin.Get("/projects", listProjects)
	admin.Post("/projects", createProject)

	log.Info("Routes registered",
		zap.Strings("groups", []string{
			"/auth/v1", "/rest/v1", "/graphql/v1", "/storage/v1",
			"/realtime/v1", "/functions/v1", "/meta", "/admin/v1",
		}),
	)
}

// Placeholder handlers — will be moved to proper service in Phase 6
func listProjects(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"projects": []interface{}{}})
}

func createProject(c *fiber.Ctx) error {
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Project creation coming in Phase 6"})
}
