package utils

import (
	"html"
	"html/template"
	"regexp"
	"strings"
)

var (
	scriptTagPattern = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	htmlTagPattern   = regexp.MustCompile(`<[^>]+>`)
)

// SanitizeHTML escapes HTML special characters to prevent XSS
func SanitizeHTML(input string) string {
	return html.EscapeString(input)
}

// SanitizeHTMLTemplate uses html/template for more robust escaping
func SanitizeHTMLTemplate(input string) template.HTML {
	return template.HTML(html.EscapeString(input))
}

// SanitizeURL removes javascript: and data: URLs
func SanitizeURL(url string) string {
	url = strings.TrimSpace(url)
	lower := strings.ToLower(url)

	dangerousProtocols := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
	}

	for _, protocol := range dangerousProtocols {
		if strings.HasPrefix(lower, protocol) {
			return ""
		}
	}

	return url
}

// SanitizeUserInput removes dangerous characters from user input
func SanitizeUserInput(input string) string {
	input = strings.ReplaceAll(input, "\x00", "")

	var result strings.Builder
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\t' {
			result.WriteRune(r)
		}
	}

	return strings.TrimSpace(result.String())
}

// SanitizeForJSON escapes characters that could break JSON strings
func SanitizeForJSON(input string) string {
	input = strings.ReplaceAll(input, "\\", "\\\\")
	input = strings.ReplaceAll(input, "\"", "\\\"")
	input = strings.ReplaceAll(input, "\n", "\\n")
	input = strings.ReplaceAll(input, "\r", "\\r")
	input = strings.ReplaceAll(input, "\t", "\\t")
	return input
}

// StripTags removes all HTML tags from input
func StripTags(input string) string {
	return htmlTagPattern.ReplaceAllString(input, "")
}

// RemoveScriptTags removes script tags from input
func RemoveScriptTags(input string) string {
	return scriptTagPattern.ReplaceAllString(input, "")
}
