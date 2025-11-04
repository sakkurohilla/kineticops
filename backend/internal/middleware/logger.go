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
		// Only log errors and important requests in production
		if c.Method() != "GET" || c.Response().StatusCode() >= 400 {
			LoggerInstance.Info("request",
				zap.String("method", c.Method()),
				zap.String("path", c.Path()),
				zap.String("ip", c.IP()),
				zap.Int("status", c.Response().StatusCode()),
			)
		}
		return c.Next()
	}
}
