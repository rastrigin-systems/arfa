package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/internal/middleware"
	"github.com/sergeirastrigin/ubik-enterprise/tests/testutil"
)

// ============================================================================
// List Team Agent Configs Integration Tests
// ============================================================================

func TestListTeamAgentConfigs_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agents from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")

	// Create team agent configs
	config1JSON := []byte(`{"max_tokens":4000}`)
	_, err = queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agents[0].ID, config1JSON, true))
	require.NoError(t, err)

	config2JSON := []byte(`{"temperature":0.7}`)
	_, err = queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agents[1].ID, config2JSON, false))
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/teams/{team_id}/agent-configs", handler.ListTeamAgentConfigs)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/teams/"+team.ID.String()+"/agent-configs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListTeamAgentConfigsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return 2 configs
	assert.Equal(t, 2, response.Total)
	require.Len(t, response.Configs, 2)

	// Verify config data
	assert.NotNil(t, response.Configs[0].AgentName)
	assert.NotNil(t, response.Configs[0].ConfigOverride)
	assert.NotNil(t, response.Configs[1].AgentName)
	assert.NotNil(t, response.Configs[1].ConfigOverride)

	// Verify we have both enabled and disabled configs
	enabledCount := 0
	disabledCount := 0
	for _, cfg := range response.Configs {
		if cfg.IsEnabled {
			enabledCount++
		} else {
			disabledCount++
		}
	}
	assert.Equal(t, 1, enabledCount, "Should have 1 enabled config")
	assert.Equal(t, 1, disabledCount, "Should have 1 disabled config")
}

func TestListTeamAgentConfigs_Integration_TeamNotFound(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/teams/{team_id}/agent-configs", handler.ListTeamAgentConfigs)

	// Make request with non-existent team ID
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/teams/"+nonExistentID.String()+"/agent-configs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// Create Team Agent Config Integration Tests
// ============================================================================

func TestCreateTeamAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/teams/{team_id}/agent-configs", handler.CreateTeamAgentConfig)

	// Create request
	isEnabled := true
	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: api.TeamId(agent.ID),
		ConfigOverride: map[string]interface{}{
			"max_tokens":  float64(4000),
			"temperature": 0.7,
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+team.ID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.TeamAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.NotNil(t, response.AgentName)
	assert.True(t, response.IsEnabled)
	assert.NotNil(t, response.ConfigOverride)
	assert.Equal(t, float64(4000), response.ConfigOverride["max_tokens"])
}

func TestCreateTeamAgentConfig_Integration_DuplicateAgent(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create first config
	configJSON := []byte(`{"max_tokens":4000}`)
	_, err = queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agent.ID, configJSON, true))
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/teams/{team_id}/agent-configs", handler.CreateTeamAgentConfig)

	// Try to create duplicate config
	isEnabled := true
	reqBody := api.CreateTeamAgentConfigRequest{
		AgentId: api.TeamId(agent.ID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(8000),
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/teams/"+team.ID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ============================================================================
// Get Team Agent Config Integration Tests
// ============================================================================

func TestGetTeamAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create team agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agent.ID, configJSON, true))
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/teams/{team_id}/agent-configs/{config_id}", handler.GetTeamAgentConfig)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/teams/"+team.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.TeamAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.NotNil(t, response.AgentName)
	assert.True(t, response.IsEnabled)
}

// ============================================================================
// Update Team Agent Config Integration Tests
// ============================================================================

func TestUpdateTeamAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create team agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agent.ID, configJSON, true))
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/teams/{team_id}/agent-configs/{config_id}", handler.UpdateTeamAgentConfig)

	// Update config
	isEnabled := false
	newConfig := map[string]interface{}{
		"max_tokens": float64(8000),
		"temperature": 0.9,
	}
	reqBody := api.UpdateTeamAgentConfigRequest{
		ConfigOverride: &newConfig,
		IsEnabled:      &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/teams/"+team.ID.String()+"/agent-configs/"+config.ID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.TeamAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsEnabled)
	assert.Equal(t, float64(8000), response.ConfigOverride["max_tokens"])
}

// ============================================================================
// Delete Team Agent Config Integration Tests
// ============================================================================

func TestDeleteTeamAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create test team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create team agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateTeamAgentConfig(ctx, testutil.CreateTeamAgentConfigParams(team.ID, agent.ID, configJSON, true))
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewTeamAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Delete("/teams/{team_id}/agent-configs/{config_id}", handler.DeleteTeamAgentConfig)

	// Delete config
	req := httptest.NewRequest(http.MethodDelete, "/teams/"+team.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify config is deleted
	router2 := chi.NewRouter()
	router2.Use(middleware.JWTAuth(queries))
	router2.Get("/teams/{team_id}/agent-configs/{config_id}", handler.GetTeamAgentConfig)

	req2 := httptest.NewRequest(http.MethodGet, "/teams/"+team.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()

	router2.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusNotFound, rec2.Code)
}
