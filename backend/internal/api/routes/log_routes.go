package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

// ✅ CHANGE THIS FUNCTION NAME
func RegisterLogRoutes(app *fiber.App) { // ← Was RegisterAlertRoutes
	api := app.Group("/api/v1/logs", middleware.AuthRequired())
	api.Post("/", handlers.CollectLog)
	api.Get("/", handlers.SearchLogs)
}
