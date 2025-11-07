package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	redisrepo "github.com/sakkurohilla/kineticops/backend/internal/repository/redis"
)

// AgentRateLimit limits agent-origin requests using Redis counters.
// If Redis is unreachable we fall back to allowing the request (fail-open).
func AgentRateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only apply to agent-authenticated requests
		at := c.Locals("agent_token")
		if at == nil || at != true {
			return c.Next()
		}

		// Build key using agent_id when available, otherwise fall back to token header
		var key string
		if aid := c.Locals("agent_id"); aid != nil {
			key = fmt.Sprintf("ratelimit:agent:%v", aid)
		} else {
			token := c.Get("X-Agent-Token")
			key = fmt.Sprintf("ratelimit:agent:token:%s", token)
		}

		// Limits: 120 requests per minute by default (tunable)
		limit := int64(120)
		window := time.Minute

		ctx := context.Background()
		r := redisrepo.GetRedisClient()
		if r == nil {
			// No redis: allow (fail-open)
			return c.Next()
		}

		count, err := r.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			r.Expire(ctx, key, window)
		}
		if count > limit {
			// Rate limited
			return c.Status(429).JSON(fiber.Map{"error": "rate limit exceeded"})
		}
		// Optionally set headers
		c.Set("X-Ratelimit-Limit", fmt.Sprintf("%d", limit))
		c.Set("X-Ratelimit-Remaining", fmt.Sprintf("%d", limit-count))

		return c.Next()
	}
}
