package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
)

func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		status := fiber.Map{"status": "ok"}

		// PostgreSQL health check
		sqlDB, err := postgres.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			status["postgres"] = "unreachable"
		} else {
			status["postgres"] = "connected"
		}

		// Redis health check
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		if err := redisrepo.Client.Ping(ctx).Err(); err != nil {
			status["redis"] = "unreachable"
		} else {
			status["redis"] = "connected"
		}

		// MongoDB health check (optional - if you want to add it)
		// Add MongoDB check here if needed

		return c.JSON(status)
	})
}
