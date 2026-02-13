package unit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/pipeline-arch/app/internal/api"
	"github.com/pipeline-arch/app/internal/config"
	"github.com/pipeline-arch/app/pkg/logger"
	"github.com/pipeline-arch/app/pkg/metrics"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestHandler() (*api.Handlers, *chi.Mux) {
	cfg := &config.Config{
		Host:        "0.0.0.0",
		Port:        8080,
		Environment: "test",
		LogLevel:    "debug",
		MetricsPort: 9090,
	}

	log := logger.New("debug")
	m := metrics.New("pipeline-arch-test", 9091)

	handlers := api.NewHandlers(cfg, m, log)
	router := chi.NewRouter()

	// Setup test routes
	router.Get("/healthz", handlers.Healthz)
	router.Get("/readyz", handlers.Readyz)
	router.Get("/api/v1/users", handlers.ListUsers)
	router.Post("/api/v1/users", handlers.CreateUser)
	router.Get("/api/v1/users/{id}", handlers.GetUser)
	router.Put("/api/v1/users/{id}", handlers.UpdateUser)
	router.Delete("/api/v1/users/{id}", handlers.DeleteUser)

	return handlers, router
}

func TestHealthz(t *testing.T) {
	handlers, router := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestReadyz(t *testing.T) {
	handlers, router := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())
}

func TestListUsers(t *testing.T) {
	_, router := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	users, ok := response["users"].([]interface{})
	assert.True(t, ok, "users should be an array")
	assert.GreaterOrEqual(t, len(users), 1, "should have at least one user")
}

func TestGetUser(t *testing.T) {
	_, router := setupTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.Equal(t, "123", user["id"])
	assert.Equal(t, "user@example.com", user["email"])
	assert.Equal(t, "Test User", user["name"])
}

func TestCreateUser(t *testing.T) {
	_, router := setupTestHandler()

	body := map[string]string{
		"email": "newuser@example.com",
		"name":  "New User",
		"role":  "user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var user map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.NotEmpty(t, user["id"])
	assert.Equal(t, "newuser@example.com", user["email"])
	assert.Equal(t, "New User", user["name"])
	assert.Equal(t, "user", user["role"])
}

func TestCreateUserValidation(t *testing.T) {
	_, router := setupTestHandler()

	// Test with missing email
	body := map[string]string{
		"name": "Test User",
		"role": "user",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateUser(t *testing.T) {
	_, router := setupTestHandler()

	body := map[string]string{
		"name": "Updated User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/123", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user)
	require.NoError(t, err)

	assert.Equal(t, "123", user["id"])
	assert.Equal(t, "Updated User", user["name"])
}

func TestDeleteUser(t *testing.T) {
	_, router := setupTestHandler()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "User deleted successfully", response["message"])
}

func TestUserResponseToJSON(t *testing.T) {
	cfg := &config.Config{
		Host:        "0.0.0.0",
		Port:        8080,
		Environment: "test",
		LogLevel:    "debug",
		MetricsPort: 9090,
	}

	log := logger.New("debug")
	_ = metrics.New("pipeline-arch-test", 9091)
	handlers := api.NewHandlers(cfg, nil, log)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handlers.Index(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "Pipeline Architecture API", response["name"])
	assert.Equal(t, "1.0.0", response["version"])
	assert.Equal(t, "running", response["status"])
}