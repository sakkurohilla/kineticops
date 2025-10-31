package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAlertRoutes(app *fiber.App) {
	api := app.Group("/api/v1/alerts")
	
	// Public endpoint for stats (no auth required for dashboard)
	api.Get("/stats", handlers.GetAlertStats)
	
	// Protected endpoints
	apiAuth := app.Group("/api/v1/alerts", middleware.AuthRequired())
	apiAuth.Post("/rules", handlers.CreateAlertRule)
	apiAuth.Get("/rules", handlers.ListAlertRules)
	apiAuth.Get("/", handlers.ListAlerts)
	apiAuth.Patch("/:id", handlers.UpdateAlert)
	apiAuth.Post("/:id/acknowledge", handlers.UpdateAlert)
	apiAuth.Post("/:id/silence", handlers.UpdateAlert)
	apiAuth.Post("/:id/resolve", handlers.UpdateAlert)
}
