package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_ListAgents(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/agents", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := ListAgentsResponse{
			Agents: []Agent{
				{
					ID:          "agent-1",
					Name:        "Claude Code",
					Provider:    "anthropic",
					Description: "AI coding assistant",
				},
				{
					ID:          "agent-2",
					Name:        "Cursor",
					Provider:    "cursor",
					Description: "AI code editor",
				},
			},
			Total: 2,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")

	svc := NewService(client, nil)
	ctx := context.Background()
	agents, err := svc.ListAgents(ctx)

	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, "Claude Code", agents[0].Name)
	assert.Equal(t, "Cursor", agents[1].Name)
}

func TestService_GetAgent(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/agents/agent-1", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		agent := Agent{
			ID:          "agent-1",
			Name:        "Claude Code",
			Provider:    "anthropic",
			Description: "AI coding assistant",
			PricingTier: "enterprise",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(agent)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")

	svc := NewService(client, nil)
	ctx := context.Background()
	agent, err := svc.GetAgent(ctx, "agent-1")

	require.NoError(t, err)
	assert.Equal(t, "Claude Code", agent.Name)
	assert.Equal(t, "anthropic", agent.Provider)
	assert.Equal(t, "enterprise", agent.PricingTier)
}

func TestService_ListEmployeeAgentConfigs(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/employees/emp-1/agent-configs", r.URL.Path)
		assert.Equal(t, "GET", r.Method)

		resp := ListEmployeeAgentConfigsResponse{
			AgentConfigs: []EmployeeAgentConfig{
				{
					ID:         "config-1",
					EmployeeID: "emp-1",
					AgentID:    "agent-1",
					AgentName:  "Claude Code",
					IsEnabled:  true,
				},
			},
			Total: 1,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")

	svc := NewService(client, nil)
	ctx := context.Background()
	configs, err := svc.ListEmployeeAgentConfigs(ctx, "emp-1")

	require.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, "Claude Code", configs[0].AgentName)
	assert.True(t, configs[0].IsEnabled)
}

func TestService_RequestAgent(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/employees/emp-1/agent-configs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Decode request body
		var reqBody CreateEmployeeAgentConfigRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		assert.Equal(t, "agent-1", reqBody.AgentID)
		assert.True(t, reqBody.IsEnabled)

		// Return success
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")

	svc := NewService(client, nil)
	ctx := context.Background()
	err := svc.RequestAgent(ctx, "emp-1", "agent-1")

	require.NoError(t, err)
}

func TestService_CheckForUpdates(t *testing.T) {
	// Skip this test for now - requires mocking home directory
	// TODO: Implement proper mocking of os.UserHomeDir()
	t.Skip("Requires HOME directory mocking - implement later")
}

func TestService_CheckForUpdates_NoUpdates(t *testing.T) {
	// Skip this test for now - requires mocking home directory
	t.Skip("Requires HOME directory mocking - implement later")
}

func TestService_GetLocalAgents(t *testing.T) {
	// Skip this test for now - requires mocking home directory
	t.Skip("Requires HOME directory mocking - implement later")
}
