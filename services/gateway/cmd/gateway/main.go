package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/machinelearningprodigy/OmniBase/gateway/internal/apilogs"
	"github.com/machinelearningprodigy/OmniBase/gateway/internal/router"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
	"go.uber.org/zap"
)

//	@title			OmniBase API Gateway
//	@version		1.0
//	@description	The single entry point for all OmniBase services
//	@termsOfService	https://omnibase.dev/terms
//	@contact.name	OmniBase Support
//	@contact.url	https://omnibase.dev/support
//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html
//	@host			localhost:8000
//	@BasePath		/

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log, err := logger.New(cfg.LogLevel, cfg.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	app := fiber.New(fiber.Config{
		AppName:               "OmniBase Gateway v1.0",
		ReadTimeout:           30 * time.Second,
		WriteTimeout:          30 * time.Second,
		IdleTimeout:           120 * time.Second,
		EnableTrustedProxyCheck: true,
		// Custom error handler that returns OmniBase-formatted errors
		ErrorHandler: errorHandler(log),
	})

	// ─── Global Middleware ────────────────────────────────────────────────────
	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Env != "production",
	}))
	app.Use(helmet.New())
	app.Use(compress.New(compress.Config{Level: compress.LevelBestSpeed}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*", // Configured per-project in Phase 6
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-OmniBase-Key, Accept-Profile, Content-Profile, Prefer",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: false,
	}))
	app.Use(limiter.New(limiter.Config{
		Max:        cfg.RateLimitRPS,
		Expiration: 1 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			// Rate limit per API key if present, otherwise per IP
			if key := c.Get("X-OmniBase-Key"); key != "" {
				return "key:" + key
			}
			return "ip:" + c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"code":    "rate_limit_exceeded",
				"message": "Too many requests. Please slow down.",
			})
		},
	}))

	// ─── Request Tracing Middleware ───────────────────────────────────────────
	app.Use(func(c *fiber.Ctx) error {
		traceID := c.Get("X-Request-ID")
		if traceID == "" {
			traceID = generateTraceID()
		}
		c.Set("X-Request-ID", traceID)
		c.Locals("trace_id", traceID)

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		log.Info("request",
			zap.String("method", c.Method()),
			zap.String("path", c.Path()),
			zap.Int("status", c.Response().StatusCode()),
			zap.Duration("latency", latency),
			zap.String("ip", c.IP()),
			zap.String("trace_id", traceID),
		)

		apilogs.AddEntry(apilogs.APILogEntry{
			ID:        traceID,
			Method:    c.Method(),
			Path:      c.Path(),
			Status:    c.Response().StatusCode(),
			LatencyMs: float64(latency.Microseconds()) / 1000.0,
			IP:        c.IP(),
			Timestamp: start,
		})

		return err
	})

	// ─── Routes ────────────────────────────────────────────────────────────────
	router.Register(app, cfg, log)

	// ─── Health Check ──────────────────────────────────────────────────────────
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "OmniBase API Gateway",
			"status":      "ok",
			"dashboard":   "http://localhost:3001",
			"health":      "/health",
			"ready":       "/ready",
			"auth":        "/auth/v1",
			"rest":        "/rest/v1",
			"graphql":     "/graphql/v1",
			"storage":     "/storage/v1",
			"realtime":    "/realtime/v1/websocket",
			"functions":   "/functions/v1",
			"admin":       "/admin/v1",
			"description": "Use the dashboard on port 3001 or call these APIs directly from your app.",
		})
	})

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"version": "1.0.0",
			"service": "gateway",
		})
	})

	// /health/services — checks all internal services and returns aggregated status.
	// Used by the dashboard (browser cannot directly reach internal service ports).
	app.Get("/health/services", func(c *fiber.Ctx) error {
		type ServiceCheck struct {
			Name    string `json:"name"`
			URL     string `json:"url"`
			Status  string `json:"status"`
			Latency int64  `json:"latency_ms"`
		}

		services := []struct {
			name string
			url  string
		}{
			{"Auth Service", cfg.AuthServiceURL + "/health"},
			{"Storage Service", cfg.StorageServiceURL + "/health"},
			{"Realtime Service", cfg.RealtimeServiceURL + "/health"},
			{"Functions Service", cfg.FunctionsServiceURL + "/health"},
			{"Database Service", cfg.DatabaseServiceURL + "/health"},
			{"PostgREST", cfg.PostgRESTURL + "/"},
		}

		results := make([]ServiceCheck, len(services))
		var wg sync.WaitGroup
		for i, svc := range services {
			wg.Add(1)
			go func(idx int, name, url string) {
				defer wg.Done()
				start := time.Now()
				ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
				defer cancel()
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
				if err != nil {
					results[idx] = ServiceCheck{Name: name, URL: url, Status: "down", Latency: 0}
					return
				}
				resp, err := http.DefaultClient.Do(req)
				latency := time.Since(start).Milliseconds()
				if err != nil || resp == nil {
					results[idx] = ServiceCheck{Name: name, URL: url, Status: "down", Latency: 0}
					return
				}
				defer resp.Body.Close()
				status := "healthy"
				if resp.StatusCode >= 500 {
					status = "degraded"
				} else if resp.StatusCode >= 400 {
					status = "degraded"
				}
				results[idx] = ServiceCheck{Name: name, URL: url, Status: status, Latency: latency}
			}(i, svc.name, svc.url)
		}
		wg.Wait()

		healthyCount := 0
		for _, r := range results {
			if r.Status == "healthy" {
				healthyCount++
			}
		}
		overall := "healthy"
		if healthyCount == 0 {
			overall = "down"
		} else if healthyCount < len(results) {
			overall = "degraded"
		}
		return c.JSON(fiber.Map{
			"status":   overall,
			"services": results,
			"healthy":  healthyCount,
			"total":    len(results),
		})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.AuthServiceURL+"/health", nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil || resp == nil || resp.StatusCode != http.StatusOK {
			if resp != nil {
				resp.Body.Close()
			}
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"status": "not_ready", "reason": "auth_unavailable"})
		}
		resp.Body.Close()
		return c.JSON(fiber.Map{"status": "ready"})
	})

	// ─── Graceful Shutdown ────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf("%s:%d", cfg.GatewayHost, cfg.GatewayPort)
		log.Info("OmniBase Gateway starting", zap.String("addr", addr))
		if err := app.Listen(addr); err != nil {
			log.Fatal("Gateway failed to start", zap.Error(err))
		}
	}()

	<-quit
	log.Info("Shutting down gateway...")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Error("Forced shutdown", zap.Error(err))
	}
	log.Info("Gateway stopped cleanly")
}

func errorHandler(log *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		log.Error("unhandled error",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.Int("status", code),
		)

		return c.Status(code).JSON(fiber.Map{
			"code":    "internal_error",
			"message": err.Error(),
		})
	}
}

var traceCounter uint64

func generateTraceID() string {
	// Simple trace ID — in Phase 4 this becomes OpenTelemetry trace ID
	return fmt.Sprintf("omni-%d-%d", time.Now().UnixNano(), traceCounter)
}
