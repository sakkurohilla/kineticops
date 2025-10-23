package routes

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/mongodb"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
)

func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		dbs := map[string]string{
			"postgres": "OK",
			"mongo":    "OK",
			"redis":    "OK",
		}

		// PostgreSQL
		sqlDB, err := postgres.DB.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbs["postgres"] = "FAIL"
		}

		// MongoDB
		if err := mongodb.Client.Ping(context.Background(), nil); err != nil {
			dbs["mongo"] = "FAIL"
		}

		// Redis
		if err := redisrepo.Client.Ping(context.Background()).Err(); err != nil {
			dbs["redis"] = "FAIL"
		}

		return c.JSON(fiber.Map{
			"status":  "OK",
			"details": dbs,
		})
	})
}
