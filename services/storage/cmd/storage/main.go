package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/jackc/pgx/v5"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
	"github.com/machinelearningprodigy/OmniBase/shared/jwt"
	"github.com/machinelearningprodigy/OmniBase/shared/logger"
	"github.com/machinelearningprodigy/OmniBase/storage/internal/backends"
	"github.com/machinelearningprodigy/OmniBase/storage/internal/handlers"
	"github.com/machinelearningprodigy/OmniBase/storage/internal/services"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log := logger.MustNew(cfg.LogLevel, cfg.Env)
	defer log.Sync()

	// Connect to PostgreSQL metadata database
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		cfg.DatabaseUser, cfg.DatabasePassword, cfg.DatabaseHost, cfg.DatabasePort, cfg.DatabaseName)
	
	ctx := context.Background()
	db, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close(ctx)

	var backend backends.StorageBackend
	if cfg.StorageBackend == "minio" {
		backend, err = backends.NewMinIOBackend(
			cfg.MinIOEndpoint,
			cfg.MinIOAccessKey,
			cfg.MinIOSecretKey,
			cfg.MinIOUseSSL,
			log,
		)
		if err != nil {
			log.Fatal("failed to initialize MinIO backend", zap.Error(err))
		}
	} else {
		log.Fatal("unsupported storage backend requested")
	}

	jwtManager := jwt.NewManager(cfg.JWTSecret, 0, 0)
	storageService := services.NewStorageService(db, backend, log)
	storageHandler := handlers.NewStorageHandler(storageService, jwtManager, log)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Storage API Routes
	api := app.Group("/storage/v1")
	
	// Bucket Operations
	api.Get("/bucket", storageHandler.RequireAuth, storageHandler.ListBuckets)
	api.Post("/bucket", storageHandler.RequireAuth, storageHandler.CreateBucket)
	api.Delete("/bucket/:id", storageHandler.RequireAuth, storageHandler.DeleteBucket)
	
	// Object Operations
	api.Post("/object/:bucket/*", storageHandler.RequireAuth, storageHandler.UploadObject)
	api.Put("/object/:bucket/*", storageHandler.RequireAuth, storageHandler.UploadObject)
	api.Delete("/object/:bucket", storageHandler.RequireAuth, storageHandler.DeleteObjects)
	api.Get("/object/public/:bucket/*", storageHandler.GetPublicObject)
	
	// Pre-signed URLs for private access
	api.Post("/object/sign/:bucket/*", storageHandler.RequireAuth, storageHandler.CreateSignedURL)
	api.Get("/object/sign/:bucket/*", storageHandler.GetSignedObject)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down storage service...")
		app.Shutdown()
	}()

	port := fmt.Sprintf("%d", cfg.StoragePort)
	log.Info("Storage service starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Storage service stopped", zap.Error(err))
	}
}
