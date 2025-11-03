package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterAPMRoutes(app *fiber.App) {
	apm := app.Group("/api/v1/apm", middleware.AuthRequired())

	apm.Get("/applications", handlers.GetAPMApplications)
	apm.Post("/applications", handlers.CreateAPMApplication)
	apm.Get("/applications/:id/stats", handlers.GetAPMStats)
	apm.Get("/applications/:id/errors", handlers.GetAPMErrors)
	apm.Get("/applications/:id/traces", handlers.GetAPMTraces)
	apm.Get("/applications/:id/metrics", handlers.GetAPMMetrics)
}