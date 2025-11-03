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
	RegisterLogRoutes(app)
	RegisterAlertRoutes(app)
	RegisterWorkflowRoutes(app)
	RegisterAPMRoutes(app)
	RegisterSyntheticsRoutes(app)
	RegisterAgentRoutes(app)
	RegisterInstallRoutes(app)
	RegisterInternalRoutes(app)
}
