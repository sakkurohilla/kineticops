package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAuthRoutes(app *fiber.App) {
	// Rate limiter for public auth endpoints
	rl := limiter.New(limiter.Config{
		Max:        20,
		Expiration: 60 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Too many requests. Please try again in a minute.",
			})
		},
	})

	// PUBLIC auth routes (with rate limiting)
	auth := app.Group("/api/v1/auth", rl)

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
