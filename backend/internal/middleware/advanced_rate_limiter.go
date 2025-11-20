package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

type RateLimiterConfig struct {
	// Global limits
	GlobalRPM int // Requests per minute globally
	GlobalRPH int // Requests per hour globally

	// Per-IP limits
	IPLimitRPM int
	IPLimitRPH int

	// Per-User limits
	UserLimitRPM int
	UserLimitRPH int

	// Per-Endpoint limits (specific endpoints can have custom limits)
	EndpointLimits map[string]EndpointLimit

	// Redis client
	RedisClient *redis.Client

	// Bypass for certain IPs (internal services, health checks)
	BypassIPs map[string]bool
}

type EndpointLimit struct {
	RPM int // Requests per minute
	RPH int // Requests per hour
}

var defaultConfig = RateLimiterConfig{
	GlobalRPM:    10000,
	GlobalRPH:    100000,
	IPLimitRPM:   60,
	IPLimitRPH:   1000,
	UserLimitRPM: 120,
	UserLimitRPH: 2000,
	EndpointLimits: map[string]EndpointLimit{
		"/api/v1/metrics/collect": {RPM: 120, RPH: 5000}, // Higher limit for metric collection
		"/api/v1/auth/login":      {RPM: 5, RPH: 20},     // Lower limit for login
		"/api/v1/agents/register": {RPM: 10, RPH: 50},    // Lower limit for agent registration
	},
	BypassIPs: map[string]bool{
		"127.0.0.1": true,
		"::1":       true,
	},
}

func AdvancedRateLimiter(redisClient *redis.Client) fiber.Handler {
	config := defaultConfig
	config.RedisClient = redisClient

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		// Bypass check
		if config.BypassIPs[ip] {
			return c.Next()
		}

		userID := c.Locals("user_id")
		endpoint := c.Path()

		ctx := context.Background()
		now := time.Now()

		// Check multiple rate limits
		checks := []struct {
			key     string
			limit   int
			window  time.Duration
			message string
		}{
			// Global limits
			{
				key:     fmt.Sprintf("ratelimit:global:minute:%d", now.Unix()/60),
				limit:   config.GlobalRPM,
				window:  time.Minute,
				message: "Global rate limit exceeded",
			},
			{
				key:     fmt.Sprintf("ratelimit:global:hour:%d", now.Unix()/3600),
				limit:   config.GlobalRPH,
				window:  time.Hour,
				message: "Global rate limit exceeded",
			},
			// Per-IP limits
			{
				key:     fmt.Sprintf("ratelimit:ip:%s:minute:%d", ip, now.Unix()/60),
				limit:   config.IPLimitRPM,
				window:  time.Minute,
				message: "IP rate limit exceeded. Please slow down.",
			},
			{
				key:     fmt.Sprintf("ratelimit:ip:%s:hour:%d", ip, now.Unix()/3600),
				limit:   config.IPLimitRPH,
				window:  time.Hour,
				message: "IP rate limit exceeded. Please try again later.",
			},
		}

		// Per-User limits (if authenticated)
		if userID != nil {
			userIDStr := fmt.Sprintf("%v", userID)
			checks = append(checks, []struct {
				key     string
				limit   int
				window  time.Duration
				message string
			}{
				{
					key:     fmt.Sprintf("ratelimit:user:%s:minute:%d", userIDStr, now.Unix()/60),
					limit:   config.UserLimitRPM,
					window:  time.Minute,
					message: "User rate limit exceeded",
				},
				{
					key:     fmt.Sprintf("ratelimit:user:%s:hour:%d", userIDStr, now.Unix()/3600),
					limit:   config.UserLimitRPH,
					window:  time.Hour,
					message: "User rate limit exceeded",
				},
			}...)
		}

		// Per-Endpoint limits
		if endpointLimit, exists := config.EndpointLimits[endpoint]; exists {
			checks = append(checks, []struct {
				key     string
				limit   int
				window  time.Duration
				message string
			}{
				{
					key:     fmt.Sprintf("ratelimit:endpoint:%s:minute:%d", endpoint, now.Unix()/60),
					limit:   endpointLimit.RPM,
					window:  time.Minute,
					message: fmt.Sprintf("Rate limit exceeded for %s", endpoint),
				},
				{
					key:     fmt.Sprintf("ratelimit:endpoint:%s:hour:%d", endpoint, now.Unix()/3600),
					limit:   endpointLimit.RPH,
					window:  time.Hour,
					message: fmt.Sprintf("Rate limit exceeded for %s", endpoint),
				},
			}...)
		}

		// Perform all checks
		for _, check := range checks {
			count, err := config.RedisClient.Incr(ctx, check.key).Result()
			if err != nil {
				logging.Errorf("Rate limiter Redis error: %v", err)
				// Fail open - allow request if Redis is down
				continue
			}

			// Set expiration on first increment
			if count == 1 {
				config.RedisClient.Expire(ctx, check.key, check.window)
			}

			// Check if limit exceeded
			if int(count) > check.limit {
				// Get remaining time
				ttl, _ := config.RedisClient.TTL(ctx, check.key).Result()
				retryAfter := int(ttl.Seconds())

				c.Set("X-RateLimit-Limit", strconv.Itoa(check.limit))
				c.Set("X-RateLimit-Remaining", "0")
				c.Set("X-RateLimit-Reset", strconv.Itoa(retryAfter))
				c.Set("Retry-After", strconv.Itoa(retryAfter))

				logging.Warnf("Rate limit exceeded for IP=%s, endpoint=%s, limit=%d", ip, endpoint, check.limit)

				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": fiber.Map{
						"code":        "RATE_LIMIT_EXCEEDED",
						"message":     check.message,
						"retry_after": retryAfter,
					},
				})
			}

			// Set rate limit headers
			remaining := check.limit - int(count)
			c.Set("X-RateLimit-Limit", strconv.Itoa(check.limit))
			c.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		}

		return c.Next()
	}
}
