package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAgentRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Agent data collection endpoints (no auth required for agents)
	// Apply agent rate limiting middleware to protect ingestion endpoints
	api.Post("/agent/data", middleware.AgentRateLimit(), handlers.ReceiveAgentData)
	api.Post("/metrics/collect", middleware.AgentRateLimit(), handlers.ReceiveAgentData) // Alias for compatibility
	api.Post("/metrics/bulk", middleware.AgentRateLimit(), handlers.BulkIngestMetrics)   // High-throughput bulk ingestion (COPY)
	api.Post("/logs/collect", middleware.AgentRateLimit(), handlers.ReceiveAgentData)    // Alias for compatibility

	// Bootstrap endpoint for installer to request per-host Loki URL and token
	api.Post("/agents/bootstrap", handlers.BootstrapAgent)

	// Admin agent management endpoints (require user auth)
	agents := app.Group("/api/v1/agents", middleware.AuthRequired())
	agents.Post(":id/revoke", handlers.RevokeAgent)
	agents.Post(":id/unrevoke", handlers.UnrevokeAgent)
}
