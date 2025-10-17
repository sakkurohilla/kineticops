package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kineticops/backend/internal/auth"
	"github.com/kineticops/backend/internal/handlers"
	"github.com/kineticops/backend/internal/middleware"
)

func RegisterAuthRoutes(app *fiber.App, h *handlers.AuthHandler, jwtService *auth.JWTService) {
	authGroup := app.Group("/api/v1/auth")

	authGroup.Post("/register", h.Register)
	authGroup.Post("/login", h.Login)

	protected := authGroup.Group("/", middleware.Protected(jwtService))

	protected.Get("/me", func(c *fiber.Ctx) error {
		userID := c.Locals("user_id")
		return c.JSON(fiber.Map{"user_id": userID})
	})
}
