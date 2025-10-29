package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: These tests require Docker to be running
// They are integration tests and should be run separately

func TestNewDockerClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	assert.NotNil(t, client)
	assert.NotNil(t, client.cli)
}

func TestDockerClient_Ping(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	err = client.Ping()
	assert.NoError(t, err, "Docker daemon should be accessible")
}

func TestDockerClient_GetVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	version, err := client.GetVersion()
	assert.NoError(t, err)
	assert.NotEmpty(t, version)
}

func TestDockerClient_NetworkExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	// Check for default bridge network
	exists, err := client.NetworkExists("bridge")
	assert.NoError(t, err)
	assert.True(t, exists, "bridge network should always exist")

	// Check for non-existent network
	exists, err = client.NetworkExists("nonexistent-network-12345")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestDockerClient_ListContainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	client, err := NewDockerClient()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer client.Close()

	containers, err := client.ListContainers(true, nil)
	assert.NoError(t, err)
	// Can't assert exact count, but should be a list
	assert.NotNil(t, containers)
}
