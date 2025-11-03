package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAuthRoutes(app *fiber.App) {
	// PUBLIC auth routes (rate limiting disabled)
	auth := app.Group("/api/v1/auth")

	// Registration & Login
	auth.Post("/register", handlers.Register)
	auth.Post("/login", handlers.Login)
	auth.Post("/refresh", handlers.RefreshToken)

	// Password reset routes (public)
	auth.Post("/forgot-password", handlers.ForgotPassword)
	auth.Post("/verify-reset-token", handlers.VerifyResetToken)
	auth.Post("/reset-password", handlers.ResetPassword)

	// PROTECTED auth routes (authenticated user's own operations)
	authProtected := app.Group("/api/v1/auth")
	authProtected.Use(middleware.AuthMiddleware)

	// Current user profile
	authProtected.Get("/me", handlers.GetCurrentUser)
	authProtected.Put("/me", handlers.UpdateUser)
	authProtected.Delete("/me", handlers.DeleteUser)

	// Password management for authenticated user
	authProtected.Post("/change-password", handlers.ChangePassword)
}
