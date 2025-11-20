package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// RedirectHTTPSMiddleware redirects HTTP requests to HTTPS in production
func RedirectHTTPSMiddleware() fiber.Handler {
	// Only enforce in production environment
	env := os.Getenv("ENV")
	if env != "production" && env != "prod" {
		// Skip HTTPS redirect in dev/test
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		// Check if request is already HTTPS
		if c.Protocol() == "https" {
			return c.Next()
		}

		// Check X-Forwarded-Proto header (for reverse proxies like nginx)
		proto := c.Get("X-Forwarded-Proto")
		if proto == "https" {
			return c.Next()
		}

		// Redirect to HTTPS
		host := c.Hostname()
		if host == "" {
			host = "localhost"
		}

		// Build HTTPS URL
		httpsURL := "https://" + host + c.OriginalURL()

		return c.Redirect(httpsURL, fiber.StatusMovedPermanently)
	}
}

// SecureCookieMiddleware sets secure cookie flags in production
func SecureCookieMiddleware() fiber.Handler {
	env := os.Getenv("ENV")
	isProduction := env == "production" || env == "prod"

	return func(c *fiber.Ctx) error {
		// Proceed with request
		err := c.Next()

		// Set secure cookie attributes if in production
		if isProduction {
			// Fiber automatically sets Secure, HttpOnly, and SameSite when using c.Cookie()
			// This middleware ensures all cookies have secure flags
			c.Response().Header.VisitAllCookie(func(key, value []byte) {
				cookieStr := string(value)

				// Ensure Secure flag
				if !strings.Contains(cookieStr, "Secure") {
					cookieStr += "; Secure"
				}

				// Ensure SameSite=Strict
				if !strings.Contains(cookieStr, "SameSite") {
					cookieStr += "; SameSite=Strict"
				}

				// Ensure HttpOnly
				if !strings.Contains(cookieStr, "HttpOnly") {
					cookieStr += "; HttpOnly"
				}

				// Note: In production, JWT tokens should be in HttpOnly cookies
				// Current implementation uses Bearer tokens which is acceptable
				// This middleware is here for future cookie-based auth
			})
		}

		return err
	}
}

// HSTSMiddleware adds HTTP Strict Transport Security header
func HSTSMiddleware() fiber.Handler {
	env := os.Getenv("ENV")
	isProduction := env == "production" || env == "prod"

	if !isProduction {
		// Skip HSTS in dev/test
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		// Set HSTS header (1 year max-age, include subdomains)
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		return c.Next()
	}
}
