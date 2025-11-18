package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ConfigError
		wantText string
	}{
		{
			name: "missing field error",
			err: &ConfigError{
				Field:   "database_url",
				Type:    ErrMissingRequired,
				details: "this should not be exposed",
			},
			wantText: "configuration error: database_url field is invalid",
		},
		{
			name: "invalid format error",
			err: &ConfigError{
				Field:   "port",
				Type:    ErrInvalidFormat,
				details: "internal details here",
			},
			wantText: "configuration error: port field is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantText {
				t.Errorf("ConfigError.Error() = %q, want %q", got, tt.wantText)
			}

			// Verify that internal details are NOT exposed in Error()
			if strings.Contains(got, tt.err.details) {
				t.Errorf("ConfigError.Error() leaked internal details: %q", got)
			}
		})
	}
}

func TestConfigError_Details(t *testing.T) {
	err := &ConfigError{
		Field:   "test_field",
		Type:    ErrInvalidFormat,
		details: "detailed error information",
	}

	details := err.Details()
	if !strings.Contains(details, "detailed error information") {
		t.Errorf("ConfigError.Details() should contain detailed information, got: %q", details)
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	baseErr := ErrMissingRequired
	configErr := &ConfigError{
		Field: "test",
		Type:  baseErr,
	}

	if !errors.Is(configErr, ErrMissingRequired) {
		t.Error("ConfigError.Unwrap() should support errors.Is()")
	}
}

func TestNewConfigError(t *testing.T) {
	err := NewConfigError("test_field", ErrInvalidFormat, "test details")

	if err.Field != "test_field" {
		t.Errorf("Field = %q, want %q", err.Field, "test_field")
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Error("NewConfigError should wrap the error type")
	}

	if err.details != "test details" {
		t.Errorf("details = %q, want %q", err.details, "test details")
	}
}

func TestNewMissingFieldError(t *testing.T) {
	err := NewMissingFieldError("DATABASE_URL")

	if err.Field != "DATABASE_URL" {
		t.Errorf("Field = %q, want %q", err.Field, "DATABASE_URL")
	}

	if !errors.Is(err, ErrMissingRequired) {
		t.Error("NewMissingFieldError should wrap ErrMissingRequired")
	}

	// Verify error message is safe (no internal details leaked)
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "DATABASE_URL") {
		t.Errorf("Error message should mention the field name: %q", errorMsg)
	}

	// But details should contain more information
	details := err.Details()
	if !strings.Contains(details, "not set") {
		t.Errorf("Details should contain additional information: %q", details)
	}
}

func TestNewInvalidFormatError(t *testing.T) {
	err := NewInvalidFormatError("port", "must be numeric")

	if err.Field != "port" {
		t.Errorf("Field = %q, want %q", err.Field, "port")
	}

	if !errors.Is(err, ErrInvalidFormat) {
		t.Error("NewInvalidFormatError should wrap ErrInvalidFormat")
	}

	details := err.Details()
	if !strings.Contains(details, "must be numeric") {
		t.Errorf("Details should contain reason: %q", details)
	}
}

func TestErrorDoesNotLeakSensitiveInfo(t *testing.T) {
	sensitiveDetails := "database password: supersecret123"
	err := NewConfigError("database_url", ErrInvalidFormat, sensitiveDetails)

	// The public Error() method should NOT contain sensitive details
	publicMsg := err.Error()
	if strings.Contains(publicMsg, "supersecret123") {
		t.Errorf("Error() leaked sensitive information: %q", publicMsg)
	}

	if strings.Contains(publicMsg, "password") {
		t.Errorf("Error() leaked sensitive field name: %q", publicMsg)
	}

	// But Details() (for logging) should contain them
	details := err.Details()
	if !strings.Contains(details, sensitiveDetails) {
		t.Errorf("Details() should contain full error information for logging: %q", details)
	}
}

func TestSentinelErrors(t *testing.T) {
	// Verify all sentinel errors are defined
	sentinels := []error{
		ErrConfigurationInvalid,
		ErrMissingRequired,
		ErrInvalidFormat,
		ErrFileNotFound,
		ErrFilePermissions,
	}

	for i, err := range sentinels {
		if err == nil {
			t.Errorf("Sentinel error %d is nil", i)
		}
	}

	// Verify they're all different
	for i, err1 := range sentinels {
		for j, err2 := range sentinels {
			if i != j && errors.Is(err1, err2) {
				t.Errorf("Sentinel errors %d and %d should be different", i, j)
			}
		}
	}
}
