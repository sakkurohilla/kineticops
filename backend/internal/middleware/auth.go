package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
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
	// Note: tenant_id currently maps to the owning user id until a separate
	// tenant model is introduced. This keeps existing behavior while allowing
	// a proper tenant field to be added later.
	c.Locals("tenant_id", claims.UserID)
	return c.Next()
}

// AuthOptional returns a middleware that will parse and validate a JWT if
// one is provided in the Authorization header. If no token is present the
// request proceeds unauthenticated (no tenant/user locals are set). This is
// used for endpoints that should be public but still want to know the
// authenticated user when a token is supplied (for tenant-scoped responses).
func AuthOptional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No token provided: continue without setting locals
			return c.Next()
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			// Token invalid - treat as unauthenticated (do not fail public requests)
			return c.Next()
		}
		// Store user info in context for downstream handlers
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		// See note in AuthMiddleware: tenant is currently mapped to user id
		c.Locals("tenant_id", claims.UserID)
		return c.Next()
	}
}

// AgentOrUserAuth allows either a valid JWT (regular user) or an agent token
// provided in the X-Agent-Token header. When an agent token is used we set a
// context flag "agent_token" so handlers can map host -> tenant as needed.
func AgentOrUserAuth() fiber.Handler {
	cfg := config.Load()
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := auth.ValidateJWT(tokenStr)
			if err != nil {
				return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
			}
			c.Locals("user_id", claims.UserID)
			c.Locals("username", claims.Username)
			c.Locals("tenant_id", claims.UserID)
			return c.Next()
		}

		// Fallback: check agent token header
		agentToken := c.Get("X-Agent-Token")
		if agentToken != "" && cfg.AgentToken != "" && agentToken == cfg.AgentToken {
			c.Locals("agent_token", true)
			return c.Next()
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid credentials"})
	}
}
