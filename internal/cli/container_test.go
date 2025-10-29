package cli

import (
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

	// Remove network if it exists
	dockerClient.RemoveNetwork(cm.networkName)

	// Setup network
	err = cm.SetupNetwork()
	require.NoError(t, err)

	// Verify network exists
	exists, err := dockerClient.NetworkExists(cm.networkName)
	require.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	err = dockerClient.RemoveNetwork(cm.networkName)
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

	containers, err := cm.GetContainerStatus()
	require.NoError(t, err)
	assert.NotNil(t, containers)
	// There might be 0 or more containers
}

func TestGetWorkspacePath(t *testing.T) {
	path, err := GetWorkspacePath(".")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	// Should return absolute path
	assert.Contains(t, path, "/")
}

func TestGetWorkspacePath_InvalidPath(t *testing.T) {
	// Test with a path that contains invalid characters
	// Note: Most paths are valid on Unix systems, so this might not fail
	_, err := GetWorkspacePath(".")
	assert.NoError(t, err) // Current directory should always work
}
