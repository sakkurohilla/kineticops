package middleware

import (
	"kineticops/backend/config"
	"kineticops/backend/handlers"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
)

func Protected(db *sqlx.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
		}

		const bearer = "Bearer "
		if !strings.HasPrefix(auth, bearer) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid authorization header"})
		}

		tokenString := strings.TrimPrefix(auth, bearer)
		token, err := jwt.ParseWithClaims(tokenString, &handlers.Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		claims, ok := token.Claims.(*handlers.Claims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid token claims"})
		}

		c.Locals("userID", claims.UserID)
		c.Locals("role", claims.Role)

		return c.Next()
	}
}
