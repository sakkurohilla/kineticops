package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/api/routes"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"

	// DB imports
	"github.com/sakkurohilla/kineticops/backend/internal/repository/mongodb"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
)

func main() {
	// Load Config
	cfg := config.Load()

	// Initialize Databases
	if err := postgres.Init(); err != nil {
		log.Fatalf("PostgreSQL connection error: %v", err)
	}
	if err := mongodb.Init(); err != nil {
		log.Fatalf("MongoDB connection error: %v", err)
	}
	if err := redisrepo.Init(); err != nil {
		log.Fatalf("Redis connection error: %v", err)
	}

	log.Println("All DBs initialized.")

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// Middlewares
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())
	app.Use(middleware.RateLimiter())

	// Health route (now checks DBs too)
	routes.RegisterHealthRoutes(app)

	log.Printf("Starting server on port %s...", cfg.AppPort)
	if err := app.Listen(":" + cfg.AppPort); err != nil {
		log.Fatal(err)
	}
}
