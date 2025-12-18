package sync

import (
	"context"
	"os"
	"testing"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_SetDockerClient(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(tempDir + "/config.json")
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Initially no Docker client or container manager
	assert.Nil(t, syncService.dockerClient)
	assert.Nil(t, syncService.containerManager)

	// Set Docker client with nil (unit test)
	syncService.SetDockerClient(nil)
	assert.Nil(t, syncService.dockerClient)
}

func TestService_StartContainers_NoDockerClient(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(tempDir + "/config.json")
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	ctx := context.Background()
	// Try to start containers without Docker client
	err := syncService.StartContainers(ctx, "/tmp", "test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Docker client not configured")
}

func TestService_StopContainers_NoContainerManager(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(tempDir + "/config.json")
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	ctx := context.Background()
	// Try to stop containers without container manager
	err := syncService.StopContainers(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container manager not configured")
}

func TestService_GetContainerStatus_NoContainerManager(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(tempDir + "/config.json")
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	ctx := context.Background()
	// Try to get status without container manager
	status, err := syncService.GetContainerStatus(ctx)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "container manager not configured")
}

func TestService_StartContainers_NoConfigs(t *testing.T) {
	t.Skip("Docker integration test - requires Docker and DockerClient from cli package")
}

func TestService_GetContainerStatus_WithDocker(t *testing.T) {
	t.Skip("Docker integration test - requires Docker and DockerClient from cli package")
}

func TestConvertMCPServers(t *testing.T) {
	configs := []api.MCPServerConfig{
		{
			ServerID:   "mcp-1",
			ServerName: "Filesystem",
			ServerType: "filesystem",
			IsEnabled:  true,
			Config:     map[string]interface{}{"root": "/workspace"},
		},
		{
			ServerID:   "mcp-2",
			ServerName: "Git",
			ServerType: "git",
			IsEnabled:  true,
			Config:     map[string]interface{}{"repo": "/workspace"},
		},
	}

	specs := convertMCPServers(configs)

	assert.Len(t, specs, 2)

	// Check first spec
	assert.Equal(t, "mcp-1", specs[0].ServerID)
	assert.Equal(t, "Filesystem", specs[0].ServerName)
	assert.Equal(t, "filesystem", specs[0].ServerType)
	assert.Equal(t, "ubik/mcp-filesystem:latest", specs[0].Image)
	assert.Equal(t, 8001, specs[0].Port)

	// Check second spec (port should be incremented)
	assert.Equal(t, "mcp-2", specs[1].ServerID)
	assert.Equal(t, "Git", specs[1].ServerName)
	assert.Equal(t, "git", specs[1].ServerType)
	assert.Equal(t, "ubik/mcp-git:latest", specs[1].Image)
	assert.Equal(t, 8002, specs[1].Port)
}

func TestConvertMCPServers_Empty(t *testing.T) {
	configs := []api.MCPServerConfig{}
	specs := convertMCPServers(configs)

	assert.Empty(t, specs)
	assert.NotNil(t, specs)
}

// Integration test: Full lifecycle with Docker
func TestService_FullLifecycle_Integration(t *testing.T) {
	t.Skip("Docker integration test - requires Docker and DockerClient from cli package")
}

// TestService_SaveAgentConfigs_CleanupOld tests that old agent configs are cleaned up
func TestService_SaveAgentConfigs_CleanupOld(t *testing.T) {
	tempDir := t.TempDir()

	cm := config.NewManagerWithPath(tempDir + "/config.json")
	pc := api.NewClient("https://test.example.com")
	authService := auth.NewService(cm, pc)
	syncService := NewService(cm, pc, authService)

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Save initial configs
	configs1 := []api.AgentConfig{
		{AgentID: "agent-1", AgentName: "Agent 1", IsEnabled: true},
		{AgentID: "agent-2", AgentName: "Agent 2", IsEnabled: true},
	}
	err := syncService.saveAgentConfigs(configs1)
	require.NoError(t, err)

	// Verify both agents exist
	loaded, err := syncService.GetLocalAgentConfigs()
	require.NoError(t, err)
	assert.Len(t, loaded, 2)

	// Save new configs with only agent-1
	configs2 := []api.AgentConfig{
		{AgentID: "agent-1", AgentName: "Agent 1 Updated", IsEnabled: true},
	}
	err = syncService.saveAgentConfigs(configs2)
	require.NoError(t, err)

	// Verify only agent-1 exists now
	loaded, err = syncService.GetLocalAgentConfigs()
	require.NoError(t, err)
	assert.Len(t, loaded, 1)
	assert.Equal(t, "agent-1", loaded[0].AgentID)
	assert.Equal(t, "Agent 1 Updated", loaded[0].AgentName)
}
