package validation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		allowedSchemes []string
		wantErr        bool
	}{
		{
			name:           "valid HTTP URL",
			url:            "http://example.com",
			allowedSchemes: []string{"http", "https"},
			wantErr:        false,
		},
		{
			name:           "valid HTTPS URL",
			url:            "https://example.com:8080/path",
			allowedSchemes: []string{"http", "https"},
			wantErr:        false,
		},
		{
			name:           "valid postgres URL",
			url:            "postgres://user:pass@localhost:5432/db",
			allowedSchemes: []string{"postgres", "postgresql"},
			wantErr:        false,
		},
		{
			name:           "empty URL",
			url:            "",
			allowedSchemes: []string{"http"},
			wantErr:        true,
		},
		{
			name:           "invalid scheme",
			url:            "ftp://example.com",
			allowedSchemes: []string{"http", "https"},
			wantErr:        true,
		},
		{
			name:           "no scheme",
			url:            "example.com",
			allowedSchemes: []string{"http"},
			wantErr:        true,
		},
		{
			name:           "no host",
			url:            "http://",
			allowedSchemes: []string{"http"},
			wantErr:        true,
		},
		{
			name:           "malformed URL",
			url:            "ht!tp://exa mple.com",
			allowedSchemes: []string{"http"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.url, tt.allowedSchemes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"valid port 80", "80", false},
		{"valid port 8080", "8080", false},
		{"valid port 443", "443", false},
		{"valid port 65535", "65535", false},
		{"valid port 1", "1", false},
		{"empty port", "", true},
		{"port too high", "65536", true},
		{"port zero", "0", true},
		{"negative port", "-1", true},
		{"non-numeric port", "abc", true},
		{"port with spaces", " 8080 ", true}, // Should fail because of spaces
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
	}{
		{
			name:      "no changes needed",
			input:     "hello",
			maxLength: 10,
			want:      "hello",
		},
		{
			name:      "trim whitespace",
			input:     "  hello  ",
			maxLength: 10,
			want:      "hello",
		},
		{
			name:      "truncate long string",
			input:     "this is a very long string",
			maxLength: 10,
			want:      "this is a ",
		},
		{
			name:      "no max length",
			input:     "  test  ",
			maxLength: 0,
			want:      "test",
		},
		{
			name:      "empty string",
			input:     "",
			maxLength: 10,
			want:      "",
		},
		{
			name:      "only whitespace",
			input:     "   ",
			maxLength: 10,
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input, tt.maxLength)
			if got != tt.want {
				t.Errorf("SanitizeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidateClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		wantErr  bool
	}{
		{"valid client ID", "1234567890-abcdefghijklmnop.apps.googleusercontent.com", false},
		{"empty client ID", "", false}, // Optional field
		{"too short", "123456789", true},
		{"minimum valid length", "1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientID(tt.clientID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClientID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateClientSecret(t *testing.T) {
	tests := []struct {
		name   string
		secret string
		wantErr bool
	}{
		{"valid secret", "GOCSPX-1234567890abcdefghijk", false},
		{"empty secret", "", false}, // Optional field
		{"too short", "123456789", true},
		{"minimum valid length", "1234567890", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateClientSecret(tt.secret)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateClientSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCheckFilePermissions(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		permissions os.FileMode
		wantErr     bool
	}{
		{"secure permissions 0600", 0600, false},
		{"secure permissions 0400", 0400, false},
		{"insecure permissions 0644", 0644, true},
		{"insecure permissions 0666", 0666, true},
		{"insecure permissions 0777", 0777, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test file with specific permissions
			testFile := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(testFile, []byte("test"), tt.permissions); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			err := CheckFilePermissions(testFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckFilePermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDatabaseURL(t *testing.T) {
	tests := []struct {
		name    string
		dbURL   string
		wantErr bool
	}{
		{"valid postgres URL", "postgres://user:pass@localhost:5432/dbname", false},
		{"valid postgresql URL", "postgresql://user:pass@localhost:5432/dbname", false},
		{"valid mysql URL", "mysql://user:pass@localhost:3306/dbname", false},
		{"valid sqlite URL", "sqlite://./data.db", false},
		{"valid sqlite3 URL", "sqlite3://./data.db", false},
		{"sqlite file path", "test.db", false},
		{"empty database URL", "", true},
		{"invalid scheme", "http://localhost/db", true},
		{"no scheme", "localhost:5432/db", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDatabaseURL(tt.dbURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDatabaseURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid HTTP URL", "http://localhost:8080", false},
		{"valid HTTPS URL", "https://api.example.com/v1", false},
		{"empty URL", "", true},
		{"invalid scheme", "ftp://example.com", true},
		{"no scheme", "example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHTTPURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHTTPURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Security tests for injection attempts
func TestValidationAgainstInjection(t *testing.T) {
	injectionTests := []struct {
		name     string
		input    string
		testFunc func(string) error
		wantErr  bool
	}{
		{
			name:     "SQL injection in URL",
			input:    "postgres://user'; DROP TABLE users;--@localhost/db",
			testFunc: ValidateDatabaseURL,
			wantErr:  true, // Invalid URL format due to special characters
		},
		{
			name:     "XSS in URL",
			input:    "http://example.com/<script>alert('xss')</script>",
			testFunc: ValidateHTTPURL,
			wantErr:  false, // URL parsing handles this, but it's escaped in the path
		},
		{
			name:     "Command injection in port",
			input:    "8080; rm -rf /",
			testFunc: ValidatePort,
			wantErr:  true,
		},
		{
			name:     "Path traversal in database URL",
			input:    "sqlite://../../etc/passwd",
			testFunc: ValidateDatabaseURL,
			wantErr:  false, // Path traversal is file system concern, not URL validation
		},
	}

	for _, tt := range injectionTests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Injection test %s: error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}
