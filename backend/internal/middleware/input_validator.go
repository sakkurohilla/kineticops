package middleware

import (
	"html"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/sakkurohilla/kineticops/backend/internal/logging"
)

// Validation patterns
var (
	// SQL injection patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select)`),
		regexp.MustCompile(`(?i)(insert\s+into)`),
		regexp.MustCompile(`(?i)(delete\s+from)`),
		regexp.MustCompile(`(?i)(drop\s+table)`),
		regexp.MustCompile(`(?i)(update\s+.+\s+set)`),
		regexp.MustCompile(`(?i)(exec(\s|\()+)`),
		regexp.MustCompile(`(?i)(execute(\s|\()+)`),
		regexp.MustCompile(`(?i)(select\s+.+\s+from)`),
		regexp.MustCompile(`(?i)(-{2}|;|\/\*|\*\/)`), // SQL comments
		regexp.MustCompile(`(?i)(xp_cmdshell)`),
		regexp.MustCompile(`(?i)(sp_executesql)`),
	}

	// XSS patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // Event handlers like onclick, onerror
		regexp.MustCompile(`(?i)<iframe`),
		regexp.MustCompile(`(?i)<object`),
		regexp.MustCompile(`(?i)<embed`),
		regexp.MustCompile(`(?i)<img[^>]+src[^>]*>`),
		regexp.MustCompile(`(?i)eval\s*\(`),
		regexp.MustCompile(`(?i)expression\s*\(`),
	}

	// Safe patterns
	emailPattern    = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	ipv4Pattern     = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$`)
	hostnamePattern = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?)*$`)
	alphanumPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-]+$`)
	uuidPattern     = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
)

// SanitizeInput removes dangerous characters and patterns
func SanitizeInput(input string) string {
	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// HTML escape to prevent XSS
	input = html.EscapeString(input)

	// Limit length to prevent DoS
	if len(input) > 10000 {
		input = input[:10000]
	}

	return input
}

// CheckSQLInjection detects potential SQL injection attempts
func CheckSQLInjection(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// CheckXSS detects potential XSS attempts
func CheckXSS(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// ValidateEmail checks if email is valid
func ValidateEmail(email string) bool {
	if len(email) > 254 {
		return false
	}
	return emailPattern.MatchString(email)
}

// ValidateIP checks if IP address is valid
func ValidateIP(ip string) bool {
	return ipv4Pattern.MatchString(ip)
}

// ValidateHostname checks if hostname is valid
func ValidateHostname(hostname string) bool {
	if len(hostname) > 253 {
		return false
	}
	return hostnamePattern.MatchString(hostname)
}

// ValidateAlphanumeric checks if string contains only alphanumeric characters
func ValidateAlphanumeric(input string) bool {
	return alphanumPattern.MatchString(input)
}

// ValidateUUID checks if string is a valid UUID
func ValidateUUID(input string) bool {
	return uuidPattern.MatchString(input)
}

// ValidateStringField validates a string field with custom rules
func ValidateStringField(value string, fieldName string, minLen, maxLen int, pattern *regexp.Regexp) error {
	value = strings.TrimSpace(value)

	if len(value) < minLen {
		return fiber.NewError(fiber.StatusBadRequest,
			fieldName+" must be at least "+string(rune(minLen))+" characters")
	}

	if len(value) > maxLen {
		return fiber.NewError(fiber.StatusBadRequest,
			fieldName+" must be at most "+string(rune(maxLen))+" characters")
	}

	if pattern != nil && !pattern.MatchString(value) {
		return fiber.NewError(fiber.StatusBadRequest,
			fieldName+" contains invalid characters")
	}

	if CheckSQLInjection(value) {
		return fiber.NewError(fiber.StatusBadRequest,
			fieldName+" contains potentially dangerous content")
	}

	if CheckXSS(value) {
		return fiber.NewError(fiber.StatusBadRequest,
			fieldName+" contains potentially dangerous content")
	}

	return nil
}

// ValidateAndSanitizeInput is a middleware that validates and sanitizes all inputs
func ValidateAndSanitizeInput() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip validation for health checks
		if strings.HasPrefix(c.Path(), "/health") {
			return c.Next()
		}

		// Check query parameters
		c.Request().URI().QueryArgs().VisitAll(func(key, value []byte) {
			keyStr := string(key)
			valueStr := string(value)

			// Check for SQL injection
			if CheckSQLInjection(valueStr) {
				logging.Warnf("SQL injection attempt detected in query param %s: %s from IP %s", keyStr, valueStr, c.IP())
			}

			// Check for XSS
			if CheckXSS(valueStr) {
				logging.Warnf("XSS attempt detected in query param %s: %s from IP %s", keyStr, valueStr, c.IP())
			}
		})

		// Check form data
		if c.Method() == "POST" || c.Method() == "PUT" || c.Method() == "PATCH" {
			contentType := string(c.Request().Header.ContentType())

			// For form data
			if strings.Contains(contentType, "application/x-www-form-urlencoded") ||
				strings.Contains(contentType, "multipart/form-data") {
				c.Request().PostArgs().VisitAll(func(key, value []byte) {
					keyStr := string(key)
					valueStr := string(value)

					if CheckSQLInjection(valueStr) {
						logging.Warnf("SQL injection attempt in form field %s from IP %s", keyStr, c.IP())
					}

					if CheckXSS(valueStr) {
						logging.Warnf("XSS attempt in form field %s from IP %s", keyStr, c.IP())
					}
				})
			}

			// For JSON data - body is already parsed by Fiber
			// Validation will be done at handler level using struct tags
		}

		return c.Next()
	}
}

// SanitizeOutput sanitizes output data to prevent XSS
func SanitizeOutput(data interface{}) interface{} {
	switch v := data.(type) {
	case string:
		return html.EscapeString(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[key] = SanitizeOutput(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = SanitizeOutput(val)
		}
		return result
	default:
		return data
	}
}
