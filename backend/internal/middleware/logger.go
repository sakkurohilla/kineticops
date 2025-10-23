package middleware

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var LoggerInstance *zap.Logger

func Logger() fiber.Handler {
	logger, _ := zap.NewProduction()
	LoggerInstance = logger

	return func(c *fiber.Ctx) error {
		LoggerInstance.Info("request received",
			zap.String("method", c.Method()),
			zap.String("url", c.OriginalURL()),
			zap.String("ip", c.IP()),
		)
		return c.Next()
	}
}
