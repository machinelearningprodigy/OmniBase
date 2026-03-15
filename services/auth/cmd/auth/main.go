package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/handlers"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/identity"
	"github.com/machinelearningprodigy/OmniBase/auth/internal/services"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
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

	// Initialize Identity layer
	idSvc := identity.NewService(authSvc.DB(), l, cfg, authSvc)
	idHandler := identity.NewHandler(idSvc, l)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "OmniBase Auth v1.0",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	})

	app.Use(recover.New())

	app.Use(idSvc.AttackGuardMiddleware())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "service": "auth"})
	})

	// Register auth routes at /auth/v1
	h := handlers.NewAuthHandler(authSvc, l)
	v1 := app.Group("/auth/v1")

	// ─── Public Routes ────────────────────────────────────────────────────────
	v1.Post("/signup", h.SignUp)
	v1.Post("/token", h.SignIn) // grant_type=password or refresh_token
	v1.Post("/logout", h.SignOut)
	v1.Post("/recover", h.RecoverPassword)
	v1.Post("/magiclink", h.MagicLink)
	v1.Post("/reset-password", h.ResetPassword)
	v1.Get("/verify", h.VerifyEmail)
	v1.Get("/providers", h.ListProviders)
	v1.Put("/user", h.UpdateUser)
	v1.Get("/user", h.GetUser)

	// ─── OAuth2 & Web3 Routes ─────────────────────────────────────────────────
	v1.Get("/authorize", h.OAuthAuthorize)
	v1.Get("/callback", h.OAuthCallback)
	v1.Post("/web3/nonce", idHandler.Web3Nonce)
	v1.Post("/web3/verify", idHandler.Web3Verify)

	// ─── Passkeys, MFA & Sessions & SAML ────────────────────────────────────────────────────────
	v1.Get("/sessions", idHandler.ListSessions)
	v1.Delete("/sessions/:id", idHandler.RevokeSession)
	v1.Post("/mfa/enroll", idHandler.EnrollMFA)
	v1.Get("/mfa", idHandler.ListMFA)
	v1.Post("/mfa/verify", idHandler.VerifyMFA)
	v1.Delete("/mfa/:id", idHandler.DeleteMFA)

	// Single Sign-On Consumer (ACS / Redirect)
	v1.Get("/saml/authorize", idHandler.GetSAMLAuthorize)
	v1.Post("/saml/acs/:project_id/:provider", idHandler.PostSAMLAssertion)

	// ─── Admin Routes (service role only) ─────────────────────────────────────
	admin := v1.Group("/admin")
	admin.Use(h.RequireAdmin)
	admin.Get("/users", idHandler.AdminListUsers)
	admin.Get("/users/:id", idHandler.AdminGetUser)
	admin.Put("/users/:id", idHandler.AdminUpdateUser)
	admin.Delete("/users/:id", idHandler.AdminDeleteUser)
	admin.Post("/users/:id/ban", idHandler.AdminBanUser)
	admin.Post("/invite", idHandler.AdminInviteUser)
	admin.Get("/users/:id/audit", idHandler.AdminGetUserAuditLogs)
	admin.Post("/generate-link", h.AdminGenerateLink)
	admin.Get("/audit", idHandler.AdminListAuditLogs)
	admin.Get("/oauth/apps", idHandler.ListOAuthApps)
	admin.Post("/oauth/apps", idHandler.SaveOAuthApp)

	// Config, Webhooks, Templates
	admin.Get("/config/templates", idHandler.GetEmailTemplates)
	admin.Post("/config/templates", idHandler.SaveEmailTemplate)
	admin.Get("/hooks", idHandler.GetAuthHooks)
	admin.Post("/hooks", idHandler.SaveAuthHook)
	admin.Delete("/hooks/:id", idHandler.DeleteAuthHook)
	admin.Post("/config/settings/:key", idHandler.SaveSettings)

	// SAML, Banned IPs, Auth OAuth Server
	admin.Get("/saml/providers", idHandler.ListIdentityProviders)
	admin.Post("/saml/providers", idHandler.SaveIdentityProvider)
	admin.Delete("/saml/providers/:id", idHandler.DeleteIdentityProvider)
	admin.Get("/oauth/apps", idHandler.ListOAuthApps)
	admin.Get("/protection/banned-ips", idHandler.GetBannedIPs)
	admin.Post("/protection/banned-ips", idHandler.BanIP)
	admin.Delete("/protection/banned-ips/:ip", idHandler.UnbanIP)
	admin.Get("/server/clients", idHandler.ListOAuthClients)
	admin.Get("/identity/rules", idHandler.ListIdentityRules)
	admin.Post("/identity/rules", idHandler.SaveIdentityRule)
	admin.Delete("/identity/rules/:id", idHandler.DeleteIdentityRule)
	admin.Get("/config/keys", idHandler.GetProjectKeys)

	// Passkeys WebAuthn endpoints
	v1.Post("/passkeys/challenge", idHandler.GeneratePasskeyChallenge)
	v1.Post("/passkeys/verify", idHandler.VerifyPasskey)

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
