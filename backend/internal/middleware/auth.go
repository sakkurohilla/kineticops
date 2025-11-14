package middleware

//nolint:staticcheck // intentional usage of deprecated package for compatibility; see repository/postgres/connection.go
import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/config"
	"github.com/sakkurohilla/kineticops/backend/internal/auth"
	"github.com/sakkurohilla/kineticops/backend/internal/repository/postgres"
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

		// Fallback: check agent token header. Support both a global AGENT_TOKEN
		// (for simple deployments) and per-agent tokens stored in Postgres.
		agentToken := c.Get("X-Agent-Token")
		if agentToken != "" {
			// 1) Global token match (legacy/simple mode)
			if cfg.AgentToken != "" && agentToken == cfg.AgentToken {
				c.Locals("agent_token", true)
				// tenant id not known for global token - leave tenant unset
				return c.Next()
			}

			// 2) Per-agent token: look up the agent and resolve its host -> tenant
			if agent, err := postgres.GetAgentByToken(agentToken); err == nil && agent != nil {
				// Reject revoked tokens immediately
				if agent.Revoked {
					return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Agent token revoked"})
				}
				// Resolve host to find tenant id
				if host, herr := postgres.GetHost(postgres.DB, int64(agent.HostID)); herr == nil && host != nil {
					c.Locals("agent_token", true)
					c.Locals("tenant_id", host.TenantID)
					c.Locals("agent_id", agent.ID)
					return c.Next()
				}
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid credentials"})
	}
}
