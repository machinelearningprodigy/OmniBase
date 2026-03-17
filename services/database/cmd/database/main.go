package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/machinelearningprodigy/OmniBase/database/internal/handlers"
	"github.com/machinelearningprodigy/OmniBase/database/internal/services"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.MustNew(cfg.LogLevel, cfg.Env)
	defer log.Sync()

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Since we are validating tokens, we need the secret
	jwtManager := jwt.NewManager(cfg.JWTSecret, 15*time.Minute, 30*24*time.Hour)
	metaService := services.NewMetaService(db, log)
	metaHandler := handlers.NewMetaHandler(metaService, jwtManager, log, cfg.Env)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Public API (accessible via Gateway with any role)
	public := app.Group("/api")
	public.Post("/graphql", metaHandler.ResolveGraphQL)

	// Meta API configuration (raw queries, schema fetching for dashboard)
	api := app.Group("/pg")
	api.Use(metaHandler.RequireAdmin)

	// Get tables metadata
	api.Get("/tables", metaHandler.ListTables)
	api.Get("/columns", metaHandler.ListTableColumns)

	// Create a new table
	api.Post("/tables", metaHandler.CreateTable)
	api.Delete("/tables", metaHandler.DeleteTable)
	api.Post("/rls", metaHandler.ToggleRLS)

	// Create a new project (schema)
	api.Post("/projects", metaHandler.CreateProject)

	// Functions metadata
	api.Get("/functions", metaHandler.ListFunctions)

	// Policies metadata
	api.Get("/policies", metaHandler.ListPolicies)

	// Schemas metadata
	api.Get("/schemas", metaHandler.ListSchemas)
	api.Get("/full-schema", metaHandler.GetFullSchema)

	// Triggers
	api.Get("/triggers", metaHandler.ListTriggers)
	api.Post("/triggers", metaHandler.CreateTrigger)
	api.Delete("/triggers", metaHandler.DeleteTrigger)

	// Foreign Keys
	api.Get("/foreign-keys", metaHandler.ListForeignKeys)

	// Enum Types
	api.Get("/enums", metaHandler.ListEnums)
	api.Post("/enums", metaHandler.CreateEnum)
	api.Patch("/enums", metaHandler.UpdateEnum)
	api.Delete("/enums", metaHandler.DeleteEnum)

	// Indexes
	api.Get("/indexes", metaHandler.ListIndexes)
	api.Post("/indexes", metaHandler.CreateIndex)
	api.Delete("/indexes", metaHandler.DeleteIndex)

	// Extensions
	api.Get("/extensions", metaHandler.ListExtensions)
	api.Post("/extensions/enable", metaHandler.EnableExtension)
	api.Delete("/extensions/disable", metaHandler.DisableExtension)

	// Trigger schema reload
	api.Post("/reload", metaHandler.ReloadSchema)

	// Run arbitrary SQL queries
	api.Post("/query", metaHandler.RunQuery)
	api.Get("/migrations", metaHandler.ListMigrations)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down database service...")
		app.Shutdown()
	}()

	port := fmt.Sprintf("%d", cfg.DatabaseMetaPort)
	log.Info("Database meta service starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Database meta service stopped", zap.Error(err))
	}
}
