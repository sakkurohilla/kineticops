package middleware

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gofiber/fiber/v2"
)

const csrfTokenKey = "csrf_token"

// CSRFMiddleware provides CSRF protection using double-submit cookie pattern
func CSRFMiddleware(c *fiber.Ctx) error {
	// Skip CSRF for authentication endpoints and agent data collection
	path := c.Path()
	if path == "/api/v1/auth/login" ||
		path == "/api/v1/auth/register" ||
		path == "/api/v1/auth/refresh" ||
		path == "/api/v1/auth/forgot-password" ||
		path == "/api/v1/auth/reset-password" ||
		path == "/api/v1/auth/verify-reset-token" ||
		path == "/api/v1/metrics/collect" ||
		path == "/api/v1/logs/collect" ||
		path == "/api/v1/traces/collect" ||
		len(path) >= 21 && path[:21] == "/api/v1/alert-rules" {
		return c.Next()
	}

	// Skip CSRF for workflow session endpoints (uses X-Session-Token authentication)
	if c.Method() == "POST" && (path == "/api/v1/workflow/session" ||
		len(path) > 17 && path[:17] == "/api/v1/workflow/") {
		return c.Next()
	}
	if c.Method() == "DELETE" && path == "/api/v1/workflow/session" {
		return c.Next()
	}

	// Skip CSRF for safe methods
	if c.Method() == "GET" || c.Method() == "HEAD" || c.Method() == "OPTIONS" {
		token := c.Cookies(csrfTokenKey)
		if token == "" {
			token = generateCSRFToken()
			c.Cookie(&fiber.Cookie{
				Name:     csrfTokenKey,
				Value:    token,
				HTTPOnly: true,
				Secure:   false, // Allow HTTP for local development
				SameSite: "Lax",
				MaxAge:   86400, // 24 hours
			})
		}
		return c.Next()
	}

	// Validate CSRF token for unsafe methods
	cookieToken := c.Cookies(csrfTokenKey)
	headerToken := c.Get("X-CSRF-Token")

	if cookieToken == "" || headerToken == "" || cookieToken != headerToken {
		return c.Status(403).JSON(fiber.Map{
			"error": "CSRF token validation failed",
		})
	}

	return c.Next()
}

func generateCSRFToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}
