package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/machinelearningprodigy/OmniBase/shared/config"
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

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	api := app.Group("/functions/v1")
	api.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "OmniBase Edge Functions",
			"status":  "experimental",
		})
	})

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down functions service...")
		app.Shutdown()
	}()

	port := "9006"
	log.Info("Functions service starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Functions service stopped", zap.Error(err))
	}
}
