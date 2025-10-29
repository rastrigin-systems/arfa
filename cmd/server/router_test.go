package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
)

// setupTestRouter creates a router with all handlers wired up for testing
func setupTestRouter(queries *db.Queries) chi.Router {
	router := chi.NewRouter()

	// Create handlers
	authHandler := handlers.NewAuthHandler(queries)
	teamsHandler := handlers.NewTeamsHandler(queries)
	orgAgentConfigsHandler := handlers.NewOrgAgentConfigsHandler(queries)
	teamAgentConfigsHandler := handlers.NewTeamAgentConfigsHandler(queries)
	employeeAgentConfigsHandler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router.Route("/api/v1", func(r chi.Router) {
		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Get("/me", authHandler.GetMe)
		})

		// Teams routes
		r.Route("/teams", func(r chi.Router) {
			r.Get("/", teamsHandler.ListTeams)
			r.Post("/", teamsHandler.CreateTeam)
			r.Get("/{team_id}", teamsHandler.GetTeam)
			r.Patch("/{team_id}", teamsHandler.UpdateTeam)
			r.Delete("/{team_id}", teamsHandler.DeleteTeam)

			// Team agent configs
			r.Route("/{team_id}/agent-configs", func(r chi.Router) {
				r.Get("/", teamAgentConfigsHandler.ListTeamAgentConfigs)
				r.Post("/", teamAgentConfigsHandler.CreateTeamAgentConfig)
				r.Get("/{config_id}", teamAgentConfigsHandler.GetTeamAgentConfig)
				r.Patch("/{config_id}", teamAgentConfigsHandler.UpdateTeamAgentConfig)
				r.Delete("/{config_id}", teamAgentConfigsHandler.DeleteTeamAgentConfig)
			})
		})

		// Organizations routes
		r.Route("/organizations/current/agent-configs", func(r chi.Router) {
			r.Get("/", orgAgentConfigsHandler.ListOrgAgentConfigs)
			r.Post("/", orgAgentConfigsHandler.CreateOrgAgentConfig)
			r.Get("/{config_id}", orgAgentConfigsHandler.GetOrgAgentConfig)
			r.Patch("/{config_id}", orgAgentConfigsHandler.UpdateOrgAgentConfig)
			r.Delete("/{config_id}", orgAgentConfigsHandler.DeleteOrgAgentConfig)
		})

		// Employee agent configs routes
		r.Route("/employees/{employee_id}/agent-configs", func(r chi.Router) {
			r.Get("/", employeeAgentConfigsHandler.ListEmployeeAgentConfigs)
			r.Post("/", employeeAgentConfigsHandler.CreateEmployeeAgentConfig)
			r.Get("/resolved", orgAgentConfigsHandler.GetEmployeeResolvedAgentConfigs)
			r.Get("/{config_id}", employeeAgentConfigsHandler.GetEmployeeAgentConfig)
			r.Patch("/{config_id}", employeeAgentConfigsHandler.UpdateEmployeeAgentConfig)
			r.Delete("/{config_id}", employeeAgentConfigsHandler.DeleteEmployeeAgentConfig)
		})
	})

	return router
}

// Test helper to create a test database connection
func setupTestDB(t *testing.T) (*pgxpool.Pool, *db.Queries, func()) {
	// Use the same test database setup as integration tests
	ctx := context.Background()
	dbURL := "postgres://pivot:pivot_dev_password@localhost:5432/pivot_test?sslmode=disable"

	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to test database")

	queries := db.New(pool)

	cleanup := func() {
		pool.Close()
	}

	return pool, queries, cleanup
}

// TestTeamsRoutes_List verifies GET /teams route is registered
func TestTeamsRoutes_List(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should not be 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "Teams route should be registered")

	// Should be 401 (unauthorized) or 200 (if no auth required)
	// We'll implement auth later, so for now expect 401 or 200
	assert.Contains(t, []int{http.StatusOK, http.StatusUnauthorized}, rec.Code)
}

// TestTeamsRoutes_Create verifies POST /teams route is registered
func TestTeamsRoutes_Create(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	reqBody := api.CreateTeamRequest{
		Name:        "Engineering",
		Description: stringPtr("Software development team"),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should not be 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "POST /teams route should be registered")
}

// TestTeamsRoutes_GetByID verifies GET /teams/{team_id} route is registered
func TestTeamsRoutes_GetByID(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	teamID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should not be 404 for route not found
	// Will be 404 for team not found, or 401 unauthorized, or 400 bad request
	assert.NotEqual(t, http.StatusNotFound, rec.Code, "GET /teams/{id} route should be registered")
}

// TestTeamsRoutes_Update verifies PATCH /teams/{team_id} route is registered
func TestTeamsRoutes_Update(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	teamID := "550e8400-e29b-41d4-a716-446655440000"
	reqBody := api.UpdateTeamRequest{
		Name: stringPtr("Engineering Updated"),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/teams/"+teamID, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "PATCH /teams/{id} route should be registered")
}

// TestTeamsRoutes_Delete verifies DELETE /teams/{team_id} route is registered
func TestTeamsRoutes_Delete(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	teamID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/teams/"+teamID, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "DELETE /teams/{id} route should be registered")
}

// TestTeamAgentConfigRoutes_List verifies GET /teams/{team_id}/agent-configs route is registered
func TestTeamAgentConfigRoutes_List(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	teamID := "550e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/teams/"+teamID+"/agent-configs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "GET /teams/{id}/agent-configs route should be registered")
}

// TestTeamAgentConfigRoutes_Create verifies POST /teams/{team_id}/agent-configs route is registered
func TestTeamAgentConfigRoutes_Create(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	teamID := "550e8400-e29b-41d4-a716-446655440000"
	agentIDStr := "660e8400-e29b-41d4-a716-446655440000"
	agentUUID := mustParseUUID(agentIDStr)

	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: agentUUID,
		ConfigOverride: map[string]interface{}{
			"max_tokens": 4000,
		},
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/teams/"+teamID+"/agent-configs", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "POST /teams/{id}/agent-configs route should be registered")
}

// TestEmployeeAgentConfigRoutes_List verifies GET /employees/{employee_id}/agent-configs route is registered
func TestEmployeeAgentConfigRoutes_List(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	employeeID := "770e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+employeeID+"/agent-configs", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "GET /employees/{id}/agent-configs route should be registered")
}

// TestEmployeeAgentConfigRoutes_Resolved verifies GET /employees/{employee_id}/agent-configs/resolved route is registered
func TestEmployeeAgentConfigRoutes_Resolved(t *testing.T) {
	t.Skip("Integration test - requires database setup")

	_, queries, cleanup := setupTestDB(t)
	defer cleanup()

	router := setupTestRouter(queries)

	employeeID := "770e8400-e29b-41d4-a716-446655440000"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/employees/"+employeeID+"/agent-configs/resolved", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.NotEqual(t, http.StatusNotFound, rec.Code, "GET /employees/{id}/agent-configs/resolved route should be registered")
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func mustParseUUID(s string) api.TeamId {
	// Parse UUID from string
	parsed, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return api.TeamId(parsed)
}
