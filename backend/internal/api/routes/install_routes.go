package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterInstallRoutes(app *fiber.App) {
	install := app.Group("/api/v1/install")
	
	// Generate installation token (protected)
	install.Post("/token", middleware.AuthRequired(), handlers.GenerateInstallationToken)
	
	// Serve installation script (public)
	install.Get("/agent.sh", handlers.ServeInstallScript)
	
	// Serve agent binary (public)
	install.Get("/agent-:os-:arch", handlers.ServeAgentBinary)
}