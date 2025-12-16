package docker

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: These tests require Docker to be running
// They are integration tests and should be run separately

func TestNewClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.cli)
}

func TestClient_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// Should close without error
	err = client.Close()
	assert.NoError(t, err)
}

func TestClient_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	err = client.Ping(ctx)
	assert.NoError(t, err, "Docker daemon should be accessible")
}

func TestClient_GetVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	version, err := client.GetVersion(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
	t.Logf("Docker version: %s", version)
}

func TestClient_NetworkExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// Check for default bridge network
	exists, err := client.NetworkExists(ctx, "bridge")
	assert.NoError(t, err)
	assert.True(t, exists, "bridge network should always exist")

	// Check for non-existent network
	exists, err = client.NetworkExists(ctx, "nonexistent-network-12345")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestClient_CreateAndRemoveNetwork(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	networkName := "test-network-12345"

	// Clean up if exists
	client.RemoveNetwork(ctx, networkName)

	// Create network
	networkID, err := client.CreateNetwork(ctx, networkName)
	require.NoError(t, err)
	assert.NotEmpty(t, networkID)

	// Verify it exists
	exists, err := client.NetworkExists(ctx, networkName)
	require.NoError(t, err)
	assert.True(t, exists)

	// Remove network
	err = client.RemoveNetwork(ctx, networkName)
	require.NoError(t, err)

	// Verify it's gone
	exists, err = client.NetworkExists(ctx, networkName)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestClient_ListContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()

	// List all containers (running and stopped)
	containers, err := client.ListContainers(ctx, true, nil)
	assert.NoError(t, err)
	assert.NotNil(t, containers)
	t.Logf("Found %d containers (all)", len(containers))

	// List only running containers
	runningContainers, err := client.ListContainers(ctx, false, nil)
	assert.NoError(t, err)
	assert.NotNil(t, runningContainers)
	t.Logf("Found %d containers (running)", len(runningContainers))

	// Running containers should be <= all containers
	assert.LessOrEqual(t, len(runningContainers), len(containers))
}

func TestClient_ContainerInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	containers, err := client.ListContainers(ctx, true, nil)
	require.NoError(t, err)

	if len(containers) > 0 {
		// Test first container info
		c := containers[0]
		assert.NotEmpty(t, c.ID)
		assert.NotEmpty(t, c.Image)
		assert.NotEmpty(t, c.State)
		assert.NotZero(t, c.Created)
		t.Logf("Container: %s (image: %s, state: %s)", c.Name, c.Image, c.State)
	}
}

func TestClient_PullImage_Error(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	// Try to pull non-existent image
	err = client.PullImage(ctx, "this-image-definitely-does-not-exist-12345:latest")
	assert.Error(t, err, "Should fail to pull non-existent image")
}
