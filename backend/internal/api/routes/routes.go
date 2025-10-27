package routes

import (
	"github.com/gofiber/fiber/v2"
)

// RegisterAllRoutes registers all application routes
func RegisterAllRoutes(app *fiber.App) {
	// Health check (no auth required)
	RegisterHealthRoutes(app)

	// Authentication routes (public)
	RegisterAuthRoutes(app)

	// User management routes (protected)
	RegisterUserRoutes(app)

	// Resource routes (protected)
	RegisterHostRoutes(app)
	RegisterMetricRoutes(app)
	RegisterLogRoutes(app) // ✅ This should now work
	RegisterAlertRoutes(app)
}
