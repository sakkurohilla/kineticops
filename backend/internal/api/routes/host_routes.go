package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

// RegisterHostRoutes registers host-related routes.
func RegisterHostRoutes(app *fiber.App) {
	// Agent heartbeat endpoint (public - agents authenticate with token)
	app.Post("/api/v1/agents/heartbeat", handlers.AgentHeartbeat)

	// Protected host routes (all require authentication)
	hosts := app.Group("/api/v1/hosts", middleware.AuthRequired())

	hosts.Post("/", handlers.CreateHost)
	hosts.Put("/:id", handlers.UpdateHost)
	hosts.Delete("/:id", handlers.DeleteHost)
	hosts.Post("/:id/heartbeat", handlers.HostHeartbeat)

	// Host management routes
	hosts.Get("/", handlers.ListHosts)
	hosts.Get("/:id", handlers.GetHost)
	hosts.Get("/:id/metrics", handlers.GetHostMetrics)
	hosts.Get("/:id/metrics/latest", handlers.GetHostLatestMetrics)
	// multi-host latest metrics (tenant-scoped)
	hosts.Get("/metrics/latest/all", handlers.GetAllHostsLatestMetrics)
	hosts.Get("/:id/metrics/range", handlers.GetHostMetricsTimeRange)
	hosts.Post("/test-ssh", handlers.TestSSHConnection)
	hosts.Post("/with-agent", handlers.CreateHostWithAgent)
	hosts.Get("/:id/services", handlers.GetHostServices)

}
