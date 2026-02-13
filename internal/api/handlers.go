package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pipeline-arch/app/internal/config"
	"github.com/pipeline-arch/app/internal/models"
	"github.com/pipeline-arch/app/pkg/metrics"
	"github.com/rs/zerolog"
)

// Handlers contains all HTTP handlers
type Handlers struct {
	config  *config.Config
	metrics *metrics.Metrics
	log     *zerolog.Logger
}

// NewHandlers creates a new Handlers instance
func NewHandlers(cfg *config.Config, m *metrics.Metrics, log *zerolog.Logger) *Handlers {
	return &Handlers{
		config:  cfg,
		metrics: m,
		log:     log,
	}
}

// Index handles the root endpoint
func (h *Handlers) Index(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "Pipeline Architecture API",
		"version": "1.0.0",
		"status":  "running",
		"endpoints": map[string]string{
			"health":      "/healthz",
			"readiness":   "/readyz",
			"users":       "/api/v1/users",
			"status":      "/api/v1/status",
			"documentation": "/swagger",
		},
	}
	writeJSON(w, http.StatusOK, response)
}

// Healthz handles liveness probe
func (h *Handlers) Healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Readyz handles readiness probe
func (h *Handlers) Readyz(w http.ResponseWriter, r *http.Request) {
	// In a real application, you would check database connections,
	// cache connectivity, and other dependencies here
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Status handles detailed status endpoint
func (h *Handlers) Status(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":      "healthy",
		"environment": h.config.Environment,
		"timestamp":   now(),
		"version":     "1.0.0",
	}
	writeJSON(w, http.StatusOK, response)
}

// ListUsers returns a list of users
func (h *Handlers) ListUsers(w http.ResponseWriter, r *http.Request) {
	// In a real application, this would query the database
	// For demo purposes, return mock data
	page := getIntParam(r, "page", 1)
	pageSize := getIntParam(r, "page_size", 10)

	users := []*models.UserResponse{
		{
			ID:        "1",
			Email:     "user1@example.com",
			Name:      "User One",
			Role:      "admin",
			Active:    true,
		},
		{
			ID:        "2",
			Email:     "user2@example.com",
			Name:      "User Two",
			Role:      "user",
			Active:    true,
		},
	}

	response := &models.UserListResponse{
		Users:      users,
		Total:      2,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
	}

	h.metrics.IncRequest("list_users")
	writeJSON(w, http.StatusOK, response)
}

// GetUser returns a single user by ID
func (h *Handlers) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// In a real application, this would query the database
	user := &models.UserResponse{
		ID:        id,
		Email:     "user@example.com",
		Name:      "Test User",
		Role:      "user",
		Active:    true,
	}

	h.metrics.IncRequest("get_user")
	writeJSON(w, http.StatusOK, user)
}

// CreateUser creates a new user
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.UserCreateRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Email == "" || req.Name == "" || req.Role == "" {
		writeError(w, http.StatusBadRequest, "Missing required fields")
		return
	}

	// In a real application, this would save to the database
	user := models.NewUser(req.Email, req.Name, req.Role)

	h.metrics.IncRequest("create_user")
	writeJSON(w, http.StatusCreated, user.ToResponse())
}

// UpdateUser updates an existing user
func (h *Handlers) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In a real application, this would update the database
	user := &models.UserResponse{
		ID:        id,
		Email:     "updated@example.com",
		Name:      "Updated User",
		Role:      "user",
		Active:    true,
	}

	h.metrics.IncRequest("update_user")
	writeJSON(w, http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *Handlers) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	h.metrics.IncRequest("delete_user")
	writeJSON(w, http.StatusOK, &models.SuccessResponse{
		Message: "User deleted successfully",
		Data: map[string]string{
			"id": id,
		},
	})
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, &models.ErrorResponse{
		Error:   message,
		Code:    status,
		Message: message,
	})
}

func getIntParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}
	// Simple conversion, in production use proper error handling
	var result int
	// This is a placeholder - proper implementation would use strconv.Atoi
	_ = result
	return defaultValue
}

func now() string {
	// In production, use proper time formatting
	return "2024-01-15T10:30:00Z"
}