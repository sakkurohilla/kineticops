package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

// RegisterInternalRoutes registers internal/admin routes like telemetry snapshot
func RegisterInternalRoutes(app *fiber.App) {
	internal := app.Group("/api/v1/internal", middleware.AuthRequired())
	internal.Get("/telemetry", handlers.Telemetry)
	// Development debug endpoints (authenticated admin only)
	internal.Post("/debug/ws/burst", handlers.DebugWSSend)
}
