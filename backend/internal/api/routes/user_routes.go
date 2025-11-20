package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterUserRoutes(app *fiber.App) {
	// Admin-only user management routes
	users := app.Group("/api/v1/users", middleware.AuthMiddleware)

	users.Get("/", handlers.ListUsers)            // List all users
	users.Get("/:id", handlers.GetUserByID)       // Get specific user
	users.Put("/:id", handlers.UpdateUserByID)    // Update user
	users.Delete("/:id", handlers.DeleteUserByID) // Delete user

	// Current user profile routes
	profile := app.Group("/api/v1/profile", middleware.AuthMiddleware)
	profile.Get("/", handlers.GetCurrentUser) // Get current user profile
	profile.Put("/", handlers.UpdateUser)     // Update current user profile

	// Password management routes
	password := app.Group("/api/v1/password", middleware.AuthMiddleware)
	password.Put("/change", handlers.ChangePassword) // Change password

	// Settings routes
	settings := app.Group("/api/v1/settings", middleware.AuthMiddleware)
	settings.Get("/", handlers.GetSettings)    // Get user settings
	settings.Put("/", handlers.UpdateSettings) // Update user settings
}
