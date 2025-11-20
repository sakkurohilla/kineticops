package logging

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger        *zap.Logger
	once          sync.Once
	correlationMu sync.RWMutex
	correlations  = make(map[string]string)
)

// InitStructuredLogger initializes the structured logger with correlation ID support
func InitStructuredLogger() {
	once.Do(func() {
		config := zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.CallerKey = "caller"
		config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

		// Add custom fields
		config.EncoderConfig.MessageKey = "message"
		config.EncoderConfig.LevelKey = "level"
		config.EncoderConfig.StacktraceKey = "stacktrace"

		// Set log level
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

		// Build logger
		var err error
		logger, err = config.Build(
			zap.AddCaller(),
			zap.AddCallerSkip(1),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize logger: %v", err))
		}
	})
}

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	correlationMu.Lock()
	defer correlationMu.Unlock()

	key := fmt.Sprintf("ctx_%p", ctx)
	correlations[key] = correlationID

	return ctx
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	correlationMu.RLock()
	defer correlationMu.RUnlock()

	key := fmt.Sprintf("ctx_%p", ctx)
	if id, ok := correlations[key]; ok {
		return id
	}
	return ""
}

// InfoWithContext logs info with correlation ID
func InfoWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitStructuredLogger()
	}

	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	logger.Info(msg, fields...)
}

// ErrorWithContext logs error with correlation ID
func ErrorWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitStructuredLogger()
	}

	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	// Add caller information
	_, file, line, ok := runtime.Caller(1)
	if ok {
		fields = append(fields, zap.String("source", fmt.Sprintf("%s:%d", file, line)))
	}

	logger.Error(msg, fields...)
}

// WarnWithContext logs warning with correlation ID
func WarnWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitStructuredLogger()
	}

	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	logger.Warn(msg, fields...)
}

// DebugWithContext logs debug with correlation ID
func DebugWithContext(ctx context.Context, msg string, fields ...zap.Field) {
	if logger == nil {
		InitStructuredLogger()
	}

	correlationID := GetCorrelationID(ctx)
	if correlationID != "" {
		fields = append(fields, zap.String("correlation_id", correlationID))
	}

	logger.Debug(msg, fields...)
}

// LogRequest logs HTTP request details
func LogRequest(ctx context.Context, method, path, ip, userAgent string, duration time.Duration, statusCode int) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("ip", ip),
		zap.String("user_agent", userAgent),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	}

	InfoWithContext(ctx, "HTTP Request", fields...)
}

// LogError logs error with stack trace
func LogError(ctx context.Context, err error, msg string, fields ...zap.Field) {
	if err != nil {
		fields = append(fields, zap.Error(err))

		// Add stack trace
		stack := make([]byte, 4096)
		length := runtime.Stack(stack, false)
		fields = append(fields, zap.String("stack", string(stack[:length])))
	}

	ErrorWithContext(ctx, msg, fields...)
}

// LogMetric logs a metric event
func LogMetric(ctx context.Context, metricName string, value float64, labels map[string]string) {
	fields := []zap.Field{
		zap.String("metric_name", metricName),
		zap.Float64("value", value),
	}

	for k, v := range labels {
		fields = append(fields, zap.String(k, v))
	}

	InfoWithContext(ctx, "Metric", fields...)
}

// Flush flushes any buffered log entries
func Flush() {
	if logger != nil {
		logger.Sync()
	}
}
