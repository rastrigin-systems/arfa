package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentService_ListAgents(t *testing.T) {
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

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	svc := NewAgentService(client, nil)
	agents, err := svc.ListAgents()

	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, "Claude Code", agents[0].Name)
	assert.Equal(t, "Cursor", agents[1].Name)
}

func TestAgentService_GetAgent(t *testing.T) {
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

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	svc := NewAgentService(client, nil)
	agent, err := svc.GetAgent("agent-1")

	require.NoError(t, err)
	assert.Equal(t, "Claude Code", agent.Name)
	assert.Equal(t, "anthropic", agent.Provider)
	assert.Equal(t, "enterprise", agent.PricingTier)
}

func TestAgentService_ListEmployeeAgentConfigs(t *testing.T) {
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

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	svc := NewAgentService(client, nil)
	configs, err := svc.ListEmployeeAgentConfigs("emp-1")

	require.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.Equal(t, "Claude Code", configs[0].AgentName)
	assert.True(t, configs[0].IsEnabled)
}

func TestAgentService_RequestAgent(t *testing.T) {
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

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	svc := NewAgentService(client, nil)
	err := svc.RequestAgent("emp-1", "agent-1")

	require.NoError(t, err)
}

func TestAgentService_CheckForUpdates(t *testing.T) {
	// Create temp config directory
	tmpDir, err := os.MkdirTemp("", "ubik-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create local config with one agent
	localConfig := LocalConfig{
		PlatformURL:  "http://localhost:3001",
		EmployeeID:   "emp-1",
		Token:        "test-token",
		TokenExpires: "2025-12-31T00:00:00Z",
		Agents: []AgentConfig{
			{
				AgentID:   "agent-1",
				AgentName: "Claude Code",
			},
		},
	}

	configPath := filepath.Join(tmpDir, "config.json")
	data, err := json.MarshalIndent(localConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Setup mock server with 2 agents (local has 1, so updates available)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/emp-1/agent-configs/resolved" {
			resp := ResolvedConfigsResponse{
				Configs: []AgentConfigAPIResponse{
					{AgentID: "agent-1", AgentName: "Claude Code"},
					{AgentID: "agent-2", AgentName: "Cursor"}, // New agent!
				},
				Total: 2,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	configManager := &ConfigManager{configPath: configPath}
	svc := NewAgentService(client, configManager)

	hasUpdates, err := svc.CheckForUpdates("emp-1")

	require.NoError(t, err)
	assert.True(t, hasUpdates, "Should detect updates (2 remote vs 1 local)")
}

func TestAgentService_CheckForUpdates_NoUpdates(t *testing.T) {
	// Create temp config directory
	tmpDir, err := os.MkdirTemp("", "ubik-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create local config with one agent
	localConfig := LocalConfig{
		PlatformURL:  "http://localhost:3001",
		EmployeeID:   "emp-1",
		Token:        "test-token",
		TokenExpires: "2025-12-31T00:00:00Z",
		Agents: []AgentConfig{
			{AgentID: "agent-1", AgentName: "Claude Code"},
		},
	}

	configPath := filepath.Join(tmpDir, "config.json")
	data, err := json.MarshalIndent(localConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	// Setup mock server with same agents
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/emp-1/agent-configs/resolved" {
			resp := ResolvedConfigsResponse{
				Configs: []AgentConfigAPIResponse{
					{AgentID: "agent-1", AgentName: "Claude Code"},
				},
				Total: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	configManager := &ConfigManager{configPath: configPath}
	svc := NewAgentService(client, configManager)

	hasUpdates, err := svc.CheckForUpdates("emp-1")

	require.NoError(t, err)
	assert.False(t, hasUpdates, "Should not detect updates when configs match")
}

func TestAgentService_GetLocalAgents(t *testing.T) {
	// Create temp config directory
	tmpDir, err := os.MkdirTemp("", "ubik-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create local config
	localConfig := LocalConfig{
		PlatformURL:  "http://localhost:3001",
		EmployeeID:   "emp-1",
		Token:        "test-token",
		TokenExpires: "2025-12-31T00:00:00Z",
		Agents: []AgentConfig{
			{AgentID: "agent-1", AgentName: "Claude Code"},
			{AgentID: "agent-2", AgentName: "Cursor"},
		},
	}

	configPath := filepath.Join(tmpDir, "config.json")
	data, err := json.MarshalIndent(localConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0644)
	require.NoError(t, err)

	configManager := &ConfigManager{configPath: configPath}
	svc := NewAgentService(nil, configManager)

	agents, err := svc.GetLocalAgents()

	require.NoError(t, err)
	assert.Len(t, agents, 2)
	assert.Equal(t, "Claude Code", agents[0].AgentName)
	assert.Equal(t, "Cursor", agents[1].AgentName)
}
