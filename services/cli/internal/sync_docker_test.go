package cli

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncService_SetDockerClient(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Initially no Docker client
	assert.Nil(t, syncService.dockerClient)
	assert.Nil(t, syncService.containerManager)

	// Set Docker client (using nil for unit test)
	if testing.Short() {
		// In short mode, just test the setter with nil
		syncService.SetDockerClient(nil)
		assert.Nil(t, syncService.dockerClient)
		return
	}

	// In full mode, test with real Docker client
	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	syncService.SetDockerClient(dockerClient)
	assert.NotNil(t, syncService.dockerClient)
	assert.NotNil(t, syncService.containerManager)
}

func TestSyncService_StartContainers_NoDockerClient(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	ctx := context.Background()
	// Try to start containers without Docker client
	err := syncService.StartContainers(ctx, "/tmp", "test-key")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Docker client not configured")
}

func TestSyncService_StopContainers_NoContainerManager(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	ctx := context.Background()
	// Try to stop containers without container manager
	err := syncService.StopContainers(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container manager not configured")
}

func TestSyncService_GetContainerStatus_NoContainerManager(t *testing.T) {
	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	ctx := context.Background()
	// Try to get status without container manager
	status, err := syncService.GetContainerStatus(ctx)
	assert.Error(t, err)
	assert.Nil(t, status)
	assert.Contains(t, err.Error(), "container manager not configured")
}

func TestSyncService_StartContainers_NoConfigs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	tempDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Set Docker client
	syncService.SetDockerClient(dockerClient)

	ctx := context.Background()
	// Try to start containers with no configs
	err = syncService.StartContainers(ctx, tempDir, "test-key")
	// Should not error, but should print "No agent configs to start"
	assert.NoError(t, err)
}

func TestSyncService_GetContainerStatus_WithDocker(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	tempDir := t.TempDir()

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Set Docker client
	syncService.SetDockerClient(dockerClient)

	ctx := context.Background()
	// Get container status
	status, err := syncService.GetContainerStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	t.Logf("Found %d ubik-managed containers", len(status))
}

func TestConvertMCPServers(t *testing.T) {
	configs := []MCPServerConfig{
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
	configs := []MCPServerConfig{}
	specs := convertMCPServers(configs)

	assert.Empty(t, specs)
	assert.NotNil(t, specs)
}

// Integration test: Full lifecycle with Docker
func TestSyncService_FullLifecycle_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	dockerClient, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerClient.Close()

	tempDir := t.TempDir()

	// Override home directory for testing
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	cm := &ConfigManager{
		configPath: tempDir + "/config.json",
	}
	pc := NewPlatformClient("https://test.example.com")
	authService := NewAuthService(cm, pc)
	syncService := NewSyncService(cm, pc, authService)

	// Set Docker client
	syncService.SetDockerClient(dockerClient)

	ctx := context.Background()

	// Test 1: Get container status (should be empty initially)
	status, err := syncService.GetContainerStatus(ctx)
	require.NoError(t, err)
	initialCount := len(status)
	t.Logf("Initial container count: %d", initialCount)

	// Test 2: Stop containers (should not error even if none exist)
	err = syncService.StopContainers(ctx)
	assert.NoError(t, err)

	// Test 3: Get status again
	status, err = syncService.GetContainerStatus(ctx)
	require.NoError(t, err)
	t.Logf("Final container count: %d", len(status))
}
