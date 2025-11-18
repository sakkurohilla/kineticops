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

	// Serve agent binary (public).
	// Support both the older /agent-<os>-<arch> pattern and arbitrary
	// artifact names (e.g. kineticops-agent-linux-amd64) by exposing a
	// generic name route that delegates to the same handler. Keep the
	// specific pattern for backward compatibility.
	install.Get("/agent-:os-:arch", handlers.ServeAgentBinary)
	// Generic artifact name route (matches e.g. /kineticops-agent-linux-amd64)
	install.Get("/:name", handlers.ServeAgentBinary)

	// Serve arbitrary artifact by name (allows checksums like agent-... .sha256)
	install.Get("/file/:name", handlers.ServeArtifact)
}
