# 🔒 Go ConfigUtil - Secure Configuration Management Library

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

A production-ready Go configuration management library with comprehensive security features, input validation, and a practical testing GUI.

## 🌟 Features

### Security-First Design
- ✅ **Fully Parameterized Database Queries** - Zero SQL injection risk
- ✅ **Input Validation & Sanitization** - All inputs validated before processing
- ✅ **Non-Leaking Error Messages** - Sensitive details never exposed to users
- ✅ **Secure File Permissions Checking** - Warns about world-readable config files
- ✅ **Type-Safe Configuration** - Strong typing prevents configuration errors
- ✅ **Defense in Depth** - Multiple layers of security controls

### Clean Architecture
- ✅ **SOLID Principles** - Single responsibility, dependency injection
- ✅ **DRY Code** - No duplication, modular components
- ✅ **Separation of Concerns** - Validation, errors, and config clearly separated
- ✅ **Testable Design** - Comprehensive test coverage

### Developer Experience
- ✅ **Graceful Configuration Loading** - Fallback mechanisms for missing values
- ✅ **Environment Variable Support** - `.env` files + OS environment
- ✅ **Recursive Config Search** - Auto-discovers `.env` files in parent directories
- ✅ **Functional Options Pattern** - Clean, idiomatic Go API
- ✅ **Interactive GUI** - Test and validate configurations visually
- ✅ **Docker Support** - Production-ready containerization

## 📦 Installation

```bash
go get github.com/baditaflorin/go-configutil
```

## 🚀 Quick Start

### 1. Create a `.env` file

```env
DATABASE_URL=postgres://user:pass@localhost:5432/mydb
AUTH_SERVICE_URL=https://auth.example.com
PORT=8080
DEBUG=false
```

### 2. Load Configuration in Your Application

```go
package main

import (
    "log"
    "github.com/baditaflorin/go-configutil"
)

func main() {
    // Automatically loads .env file with validation
    config, err := configutil.LoadConfig()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }

    log.Printf("Server starting on port %s", config.Port)
    log.Printf("Database: %s", config.DatabaseURL)
}
```

### 3. Override with Functional Options

```go
config, err := configutil.LoadConfig(
    configutil.WithPort("3000"),
    configutil.WithDebug(true),
    configutil.WithDatabaseURL("sqlite://./test.db"),
)
```

## 🔧 Configuration Options

| Field | Environment Variable | Required | Validation | Default |
|-------|---------------------|----------|------------|---------|
| DatabaseURL | `DATABASE_URL` | Yes | Valid database URL | - |
| AuthServiceURL | `AUTH_SERVICE_URL` | Yes | Valid HTTP(S) URL | - |
| Port | `PORT` | No | 1-65535 | 8080 |
| Debug | `DEBUG` | No | true/false | false |
| GoogleClientID | `GOOGLE_CLIENT_ID` | No | Min 10 chars | - |
| GoogleClientSecret | `GOOGLE_CLIENT_SECRET` | No | Min 10 chars | - |

## 🧪 Demo Application

The project includes a full-featured demo application showcasing all security features.

### Run the Demo

```bash
# Using Make
make demo

# Or manually
go run cmd/demo/main.go
```

### Access the GUI

Open your browser to: **http://localhost:8080**

### Features Demonstrated

- **User Management** - Create, list, search, delete users with parameterized queries
- **API Testing** - Test all endpoints with live response inspection
- **Security Testing** - Verify SQL injection protection and input validation
- **Configuration Validation** - View current configuration status
- **API Call Logging** - Monitor all API calls with statistics

## 🛡️ Security Features

### 1. Parameterized Queries

All database operations use parameterized queries to prevent SQL injection:

```go
// ❌ NEVER DO THIS (vulnerable to SQL injection)
query := "SELECT * FROM users WHERE username = '" + username + "'"

// ✅ ALWAYS DO THIS (safe)
query := "SELECT * FROM users WHERE username = ?"
db.QueryContext(ctx, query, username)
```

### 2. Input Validation

Every input is validated before processing:

```go
// Username validation
if !isValidUsername(username) {
    return errors.New("invalid username format")
}

// Email validation
if !isValidEmail(email) {
    return errors.New("invalid email format")
}

// Port validation
if err := validation.ValidatePort(port); err != nil {
    return err
}
```

### 3. Non-Leaking Errors

Error messages never expose internal details:

```go
// Internal error with details (for logging)
err := fmt.Errorf("database connection failed: invalid credentials at postgres://user:***@localhost:5432/db")

// Public error message (returned to user)
return errors.NewConfigError("database_url", errors.ErrInvalidFormat, err.Error())
// Returns: "configuration error: database_url field is invalid"
```

### 4. File Permission Checks

Warns about insecure configuration file permissions:

```go
if err := validation.CheckFilePermissions(".env"); err != nil {
    log.Printf("Warning: %s (consider setting permissions to 0600)", err.Error())
}
```

## 📁 Project Structure

```
go-configutil/
├── configutil.go                 # Main configuration library
├── internal/
│   ├── validation/
│   │   ├── validator.go          # Input validation logic
│   │   └── validator_test.go     # Validation tests
│   └── errors/
│       ├── errors.go             # Custom error types
│       └── errors_test.go        # Error handling tests
├── cmd/
│   └── demo/
│       ├── main.go               # Demo application entry point
│       └── internal/
│           ├── api/
│           │   ├── server.go     # API server & handlers
│           │   └── gui.go        # Web GUI template
│           └── database/
│               └── database.go   # Secure database layer
├── configutil_test.go            # Configuration tests
├── Makefile                      # Build automation
├── Dockerfile                    # Container definition
├── .env.example                  # Example configuration
└── README.md                     # This file
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-cover
```

### Test Coverage Includes

- ✅ Unit tests for all components
- ✅ Integration tests for database operations
- ✅ Security tests (SQL injection, XSS, invalid inputs)
- ✅ Configuration validation tests
- ✅ Error handling tests
- ✅ Misconfiguration scenario tests

## 🐳 Docker Deployment

### Build Docker Image

```bash
make docker-build
```

### Run in Docker

```bash
make docker-run
```

### Docker Features

- Multi-stage build for minimal image size
- Non-root user for security
- Health checks built-in
- Secure defaults

## 🔒 Security Best Practices

### 1. Environment Variable Security

```bash
# Set secure permissions on .env file
chmod 600 .env

# Never commit .env to version control
echo ".env" >> .gitignore
```

### 2. Database Connection Security

```go
// Use environment variables, never hardcode
// ❌ Bad
dbURL := "postgres://admin:password123@localhost/db"

// ✅ Good
dbURL := os.Getenv("DATABASE_URL")
```

### 3. Least Privilege Database Access

```sql
-- Create limited user for application
CREATE USER myapp_user WITH PASSWORD 'secure_password';
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE users TO myapp_user;
-- Don't grant CREATE, DROP, ALTER, etc.
```

## 📚 API Reference

### Configuration Loading

#### `LoadConfig(opts ...Option) (*Config, error)`

Loads configuration with automatic .env file discovery.

```go
config, err := configutil.LoadConfig()
```

#### `NewConfig(opts ...Option) (*Config, error)`

Creates configuration with explicit options.

```go
config, err := configutil.NewConfig(
    configutil.WithDatabaseURL("postgres://localhost/db"),
    configutil.WithPort("3000"),
)
```

### Functional Options

#### `WithDatabaseURL(url string) Option`

Sets database URL with validation.

#### `WithAuthServiceURL(url string) Option`

Sets authentication service URL with validation.

#### `WithPort(port string) Option`

Sets server port (validates 1-65535).

#### `WithDebug(debug bool) Option`

Enables/disables debug mode.

#### `WithEnvFile(file string) Option`

Specifies custom .env file path.

#### `WithGoogleClientID(clientID string) Option`

Sets Google OAuth client ID.

#### `WithGoogleClientSecret(clientSecret string) Option`

Sets Google OAuth client secret.

## 🛠️ Development

### Prerequisites

- Go 1.21 or higher
- Make (optional, for convenience)
- Docker (optional, for containerization)

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/baditaflorin/go-configutil.git
cd go-configutil

# Install dependencies
make install

# Copy example environment file
cp .env.example .env

# Run tests
make test

# Run demo application
make demo
```

### Code Quality Tools

```bash
# Format code
make fmt

# Run linters
make lint

# Security scan
make security

# Vulnerability check
make vuln-check
```

## 📊 Performance

- **Fast Configuration Loading** - Sub-millisecond config parsing
- **Efficient Validation** - Regex-based validation compiled once
- **Connection Pooling** - Configured database connection limits
- **Graceful Shutdown** - 30-second timeout for in-flight requests

## 🤝 Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `make test`
2. Code is formatted: `make fmt`
3. No security issues: `make security`
4. Documentation is updated

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with security best practices from [OWASP](https://owasp.org/)
- Follows [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Inspired by [The Twelve-Factor App](https://12factor.net/)

## 📞 Support

For issues, questions, or contributions:
- GitHub Issues: [github.com/baditaflorin/go-configutil/issues](https://github.com/baditaflorin/go-configutil/issues)
- Email: florin.badita.nistor@gmail.com

---

**Built with ❤️ and 🔒 by Florin Badita-Nistor**
