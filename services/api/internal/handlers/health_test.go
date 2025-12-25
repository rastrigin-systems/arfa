package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// Health Check Tests
// ============================================================================

// TDD Lesson: Testing health check endpoint
// This endpoint is special - it doesn't need database access or authentication
// It's used by monitoring systems to verify the API is running
func TestHealthCheck_Success(t *testing.T) {
	// No mock DB needed - health check doesn't use database
	handler := handlers.NewHealthHandler()

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.HealthCheck(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var resp api.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, "healthy", resp.Status)

	// Verify timestamp is recent (within last 5 seconds)
	assert.WithinDuration(t, time.Now(), resp.Timestamp, 5*time.Second)
}

// TDD Lesson: Test that timestamp is in ISO8601 format (RFC3339)
func TestHealthCheck_TimestampFormat(t *testing.T) {
	handler := handlers.NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	var resp api.HealthResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify timestamp is not zero value
	assert.False(t, resp.Timestamp.IsZero(), "Timestamp should be set")

	// Verify it's recent (time.Time marshals to RFC3339 automatically)
	assert.WithinDuration(t, time.Now(), resp.Timestamp, 5*time.Second)
}

// TDD Lesson: Test that health check works with any HTTP method (GET, HEAD, etc.)
// Many monitoring systems use HEAD requests to minimize bandwidth
func TestHealthCheck_SupportsGET(t *testing.T) {
	handler := handlers.NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Body.String())
}
