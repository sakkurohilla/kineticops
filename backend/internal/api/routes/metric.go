package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterMetricRoutes(app *fiber.App) {
	api := app.Group("/api/v1/metrics", middleware.AuthRequired())
	api.Post("/collect", handlers.CollectMetric)
	api.Get("/", handlers.ListMetrics)
	api.Get("/latest", handlers.LatestMetric)
	api.Get("/prometheus", handlers.PrometheusExport)
	// For Prometheus export (plain text), see below
}
