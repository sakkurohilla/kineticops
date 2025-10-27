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
		return
	}
	if pingErr := mongoClient.Ping(ctx, nil); pingErr != nil {
		log.Printf("[WARN] MongoDB ping error: %v", pingErr)
		return
	}
	models.LogCollection = mongoClient.Database("kineticops").Collection("logs")
	log.Println("MongoDB (logs) connected.")

	log.Println("All DBs initialized.")
}

func main() {
	cfg := config.Load()
	fmt.Println("[INFO] Starting KineticOps Server...")
	fmt.Println("[DEBUG] JWTSecret loaded:", cfg.JWTSecret != "")

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

	// Global middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimiter())

	// ✅ Register ALL routes through unified router
	routes.RegisterAllRoutes(app)

	// Protected demo route
	app.Get("/protected", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"msg":      "Access granted",
			"user_id":  c.Locals("user_id"),
			"username": c.Locals("username"),
		})
	})

	// WebSocket JWT-enabled endpoint
	app.Get("/ws", websocket.New(ws.WsHandler(wsHub, cfg.JWTSecret)))

	log.Printf("✅ Server started successfully on port %s", cfg.AppPort)
	log.Printf("📊 Health check: http://localhost:%s/health", cfg.AppPort)
	log.Printf("🔐 API Base: http://localhost:%s/api/v1", cfg.AppPort)

	log.Printf("✅ Server started successfully on port %s", cfg.AppPort)
	log.Printf("📊 Health check: http://localhost:%s/health", cfg.AppPort)
	log.Printf("🔐 API Base: http://localhost:%s/api/v1", cfg.AppPort)
	log.Printf("🌐 Network: http://0.0.0.0:%s", cfg.AppPort)

	// ✅ IMPORTANT: Listen on all interfaces (0.0.0.0)
	if err := app.Listen("0.0.0.0:" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}

}
