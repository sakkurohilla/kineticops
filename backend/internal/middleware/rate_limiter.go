package middleware

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// RateLimiter applies a per-IP limiter but only for API endpoints (paths
// starting with /api/). This avoids rate-limiting static assets, the SPA
// frontend, or the WebSocket upgrade path which can cause confusing client
// behavior in development.
func RateLimiter() fiber.Handler {
	// create limiter handler once
	lim := limiter.New(limiter.Config{
		Max:        100,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(429).JSON(fiber.Map{
				"error": "Rate limit exceeded",
			})
		},
	})

	return func(c *fiber.Ctx) error {
		// Only apply to API endpoints (e.g. /api/...). Skip for WS and static files.
		p := c.Path()
		if !strings.HasPrefix(p, "/api/") {
			return c.Next()
		}
		return lim(c)
	}
}
