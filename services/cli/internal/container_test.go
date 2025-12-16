package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContainerManager(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	assert.NotNil(t, cm)
	assert.Equal(t, "ubik-network", cm.networkName)
	assert.NotNil(t, cm.dockerClient)
}

func TestContainerManager_SetupNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	ctx := context.Background()

	// Remove network if it exists
	dockerClient.RemoveNetwork(ctx, cm.networkName)

	// Setup network (first time)
	err = cm.SetupNetwork(ctx)
	require.NoError(t, err)

	// Verify network exists
	exists, err := dockerClient.NetworkExists(ctx, cm.networkName)
	require.NoError(t, err)
	assert.True(t, exists)

	// Setup network again (should be idempotent)
	err = cm.SetupNetwork(ctx)
	require.NoError(t, err)

	// Cleanup
	err = dockerClient.RemoveNetwork(ctx, cm.networkName)
	assert.NoError(t, err)
}

func TestContainerManager_GetContainerStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	ctx := context.Background()

	containers, err := cm.GetContainerStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, containers)
	t.Logf("Found %d ubik-managed containers", len(containers))
}

func TestContainerManager_StopContainers_NoContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	ctx := context.Background()

	// Should handle case with no containers gracefully
	err = cm.StopContainers(ctx)
	assert.NoError(t, err)
}

func TestContainerManager_CleanupContainers_NoContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	ctx := context.Background()

	// Should handle case with no stopped containers gracefully
	err = cm.CleanupContainers(ctx)
	assert.NoError(t, err)
}

func TestGetWorkspacePath_CurrentDirectory(t *testing.T) {
	path, err := GetWorkspacePath(".")
	require.NoError(t, err)
	assert.NotEmpty(t, path)

	// Should return absolute path
	assert.True(t, filepath.IsAbs(path), "Should return absolute path")

	// Should match current working directory
	cwd, _ := os.Getwd()
	assert.Equal(t, cwd, path)
}

func TestGetWorkspacePath_RelativePath(t *testing.T) {
	path, err := GetWorkspacePath("../")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.True(t, filepath.IsAbs(path), "Should return absolute path")
}

func TestGetWorkspacePath_AbsolutePath(t *testing.T) {
	absPath := "/tmp"
	path, err := GetWorkspacePath(absPath)
	require.NoError(t, err)
	assert.Equal(t, absPath, path)
}

func TestMCPServerSpec_Validation(t *testing.T) {
	spec := MCPServerSpec{
		ServerID:   "test-server",
		ServerName: "Test Server",
		ServerType: "filesystem",
		Image:      "ubik/mcp-filesystem:latest",
		Port:       8001,
		Config:     map[string]interface{}{"root": "/workspace"},
	}

	assert.Equal(t, "test-server", spec.ServerID)
	assert.Equal(t, "Test Server", spec.ServerName)
	assert.Equal(t, "filesystem", spec.ServerType)
	assert.Equal(t, 8001, spec.Port)
	assert.NotNil(t, spec.Config)
}

func TestAgentSpec_Validation(t *testing.T) {
	spec := AgentSpec{
		AgentID:       "test-agent",
		AgentName:     "Claude Code",
		AgentType:     "claude-code",
		Image:         "ubik/claude-code:latest",
		Configuration: map[string]interface{}{"model": "claude-3-5-sonnet"},
		MCPServers:    []MCPServerSpec{},
		APIKey:        "sk-ant-test",
	}

	assert.Equal(t, "test-agent", spec.AgentID)
	assert.Equal(t, "Claude Code", spec.AgentName)
	assert.Equal(t, "claude-code", spec.AgentType)
	assert.Equal(t, "sk-ant-test", spec.APIKey)
	assert.NotNil(t, spec.Configuration)
	assert.Empty(t, spec.MCPServers)
}

// Integration test: Create and start a simple container
func TestContainerManager_StartMCPServer_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	cm := NewContainerManager(dockerClient)
	ctx := context.Background()

	// Setup network
	err = cm.SetupNetwork(ctx)
	require.NoError(t, err)
	defer dockerClient.RemoveNetwork(ctx, cm.networkName)

	// Create temp workspace
	tmpDir := t.TempDir()

	// Note: This will fail if the image doesn't exist
	// We're testing the code path, not actually starting a real MCP server
	spec := MCPServerSpec{
		ServerID:   "test-mcp",
		ServerName: "Test MCP",
		ServerType: "filesystem",
		Image:      "alpine:latest", // Use alpine instead of actual MCP image
		Port:       8001,
		Config:     map[string]interface{}{"root": "/workspace"},
	}

	// This will try to start but will likely fail since alpine doesn't have MCP server
	// We're just testing that the function doesn't panic
	containerID, err := cm.StartMCPServer(ctx, spec, tmpDir)

	// Clean up if container was created
	if containerID != "" {
		dockerClient.StopContainer(ctx, containerID, nil)
		dockerClient.RemoveContainer(ctx, containerID, true)
	}

	// We expect it to work or fail gracefully
	if err != nil {
		t.Logf("Expected behavior: %v", err)
	}
}
