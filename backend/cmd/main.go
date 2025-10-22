package main

import (
	"kineticops/backend/config"
	"kineticops/backend/handlers"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.Load()
	log.Println("JWT Secret length:", len(cfg.JwtSecret))

	db, err := sqlx.Connect("postgres", cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()
	log.Println("Connected to Postgres")

	app := fiber.New()
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/api/register", handlers.RegisterUser(db, cfg))
	app.Post("/api/login", handlers.LoginUser(db, cfg))
	app.Get("/api/profile", handlers.ProtectedRoute(db, cfg))

	log.Println("Starting server at :8080")
	log.Fatal(app.Listen(":8080"))
}
