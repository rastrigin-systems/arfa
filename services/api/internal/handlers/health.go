package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
)

// HealthHandler handles health check requests
//
// TDD Lesson: This handler is special - it doesn't need database access
// or authentication. It's used by monitoring systems to verify the API is running.
type HealthHandler struct {
	// No database needed for health checks
}

// NewHealthHandler creates a new health check handler
//
// TDD Lesson: We create this constructor for consistency with other handlers,
// even though this handler doesn't need any dependencies.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck returns the current health status of the API
//
// TDD Lesson: This implementation is driven by our tests.
// The tests require:
// 1. Return 200 OK
// 2. Return JSON with "status": "healthy"
// 3. Include timestamp in ISO8601 format
//
// Implementation note: This endpoint should be fast and not depend on
// external services. If you need to check database health, create a
// separate /health/deep endpoint.
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Build response
	timestamp := time.Now()
	resp := api.HealthResponse{
		Status:    "healthy",
		Timestamp: timestamp,
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
