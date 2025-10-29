package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListAgents_Integration_WithSeedData tests listing agents with real database
func TestListAgents_Integration_WithSeedData(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee for authentication
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Create test agents in the catalog with unique names
	timestamp := time.Now().UnixNano()
	agent1Name := testutil.GenerateUniqueName("Claude Code", timestamp)
	agent2Name := testutil.GenerateUniqueName("Cursor", timestamp+1)
	privateAgentName := testutil.GenerateUniqueName("Private Agent", timestamp+2)

	agent1, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		Name:          agent1Name,
		Type:          "claude-code",
		Description:   "AI-powered code assistant",
		Provider:      "anthropic",
		DefaultConfig: []byte(`{"model":"claude-3-5-sonnet-20241022"}`),
		Capabilities:  []byte(`["code_generation","code_review"]`),
		LlmProvider:   "anthropic",
		LlmModel:      "claude-3-5-sonnet-20241022",
		IsPublic:      true,
	})
	require.NoError(t, err)

	agent2, err := queries.CreateAgent(ctx, db.CreateAgentParams{
		Name:          agent2Name,
		Type:          "cursor",
		Description:   "AI code editor",
		Provider:      "openai",
		DefaultConfig: []byte(`{"model":"gpt-4"}`),
		Capabilities:  []byte(`["inline_suggestions"]`),
		LlmProvider:   "openai",
		LlmModel:      "gpt-4",
		IsPublic:      true,
	})
	require.NoError(t, err)

	// Create a non-public agent (should not be returned)
	_, err = queries.CreateAgent(ctx, db.CreateAgentParams{
		Name:          privateAgentName,
		Type:          "private",
		Description:   "Private agent",
		Provider:      "custom",
		DefaultConfig: []byte(`{}`),
		Capabilities:  []byte(`[]`),
		LlmProvider:   "custom",
		LlmModel:      "custom-model",
		IsPublic:      false, // Not public
	})
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewAgentsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/agents", handler.ListAgents)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListAgentsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return public agents (including our test agents + any seed data)
	assert.Greater(t, response.Total, 0, "Should have at least some agents")
	assert.Greater(t, len(response.Agents), 0, "Should have at least some agents")

	// Find our test agents in the response (they're sorted alphabetically)
	var foundAgent1, foundAgent2, foundPrivateAgent bool
	for _, agent := range response.Agents {
		if agent.Id.String() == agent1.ID.String() {
			foundAgent1 = true
			assert.Equal(t, agent1Name, agent.Name)
			assert.Equal(t, "claude-code", agent.Type)
			assert.True(t, agent.IsPublic)

			// Verify capabilities and default_config are properly deserialized
			assert.NotNil(t, agent.Capabilities)
			assert.Contains(t, *agent.Capabilities, "code_generation")

			assert.NotNil(t, agent.DefaultConfig)
			config := *agent.DefaultConfig
			assert.Equal(t, "claude-3-5-sonnet-20241022", config["model"])
		}
		if agent.Id.String() == agent2.ID.String() {
			foundAgent2 = true
			assert.Equal(t, agent2Name, agent.Name)
			assert.Equal(t, "cursor", agent.Type)
			assert.True(t, agent.IsPublic)
		}
		// Check that private agent is NOT in the response
		if agent.Name == privateAgentName {
			foundPrivateAgent = true
		}
	}

	assert.True(t, foundAgent1, "Public Agent 1 should be in response")
	assert.True(t, foundAgent2, "Public Agent 2 should be in response")
	assert.False(t, foundPrivateAgent, "Private agent should NOT be in response")
}

// TestListAgents_Integration_Authentication tests that authentication is required
func TestListAgents_Integration_Authentication(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	// Setup handler with middleware (no authentication)
	handler := handlers.NewAgentsHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/agents", handler.ListAgents)

	// Make request without authentication
	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Verify we got an error response (body is not empty)
	assert.NotEmpty(t, rec.Body.String(), "Should return error response")
}
