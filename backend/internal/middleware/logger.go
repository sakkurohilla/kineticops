package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var LoggerInstance *zap.Logger

func Logger() fiber.Handler {
	// Build a production config but override the time encoder so timestamps
	// are emitted in the local timezone (time.Local). This makes log
	// timestamps consistent with the standard library logger where we've
	// set time.Local = Asia/Kolkata in main.go.
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(time.Local).Format(time.RFC3339))
	})

	// Build the logger. Default output goes to stderr which Docker also captures.
	logger, _ := cfg.Build()
	LoggerInstance = logger

	return func(c *fiber.Ctx) error {
		// Only log non-GET requests or requests with status >= 400 to limit noise.
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
