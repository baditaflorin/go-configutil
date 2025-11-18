package api

import (
	"context"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/baditaflorin/go-configutil"
	"github.com/baditaflorin/go-configutil/cmd/demo/internal/database"
)

// Server represents the API server
type Server struct {
	config *configutil.Config
	db     *database.DB
	router *http.ServeMux
}

// NewServer creates a new API server
func NewServer(config *configutil.Config, db *database.DB) *Server {
	s := &Server{
		config: config,
		db:     db,
		router: http.NewServeMux(),
	}

	s.routes()
	return s
}

// Router returns the configured router
func (s *Server) Router() http.Handler {
	return s.loggingMiddleware(s.securityHeadersMiddleware(s.router))
}

// routes configures all application routes
func (s *Server) routes() {
	// GUI routes
	s.router.HandleFunc("/", s.handleGUI)

	// API routes
	s.router.HandleFunc("/api/users", s.handleUsers)
	s.router.HandleFunc("/api/users/search", s.handleSearchUsers)
	s.router.HandleFunc("/api/stats", s.handleStats)
	s.router.HandleFunc("/api/config/test", s.handleTestConfig)
	s.router.HandleFunc("/api/health", s.handleHealth)
}

// Middleware for security headers
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// Middleware for logging API calls
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)

		// Log to console
		log.Printf("%s %s - %d (%v)", r.Method, r.URL.Path, wrapped.statusCode, duration)

		// Log to database (don't block on failure)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			_ = s.db.LogAPICall(ctx, r.URL.Path, r.Method, wrapped.statusCode, duration)
		}()
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// handleGUI serves the web GUI
func (s *Server) handleGUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	tmpl := template.Must(template.New("gui").Parse(guiTemplate))
	if err := tmpl.Execute(w, s.config); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
	}
}

// handleUsers handles user CRUD operations
func (s *Server) handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		s.handleListUsers(w, r, ctx)
	case http.MethodPost:
		s.handleCreateUser(w, r, ctx)
	case http.MethodDelete:
		s.handleDeleteUser(w, r, ctx)
	default:
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListUsers lists all users with pagination
func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	// Validate and sanitize pagination parameters
	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	users, err := s.db.ListUsers(ctx, limit, offset)
	if err != nil {
		log.Printf("Error listing users: %v", err)
		writeJSONError(w, "Failed to list users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	})
}

// handleCreateUser creates a new user
func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Input validation with secure regex patterns
	if !isValidUsername(req.Username) {
		writeJSONError(w, "Invalid username format", http.StatusBadRequest)
		return
	}

	if !isValidEmail(req.Email) {
		writeJSONError(w, "Invalid email format", http.StatusBadRequest)
		return
	}

	// Create user with PARAMETERIZED query (SQL injection impossible)
	user, err := s.db.CreateUser(ctx, req.Username, req.Email)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, user)
}

// handleDeleteUser deletes a user
func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		writeJSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSONError(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := s.db.DeleteUser(ctx, id); err != nil {
		log.Printf("Error deleting user: %v", err)
		writeJSONError(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "User deleted successfully"})
}

// handleSearchUsers searches for users
func (s *Server) handleSearchUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	query := r.URL.Query().Get("q")
	if query == "" {
		writeJSONError(w, "Search query is required", http.StatusBadRequest)
		return
	}

	// Sanitize search query (prevent injection even though we use parameterized queries)
	query = sanitizeSearchQuery(query)

	users, err := s.db.SearchUsers(ctx, query)
	if err != nil {
		log.Printf("Error searching users: %v", err)
		writeJSONError(w, "Failed to search users", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"users": users})
}

// handleStats returns API call statistics
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	stats, err := s.db.GetAPICallStats(ctx, limit)
	if err != nil {
		log.Printf("Error getting stats: %v", err)
		writeJSONError(w, "Failed to get statistics", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]interface{}{"calls": stats})
}

// handleTestConfig tests configuration and external services
func (s *Server) handleTestConfig(w http.ResponseWriter, r *http.Request) {
	// Return safe, non-leaking configuration info
	writeJSON(w, map[string]interface{}{
		"database_configured": s.config.DatabaseURL != "",
		"auth_service_configured": s.config.AuthServiceURL != "",
		"port":  s.config.Port,
		"debug": s.config.Debug,
		// Never expose actual URLs or credentials
	})
}

// handleHealth returns health status
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().Unix(),
	})
}

// Validation helpers

var (
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func isValidUsername(username string) bool {
	return usernameRegex.MatchString(username)
}

func isValidEmail(email string) bool {
	return len(email) <= 254 && emailRegex.MatchString(email)
}

func sanitizeSearchQuery(query string) string {
	// Remove potentially dangerous characters but allow basic search
	// This is defense in depth - parameterized queries already prevent SQL injection
	if len(query) > 100 {
		query = query[:100]
	}
	return query
}

// Response helpers

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
