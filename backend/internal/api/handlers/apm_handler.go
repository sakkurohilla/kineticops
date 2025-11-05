package handlers

import (
	"github.com/gofiber/fiber/v2"
)

func GetAPMApplications(c *fiber.Ctx) error {
	return c.JSON([]map[string]interface{}{})
}

func CreateAPMApplication(c *fiber.Ctx) error {
	var app map[string]interface{}
	if err := c.BodyParser(&app); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	app["id"] = 1
	return c.JSON(app)
}

func GetAPMStats(c *fiber.Ctx) error {
	return c.JSON(map[string]interface{}{
		"avg_response_time": 0,
		"throughput":        0,
		"error_rate":        0,
		"apdex":             0,
	})
}

func GetAPMErrors(c *fiber.Ctx) error {
	return c.JSON([]map[string]interface{}{})
}

func GetAPMTraces(c *fiber.Ctx) error {
	return c.JSON([]map[string]interface{}{})
}

func GetAPMMetrics(c *fiber.Ctx) error {
	return c.JSON([]map[string]interface{}{})
}
