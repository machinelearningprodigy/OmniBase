package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/handlers"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	l, err := logger.New(cfg.LogLevel, cfg.Env)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer l.Sync()

	// Initialize services
	authSvc, err := services.NewAuthService(cfg, l)
	if err != nil {
		l.Fatal("Failed to create auth service", zap.Error(err))
	}
	defer authSvc.Close()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "OmniBase Auth v1.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(recover.New())
	app.Use(cors.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "auth"})
	})

	// Register auth routes at /auth/v1
	h := handlers.NewAuthHandler(authSvc, l)
	v1 := app.Group("/auth/v1")

	// ─── Public Routes ────────────────────────────────────────────────────────
	v1.Post("/signup", h.SignUp)
	v1.Post("/token", h.SignIn)              // grant_type=password or refresh_token
	v1.Post("/logout", h.SignOut)
	v1.Post("/recover", h.RecoverPassword)   // Send password reset email
	v1.Get("/verify", h.VerifyEmail)         // Email verification link
	v1.Put("/user", h.UpdateUser)            // Update current user
	v1.Get("/user", h.GetUser)               // Get current user

	// ─── OAuth2 Routes ────────────────────────────────────────────────────────
	v1.Get("/authorize", h.OAuthAuthorize)   // Redirect to provider
	v1.Get("/callback", h.OAuthCallback)     // Handle provider callback

	// ─── Admin Routes (service role only) ─────────────────────────────────────
	admin := v1.Group("/admin")
	admin.Get("/users", h.AdminListUsers)
	admin.Get("/users/:id", h.AdminGetUser)
	admin.Put("/users/:id", h.AdminUpdateUser)
	admin.Delete("/users/:id", h.AdminDeleteUser)
	admin.Post("/users/:id/ban", h.AdminBanUser)
	admin.Post("/generate-link", h.AdminGenerateLink)

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.AuthPort)
	l.Info("Auth service starting", zap.String("addr", addr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Listen(addr); err != nil {
			l.Fatal("Auth service failed", zap.Error(err))
		}
	}()

	<-quit
	l.Info("Auth service shutting down...")
	app.ShutdownWithTimeout(10 * time.Second)
}
