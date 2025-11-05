package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterWorkflowRoutes(app *fiber.App) {
	api := app.Group("/api/v1")

	// Workflow session management
	api.Post("/workflow/session", middleware.AuthRequired(), handlers.CreateWorkflowSession)
	api.Delete("/workflow/session", handlers.CloseWorkflowSession)

	// Service discovery and control
	api.Post("/workflow/:hostId/discover", handlers.DiscoverServices)
	api.Get("/hosts/:hostId/workflow", handlers.GetWorkflowData)
	api.Post("/services/:serviceId/control", middleware.AuthRequired(), handlers.ControlService)
	api.Get("/services/:serviceId/status", handlers.GetServiceStatus)
}
