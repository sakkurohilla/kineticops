package routes

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/messaging/redpanda"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
	ws "github.com/sakkurohilla/kineticops/backend/internal/websocket"
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

		// Kafka/Redpanda health check (attempt to dial first broker)
		// We use a short timeout so the endpoint remains fast
		if err := redpanda.ProducerPing([]string{"localhost:9092"}, 500*time.Millisecond); err != nil {
			status["kafka"] = "unreachable"
		} else {
			status["kafka"] = "connected"
		}

		// WebSocket clients count
		status["ws_clients"] = ws.GetGlobalClientCount()

		// MongoDB health check (optional - if you want to add it)
		// Add MongoDB check here if needed

		return c.JSON(status)
	})
}
