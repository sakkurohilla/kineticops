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

	// Health check route
	routes.RegisterHealthRoutes(app)

	// Auth routes (login, register, hashing, etc.)
	routes.RegisterAuthRoutes(app)

	// Host routes ( CRUD )
	routes.RegisterHostRoutes(app)

	// Metric routes
	routes.RegisterMetricRoutes(app)

	// A sample protected route (requires valid JWT)
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
