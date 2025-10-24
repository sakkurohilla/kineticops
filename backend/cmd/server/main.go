package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"

	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/api/routes"
	kafkaevents "github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongoClient *mongo.Client

func initDBs(cfg *config.Config) {
	// PostgreSQL
	if err := postgres.Init(); err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}

	// Redis
	if err := redisrepo.Init(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// MongoDB (optional, for logs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Printf("[WARN] MongoDB connection error: %v", err)
		return // Don't fatal if nonessential!
	}
	if pingErr := mongoClient.Ping(ctx, nil); pingErr != nil {
		log.Printf("[WARN] MongoDB ping error: %v", pingErr)
		return // Don't fatal!
	}
	models.LogCollection = mongoClient.Database("kineticops").Collection("logs")
	log.Println("MongoDB (logs) connected.")

	log.Println("All DBs initialized.")
}

func main() {
	cfg := config.Load()
	fmt.Println("[DEBUG] JWTSecret from config (main.go):", cfg.JWTSecret)

	initDBs(cfg)

	// Set up Redpanda/Kafka
	brokers := []string{"localhost:9092"}
	topic := "metrics-events"

	kafkaevents.InitProducer(brokers, topic)

	// WebSocket hub
	wsHub := ws.NewHub()
	go wsHub.Run()

	// Kafka consumer broadcast to WebSocket clients
	kafkaevents.StartConsumer(brokers, topic, func(msg []byte) {
		fmt.Println("[DEBUG] Broadcasting to WebSocket clients:", string(msg))
		wsHub.Broadcast(msg)
	})

	// Fiber app with error handler
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimiter())

	// Health route (robust, production-safe)
	app.Get("/health", func(c *fiber.Ctx) error {
		status := fiber.Map{"status": "ok"}
		// Optionally test DB connections
		if mongoClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			if err := mongoClient.Ping(ctx, nil); err != nil {
				status["mongo"] = "unreachable"
			} else {
				status["mongo"] = "connected"
			}
		} else {
			status["mongo"] = "not_initialized"
		}
		// Other DB checks (add as needed)
		return c.JSON(status)
	})

	// Register all API routes (previous days)
	routes.RegisterAuthRoutes(app)
	routes.RegisterHostRoutes(app)
	routes.RegisterMetricRoutes(app)
	routes.RegisterLogRoutes(app)
	routes.RegisterAlertRoutes(app)

	// Protected route
	app.Get("/protected", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"msg":      "Access granted",
			"user_id":  c.Locals("user_id"),
			"username": c.Locals("username"),
		})
	})

	// WebSocket JWT-enabled endpoint
	app.Get("/ws", websocket.New(ws.WsHandler(wsHub, cfg.JWTSecret)))

	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
