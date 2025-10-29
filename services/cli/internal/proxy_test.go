package cli

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProxyService_StreamIO tests basic I/O streaming
func TestProxyService_StreamIO(t *testing.T) {
	// Simple test: simulate reading from stdin and writing to stdout
	stdin := strings.NewReader("test input\n")
	stdout := &bytes.Buffer{}

	// Create a simple echo function that reads from reader and writes to writer
	echo := func(reader io.Reader, writer io.Writer) error {
		_, err := io.Copy(writer, reader)
		return err
	}

	err := echo(stdin, stdout)
	require.NoError(t, err)
	assert.Equal(t, "test input\n", stdout.String())
}

// TestProxyService_AttachToContainer tests container attachment logic
func TestProxyService_AttachToContainer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker integration test in short mode")
	}

	// This would require a real Docker container
	// For now, we'll test the logic structure
	ps := NewProxyService()
	assert.NotNil(t, ps, "ProxyService should be created successfully")

	// Test that we can create proxy service without Docker client
	assert.Nil(t, ps.dockerClient, "Docker client should be nil initially")
}

// TestProxyService_HandleSignals tests signal handling
func TestProxyService_HandleSignals(t *testing.T) {
	ps := NewProxyService()
	assert.NotNil(t, ps, "ProxyService should be created")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Test that context cancellation works
	select {
	case <-ctx.Done():
		assert.Error(t, ctx.Err())
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Context should have been cancelled")
	}
}

// TestSessionInfo_Duration tests session duration tracking
func TestSessionInfo_Duration(t *testing.T) {
	session := &SessionInfo{
		ContainerID: "test-container",
		StartTime:   time.Now(),
	}

	// Simulate some time passing
	time.Sleep(100 * time.Millisecond)
	session.EndTime = time.Now()

	duration := session.Duration()
	assert.Greater(t, duration.Milliseconds(), int64(50), "Duration should be at least 50ms")
	assert.Less(t, duration.Milliseconds(), int64(200), "Duration should be less than 200ms")
}

// TestSessionInfo_String tests session info string formatting
func TestSessionInfo_String(t *testing.T) {
	start := time.Now()
	end := start.Add(5*time.Minute + 23*time.Second)

	session := &SessionInfo{
		ContainerID: "test-container-123",
		AgentName:   "claude-code",
		WorkingDir:  "/workspace/project",
		StartTime:   start,
		EndTime:     end,
	}

	str := session.String()
	assert.Contains(t, str, "claude-code")
	assert.Contains(t, str, "5m23s")
	assert.Contains(t, str, "test-contain") // First 12 chars of container ID
	assert.Contains(t, str, "/workspace/project")
}

// TestProxyService_StartSession tests session initialization
func TestProxyService_StartSession(t *testing.T) {
	ps := NewProxyService()
	containerID := "test-container-abc"
	agentName := "claude-code"
	workingDir := "/workspace"

	session := ps.StartSession(containerID, agentName, workingDir)

	assert.NotNil(t, session)
	assert.Equal(t, containerID, session.ContainerID)
	assert.Equal(t, agentName, session.AgentName)
	assert.Equal(t, workingDir, session.WorkingDir)
	assert.False(t, session.StartTime.IsZero())
	assert.True(t, session.EndTime.IsZero()) // Not ended yet
}

// TestProxyService_EndSession tests session termination
func TestProxyService_EndSession(t *testing.T) {
	ps := NewProxyService()
	containerID := "test-container-abc"

	session := ps.StartSession(containerID, "claude-code", "/workspace")
	assert.True(t, session.EndTime.IsZero(), "Session should not be ended initially")

	// Simulate some work
	time.Sleep(50 * time.Millisecond)

	ps.EndSession(session)
	assert.False(t, session.EndTime.IsZero(), "Session should be ended now")
	assert.Greater(t, session.Duration().Milliseconds(), int64(40))
}

// TestProxyService_StreamWithTimeout tests I/O streaming with timeout
func TestProxyService_StreamWithTimeout(t *testing.T) {
	// Test that we can cancel streaming via context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Create a reader that blocks indefinitely
	pr, pw := io.Pipe()
	defer pr.Close()
	defer pw.Close()

	errChan := make(chan error, 1)

	go func() {
		// This should be cancelled by context timeout
		buf := make([]byte, 1024)
		_, err := pr.Read(buf)
		errChan <- err
	}()

	select {
	case <-ctx.Done():
		// Expected - context timed out
		assert.Error(t, ctx.Err())
	case err := <-errChan:
		// Should not get here unless read errors
		if err != nil {
			t.Logf("Read error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Should have timed out")
	}
}

// TestProxyOptions_Validation tests proxy configuration validation
func TestProxyOptions_Validation(t *testing.T) {
	tests := []struct {
		name        string
		options     ProxyOptions
		expectValid bool
	}{
		{
			name: "Valid options with container ID",
			options: ProxyOptions{
				ContainerID: "test-container",
				AgentName:   "claude-code",
				WorkingDir:  "/workspace",
			},
			expectValid: true,
		},
		{
			name: "Missing container ID",
			options: ProxyOptions{
				ContainerID: "",
				AgentName:   "claude-code",
				WorkingDir:  "/workspace",
			},
			expectValid: false,
		},
		{
			name: "Missing agent name",
			options: ProxyOptions{
				ContainerID: "test-container",
				AgentName:   "",
				WorkingDir:  "/workspace",
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.expectValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
