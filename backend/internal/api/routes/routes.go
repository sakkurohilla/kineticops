package routes

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterAllRoutes(app *fiber.App) {
	RegisterHealthRoutes(app)
	RegisterAuthRoutes(app)
	RegisterUserRoutes(app)
	RegisterHostRoutes(app)
	RegisterMetricRoutes(app)
	// Ensure agent ingestion routes are registered before the UI-facing log
	// routes so that un-authenticated agent POSTs (installation tokens) are
	// matched by the dedicated agent endpoints and not intercepted by the
	// UI auth middleware attached to the log routes.
	RegisterAgentRoutes(app)
	RegisterLogRoutes(app)
	RegisterAlertRoutes(app)
	RegisterWorkflowRoutes(app)
	RegisterAPMRoutes(app)
	RegisterSyntheticsRoutes(app)
	RegisterInstallRoutes(app)
	RegisterInternalRoutes(app)
}
