package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ErrorHandler(c *fiber.Ctx, err error) error {
	LoggerInstance.Error("internal error",
		zap.Error(err),
		zap.String("method", c.Method()),
		zap.String("path", c.Path()),
		zap.String("ip", c.IP()),
	)

	// Don't expose internal errors in production
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"error": "Internal server error",
	})
}
