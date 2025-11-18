package configutil

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	configerrors "github.com/baditaflorin/go-configutil/internal/errors"
)

func TestNewConfig_Success(t *testing.T) {
	// Create temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `DATABASE_URL=postgres://user:pass@localhost:5432/testdb
AUTH_SERVICE_URL=https://auth.example.com
PORT=8080
DEBUG=true
GOOGLE_CLIENT_ID=test-client-id-1234567890
GOOGLE_CLIENT_SECRET=test-secret-1234567890
`
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config, err := NewConfig(WithEnvFile(envFile))
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify all fields are loaded correctly
	if config.DatabaseURL != "postgres://user:pass@localhost:5432/testdb" {
		t.Errorf("DatabaseURL = %q, want %q", config.DatabaseURL, "postgres://user:pass@localhost:5432/testdb")
	}

	if config.AuthServiceURL != "https://auth.example.com" {
		t.Errorf("AuthServiceURL = %q, want %q", config.AuthServiceURL, "https://auth.example.com")
	}

	if config.Port != "8080" {
		t.Errorf("Port = %q, want %q", config.Port, "8080")
	}

	if !config.Debug {
		t.Error("Debug should be true")
	}

	if config.GoogleClientID != "test-client-id-1234567890" {
		t.Errorf("GoogleClientID = %q, want %q", config.GoogleClientID, "test-client-id-1234567890")
	}

	if config.GoogleClientSecret != "test-secret-1234567890" {
		t.Errorf("GoogleClientSecret = %q, want %q", config.GoogleClientSecret, "test-secret-1234567890")
	}
}

func TestNewConfig_MissingRequired(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	tests := []struct {
		name        string
		envContent  string
		expectedErr error
	}{
		{
			name:        "missing DATABASE_URL",
			envContent:  "AUTH_SERVICE_URL=https://auth.example.com\n",
			expectedErr: configerrors.ErrMissingRequired,
		},
		{
			name:        "missing AUTH_SERVICE_URL",
			envContent:  "DATABASE_URL=postgres://localhost/db\n",
			expectedErr: configerrors.ErrMissingRequired,
		},
		{
			name:        "both missing",
			envContent:  "PORT=8080\n",
			expectedErr: configerrors.ErrMissingRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(envFile, []byte(tt.envContent), 0600); err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			_, err := NewConfig(WithEnvFile(envFile))
			if err == nil {
				t.Error("NewConfig() should have failed with missing required fields")
			}

			var configErr *configerrors.ConfigError
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error type %v, got %v", tt.expectedErr, err)
			}

			// Verify error is a ConfigError
			if !errors.As(err, &configErr) {
				t.Errorf("Error should be ConfigError type, got %T", err)
			}
		})
	}
}

func TestNewConfig_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	tests := []struct {
		name       string
		envContent string
		wantErr    bool
	}{
		{
			name: "invalid port",
			envContent: `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://auth.example.com
PORT=99999
`,
			wantErr: true,
		},
		{
			name: "invalid database URL",
			envContent: `DATABASE_URL=not-a-valid-url
AUTH_SERVICE_URL=https://auth.example.com
`,
			wantErr: true,
		},
		{
			name: "invalid auth service URL",
			envContent: `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=ftp://invalid-scheme.com
`,
			wantErr: true,
		},
		{
			name: "client ID too short",
			envContent: `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://auth.example.com
GOOGLE_CLIENT_ID=short
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(envFile, []byte(tt.envContent), 0600); err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			_, err := NewConfig(WithEnvFile(envFile))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				var configErr *configerrors.ConfigError
				if !errors.As(err, &configErr) {
					t.Errorf("Error should be ConfigError type, got %T", err)
				}
			}
		})
	}
}

func TestWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []Option
		check   func(*testing.T, *Config)
		wantErr bool
	}{
		{
			name: "WithDatabaseURL",
			options: []Option{
				WithDatabaseURL("postgres://test:test@localhost/testdb"),
				WithAuthServiceURL("https://auth.example.com"),
			},
			check: func(t *testing.T, c *Config) {
				if c.DatabaseURL != "postgres://test:test@localhost/testdb" {
					t.Errorf("DatabaseURL not set correctly: %q", c.DatabaseURL)
				}
			},
			wantErr: false,
		},
		{
			name: "WithPort",
			options: []Option{
				WithDatabaseURL("postgres://localhost/db"),
				WithAuthServiceURL("https://auth.example.com"),
				WithPort("3000"),
			},
			check: func(t *testing.T, c *Config) {
				if c.Port != "3000" {
					t.Errorf("Port = %q, want %q", c.Port, "3000")
				}
			},
			wantErr: false,
		},
		{
			name: "WithDebug",
			options: []Option{
				WithDatabaseURL("postgres://localhost/db"),
				WithAuthServiceURL("https://auth.example.com"),
				WithDebug(true),
			},
			check: func(t *testing.T, c *Config) {
				if !c.Debug {
					t.Error("Debug should be true")
				}
			},
			wantErr: false,
		},
		{
			name: "invalid port in option",
			options: []Option{
				WithDatabaseURL("postgres://localhost/db"),
				WithAuthServiceURL("https://auth.example.com"),
				WithPort("invalid"),
			},
			wantErr: true,
		},
		{
			name: "invalid database URL in option",
			options: []Option{
				WithDatabaseURL("not-a-url"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewConfig(tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, config)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory with .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `DATABASE_URL=postgres://localhost/testdb
AUTH_SERVICE_URL=https://auth.example.com
`
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Change to temp directory so LoadConfig finds the .env file
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}

	if config.DatabaseURL != "postgres://localhost/testdb" {
		t.Errorf("DatabaseURL = %q, want %q", config.DatabaseURL, "postgres://localhost/testdb")
	}
}

func TestFindEnvFilePath(t *testing.T) {
	// Create nested directory structure with .env file
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "sub", "nested", "deep")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	envFile := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envFile, []byte("TEST=true"), 0600); err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Search from nested directory
	found := FindEnvFilePath(nestedDir)
	if found != envFile {
		t.Errorf("FindEnvFilePath() = %q, want %q", found, envFile)
	}

	// Search from a completely different temp directory with no .env
	tmpDir2 := t.TempDir()
	notFound := FindEnvFilePath(tmpDir2)
	if notFound != "" {
		t.Errorf("FindEnvFilePath() should return empty string when not found, got %q", notFound)
	}
}

func TestEnvFilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		permissions os.FileMode
		wantWarning bool
	}{
		{"secure permissions", 0600, false},
		{"insecure permissions", 0644, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := filepath.Join(tmpDir, tt.name+".env")
			content := `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://auth.example.com
`
			if err := os.WriteFile(envFile, []byte(content), tt.permissions); err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			// Loading config should succeed but may log a warning
			config, err := NewConfig(WithEnvFile(envFile))
			if err != nil {
				t.Errorf("NewConfig() should succeed even with insecure permissions: %v", err)
			}

			if config == nil {
				t.Error("Config should not be nil")
			}
		})
	}
}

// Security tests
func TestSecurityInjectionAttempts(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	tests := []struct {
		name       string
		envContent string
		shouldFail bool
	}{
		{
			name: "SQL injection in database URL",
			envContent: `DATABASE_URL=postgres://user'; DROP TABLE users;--@localhost/db
AUTH_SERVICE_URL=https://auth.example.com
`,
			shouldFail: true, // Invalid URL format due to special characters
		},
		{
			name: "XSS in auth URL",
			envContent: `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://example.com/<script>alert('xss')</script>
`,
			shouldFail: false, // URL parsing handles this, but it's escaped in the path
		},
		{
			name: "command injection in port",
			envContent: `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://auth.example.com
PORT=8080; rm -rf /
`,
			shouldFail: true, // Should fail port validation
		},
		{
			name: "null bytes",
			envContent: "DATABASE_URL=postgres://localhost/db\x00invalid\nAUTH_SERVICE_URL=https://auth.example.com\n",
			shouldFail: true, // Invalid characters in URL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.WriteFile(envFile, []byte(tt.envContent), 0600); err != nil {
				t.Fatalf("Failed to create test .env file: %v", err)
			}

			_, err := NewConfig(WithEnvFile(envFile))
			if tt.shouldFail && err == nil {
				t.Error("NewConfig() should have failed for malicious input")
			} else if !tt.shouldFail && err != nil {
				t.Errorf("NewConfig() should not have failed: %v", err)
			}
		})
	}
}

func TestMisconfiguration(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		setup   func() string
		wantErr bool
	}{
		{
			name: "nonexistent env file",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.env")
			},
			wantErr: true,
		},
		{
			name: "empty env file",
			setup: func() string {
				envFile := filepath.Join(tmpDir, "empty.env")
				os.WriteFile(envFile, []byte(""), 0600)
				return envFile
			},
			wantErr: true, // Should fail validation for missing required fields
		},
		{
			name: "malformed env file",
			setup: func() string {
				envFile := filepath.Join(tmpDir, "malformed.env")
				os.WriteFile(envFile, []byte("INVALID LINE WITHOUT EQUALS\n"), 0600)
				return envFile
			},
			wantErr: true, // Should fail validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile := tt.setup()
			_, err := NewConfig(WithEnvFile(envFile))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	// Test that values are sanitized (whitespace trimmed, etc.)
	envContent := `DATABASE_URL=  postgres://localhost/db
AUTH_SERVICE_URL=  https://auth.example.com
PORT=  8080
`
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config, err := NewConfig(WithEnvFile(envFile))
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify values are sanitized (trimmed)
	if config.DatabaseURL != "postgres://localhost/db" {
		t.Errorf("DatabaseURL should be trimmed: %q", config.DatabaseURL)
	}

	if config.AuthServiceURL != "https://auth.example.com" {
		t.Errorf("AuthServiceURL should be trimmed: %q", config.AuthServiceURL)
	}

	if config.Port != "8080" {
		t.Errorf("Port should be trimmed: %q", config.Port)
	}
}

func TestDefaultPort(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")
	envContent := `DATABASE_URL=postgres://localhost/db
AUTH_SERVICE_URL=https://auth.example.com
`
	if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	config, err := NewConfig(WithEnvFile(envFile))
	if err != nil {
		t.Fatalf("NewConfig() failed: %v", err)
	}

	// Verify default port is set
	if config.Port != "8080" {
		t.Errorf("Port should default to 8080, got %q", config.Port)
	}
}
