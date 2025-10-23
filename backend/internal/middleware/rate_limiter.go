package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        100,
		Expiration: 60 * 1000, // 1 minute in ms
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})
}
