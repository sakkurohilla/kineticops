package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityHeaders adds comprehensive security headers to all responses
func SecurityHeaders() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// HTTP Strict Transport Security (HSTS)
		// Tells browsers to only connect via HTTPS for the next year
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// X-Frame-Options
		// Prevents clickjacking attacks by not allowing the page to be framed
		c.Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options
		// Prevents MIME type sniffing
		c.Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection
		// Enables XSS filter built into most browsers
		c.Set("X-XSS-Protection", "1; mode=block")

		// Content-Security-Policy (CSP)
		// Restricts sources of content that can be loaded
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss:; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Set("Content-Security-Policy", csp)

		// Referrer-Policy
		// Controls how much referrer information is included with requests
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (formerly Feature-Policy)
		// Controls which browser features can be used
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

		// X-Permitted-Cross-Domain-Policies
		// Restricts Adobe Flash and PDF cross-domain requests
		c.Set("X-Permitted-Cross-Domain-Policies", "none")

		// Remove server identification header
		c.Set("Server", "")

		// Add custom security header
		c.Set("X-Security-Headers", "enabled")

		return c.Next()
	}
}

// HTTPSRedirect redirects HTTP to HTTPS in production
func HTTPSRedirect() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request is already HTTPS
		if c.Protocol() == "https" {
			return c.Next()
		}

		// Check X-Forwarded-Proto header (for reverse proxies)
		if c.Get("X-Forwarded-Proto") == "https" {
			return c.Next()
		}

		// Skip redirect for localhost/development
		if c.Hostname() == "localhost" || c.Hostname() == "127.0.0.1" {
			return c.Next()
		}

		// Redirect to HTTPS
		return c.Redirect("https://"+c.Hostname()+c.OriginalURL(), fiber.StatusMovedPermanently)
	}
}
