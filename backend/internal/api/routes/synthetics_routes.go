package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterSyntheticsRoutes(app *fiber.App) {
	synthetics := app.Group("/api/v1/synthetics", middleware.AuthRequired())

	synthetics.Get("/monitors", handlers.GetSyntheticMonitors)
	synthetics.Post("/monitors", handlers.CreateSyntheticMonitor)
	synthetics.Get("/monitors/:id/results", handlers.GetSyntheticResults)
	synthetics.Get("/monitors/:id/stats", handlers.GetSyntheticStats)
}
