package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

// RegisterMetricRoutes registers metric-related routes under /api/v1/metrics
func RegisterMetricRoutes(app *fiber.App) {
	// Public agent endpoints (no auth required)
	app.Post("/api/v1/metrics/collect", handlers.ReceiveAgentData) // Agent data collection
	
	// Protected user endpoints
	metrics := app.Group("/api/v1/metrics", middleware.AuthRequired())

	metrics.Get("/range", handlers.GetMetricsRange)       // GET /api/v1/metrics/range?range=24h
	metrics.Post("/telegraf", handlers.IngestTelegraf)    // POST /api/v1/metrics/telegraf
	metrics.Get("/prometheus", handlers.PrometheusExport) // GET /api/v1/metrics/prometheus
}
