package middleware

import (
	"github.com/gofiber/fiber/v2"
)

func RateLimiter() fiber.Handler {
	// Disabled rate limiter - return no-op handler
	return func(c *fiber.Ctx) error {
		return c.Next()
	}
}
