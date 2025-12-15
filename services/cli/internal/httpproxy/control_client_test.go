package httpproxy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewControlClient(t *testing.T) {
	client := NewControlClient("/tmp/test.sock")
	assert.NotNil(t, client)
	assert.Equal(t, "/tmp/test.sock", client.SocketPath())
}

func TestControlClient_RegisterSession(t *testing.T) {
	// Start a control server
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	// Create client and register
	client := NewControlClient(socketPath)

	resp, err := client.RegisterSession(RegisterSessionRequest{
		SessionID:  "test-session",
		EmployeeID: "emp-123",
		AgentID:    "agent-456",
		AgentName:  "Claude Code",
		Workspace:  "/home/user/project",
	})

	require.NoError(t, err)
	assert.Equal(t, "test-session", resp.SessionID)
	assert.GreaterOrEqual(t, resp.Port, 8100)
	assert.LessOrEqual(t, resp.Port, 8109)
	assert.NotEmpty(t, resp.ProxyAddr)
}

func TestControlClient_UnregisterSession(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := NewControlClient(socketPath)

	// First register
	_, err = client.RegisterSession(RegisterSessionRequest{
		SessionID: "test-session",
	})
	require.NoError(t, err)

	// Then unregister
	err = client.UnregisterSession("test-session")
	require.NoError(t, err)

	// Verify session is gone
	assert.Nil(t, sm.GetByID("test-session"))
}

func TestControlClient_ListSessions(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := NewControlClient(socketPath)

	// Register a few sessions
	for i := 0; i < 3; i++ {
		_, err := client.RegisterSession(RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
			AgentName: "Agent " + string(rune('A'+i)),
		})
		require.NoError(t, err)
	}

	// List sessions
	sessions, err := client.ListSessions()
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
}

func TestControlClient_Health(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := NewControlClient(socketPath)

	health, err := client.Health()
	require.NoError(t, err)
	assert.Equal(t, "ok", health.Status)
	assert.True(t, health.PlatformHealthy)
}

func TestControlClient_ConnectionError(t *testing.T) {
	client := NewControlClient("/nonexistent/socket.sock")

	_, err := client.RegisterSession(RegisterSessionRequest{
		SessionID: "test",
	})
	assert.Error(t, err)
}

func TestControlClient_Timeout(t *testing.T) {
	client := NewControlClient("/nonexistent/socket.sock")
	client.SetTimeout(100 * time.Millisecond)

	_, err := client.RegisterSession(RegisterSessionRequest{
		SessionID: "test",
	})
	assert.Error(t, err)
}
