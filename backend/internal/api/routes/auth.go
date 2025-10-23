package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
)

func RegisterAuthRoutes(app *fiber.App) {
	rl := limiter.New(limiter.Config{
		Max:        5,
		Expiration: 60 * 1000, // 5 requests/minute per IP
	})

	api := app.Group("/auth", rl)

	api.Post("/register", handlers.Register)
	api.Post("/login", handlers.Login)
	api.Post("/forgot-password", handlers.ForgotPassword)
	api.Post("/refresh", handlers.RefreshToken)
}
