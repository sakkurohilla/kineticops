package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAlertRoutes(app *fiber.App) {
	api := app.Group("/api/v1/alerts", middleware.AuthRequired())
	api.Post("/rules", handlers.CreateAlertRule)
	api.Get("/rules", handlers.ListAlertRules)
	api.Get("/", handlers.ListAlerts)
	// Add more: alert status, ack, manual add etc.
}
