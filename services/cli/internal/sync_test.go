package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncService_SaveAndGetLocalAgentConfigs(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewConfigManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := NewAPIClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Create test agent configs
	configs := []AgentConfig{
		{
			AgentID:   "agent-1",
			AgentName: "Claude Code",
			AgentType: "claude-code",
			IsEnabled: true,
			Configuration: map[string]interface{}{
				"model": "claude-3-5-sonnet-20241022",
			},
			MCPServers: []MCPServerConfig{
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
			MCPServers: []MCPServerConfig{},
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
		var expected *AgentConfig
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

func TestSyncService_GetAgentConfig(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewConfigManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := NewAPIClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create test agent configs
	configs := []AgentConfig{
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

func TestSyncService_GetLocalAgentConfigs_EmptyDirectory(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewConfigManagerWithPath(filepath.Join(tempDir, "config.json"))
	pc := NewAPIClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Get configs when directory doesn't exist
	configs, err := syncService.GetLocalAgentConfigs()
	require.NoError(t, err)
	assert.Empty(t, configs)
}
