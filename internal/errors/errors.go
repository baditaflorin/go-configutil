package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrConfigurationInvalid is returned when configuration validation fails
	ErrConfigurationInvalid = errors.New("configuration is invalid")

	// ErrMissingRequired is returned when a required field is missing
	ErrMissingRequired = errors.New("required configuration field is missing")

	// ErrInvalidFormat is returned when a field has an invalid format
	ErrInvalidFormat = errors.New("configuration field has invalid format")

	// ErrFileNotFound is returned when an env file cannot be found
	ErrFileNotFound = errors.New("configuration file not found")

	// ErrFilePermissions is returned when file permissions are insecure
	ErrFilePermissions = errors.New("configuration file has insecure permissions")
)

// ConfigError represents a configuration error with sanitized details
type ConfigError struct {
	Field   string
	Type    error
	details string
}

// Error returns a safe, non-leaking error message
func (e *ConfigError) Error() string {
	// Never expose internal details to external callers
	return fmt.Sprintf("configuration error: %s field is invalid", e.Field)
}

// Details returns detailed error information for logging (not for user display)
func (e *ConfigError) Details() string {
	if e.details != "" {
		return e.details
	}
	return e.Error()
}

// Unwrap allows error wrapping
func (e *ConfigError) Unwrap() error {
	return e.Type
}

// NewConfigError creates a new configuration error
func NewConfigError(field string, errType error, details string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Type:    errType,
		details: details,
	}
}

// NewMissingFieldError creates an error for missing required fields
func NewMissingFieldError(field string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Type:    ErrMissingRequired,
		details: fmt.Sprintf("required field '%s' is not set", field),
	}
}

// NewInvalidFormatError creates an error for invalid field formats
func NewInvalidFormatError(field string, reason string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Type:    ErrInvalidFormat,
		details: fmt.Sprintf("field '%s' has invalid format: %s", field, reason),
	}
}
