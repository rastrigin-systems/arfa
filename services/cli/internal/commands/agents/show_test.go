package agents

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestEnvironment creates a temporary config directory and config file for testing
func setupTestEnvironment(t *testing.T, mockServerURL string) string {
	tempDir := t.TempDir()

	// Create .ubik directory
	ubikDir := filepath.Join(tempDir, ".ubik")
	err := os.MkdirAll(ubikDir, 0700)
	require.NoError(t, err)

	// Create config file with mock server URL and token
	configPath := filepath.Join(ubikDir, "config.json")
	config := map[string]interface{}{
		"platform_url": mockServerURL,
		"token":        "test-token",
		"employee_id":  "employee-uuid",
		"org_id":       "org-uuid",
	}
	configData, err := json.MarshalIndent(config, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, configData, 0600)
	require.NoError(t, err)

	// Set HOME to temp directory
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	t.Cleanup(func() {
		os.Setenv("HOME", originalHome)
	})

	return tempDir
}

// TestAgentsShowCommand_Success tests showing agent config cascade
func TestAgentsShowCommand_Success(t *testing.T) {
	// Setup mock API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/me":
			// Return current user
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      "employee-uuid",
				"email":   "john@example.com",
				"org_id":  "org-uuid",
				"team_id": "team-uuid",
			})

		case "/api/v1/organizations/current/agent-configs":
			// Return org-level configs
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"id":         "org-config-uuid",
						"agent_id":   "claude-code-uuid",
						"agent_name": "Claude Code",
						"config": map[string]interface{}{
							"model":                "claude-3-5-sonnet-20241022",
							"temperature":          0.2,
							"max_tokens":           4096,
							"rate_limit_per_hour":  100,
							"cost_limit_daily_usd": 50.0,
						},
						"is_enabled": true,
					},
				},
			})

		case "/api/v1/teams/team-uuid/agent-configs":
			// Return team-level configs
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"id":       "team-config-uuid",
						"agent_id": "claude-code-uuid",
						"config_override": map[string]interface{}{
							"temperature":          0.5,
							"max_tokens":           8192,
							"cost_limit_daily_usd": 75.0,
						},
						"is_enabled": true,
					},
				},
			})

		case "/api/v1/employees/employee-uuid/agent-configs":
			// Return employee-level configs
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"id":       "employee-config-uuid",
						"agent_id": "claude-code-uuid",
						"config_override": map[string]interface{}{
							"max_tokens": 16384,
						},
						"is_enabled": true,
					},
				},
			})

		case "/api/v1/employees/employee-uuid/agent-configs/resolved":
			// Return resolved config
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"agent_id":   "claude-code-uuid",
						"agent_name": "Claude Code",
						"agent_type": "claude-code",
						"provider":   "anthropic",
						"config": map[string]interface{}{
							"model":                "claude-3-5-sonnet-20241022",
							"temperature":          0.5,
							"max_tokens":           float64(16384),
							"rate_limit_per_hour":  100,
							"cost_limit_daily_usd": 75.0,
						},
						"is_enabled": true,
					},
				},
				"total": 1,
			})

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	// Setup test environment with mock server URL
	setupTestEnvironment(t, mockServer.URL)

	// Create command
	cmd := NewShowCommand()
	cmd.SetArgs([]string{"Claude Code"})

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Execute
	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify output contains key information
	assert.Contains(t, output, "Claude Code")
	assert.Contains(t, output, "Organization")
	assert.Contains(t, output, "Team")
	assert.Contains(t, output, "Personal")
	assert.Contains(t, output, "Final Resolved Configuration")
	assert.Contains(t, output, "model")
	assert.Contains(t, output, "temperature")
	assert.Contains(t, output, "max_tokens")
}

// TestAgentsShowCommand_JSONOutput tests JSON format output
func TestAgentsShowCommand_JSONOutput(t *testing.T) {
	// Setup mock API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/me":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      "employee-uuid",
				"email":   "john@example.com",
				"org_id":  "org-uuid",
				"team_id": "team-uuid",
			})
		case "/api/v1/organizations/current/agent-configs":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"agent_id": "claude-code-uuid",
						"config": map[string]interface{}{
							"model": "claude-3-5-sonnet-20241022",
						},
						"is_enabled": true,
					},
				},
			})
		case "/api/v1/teams/team-uuid/agent-configs":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{},
			})
		case "/api/v1/employees/employee-uuid/agent-configs":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{},
			})
		case "/api/v1/employees/employee-uuid/agent-configs/resolved":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{
					{
						"agent_id":   "claude-code-uuid",
						"agent_name": "Claude Code",
						"agent_type": "claude-code",
						"provider":   "anthropic",
						"config": map[string]interface{}{
							"max_tokens": 16384,
						},
						"is_enabled": true,
					},
				},
				"total": 1,
			})
		}
	}))
	defer mockServer.Close()

	// Setup test environment
	setupTestEnvironment(t, mockServer.URL)

	// Create command with --json flag
	cmd := NewShowCommand()
	cmd.SetArgs([]string{"Claude Code", "--json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify JSON output
	var result map[string]interface{}
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err, "Output should be valid JSON")

	// Verify structure
	assert.Contains(t, result, "agent")
	assert.Contains(t, result, "org_config")
	assert.Contains(t, result, "resolved_config")
}

// TestAgentsShowCommand_AgentNotFound tests error when agent doesn't exist
func TestAgentsShowCommand_AgentNotFound(t *testing.T) {
	// Setup mock API server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth/me":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":      "employee-uuid",
				"email":   "john@example.com",
				"org_id":  "org-uuid",
				"team_id": nil,
			})
		case "/api/v1/employees/employee-uuid/agent-configs/resolved":
			// Return empty list - no agents found
			json.NewEncoder(w).Encode(map[string]interface{}{
				"configs": []map[string]interface{}{},
				"total":   0,
			})
		}
	}))
	defer mockServer.Close()

	// Setup test environment
	setupTestEnvironment(t, mockServer.URL)

	cmd := NewShowCommand()
	cmd.SetArgs([]string{"NonExistentAgent"})

	var buf bytes.Buffer
	cmd.SetErr(&buf)

	err := cmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestAgentsShowCommand_NoAgentSpecified tests error when no agent name provided
func TestAgentsShowCommand_NoAgentSpecified(t *testing.T) {
	cmd := NewShowCommand()
	cmd.SetArgs([]string{}) // No agent name

	err := cmd.Execute()
	assert.Error(t, err)
	// Cobra's actual error message format
	assert.Contains(t, err.Error(), "accepts 1 arg")
}

// MockConfig for testing
type MockConfig struct {
	APIBaseURL string
	Token      string
}
