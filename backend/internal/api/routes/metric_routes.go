package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterMetricRoutes(app *fiber.App) {
	api := app.Group("/api/v1/metrics", middleware.AuthMiddleware)
	api.Post("/collect", handlers.CollectMetric)
	api.Get("/", handlers.ListMetrics)
	api.Get("/latest", handlers.LatestMetric)
	api.Get("/prometheus", handlers.PrometheusExport)

	// ✅ COMMENT THIS OUT IF AggregateMetrics handler doesn't exist yet
	api.Get("/aggregate", handlers.AggregateMetrics)
}
