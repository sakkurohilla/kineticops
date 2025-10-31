package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/telemetry"
)

// Telemetry returns a snapshot of internal in-memory telemetry counters.
func Telemetry(c *fiber.Ctx) error {
	// Protected route (AuthRequired) will ensure only authorized users can view.
	data := telemetry.GetCounters()
	return c.JSON(data)
}
