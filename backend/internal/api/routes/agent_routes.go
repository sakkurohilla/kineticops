package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
)

func RegisterAgentRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Agent data collection endpoints (no auth required for agents)
	api.Post("/agent/data", handlers.ReceiveAgentData)
	api.Post("/metrics/collect", handlers.ReceiveAgentData) // Alias for compatibility
	api.Post("/logs/collect", handlers.ReceiveAgentData)    // Alias for compatibility
}