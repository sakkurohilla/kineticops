package logging

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.In(time.Local).Format(time.RFC3339))
	})
	// Build logger; ignore build error for now (fallbacks are possible)
	l, _ := cfg.Build()
	Logger = l
}

// Sugared helpers
func Infof(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Sugar().Infof(format, args...)
	}
}

func Warnf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Sugar().Warnf(format, args...)
	}
}

func Errorf(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Sugar().Errorf(format, args...)
	}
}
