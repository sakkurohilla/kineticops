package server

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/kineticops/backend/internal/config"
)

func StartServer(cfg *config.Config) {
	app := fiber.New()

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("🚀 KineticOps API running on port %s", cfg.AppPort)
	log.Fatal(app.Listen(addr))
}
