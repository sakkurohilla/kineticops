package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/kineticops/backend/internal/auth"
	"github.com/kineticops/backend/internal/config"
	"github.com/kineticops/backend/internal/database"
	"github.com/kineticops/backend/internal/handlers"
	"github.com/kineticops/backend/internal/repository"
	"github.com/kineticops/backend/internal/routes"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	log.Println("✅ Configuration loaded")

	// Initialize database
	database.ConnectDB(cfg)
	log.Println("✅ Connected to PostgreSQL successfully")
	defer database.CloseDB()

	app := fiber.New()

	// Health check route FIRST
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Init auth services
	jwtService := auth.NewJWTService(cfg.JWTSecret)
	userRepo := repository.NewUserRepository(repository.NewBaseRepository(database.DB))
	authHandler := handlers.NewAuthHandler(userRepo, jwtService)

	// Register routes
	routes.RegisterAuthRoutes(app, authHandler, jwtService)

	// Start server last
	log.Printf("🚀 Server running on port %s\n", cfg.AppPort)
	app.Listen(":" + cfg.AppPort)
}
