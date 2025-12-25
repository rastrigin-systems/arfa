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

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

// ============================================================================
// List Employee Agent Configs Integration Tests
// ============================================================================

func TestListEmployeeAgentConfigs_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	// Create admin employee for authentication
	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agents from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")

	// Create employee agent configs
	config1JSON := []byte(`{"max_tokens":4000}`)
	syncToken := "sync-token-123"
	_, err = queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agents[0].ID,
		ConfigOverride: config1JSON,
		IsEnabled:      true,
	})
	require.NoError(t, err)

	config2JSON := []byte(`{"temperature":0.7}`)
	_, err = queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agents[1].ID,
		ConfigOverride: config2JSON,
		IsEnabled:      false,
	})
	require.NoError(t, err)

	_ = syncToken // Use it later if needed

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/"+targetEmployee.ID.String()+"/agent-configs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeAgentConfigsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return 2 configs
	assert.Equal(t, 2, response.Total)
	require.Len(t, response.Configs, 2)

	// Verify config data
	assert.NotNil(t, response.Configs[0].AgentName)
	assert.NotNil(t, response.Configs[0].ConfigOverride)
}

func TestListEmployeeAgentConfigs_Integration_EmployeeNotFound(t *testing.T) {
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
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	// Make request with non-existent employee ID
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/employees/"+nonExistentID.String()+"/agent-configs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListEmployeeAgentConfigs_Integration_OrgIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	// Create employee in org1
	employee1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "admin@org1.com",
		FullName: "Admin Org1",
		Status:   "active",
	})

	// Create employee in org2
	employee2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "admin@org2.com",
		FullName: "Admin Org2",
		Status:   "active",
	})

	// Authenticate as employee from org1
	token, _ := auth.GenerateJWT(employee1.ID, org1.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs", handler.ListEmployeeAgentConfigs)

	// Try to access employee from org2 (should be denied)
	req := httptest.NewRequest(http.MethodGet, "/employees/"+employee2.ID.String()+"/agent-configs", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// Create Employee Agent Config Integration Tests
// ============================================================================

func TestCreateEmployeeAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees/{employee_id}/agent-configs", handler.CreateEmployeeAgentConfig)

	// Create request
	isEnabled := true
	reqBody := api.CreateEmployeeAgentConfigRequest{
		AgentId: api.EmployeeId(agent.ID),
		ConfigOverride: map[string]interface{}{
			"max_tokens":  float64(4000),
			"temperature": 0.7,
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/employees/"+targetEmployee.ID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.EmployeeAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.NotNil(t, response.AgentName)
	assert.True(t, response.IsEnabled)
	assert.NotNil(t, response.ConfigOverride)
	assert.Equal(t, float64(4000), response.ConfigOverride["max_tokens"])
}

func TestCreateEmployeeAgentConfig_Integration_DuplicateAgent(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create first config
	configJSON := []byte(`{"max_tokens":4000}`)
	_, err = queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agent.ID,
		ConfigOverride: configJSON,
		IsEnabled:      true,
	})
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees/{employee_id}/agent-configs", handler.CreateEmployeeAgentConfig)

	// Try to create duplicate config
	isEnabled := true
	reqBody := api.CreateEmployeeAgentConfigRequest{
		AgentId: api.EmployeeId(agent.ID),
		ConfigOverride: map[string]interface{}{
			"max_tokens": float64(8000),
		},
		IsEnabled: &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/employees/"+targetEmployee.ID.String()+"/agent-configs", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

// ============================================================================
// Get Employee Agent Config Integration Tests
// ============================================================================

func TestGetEmployeeAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create employee agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agent.ID,
		ConfigOverride: configJSON,
		IsEnabled:      true,
	})
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}/agent-configs/{config_id}", handler.GetEmployeeAgentConfig)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/employees/"+targetEmployee.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotNil(t, response.Id)
	assert.NotNil(t, response.AgentName)
	assert.True(t, response.IsEnabled)
}

// ============================================================================
// Update Employee Agent Config Integration Tests
// ============================================================================

func TestUpdateEmployeeAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create employee agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agent.ID,
		ConfigOverride: configJSON,
		IsEnabled:      true,
	})
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/employees/{employee_id}/agent-configs/{config_id}", handler.UpdateEmployeeAgentConfig)

	// Update config
	isEnabled := false
	newConfig := map[string]interface{}{
		"max_tokens":  float64(8000),
		"temperature": 0.9,
	}
	reqBody := api.UpdateEmployeeAgentConfigRequest{
		ConfigOverride: &newConfig,
		IsEnabled:      &isEnabled,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+targetEmployee.ID.String()+"/agent-configs/"+config.ID.String(), strings.NewReader(string(bodyBytes)))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.EmployeeAgentConfig
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.IsEnabled)
	assert.Equal(t, float64(8000), response.ConfigOverride["max_tokens"])
}

// ============================================================================
// Delete Employee Agent Config Integration Tests
// ============================================================================

func TestDeleteEmployeeAgentConfig_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	targetEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "developer@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	adminEmployee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Get agent from seed data
	agents, err := queries.ListAgents(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, agents, "Need agents in seed data")
	agent := agents[0]

	// Create employee agent config
	configJSON := []byte(`{"max_tokens":4000}`)
	config, err := queries.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     targetEmployee.ID,
		AgentID:        agent.ID,
		ConfigOverride: configJSON,
		IsEnabled:      true,
	})
	require.NoError(t, err)

	// Create session for authentication
	token, _ := auth.GenerateJWT(adminEmployee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, testutil.CreateSessionParams(adminEmployee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeeAgentConfigsHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Delete("/employees/{employee_id}/agent-configs/{config_id}", handler.DeleteEmployeeAgentConfig)

	// Delete config
	req := httptest.NewRequest(http.MethodDelete, "/employees/"+targetEmployee.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify config is deleted
	router2 := chi.NewRouter()
	router2.Use(middleware.JWTAuth(queries))
	router2.Get("/employees/{employee_id}/agent-configs/{config_id}", handler.GetEmployeeAgentConfig)

	req2 := httptest.NewRequest(http.MethodGet, "/employees/"+targetEmployee.ID.String()+"/agent-configs/"+config.ID.String(), nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()

	router2.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusNotFound, rec2.Code)
}
