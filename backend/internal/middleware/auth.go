package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
)

// AuthRequired returns a middleware handler (your existing function - keep it)
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
		}
		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("tenant_id", claims.UserID)
		return c.Next()
	}
}

// AuthMiddleware - direct middleware function for use with .Use()
func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing token"})
	}
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateJWT(tokenStr)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}
	// Store user info in context
	c.Locals("user_id", claims.UserID)
	c.Locals("username", claims.Username)
	c.Locals("tenant_id", claims.UserID)
	return c.Next()
}
