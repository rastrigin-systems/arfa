package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/handlers"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/middleware"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigCascade_OrgOnly tests configuration resolution with only org-level config
func TestConfigCascade_OrgOnly(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "john@example.com",
		FullName: "John Doe",
		Status:   "active",
	})

	// Get an agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)
	agent := agents[0]

	// Create org-level config
	_, err = queries.CreateOrgAgentConfig(ctx, db.CreateOrgAgentConfigParams{
		OrgID:   org.ID,
		AgentID: agent.ID,
		Config: []byte(`{
			"model": "claude-3-5-sonnet-20241022",
			"temperature": 0.2,
			"max_tokens": 4096
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Setup router
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	handler := handlers.NewOrgAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs/resolved", handler.GetEmployeeResolvedAgentConfigs)

	// Test: Get resolved config
	url := fmt.Sprintf("/employees/%s/agent-configs/resolved", employee.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Configs []api.ResolvedAgentConfig `json:"configs"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify: Should have 1 agent config
	require.Len(t, response.Configs, 1)

	resolved := response.Configs[0]
	assert.Equal(t, agent.ID, uuid.UUID(resolved.AgentId))
	assert.True(t, resolved.IsEnabled)

	// Verify config values (all from org)
	assert.Equal(t, "claude-3-5-sonnet-20241022", resolved.Config["model"])
	assert.Equal(t, float64(0.2), resolved.Config["temperature"])
	assert.Equal(t, float64(4096), resolved.Config["max_tokens"])

	t.Logf("✅ Org-only config test passed")
	t.Logf("   Config: %+v", resolved.Config)
}

// TestConfigCascade_OrgAndTeam tests org + team configuration cascade
func TestConfigCascade_OrgAndTeam(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Frontend Team")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		TeamID:   team.ID,
		Email:    "jane@example.com",
		FullName: "Jane Smith",
		Status:   "active",
	})

	// Get an agent
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)
	agent := agents[0]

	// Create org-level config (base)
	_, err = queries.CreateOrgAgentConfig(ctx, db.CreateOrgAgentConfigParams{
		OrgID:   org.ID,
		AgentID: agent.ID,
		Config: []byte(`{
			"model": "claude-3-5-sonnet-20241022",
			"temperature": 0.2,
			"max_tokens": 4096,
			"rate_limit_per_hour": 100
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Create team-level config (override temperature and max_tokens)
	_, err = queries.CreateTeamAgentConfig(ctx, db.CreateTeamAgentConfigParams{
		TeamID:  team.ID,
		AgentID: agent.ID,
		ConfigOverride: []byte(`{
			"temperature": 0.5,
			"max_tokens": 8192
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Setup router
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	handler := handlers.NewOrgAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs/resolved", handler.GetEmployeeResolvedAgentConfigs)

	// Test: Get resolved config
	url := fmt.Sprintf("/employees/%s/agent-configs/resolved", employee.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Configs []api.ResolvedAgentConfig `json:"configs"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify: Should have 1 agent config
	require.Len(t, response.Configs, 1)

	resolved := response.Configs[0]
	assert.True(t, resolved.IsEnabled)

	// Verify config cascade: org → team
	assert.Equal(t, "claude-3-5-sonnet-20241022", resolved.Config["model"]) // from org
	assert.Equal(t, float64(0.5), resolved.Config["temperature"])           // from team (overridden!)
	assert.Equal(t, float64(8192), resolved.Config["max_tokens"])           // from team (overridden!)
	assert.Equal(t, float64(100), resolved.Config["rate_limit_per_hour"])   // from org (inherited)

	t.Logf("✅ Org + Team cascade test passed")
	t.Logf("   - model: %v (from org)", resolved.Config["model"])
	t.Logf("   - temperature: %v (from team) ⬆️", resolved.Config["temperature"])
	t.Logf("   - max_tokens: %v (from team) ⬆️", resolved.Config["max_tokens"])
	t.Logf("   - rate_limit_per_hour: %v (from org)", resolved.Config["rate_limit_per_hour"])
}

// TestConfigCascade_FullHierarchy tests the complete org → team → employee cascade
func TestConfigCascade_FullHierarchy(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Senior Engineers")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		TeamID:   team.ID,
		Email:    "senior@example.com",
		FullName: "Senior Dev",
		Status:   "active",
	})

	// Get an agent
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents)
	agent := agents[0]

	// Level 1: Org config (foundation)
	_, err = queries.CreateOrgAgentConfig(ctx, db.CreateOrgAgentConfigParams{
		OrgID:   org.ID,
		AgentID: agent.ID,
		Config: []byte(`{
			"model": "claude-3-5-sonnet-20241022",
			"temperature": 0.2,
			"max_tokens": 4096,
			"rate_limit_per_hour": 100,
			"cost_limit_daily_usd": 50.0
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Level 2: Team config (team overrides)
	_, err = queries.CreateTeamAgentConfig(ctx, db.CreateTeamAgentConfigParams{
		TeamID:  team.ID,
		AgentID: agent.ID,
		ConfigOverride: []byte(`{
			"temperature": 0.5,
			"max_tokens": 8192,
			"cost_limit_daily_usd": 75.0
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Level 3: Employee config (personal overrides)
	_, err = queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID: employee.ID,
		AgentID:    agent.ID,
		ConfigOverride: []byte(`{
			"max_tokens": 16384
		}`),
		IsEnabled: true,
	})
	require.NoError(t, err)

	// Setup router
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	handler := handlers.NewOrgAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs/resolved", handler.GetEmployeeResolvedAgentConfigs)

	// Test: Get resolved config
	url := fmt.Sprintf("/employees/%s/agent-configs/resolved", employee.ID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response struct {
		Configs []api.ResolvedAgentConfig `json:"configs"`
	}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify: Should have 1 agent config
	require.Len(t, response.Configs, 1)

	resolved := response.Configs[0]
	assert.True(t, resolved.IsEnabled)

	// Verify FULL cascade: org → team → employee
	assert.Equal(t, "claude-3-5-sonnet-20241022", resolved.Config["model"]) // from org (inherited)
	assert.Equal(t, float64(0.5), resolved.Config["temperature"])           // from team (team override)
	assert.Equal(t, float64(16384), resolved.Config["max_tokens"])          // from employee (final override!)
	assert.Equal(t, float64(100), resolved.Config["rate_limit_per_hour"])   // from org (inherited)
	assert.Equal(t, float64(75.0), resolved.Config["cost_limit_daily_usd"]) // from team (team override)

	t.Logf("✅ Full cascade (org → team → employee) test passed")
	t.Logf("   📋 Resolved config:")
	t.Logf("      - model: %v (from org)", resolved.Config["model"])
	t.Logf("      - temperature: %v (from team)", resolved.Config["temperature"])
	t.Logf("      - max_tokens: %v (from employee) ⭐", resolved.Config["max_tokens"])
	t.Logf("      - rate_limit_per_hour: %v (from org)", resolved.Config["rate_limit_per_hour"])
	t.Logf("      - cost_limit_daily_usd: %v (from team)", resolved.Config["cost_limit_daily_usd"])
}
