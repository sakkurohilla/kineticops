package routes

import (
	"github.com/gofiber/fiber/v2"
)

func RegisterHealthRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "OK",
			"details": "Fiber API healthy",
		})
	})
}
