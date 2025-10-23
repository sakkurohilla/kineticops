package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/api/routes"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"

	"github.com/sakkurohilla/kineticops/backend/internal/models"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// Load Config
	cfg := config.Load()

	// Initialize Databases
	if err := postgres.Init(); err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}
	if err := redisrepo.Init(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	// MongoDB client setup (for logs)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	// e.g. "mongodb://localhost:27017"
	if err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("MongoDB ping error: %v", err)
	}
	models.LogCollection = mongoClient.Database("kineticops").Collection("logs")
	log.Println("MongoDB (logs) connected.")

	log.Println("All DBs initialized.")

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimiter())

	// Health check route
	routes.RegisterHealthRoutes(app)

	// Auth routes (login, register, hashing, etc.)
	routes.RegisterAuthRoutes(app)

	// Host routes
	routes.RegisterHostRoutes(app)

	// Metric routes
	routes.RegisterMetricRoutes(app)

	// Log routes (NEW for Day 8)
	routes.RegisterLogRoutes(app)

	// Alert routes
	routes.RegisterAlertRoutes(app)

	// A sample protected route
	app.Get("/protected", middleware.AuthRequired(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"msg":      "Access granted",
			"user_id":  c.Locals("user_id"),
			"username": c.Locals("username"),
		})
	})

	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
