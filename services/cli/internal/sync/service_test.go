package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SaveAndGetLocalAgentConfigs(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Create test agent configs
	configs := []api.AgentConfig{
		{
			AgentID:   "agent-1",
			AgentName: "Claude Code",
			AgentType: "claude-code",
			IsEnabled: true,
			Configuration: map[string]interface{}{
				"model": "claude-3-5-sonnet-20241022",
			},
			MCPServers: []api.MCPServerConfig{
				{
					ServerID:   "mcp-1",
					ServerName: "Filesystem",
					ServerType: "filesystem",
					IsEnabled:  true,
					Config: map[string]interface{}{
						"root": "/workspace",
					},
				},
			},
		},
		{
			AgentID:   "agent-2",
			AgentName: "Aider",
			AgentType: "aider",
			IsEnabled: false,
			Configuration: map[string]interface{}{
				"version": "0.15.0",
			},
			MCPServers: []api.MCPServerConfig{},
		},
	}

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save configs
	err := syncService.saveAgentConfigs(configs)
	require.NoError(t, err)

	// Verify files were created
	agentsDir := filepath.Join(tempDir, ".ubik", "config", "agents")
	_, err = os.Stat(agentsDir)
	require.NoError(t, err)

	// Verify agent-1 config files
	agent1Dir := filepath.Join(agentsDir, "agent-1")
	_, err = os.Stat(filepath.Join(agent1Dir, "config.json"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(agent1Dir, "mcp-servers.json"))
	require.NoError(t, err)

	// Verify agent-2 config file
	agent2Dir := filepath.Join(agentsDir, "agent-2")
	_, err = os.Stat(filepath.Join(agent2Dir, "config.json"))
	require.NoError(t, err)

	// Get local configs
	loadedConfigs, err := syncService.GetLocalAgentConfigs()
	require.NoError(t, err)
	assert.Len(t, loadedConfigs, 2)

	// Verify loaded configs match
	for _, loaded := range loadedConfigs {
		var expected *api.AgentConfig
		for _, ec := range configs {
			if ec.AgentID == loaded.AgentID {
				expected = &ec
				break
			}
		}
		require.NotNil(t, expected)
		assert.Equal(t, expected.AgentName, loaded.AgentName)
		assert.Equal(t, expected.AgentType, loaded.AgentType)
		assert.Equal(t, expected.IsEnabled, loaded.IsEnabled)
	}
}

func TestService_GetAgentConfig(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create test agent configs
	configs := []api.AgentConfig{
		{
			AgentID:   "agent-1",
			AgentName: "Claude Code",
			AgentType: "claude-code",
			IsEnabled: true,
		},
		{
			AgentID:   "agent-2",
			AgentName: "Aider",
			AgentType: "aider",
			IsEnabled: false,
		},
	}

	// Save configs
	err := syncService.saveAgentConfigs(configs)
	require.NoError(t, err)

	// Test getting by ID
	config, err := syncService.GetAgentConfig("agent-1")
	require.NoError(t, err)
	assert.Equal(t, "agent-1", config.AgentID)
	assert.Equal(t, "Claude Code", config.AgentName)

	// Test getting by name
	config, err = syncService.GetAgentConfig("Aider")
	require.NoError(t, err)
	assert.Equal(t, "agent-2", config.AgentID)
	assert.Equal(t, "Aider", config.AgentName)

	// Test getting non-existent
	config, err = syncService.GetAgentConfig("non-existent")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "not found")
}

func TestService_GetLocalAgentConfigs_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Get configs when directory doesn't exist
	configs, err := syncService.GetLocalAgentConfigs()
	require.NoError(t, err)
	assert.Empty(t, configs)
}

func TestService_SaveAndGetLocalToolPolicies(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create test tool policies
	reason1 := "Shell commands are blocked for security"
	reason2 := "File writes need review"
	orgScope := api.ToolPolicyScopeOrganization
	teamScope := api.ToolPolicyScopeTeam

	policiesResp := &api.EmployeeToolPoliciesResponse{
		Policies: []api.ToolPolicy{
			{
				ID:       "policy-1",
				ToolName: "Bash",
				Action:   api.ToolPolicyActionDeny,
				Reason:   &reason1,
				Scope:    &orgScope,
			},
			{
				ID:       "policy-2",
				ToolName: "Write",
				Action:   api.ToolPolicyActionAudit,
				Reason:   &reason2,
				Scope:    &teamScope,
			},
		},
		Version:  12345,
		SyncedAt: "2024-01-15T10:30:00Z",
	}

	// Save policies
	err := syncService.saveToolPolicies(policiesResp)
	require.NoError(t, err)

	// Verify file was created
	policiesPath := filepath.Join(tempDir, ".ubik", "policies.json")
	_, err = os.Stat(policiesPath)
	require.NoError(t, err)

	// Get local policies
	loadedPolicies, err := syncService.GetLocalToolPolicies()
	require.NoError(t, err)
	assert.Len(t, loadedPolicies, 2)

	// Verify loaded policies match
	assert.Equal(t, "Bash", loadedPolicies[0].ToolName)
	assert.Equal(t, api.ToolPolicyActionDeny, loadedPolicies[0].Action)
	assert.Equal(t, "Shell commands are blocked for security", *loadedPolicies[0].Reason)

	assert.Equal(t, "Write", loadedPolicies[1].ToolName)
	assert.Equal(t, api.ToolPolicyActionAudit, loadedPolicies[1].Action)
}

func TestService_GetLocalToolPolicies_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Get policies when file doesn't exist
	policies, err := syncService.GetLocalToolPolicies()
	require.NoError(t, err)
	assert.Empty(t, policies)
}

func TestService_GetLocalToolPolicies_WithConditions(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create test policy with conditions
	employeeScope := api.ToolPolicyScopeEmployee
	reason := "Dangerous commands blocked"
	policiesResp := &api.EmployeeToolPoliciesResponse{
		Policies: []api.ToolPolicy{
			{
				ID:       "policy-1",
				ToolName: "Bash",
				Action:   api.ToolPolicyActionDeny,
				Reason:   &reason,
				Scope:    &employeeScope,
				Conditions: map[string]interface{}{
					"command": map[string]interface{}{
						"pattern": "rm -rf.*",
					},
				},
			},
		},
		Version:  67890,
		SyncedAt: "2024-01-15T11:00:00Z",
	}

	// Save policies
	err := syncService.saveToolPolicies(policiesResp)
	require.NoError(t, err)

	// Get local policies
	loadedPolicies, err := syncService.GetLocalToolPolicies()
	require.NoError(t, err)
	assert.Len(t, loadedPolicies, 1)

	// Verify conditions were preserved
	policy := loadedPolicies[0]
	assert.NotNil(t, policy.Conditions)
	assert.NotNil(t, policy.Conditions["command"])
}
