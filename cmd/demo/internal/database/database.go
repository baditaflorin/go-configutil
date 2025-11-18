package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// DB wraps the database connection with secure operations
type DB struct {
	conn *sql.DB
}

// User represents a user in the system
type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

// APICall represents an API call log entry
type APICall struct {
	ID         int64     `json:"id"`
	Endpoint   string    `json:"endpoint"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	Duration   int64     `json:"duration_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewDB creates a new database connection with proper configuration
func NewDB(databaseURL string) (*DB, error) {
	// Parse database URL and extract connection parameters
	// For demo purposes, we'll use SQLite
	// In production, you'd parse postgres:// or mysql:// URLs

	dbPath := "./demo.db"
	if databaseURL != "" {
		// Extract path from sqlite:// URL or use as-is for file path
		dbPath = extractSQLitePath(databaseURL)
	}

	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{conn: conn}, nil
}

// extractSQLitePath extracts the file path from a SQLite URL
func extractSQLitePath(url string) string {
	// Handle sqlite:// or sqlite3:// URLs
	if len(url) > 9 && url[:9] == "sqlite://" {
		return url[9:]
	}
	if len(url) > 10 && url[:10] == "sqlite3://" {
		return url[10:]
	}
	// If it's just a file path or ends with .db, use as-is
	return url
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.conn.Close()
}

// RunMigrations creates the necessary database tables
func (db *DB) RunMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
		`CREATE TABLE IF NOT EXISTS api_calls (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			endpoint TEXT NOT NULL,
			method TEXT NOT NULL,
			status_code INTEGER NOT NULL,
			duration_ms INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_api_calls_created_at ON api_calls(created_at)`,
	}

	for _, migration := range migrations {
		if _, err := db.conn.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// CreateUser creates a new user with FULLY PARAMETERIZED queries (NO SQL injection possible)
func (db *DB) CreateUser(ctx context.Context, username, email string) (*User, error) {
	// SECURITY: Using parameterized query - values are never concatenated into SQL
	query := `INSERT INTO users (username, email) VALUES (?, ?)`

	result, err := db.conn.ExecContext(ctx, query, username, email)
	if err != nil {
		// Don't leak internal database errors to the caller
		return nil, fmt.Errorf("failed to create user")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID")
	}

	return db.GetUser(ctx, id)
}

// GetUser retrieves a user by ID with PARAMETERIZED query
func (db *DB) GetUser(ctx context.Context, id int64) (*User, error) {
	// SECURITY: Parameterized query prevents SQL injection
	query := `SELECT id, username, email, created_at FROM users WHERE id = ?`

	var user User
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user")
	}

	return &user, nil
}

// GetUserByUsername retrieves a user by username with PARAMETERIZED query
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	// SECURITY: Parameterized query prevents SQL injection
	query := `SELECT id, username, email, created_at FROM users WHERE username = ?`

	var user User
	err := db.conn.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user")
	}

	return &user, nil
}

// ListUsers returns all users with pagination (PARAMETERIZED)
func (db *DB) ListUsers(ctx context.Context, limit, offset int) ([]*User, error) {
	// SECURITY: Parameterized query with limits
	query := `SELECT id, username, email, created_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`

	rows, err := db.conn.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users")
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user")
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}

// DeleteUser deletes a user by ID (PARAMETERIZED)
func (db *DB) DeleteUser(ctx context.Context, id int64) error {
	// SECURITY: Parameterized query
	query := `DELETE FROM users WHERE id = ?`

	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user")
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows")
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// LogAPICall logs an API call for monitoring (PARAMETERIZED)
func (db *DB) LogAPICall(ctx context.Context, endpoint, method string, statusCode int, duration time.Duration) error {
	// SECURITY: Parameterized query
	query := `INSERT INTO api_calls (endpoint, method, status_code, duration_ms) VALUES (?, ?, ?, ?)`

	_, err := db.conn.ExecContext(ctx, query, endpoint, method, statusCode, duration.Milliseconds())
	if err != nil {
		// Log error but don't fail the request
		return fmt.Errorf("failed to log API call")
	}

	return nil
}

// GetAPICallStats returns API call statistics (PARAMETERIZED)
func (db *DB) GetAPICallStats(ctx context.Context, limit int) ([]*APICall, error) {
	// SECURITY: Parameterized query
	query := `SELECT id, endpoint, method, status_code, duration_ms, created_at
	          FROM api_calls ORDER BY created_at DESC LIMIT ?`

	rows, err := db.conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get API call stats")
	}
	defer rows.Close()

	var calls []*APICall
	for rows.Next() {
		var call APICall
		if err := rows.Scan(&call.ID, &call.Endpoint, &call.Method, &call.StatusCode, &call.Duration, &call.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan API call")
		}
		calls = append(calls, &call)
	}

	return calls, rows.Err()
}

// SearchUsers searches users by username or email (PARAMETERIZED with LIKE)
func (db *DB) SearchUsers(ctx context.Context, searchTerm string) ([]*User, error) {
	// SECURITY: Even LIKE queries are parameterized to prevent SQL injection
	query := `SELECT id, username, email, created_at FROM users
	          WHERE username LIKE ? OR email LIKE ?
	          ORDER BY created_at DESC LIMIT 50`

	// Add wildcards in Go code, not in SQL
	searchPattern := "%" + searchTerm + "%"

	rows, err := db.conn.QueryContext(ctx, query, searchPattern, searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search users")
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user")
		}
		users = append(users, &user)
	}

	return users, rows.Err()
}
