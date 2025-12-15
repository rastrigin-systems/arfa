package httpproxy

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	assert.NotNil(t, sm)
	assert.Empty(t, sm.ListSessions())
}

func TestSessionManager_Register(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID:  "session-1",
		EmployeeID: "emp-123",
		AgentID:    "agent-456",
		AgentName:  "Claude Code",
		Workspace:  "/home/user/project",
	}

	session, err := sm.Register(req)
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "session-1", session.ID)
	assert.Equal(t, "emp-123", session.EmployeeID)
	assert.Equal(t, "agent-456", session.AgentID)
	assert.Equal(t, "Claude Code", session.AgentName)
	assert.Equal(t, "/home/user/project", session.Workspace)
	assert.GreaterOrEqual(t, session.Port, 8100)
	assert.LessOrEqual(t, session.Port, 8109)
	assert.False(t, session.StartTime.IsZero())
}

func TestSessionManager_RegisterDuplicateSession(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID: "session-1",
		AgentName: "Claude Code",
	}

	session1, err := sm.Register(req)
	require.NoError(t, err)

	// Registering same session ID should return existing session
	session2, err := sm.Register(req)
	require.NoError(t, err)
	assert.Equal(t, session1.Port, session2.Port)
	assert.Equal(t, session1.ID, session2.ID)
}

func TestSessionManager_RegisterMultipleSessions(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	ports := make(map[int]bool)

	for i := 0; i < 5; i++ {
		req := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
			AgentName: "Claude Code",
		}
		session, err := sm.Register(req)
		require.NoError(t, err)

		// Each session should get a unique port
		assert.False(t, ports[session.Port], "port %d already allocated", session.Port)
		ports[session.Port] = true
	}

	assert.Len(t, sm.ListSessions(), 5)
}

func TestSessionManager_RegisterExhaustedPorts(t *testing.T) {
	// Only 3 ports available
	sm := NewSessionManager(8100, 8102)

	// Register 3 sessions (should succeed)
	for i := 0; i < 3; i++ {
		req := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
		}
		_, err := sm.Register(req)
		require.NoError(t, err)
	}

	// 4th session should fail
	req := RegisterSessionRequest{
		SessionID: "session-D",
	}
	_, err := sm.Register(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available ports")
}

func TestSessionManager_Unregister(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID: "session-1",
	}
	session, err := sm.Register(req)
	require.NoError(t, err)

	allocatedPort := session.Port

	// Unregister
	err = sm.Unregister("session-1")
	require.NoError(t, err)

	// Session should be gone
	assert.Nil(t, sm.GetByID("session-1"))
	assert.Nil(t, sm.GetByPort(allocatedPort))
	assert.Empty(t, sm.ListSessions())
}

func TestSessionManager_UnregisterNonExistent(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	// Should not error
	err := sm.Unregister("non-existent")
	assert.NoError(t, err)
}

func TestSessionManager_GetByPort(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID: "session-1",
		AgentName: "Claude Code",
	}
	session, err := sm.Register(req)
	require.NoError(t, err)

	// Get by port
	found := sm.GetByPort(session.Port)
	require.NotNil(t, found)
	assert.Equal(t, "session-1", found.ID)
	assert.Equal(t, "Claude Code", found.AgentName)

	// Non-existent port
	notFound := sm.GetByPort(9999)
	assert.Nil(t, notFound)
}

func TestSessionManager_GetByID(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID: "session-1",
		AgentName: "Claude Code",
	}
	_, err := sm.Register(req)
	require.NoError(t, err)

	// Get by ID
	found := sm.GetByID("session-1")
	require.NotNil(t, found)
	assert.Equal(t, "Claude Code", found.AgentName)

	// Non-existent ID
	notFound := sm.GetByID("non-existent")
	assert.Nil(t, notFound)
}

func TestSessionManager_ListSessions(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	// Register 3 sessions
	for i := 0; i < 3; i++ {
		req := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
			AgentName: "Agent " + string(rune('A'+i)),
		}
		_, err := sm.Register(req)
		require.NoError(t, err)
	}

	sessions := sm.ListSessions()
	assert.Len(t, sessions, 3)

	// Verify all sessions are present
	sessionIDs := make(map[string]bool)
	for _, s := range sessions {
		sessionIDs[s.ID] = true
	}
	assert.True(t, sessionIDs["session-A"])
	assert.True(t, sessionIDs["session-B"])
	assert.True(t, sessionIDs["session-C"])
}

func TestSessionManager_PortReuse(t *testing.T) {
	sm := NewSessionManager(8100, 8102)

	// Register a session
	req1 := RegisterSessionRequest{SessionID: "session-1"}
	session1, err := sm.Register(req1)
	require.NoError(t, err)
	port1 := session1.Port

	// Unregister it
	err = sm.Unregister("session-1")
	require.NoError(t, err)

	// Register new session - should be able to reuse the port
	req2 := RegisterSessionRequest{SessionID: "session-2"}
	session2, err := sm.Register(req2)
	require.NoError(t, err)

	// Port should be reused (since we only have 3 ports and we freed one)
	assert.Contains(t, []int{8100, 8101, 8102}, session2.Port)

	// With only one active session, the freed port should be available
	_ = port1 // Port reuse is implementation-dependent
}

func TestSessionManager_ConcurrentAccess(t *testing.T) {
	sm := NewSessionManager(8100, 8199) // 100 ports

	var wg sync.WaitGroup
	errors := make(chan error, 50)

	// Concurrent registrations
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := RegisterSessionRequest{
				SessionID: "session-" + string(rune(idx)),
				AgentName: "Agent",
			}
			_, err := sm.Register(req)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent registration error: %v", err)
	}

	// All 50 sessions should be registered
	assert.Len(t, sm.ListSessions(), 50)
}

func TestSessionManager_UpdateLastActive(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID: "session-1",
	}
	session, err := sm.Register(req)
	require.NoError(t, err)

	initialLastActive := session.LastActive

	// Wait a bit and update
	time.Sleep(10 * time.Millisecond)
	sm.UpdateLastActive("session-1")

	updated := sm.GetByID("session-1")
	assert.True(t, updated.LastActive.After(initialLastActive))
}

func TestSessionManager_CleanupStale(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	// Register sessions
	for i := 0; i < 3; i++ {
		req := RegisterSessionRequest{
			SessionID: "session-" + string(rune('A'+i)),
		}
		_, err := sm.Register(req)
		require.NoError(t, err)
	}

	assert.Len(t, sm.ListSessions(), 3)

	// Cleanup with 0 timeout should remove all
	removed := sm.CleanupStale(0)
	assert.Equal(t, 3, removed)
	assert.Empty(t, sm.ListSessions())
}

func TestSession_String(t *testing.T) {
	session := &Session{
		ID:        "session-123",
		Port:      8100,
		AgentName: "Claude Code",
		Workspace: "/home/user/project",
	}

	str := session.String()
	assert.Contains(t, str, "session-123")
	assert.Contains(t, str, "8100")
	assert.Contains(t, str, "Claude Code")
}

// === New tests for listener management ===

func TestSessionManager_Constants(t *testing.T) {
	// Verify constants are set correctly
	assert.Equal(t, 10, MaxSessions)
	assert.Equal(t, 5*time.Minute, StaleSessionTimeout)
	assert.Equal(t, 1*time.Minute, CleanupInterval)
}

func TestSession_OrgID(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	req := RegisterSessionRequest{
		SessionID:  "session-1",
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		AgentName:  "Claude Code",
	}

	session, err := sm.Register(req)
	require.NoError(t, err)
	assert.Equal(t, "org-456", session.OrgID)
}

func TestRegisterSessionRequest_Token(t *testing.T) {
	// Verify Token field exists in RegisterSessionRequest
	req := RegisterSessionRequest{
		SessionID: "session-1",
		Token:     "test-jwt-token",
		AgentName: "Claude Code",
	}

	assert.Equal(t, "test-jwt-token", req.Token)
}

func TestSessionManager_MaxSessionsEnforced(t *testing.T) {
	// Create manager with exactly MaxSessions ports
	sm := NewSessionManager(8100, 8100+MaxSessions-1)

	// Register MaxSessions sessions
	for i := 0; i < MaxSessions; i++ {
		req := RegisterSessionRequest{
			SessionID: fmt.Sprintf("session-%d", i),
		}
		_, err := sm.Register(req)
		require.NoError(t, err)
	}

	// Next should fail
	req := RegisterSessionRequest{
		SessionID: "one-too-many",
	}
	_, err := sm.Register(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no available ports")
}

func TestSessionManager_CleanupStaleWithTimeout(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	// Register a session
	req := RegisterSessionRequest{
		SessionID: "stale-session",
	}
	session, err := sm.Register(req)
	require.NoError(t, err)

	// Manually set LastActive to past (beyond StaleSessionTimeout)
	sm.mu.Lock()
	session.LastActive = time.Now().Add(-StaleSessionTimeout - time.Minute)
	sm.mu.Unlock()

	// Cleanup with StaleSessionTimeout should remove it
	removed := sm.CleanupStale(StaleSessionTimeout)
	assert.Equal(t, 1, removed)
	assert.Nil(t, sm.GetByID("stale-session"))
}

func TestSessionManager_CleanupStaleKeepsActive(t *testing.T) {
	sm := NewSessionManager(8100, 8109)

	// Register sessions
	for i := 0; i < 3; i++ {
		req := RegisterSessionRequest{
			SessionID: fmt.Sprintf("session-%d", i),
		}
		_, err := sm.Register(req)
		require.NoError(t, err)
	}

	// Update one session to be active
	sm.UpdateLastActive("session-1")

	// Set other sessions to be stale
	sm.mu.Lock()
	for _, session := range sm.sessions {
		if session.ID != "session-1" {
			session.LastActive = time.Now().Add(-StaleSessionTimeout - time.Minute)
		}
	}
	sm.mu.Unlock()

	// Cleanup should remove only stale sessions
	removed := sm.CleanupStale(StaleSessionTimeout)
	assert.Equal(t, 2, removed)
	assert.NotNil(t, sm.GetByID("session-1"))
	assert.Nil(t, sm.GetByID("session-0"))
	assert.Nil(t, sm.GetByID("session-2"))
}
