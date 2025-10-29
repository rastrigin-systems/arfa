package cli

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformClient_Login(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/auth/login", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Decode request body
		var reqBody LoginRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)

		assert.Equal(t, "test@example.com", reqBody.Email)
		assert.Equal(t, "password123", reqBody.Password)

		// Return success response
		resp := LoginResponse{
			Token:     "test-token-abc123",
			ExpiresAt: "2024-12-31T23:59:59Z",
			Employee: LoginEmployeeInfo{
				ID:    "emp-123",
				OrgID: "org-456",
				Email: "test@example.com",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	resp, err := client.Login("test@example.com", "password123")

	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-token-abc123", resp.Token)
	assert.Equal(t, "emp-123", resp.Employee.ID)
	assert.Equal(t, "org-456", resp.Employee.OrgID)
	assert.Equal(t, "test@example.com", resp.Employee.Email)

	// Verify token was stored in client
	assert.Equal(t, "test-token-abc123", client.token)
}

func TestPlatformClient_Login_InvalidCredentials(t *testing.T) {
	// Setup mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid credentials"}`))
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	resp, err := client.Login("test@example.com", "wrong-password")

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "login failed")
}

func TestPlatformClient_GetEmployeeInfo(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/employees/emp-123", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Return employee info
		resp := EmployeeInfo{
			ID:       "emp-123",
			Email:    "alice@example.com",
			FullName: "Alice Smith",
			OrgID:    "org-456",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	info, err := client.GetEmployeeInfo("emp-123")

	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "emp-123", info.ID)
	assert.Equal(t, "alice@example.com", info.Email)
	assert.Equal(t, "Alice Smith", info.FullName)
	assert.Equal(t, "org-456", info.OrgID)
}

func TestPlatformClient_GetEmployeeInfo_NotFound(t *testing.T) {
	// Setup mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Employee not found"}`))
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	info, err := client.GetEmployeeInfo("emp-999")

	require.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "failed to get employee info")
}

func TestPlatformClient_GetResolvedAgentConfigs(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/employees/emp-123/agent-configs/resolved", r.URL.Path)
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		// Return resolved configs
		resp := ResolvedConfigsResponse{
			Configs: []AgentConfigAPIResponse{
				{
					AgentID:      "agent-1",
					AgentName:    "Claude Code",
					AgentType:    "claude-code",
					IsEnabled:    true,
					Config:       map[string]interface{}{"model": "claude-3-5-sonnet"},
					Provider:     "anthropic",
					SyncToken:    "sync-token-1",
					SystemPrompt: "You are a helpful coding assistant",
					LastSyncedAt: nil,
				},
				{
					AgentID:      "agent-2",
					AgentName:    "Cursor",
					AgentType:    "cursor",
					IsEnabled:    false,
					Config:       map[string]interface{}{"theme": "dark"},
					Provider:     "cursor",
					SyncToken:    "sync-token-2",
					SystemPrompt: "",
					LastSyncedAt: nil,
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

	configs, err := client.GetResolvedAgentConfigs("emp-123")

	require.NoError(t, err)
	assert.Len(t, configs, 2)

	// Check first config
	assert.Equal(t, "agent-1", configs[0].AgentID)
	assert.Equal(t, "Claude Code", configs[0].AgentName)
	assert.Equal(t, "claude-code", configs[0].AgentType)
	assert.True(t, configs[0].IsEnabled)
	assert.Equal(t, "claude-3-5-sonnet", configs[0].Configuration["model"])

	// Check second config
	assert.Equal(t, "agent-2", configs[1].AgentID)
	assert.Equal(t, "Cursor", configs[1].AgentName)
	assert.False(t, configs[1].IsEnabled)
}

func TestPlatformClient_GetResolvedAgentConfigs_Empty(t *testing.T) {
	// Setup mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ResolvedConfigsResponse{
			Configs: []AgentConfigAPIResponse{},
			Total:   0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewPlatformClient(server.URL)
	client.SetToken("test-token")

	configs, err := client.GetResolvedAgentConfigs("emp-123")

	require.NoError(t, err)
	assert.Len(t, configs, 0)
}
