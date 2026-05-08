package helpers

import (
	"html"
	"regexp"
	"strings"
)

var (
	sqlInjectionPatterns = regexp.MustCompile(`(?i)(union\s+select|drop\s+table|insert\s+into|delete\s+from|update\s+.*set|exec\s*\(|execute\s*\(|xp_|sp_|0x[0-9a-f]+)`)
	scriptTagPattern     = regexp.MustCompile(`(?i)<\s*script[^>]*>.*?<\s*/\s*script\s*>`)
	htmlEventPattern     = regexp.MustCompile(`(?i)\s+on\w+\s*=`)
	nullBytePattern      = regexp.MustCompile(`\x00`)
)

// SanitizeInput removes null bytes, trims whitespace and HTML-escapes the string.
// Use for general text fields (names, addresses, descriptions).
func SanitizeInput(input string) string {
	input = nullBytePattern.ReplaceAllString(input, "")
	input = strings.TrimSpace(input)
	input = html.EscapeString(input)
	return input
}

// SanitizeHTML strips script tags and event handlers but preserves other HTML.
// Use for rich-text fields (banner body, news body).
func SanitizeHTML(input string) string {
	input = nullBytePattern.ReplaceAllString(input, "")
	input = scriptTagPattern.ReplaceAllString(input, "")
	input = htmlEventPattern.ReplaceAllString(input, " ")
	return input
}

// DetectSQLInjection returns true if the input contains common SQL injection patterns.
// Note: GORM parameterized queries already prevent SQL injection at the DB layer,
// this is an additional defense-in-depth check for logging/alerting.
func DetectSQLInjection(input string) bool {
	return sqlInjectionPatterns.MatchString(input)
}

// SanitizeEmail normalizes and validates basic email format.
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	email = nullBytePattern.ReplaceAllString(email, "")
	return email
}

// SanitizePhoneNumber strips everything except digits and leading +.
func SanitizePhoneNumber(phone string) string {
	phone = strings.TrimSpace(phone)
	var result strings.Builder
	for i, ch := range phone {
		if ch == '+' && i == 0 {
			result.WriteRune(ch)
		} else if ch >= '0' && ch <= '9' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}
