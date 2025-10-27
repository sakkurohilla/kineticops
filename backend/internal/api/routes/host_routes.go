package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/api/handlers"
	"github.com/sakkurohilla/kineticops/backend/internal/middleware"
)

func RegisterHostRoutes(app *fiber.App) {
	api := app.Group("/api/v1/hosts", middleware.AuthMiddleware)
	api.Post("", handlers.CreateHost)
	api.Get("", handlers.ListHosts)
	api.Get("/:id", handlers.GetHost)
	api.Put("/:id", handlers.UpdateHost)
	api.Delete("/:id", handlers.DeleteHost)
	api.Post("/:id/heartbeat", handlers.HostHeartbeat)
}
