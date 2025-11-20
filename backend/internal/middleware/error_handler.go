package middleware

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
	"github.com/sakkurohilla/kineticops/backend/internal/services"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     string      `json:"error"`
	ErrorCode string      `json:"error_code,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// PanicRecovery recovers from panics and returns proper error
func PanicRecovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with stack trace
				stack := string(debug.Stack())
				logging.Errorf("PANIC RECOVERED: %v\nStack: %s", r, stack)

				// Audit log the panic
				services.AuditSvc.LogFromFiber(c, 0, "system", "panic", "system", "", "error", map[string]interface{}{
					"panic":      fmt.Sprintf("%v", r),
					"stack":      stack,
					"path":       c.Path(),
					"method":     c.Method(),
					"user_agent": c.Get("User-Agent"),
				}, fmt.Errorf("%v", r))

				// Send user-friendly error response
				c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
					Success:   false,
					Error:     "An unexpected error occurred. Our team has been notified.",
					ErrorCode: "INTERNAL_SERVER_ERROR",
					Timestamp: time.Now(),
					RequestID: c.Locals("requestid").(string),
				})
			}
		}()

		return c.Next()
	}
}

// RequestID adds a unique request ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Locals("requestid", requestID)
		c.Set("X-Request-ID", requestID)
		return c.Next()
	}
}

// ErrorLogger logs all errors with context
func ErrorLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()

		if err != nil {
			logging.Errorf("Request error: method=%s path=%s status=%d error=%v requestID=%s",
				c.Method(), c.Path(), c.Response().StatusCode(), err, c.Locals("requestid"))
		}

		return err
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// ValidationError creates a structured validation error response
func ValidationError(c *fiber.Ctx, field string, message string) error {
	return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse{
		Success:   false,
		Error:     "Validation failed",
		ErrorCode: "VALIDATION_ERROR",
		Details: map[string]string{
			"field":   field,
			"message": message,
		},
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// UnauthorizedError creates an unauthorized error response
func UnauthorizedError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorCode: "UNAUTHORIZED",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// ForbiddenError creates a forbidden error response
func ForbiddenError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusForbidden).JSON(ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorCode: "FORBIDDEN",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// NotFoundError creates a not found error response
func NotFoundError(c *fiber.Ctx, resource string) error {
	return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
		Success:   false,
		Error:     fmt.Sprintf("%s not found", resource),
		ErrorCode: "NOT_FOUND",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// InternalError creates an internal server error response
func InternalError(c *fiber.Ctx, err error) error {
	logging.Errorf("Internal error: %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{
		Success:   false,
		Error:     "An internal error occurred",
		ErrorCode: "INTERNAL_ERROR",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// ConflictError creates a conflict error response
func ConflictError(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
		Success:   false,
		Error:     message,
		ErrorCode: "CONFLICT",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

// RateLimitError creates a rate limit error response
func RateLimitError(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTooManyRequests).JSON(ErrorResponse{
		Success:   false,
		Error:     "Rate limit exceeded. Please try again later.",
		ErrorCode: "RATE_LIMIT_EXCEEDED",
		Timestamp: time.Now(),
		RequestID: getRequestID(c),
	})
}

func getRequestID(c *fiber.Ctx) string {
	if id := c.Locals("requestid"); id != nil {
		return id.(string)
	}
	return ""
}
