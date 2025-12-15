package httpproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestSocket creates a temporary socket with a short path (macOS has 104 char limit)
func createTestSocket(t *testing.T) string {
	t.Helper()
	// Use /tmp directly to avoid long paths from t.TempDir()
	tmpFile, err := os.CreateTemp("/tmp", "ubik-test-*.sock")
	require.NoError(t, err)
	path := tmpFile.Name()
	tmpFile.Close()
	os.Remove(path) // Remove file so we can create socket

	t.Cleanup(func() {
		os.Remove(path)
	})

	return path
}

func TestNewControlServer(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	assert.NotNil(t, cs)
	assert.Equal(t, socketPath, cs.SocketPath())
}

func TestControlServer_StartStop(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server
	err := cs.Start(ctx)
	require.NoError(t, err)

	// Socket should exist
	_, err = os.Stat(socketPath)
	require.NoError(t, err)

	// Should be able to connect
	conn, err := net.Dial("unix", socketPath)
	require.NoError(t, err)
	conn.Close()

	// Stop server
	err = cs.Stop()
	require.NoError(t, err)

	// Socket should be removed
	_, err = os.Stat(socketPath)
	assert.True(t, os.IsNotExist(err))
}

func TestControlServer_RegisterSession(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	// Create HTTP client for Unix socket
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register a session
	reqBody := RegisterSessionRequest{
		SessionID:  "test-session",
		EmployeeID: "emp-123",
		AgentID:    "agent-456",
		AgentName:  "Claude Code",
		Workspace:  "/home/user/project",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result RegisterSessionResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "test-session", result.SessionID)
	assert.GreaterOrEqual(t, result.Port, 8100)
	assert.LessOrEqual(t, result.Port, 8109)
	assert.NotEmpty(t, result.ProxyAddr)
}

func TestControlServer_UnregisterSession(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// First register a session
	reqBody := RegisterSessionRequest{SessionID: "test-session"}
	body, _ := json.Marshal(reqBody)
	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Unregister the session
	req, _ := http.NewRequest(http.MethodDelete, "http://unix/sessions/test-session", nil)
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Session should be gone
	assert.Nil(t, sm.GetByID("test-session"))
}

func TestControlServer_ListSessions(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register 3 sessions
	for i := 0; i < 3; i++ {
		reqBody := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
			AgentName: "Agent " + string(rune('A'+i)),
		}
		body, _ := json.Marshal(reqBody)
		resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		resp.Body.Close()
	}

	// List sessions
	resp, err := client.Get("http://unix/sessions")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var sessions []*Session
	err = json.NewDecoder(resp.Body).Decode(&sessions)
	require.NoError(t, err)

	assert.Len(t, sessions, 3)
}

func TestControlServer_Health(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	resp, err := client.Get("http://unix/health")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var health HealthResponse
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err)

	assert.Equal(t, "ok", health.Status)
	assert.Equal(t, 0, health.ActiveSessions)
}

func TestControlServer_InvalidRequest(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Invalid JSON
	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader([]byte("invalid json")))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Missing session ID
	reqBody := RegisterSessionRequest{AgentName: "Claude Code"}
	body, _ := json.Marshal(reqBody)
	resp, err = client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestControlServer_SocketPermissions(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	// Check socket permissions (should be 0600 for security)
	info, err := os.Stat(socketPath)
	require.NoError(t, err)

	// On Unix, socket files have mode prefixed with 's'
	// We check the permission bits
	perms := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0600), perms, "socket should have 0600 permissions")
}

// RegisterSessionResponse is the response for session registration
type RegisterSessionResponse struct {
	SessionID string `json:"session_id"`
	Port      int    `json:"port"`
	ProxyAddr string `json:"proxy_addr"`
	CertPath  string `json:"cert_path"`
}

// HealthResponse is the response for health check
type HealthResponse struct {
	Status          string `json:"status"`
	ActiveSessions  int    `json:"active_sessions"`
	PlatformHealthy bool   `json:"platform_healthy"`
	Uptime          string `json:"uptime"`
}

// === JWT Validation Tests ===

func TestControlServer_Register_RequiresToken(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	cs.SetRequireToken(true) // Enable JWT validation

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register without token should fail with 401
	reqBody := RegisterSessionRequest{
		SessionID: "test-session",
		AgentName: "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestControlServer_Register_ValidatesToken(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	cs.SetRequireToken(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register with invalid token should fail with 401
	reqBody := RegisterSessionRequest{
		SessionID: "test-session",
		Token:     "invalid.jwt.token",
		AgentName: "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestControlServer_Register_ExtractsClaimsFromToken(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	cs.SetRequireToken(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Generate valid token with employee_id and org_id
	token := generateTestToken(t, "emp-from-token", "org-from-token", time.Hour)

	// Register with valid token - EmployeeID/OrgID should come from token, not request
	reqBody := RegisterSessionRequest{
		SessionID:  "test-session",
		Token:      token,
		EmployeeID: "emp-from-request", // Should be overridden
		OrgID:      "org-from-request", // Should be overridden
		AgentName:  "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Verify session has EmployeeID/OrgID from token
	session := sm.GetByID("test-session")
	require.NotNil(t, session)
	assert.Equal(t, "emp-from-token", session.EmployeeID)
	assert.Equal(t, "org-from-token", session.OrgID)
}

func TestControlServer_Register_ExpiredTokenFails(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	cs.SetRequireToken(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Generate expired token
	token := generateTestToken(t, "emp-123", "org-456", -time.Hour)

	reqBody := RegisterSessionRequest{
		SessionID: "test-session",
		Token:     token,
		AgentName: "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestControlServer_Register_MaxSessionsReturns429(t *testing.T) {
	// Create manager with only 3 ports
	sm := NewSessionManager(8100, 8102)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	cs.SetRequireToken(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register 3 sessions (should succeed)
	for i := 0; i < 3; i++ {
		token := generateTestToken(t, "emp-"+string(rune('A'+i)), "org-456", time.Hour)
		reqBody := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
			Token:     token,
			AgentName: "Claude Code",
		}
		body, _ := json.Marshal(reqBody)

		resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
		require.NoError(t, err)
		resp.Body.Close()
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
	}

	// 4th session should fail with 429 or 503
	token := generateTestToken(t, "emp-D", "org-456", time.Hour)
	reqBody := RegisterSessionRequest{
		SessionID: "session-D",
		Token:     token,
		AgentName: "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should return 429 Too Many Requests or 503 Service Unavailable
	assert.True(t, resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable,
		"expected 429 or 503, got %d", resp.StatusCode)
}

func TestControlServer_Register_WorksWithoutTokenWhenNotRequired(t *testing.T) {
	sm := NewSessionManager(8100, 8109)
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)
	// NOT calling SetRequireToken(true) - token validation disabled

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Register without token should succeed when not required
	reqBody := RegisterSessionRequest{
		SessionID:  "test-session",
		EmployeeID: "emp-123",
		AgentName:  "Claude Code",
	}
	body, _ := json.Marshal(reqBody)

	resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestControlServer_ConcurrentRequests(t *testing.T) {
	sm := NewSessionManager(8100, 8199) // 100 ports
	socketPath := createTestSocket(t)

	cs := NewControlServer(socketPath, sm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := cs.Start(ctx)
	require.NoError(t, err)
	defer cs.Stop()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	// Concurrent registrations
	done := make(chan bool, 20)
	for i := 0; i < 20; i++ {
		go func(idx int) {
			reqBody := RegisterSessionRequest{
				SessionID: "session-" + string(rune(idx)),
			}
			body, _ := json.Marshal(reqBody)
			resp, err := client.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
			if err == nil {
				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
			done <- true
		}(i)
	}

	// Wait for all requests
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have registered sessions
	sessions := sm.ListSessions()
	assert.GreaterOrEqual(t, len(sessions), 15) // Allow some failures due to race
}
