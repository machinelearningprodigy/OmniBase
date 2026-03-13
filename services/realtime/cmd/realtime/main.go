package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"github.com/omnibase/omnibase/realtime/internal/hub"
	"github.com/omnibase/omnibase/realtime/internal/wal"
	"github.com/omnibase/omnibase/shared/config"
	"github.com/omnibase/omnibase/shared/jwt"
	"github.com/omnibase/omnibase/shared/logger"
	"go.uber.org/zap"
)

func main() {
	cfg := config.Load()
	log := logger.NewLogger(cfg.LogLevel)
	defer log.Sync()

	// Initialize JWT manager for validating WS connections
	jwtManager := jwt.NewJWTManager(cfg.JWTSecret, 15*time.Minute, 30*24*time.Hour)

	// Create Fiber app
	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	app.Use(recover.New())

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Create Realtime Hub
	eventHub := hub.NewHub(log)

	// Start WAL Consumer
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresHost, cfg.PostgresPort, cfg.PostgresDB)
	
	walConsumer, err := wal.NewConsumer(dsn, cfg.Get("REPLICATION_SLOT", "omnibase_realtime"), cfg.Get("REPLICATION_PUBLISHER", "omnibase_publication"), log)
	if err != nil {
		log.Fatal("failed to initialize WAL consumer", zap.Error(err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := walConsumer.Start(ctx); err != nil {
		log.Fatal("failed to start WAL consumer", zap.Error(err))
	}

	// Route WAL events to the Hub
	go func() {
		for event := range walConsumer.Events() {
			eventHub.BroadcastChange(event)
		}
	}()

	// WebSocket Upgrader Middleware
	app.Use("/realtime/v1/websocket", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket Handler
	app.Get("/realtime/v1/websocket", websocket.New(func(c *websocket.Conn) {
		clientID := uuid.New().String()
		log.Info("websocket connection opened", zap.String("client_id", clientID))

		// Try to extract user ID from token if provided
		var userID, role string
		token := c.Query("apikey")
		if token != "" {
			claims, err := jwtManager.VerifyToken(token)
			if err == nil {
				userID = claims.Subject
				if r, ok := claims.Claims["role"].(string); ok {
					role = r
				}
			}
		}
		if role == "" {
			role = "anon"
		}

		clientSend := make(chan []byte, 256)

		client := &hub.Client{
			ID:            clientID,
			UserID:        userID,
			Role:          role,
			Conn:          c,
			Subscriptions: []hub.Subscription{},
			Send:          clientSend, 
		}
		eventHub.Register(client)

		// Start a writer goroutine for this client
		go func() {
			for msg := range clientSend {
				if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
					log.Debug("websocket write error", zap.Error(err))
					break
				}
			}
		}()

		// Message read loop
		for {
			var msg map[string]interface{}
			if err := c.ReadJSON(&msg); err != nil {
				log.Debug("websocket read error", zap.Error(err))
				break
			}
			eventHub.HandleMessage(client, msg)
		}

		eventHub.Unregister(clientID)
	}))

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Info("Shutting down realtime service...")
		cancel() // Stop WAL consumer
		app.Shutdown()
	}()

	port := cfg.Get("REALTIME_PORT", "9004")
	log.Info("Realtime service starting", zap.String("port", port))
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("Realtime service stopped", zap.Error(err))
	}
}
