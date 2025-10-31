package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

// RegisterHostRoutes registers host-related routes.
// GET /api/v1/hosts and GET /api/v1/hosts/:id are public (no auth).
// Other host-management endpoints remain protected by AuthRequired.
func RegisterHostRoutes(app *fiber.App) {
	// Host endpoints require authentication to ensure tenant isolation.
	// Exposing host lists publicly caused hosts from other users to be
	// visible to new signups — switch GET endpoints back to AuthRequired.
	hostsPublic := app.Group("/api/v1/hosts", middleware.AuthRequired())
	hostsPublic.Get("/", handlers.ListHosts)
	hostsPublic.Get(":id", handlers.GetHost)

	// Protected host routes (create/update/delete, heartbeat, test-ssh, metrics)
	hosts := app.Group("/api/v1/hosts", middleware.AuthRequired())

	hosts.Post("/", handlers.CreateHost)
	hosts.Put("/:id", handlers.UpdateHost)
	hosts.Delete("/:id", handlers.DeleteHost)
	hosts.Post("/:id/heartbeat", handlers.HostHeartbeat)

	// NEW ROUTES (protected)
	hosts.Post("/test-ssh", handlers.TestSSHConnection)
	hosts.Get("/:id/metrics", handlers.GetHostMetrics)
	hosts.Get("/:id/metrics/latest", handlers.GetHostLatestMetrics)
	hosts.Get("/:id/metrics/range", handlers.GetHostMetricsTimeRange)
	hosts.Post("/:id/collect", handlers.CollectHostNow)

	// Agent setup and management routes
	hosts.Post("/with-agent", handlers.CreateHostWithAgent)
	hosts.Get("/:id/agent/status", handlers.GetAgentStatus)
	hosts.Get("/:id/services", handlers.GetHostServices)

	// Agent heartbeat endpoint (public - agents authenticate with token)
	app.Post("/api/v1/agents/heartbeat", handlers.AgentHeartbeat)
}
