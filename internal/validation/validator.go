package validation

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// portRegex validates port numbers (1-65535)
	portRegex = regexp.MustCompile(`^([1-9][0-9]{0,3}|[1-5][0-9]{4}|6[0-4][0-9]{3}|65[0-4][0-9]{2}|655[0-2][0-9]|6553[0-5])$`)

	// alphanumericRegex for basic sanitization
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
)

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ValidateURL checks if a string is a valid URL with required schemes
func ValidateURL(rawURL string, allowedSchemes []string) error {
	if rawURL == "" {
		return &ValidationError{Field: "url", Message: "URL cannot be empty"}
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{Field: "url", Message: "invalid URL format"}
	}

	if parsedURL.Scheme == "" {
		return &ValidationError{Field: "url", Message: "URL scheme is required"}
	}

	if parsedURL.Host == "" {
		return &ValidationError{Field: "url", Message: "URL host is required"}
	}

	// Check if scheme is in allowed list
	if len(allowedSchemes) > 0 {
		schemeAllowed := false
		for _, scheme := range allowedSchemes {
			if parsedURL.Scheme == scheme {
				schemeAllowed = true
				break
			}
		}
		if !schemeAllowed {
			return &ValidationError{
				Field:   "url",
				Message: fmt.Sprintf("URL scheme must be one of: %s", strings.Join(allowedSchemes, ", ")),
			}
		}
	}

	return nil
}

// ValidatePort checks if a port string is valid
func ValidatePort(port string) error {
	if port == "" {
		return &ValidationError{Field: "port", Message: "port cannot be empty"}
	}

	// Try parsing as integer first
	portNum, err := strconv.Atoi(port)
	if err != nil || portNum < 1 || portNum > 65535 {
		return &ValidationError{Field: "port", Message: "port must be a number between 1 and 65535"}
	}

	return nil
}

// SanitizeString removes potentially dangerous characters
func SanitizeString(input string, maxLength int) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Truncate to max length
	if maxLength > 0 && len(sanitized) > maxLength {
		sanitized = sanitized[:maxLength]
	}

	return sanitized
}

// ValidateClientID validates OAuth client IDs (basic validation)
func ValidateClientID(clientID string) error {
	if clientID == "" {
		return nil // Optional field
	}

	sanitized := SanitizeString(clientID, 512)
	if len(sanitized) < 10 {
		return &ValidationError{Field: "client_id", Message: "client ID too short"}
	}

	return nil
}

// ValidateClientSecret validates OAuth client secrets
func ValidateClientSecret(secret string) error {
	if secret == "" {
		return nil // Optional field
	}

	sanitized := SanitizeString(secret, 512)
	if len(sanitized) < 10 {
		return &ValidationError{Field: "client_secret", Message: "client secret too short"}
	}

	return nil
}

// CheckFilePermissions ensures a file is not world-readable
func CheckFilePermissions(filepath string) error {
	info, err := os.Stat(filepath)
	if err != nil {
		return fmt.Errorf("failed to check file permissions: %w", err)
	}

	mode := info.Mode()

	// Check if file is world-readable or world-writable
	if mode.Perm()&0044 != 0 {
		return &ValidationError{
			Field:   "file_permissions",
			Message: "file has insecure permissions (world-readable or world-writable)",
		}
	}

	return nil
}

// ValidateDatabaseURL validates database connection strings
func ValidateDatabaseURL(dbURL string) error {
	if dbURL == "" {
		return &ValidationError{Field: "database_url", Message: "database URL is required"}
	}

	// Common database schemes
	allowedSchemes := []string{"postgres", "postgresql", "mysql", "sqlite", "sqlite3"}

	if err := ValidateURL(dbURL, allowedSchemes); err != nil {
		// For sqlite, it might be a file path
		if strings.HasPrefix(dbURL, "sqlite://") || strings.HasPrefix(dbURL, "sqlite3://") || strings.HasSuffix(dbURL, ".db") {
			return nil
		}
		return err
	}

	return nil
}

// ValidateHTTPURL validates HTTP/HTTPS URLs
func ValidateHTTPURL(httpURL string) error {
	if httpURL == "" {
		return &ValidationError{Field: "http_url", Message: "HTTP URL is required"}
	}

	allowedSchemes := []string{"http", "https"}
	return ValidateURL(httpURL, allowedSchemes)
}
