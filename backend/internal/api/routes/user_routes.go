package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterUserRoutes(app *fiber.App) {
	// Admin-only user management routes
	users := app.Group("/api/v1/users", middleware.AuthMiddleware)

	// TODO: Add role-based access control (admin only)
	users.Get("/", handlers.ListUsers)            // List all users
	users.Get("/:id", handlers.GetUserByID)       // Get specific user (create this handler)
	users.Put("/:id", handlers.UpdateUserByID)    // Update user (create this handler)
	users.Delete("/:id", handlers.DeleteUserByID) // Delete user (create this handler)
}
