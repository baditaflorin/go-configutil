package configutil

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/baditaflorin/go-configutil/internal/errors"
	"github.com/baditaflorin/go-configutil/internal/validation"
	"github.com/joho/godotenv"
)

// Config holds all application configuration with validated values
type Config struct {
	DatabaseURL        string
	AuthServiceURL     string
	Debug              bool
	Port               string
	EnvFile            string
	GoogleClientID     string
	GoogleClientSecret string
}

// Option is a functional option for configuring Config
type Option func(*Config) error

// LoadConfig loads configuration with automatic .env file discovery
// It searches for .env files recursively from the current working directory
func LoadConfig(opts ...Option) (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, errors.NewConfigError("working_directory", errors.ErrConfigurationInvalid, "failed to get current working directory")
	}

	envFilePath := FindEnvFilePath(wd)

	if envFilePath != "" {
		opts = append(opts, WithEnvFile(envFilePath))
	}

	return NewConfig(opts...)
}

// WithDatabaseURL sets the database URL with validation
func WithDatabaseURL(url string) Option {
	return func(c *Config) error {
		if url == "" {
			return nil
		}

		sanitized := validation.SanitizeString(url, 2048)
		if err := validation.ValidateDatabaseURL(sanitized); err != nil {
			return errors.NewInvalidFormatError("database_url", err.Error())
		}

		c.DatabaseURL = sanitized
		return nil
	}
}

// WithAuthServiceURL sets the authentication service URL with validation
func WithAuthServiceURL(url string) Option {
	return func(c *Config) error {
		if url == "" {
			return nil
		}

		sanitized := validation.SanitizeString(url, 2048)
		if err := validation.ValidateHTTPURL(sanitized); err != nil {
			return errors.NewInvalidFormatError("auth_service_url", err.Error())
		}

		c.AuthServiceURL = sanitized
		return nil
	}
}

// WithDebug sets the debug mode
func WithDebug(debug bool) Option {
	return func(c *Config) error {
		c.Debug = debug
		return nil
	}
}

// WithPort sets the service port with validation
func WithPort(port string) Option {
	return func(c *Config) error {
		if port == "" {
			return nil
		}

		sanitized := validation.SanitizeString(port, 5)
		if err := validation.ValidatePort(sanitized); err != nil {
			return errors.NewInvalidFormatError("port", err.Error())
		}

		c.Port = sanitized
		return nil
	}
}

// WithEnvFile sets the environment file path
func WithEnvFile(file string) Option {
	return func(c *Config) error {
		if file == "" {
			return nil
		}

		// Check if file exists
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return errors.NewConfigError("env_file", errors.ErrFileNotFound, fmt.Sprintf("env file not found: %s", file))
		}

		// Check file permissions for security
		if err := validation.CheckFilePermissions(file); err != nil {
			log.Printf("Warning: %s (consider setting permissions to 0600)", err.Error())
		}

		c.EnvFile = file
		return nil
	}
}

// WithGoogleClientID sets the Google OAuth client ID with validation
func WithGoogleClientID(clientID string) Option {
	return func(c *Config) error {
		if clientID == "" {
			return nil
		}

		sanitized := validation.SanitizeString(clientID, 512)
		if err := validation.ValidateClientID(sanitized); err != nil {
			return errors.NewInvalidFormatError("google_client_id", err.Error())
		}

		c.GoogleClientID = sanitized
		return nil
	}
}

// WithGoogleClientSecret sets the Google OAuth client secret with validation
func WithGoogleClientSecret(clientSecret string) Option {
	return func(c *Config) error {
		if clientSecret == "" {
			return nil
		}

		sanitized := validation.SanitizeString(clientSecret, 512)
		if err := validation.ValidateClientSecret(sanitized); err != nil {
			return errors.NewInvalidFormatError("google_client_secret", err.Error())
		}

		c.GoogleClientSecret = sanitized
		return nil
	}
}

// NewConfig creates a new configuration with the provided options
// It loads environment variables and validates all required fields
func NewConfig(opts ...Option) (*Config, error) {
	c := &Config{}

	// Apply all options with validation
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	// Load environment variables from .env file
	envs, err := loadEnv(c.EnvFile)
	if err != nil {
		return nil, err
	}

	// Populate configuration from environment with fallbacks
	if err := c.populateFromEnv(envs); err != nil {
		return nil, err
	}

	// Validate required fields
	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

// populateFromEnv fills configuration from environment variables with validation
func (c *Config) populateFromEnv(envs map[string]string) error {
	// DatabaseURL
	if dbURL := getEnvWithFallback(envs, "DATABASE_URL", c.DatabaseURL); dbURL != "" {
		if err := WithDatabaseURL(dbURL)(c); err != nil {
			return err
		}
	}

	// AuthServiceURL
	if authURL := getEnvWithFallback(envs, "AUTH_SERVICE_URL", c.AuthServiceURL); authURL != "" {
		if err := WithAuthServiceURL(authURL)(c); err != nil {
			return err
		}
	}

	// Port
	if port := getEnvWithFallback(envs, "PORT", c.Port); port != "" {
		if err := WithPort(port)(c); err != nil {
			return err
		}
	}

	// GoogleClientID
	if clientID := getEnvWithFallback(envs, "GOOGLE_CLIENT_ID", c.GoogleClientID); clientID != "" {
		if err := WithGoogleClientID(clientID)(c); err != nil {
			return err
		}
	}

	// GoogleClientSecret
	if clientSecret := getEnvWithFallback(envs, "GOOGLE_CLIENT_SECRET", c.GoogleClientSecret); clientSecret != "" {
		if err := WithGoogleClientSecret(clientSecret)(c); err != nil {
			return err
		}
	}

	// Debug (boolean)
	c.Debug = getBoolEnvWithFallback(envs, "DEBUG", c.Debug)

	return nil
}

// validate checks that all required configuration fields are present
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return errors.NewMissingFieldError("DATABASE_URL")
	}

	if c.AuthServiceURL == "" {
		return errors.NewMissingFieldError("AUTH_SERVICE_URL")
	}

	// Port is optional, but if set, should have a default
	if c.Port == "" {
		c.Port = "8080" // Default port
		log.Printf("Info: PORT not set, using default: %s", c.Port)
	}

	return nil
}

// getEnvWithFallback retrieves environment variable with fallback
func getEnvWithFallback(envs map[string]string, key, fallback string) string {
	// First check .env file
	if value, exists := envs[key]; exists && value != "" {
		return value
	}

	// Then check OS environment
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}

	// Return fallback
	return fallback
}

// getBoolEnvWithFallback retrieves boolean environment variable with fallback
func getBoolEnvWithFallback(envs map[string]string, key string, fallback bool) bool {
	strValue := getEnvWithFallback(envs, key, strconv.FormatBool(fallback))
	boolValue, err := strconv.ParseBool(strValue)
	if err != nil {
		// Safe error message, no details leaked
		log.Printf("Warning: invalid boolean value for %s, using fallback", key)
		return fallback
	}
	return boolValue
}

// loadEnv loads environment variables from .env file
func loadEnv(envFile string) (map[string]string, error) {
	if envFile == "" {
		envFile = os.Getenv("ENV_FILE")
		if envFile == "" {
			// No env file specified, use OS environment only
			return make(map[string]string), nil
		}
	}

	envs, err := godotenv.Read(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Graceful degradation: file not found
			log.Printf("Info: .env file not found at %s, using OS environment variables only", envFile)
			return make(map[string]string), nil
		}
		// Safe error message, don't leak file system details
		return nil, errors.NewConfigError("env_file", errors.ErrConfigurationInvalid, "failed to read environment file")
	}

	return envs, nil
}

// FindEnvFilePath recursively searches for the .env file in parent directories
// Returns empty string if not found
func FindEnvFilePath(startDir string) string {
	dir := startDir
	for {
		envFilePath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envFilePath); err == nil {
			return envFilePath
		}

		// Move up one directory
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached the root directory
			break
		}
		dir = parentDir
	}
	return ""
}
