package router

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
//	/dashboard/*      → Admin Dashboard (Phase 1)
func Register(app *fiber.App, cfg *config.Config, log *zap.Logger) {
	jwtManager := jwt.NewManager(cfg.JWTSecret, cfg.JWTAccessExpiry, cfg.JWTRefreshExpiry)

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

	// ─── GraphQL API Routes ────────────── (Placeholder moved down) ──────────

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

	// Project Configuration (Tokens)
	admin.Get("/config", func(c *fiber.Ctx) error {
		// Generate anon key
		anonKey, _ := jwtManager.IssueAnonymousToken("")

		// Generate service_role key
		serviceKey, _ := jwt.NewManager(cfg.JWTSecret, time.Hour*24*365*10, time.Hour*24*365*10).IssueAccessToken("dashboard-admin", "", "service_role", "")

		return c.JSON(fiber.Map{
			"anonKey":    anonKey,
			"serviceKey": serviceKey,
			"url":        cfg.APIExternalURL,
		})
	})

	// Project management
	admin.Get("/projects", func(c *fiber.Ctx) error {
		c.Path("/pg/schemas")
		return pgMetaProxy.Forward(c)
	})
	admin.Post("/projects", func(c *fiber.Ctx) error {
		c.Path("/pg/projects")
		return pgMetaProxy.Forward(c)
	})

	log.Info("Routes registered",
		zap.Strings("groups", []string{
			"/auth/v1", "/rest/v1", "/graphql/v1", "/storage/v1",
			"/realtime/v1", "/functions/v1", "/pg", "/admin/v1",
		}),
	)
}

// Placeholder handlers — will be moved to proper service in Phase 6
