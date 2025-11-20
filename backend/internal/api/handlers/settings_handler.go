package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// GetSettings retrieves current user's settings
func GetSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	settings, err := services.GetUserSettings(userID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch settings"})
	}

	return c.JSON(settings)
}

// UpdateSettings updates current user's settings
func UpdateSettings(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int64)

	var updates map[string]interface{}
	if err := c.BodyParser(&updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Bad request"})
	}

	if err := services.UpdateUserSettings(userID, updates); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"msg": "Settings updated successfully"})
}
