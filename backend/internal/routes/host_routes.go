package routes

import (
	"kineticops/backend/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterHostRoutes(app fiber.Router, h *handlers.HostHandler) {
	group := app.Group("/api/hosts")
	h.Register(group)
}
