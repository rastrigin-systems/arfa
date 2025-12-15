package httpproxy

import (
	"fmt"
	"sync"
	"time"
)

// Session represents an active proxy session with its allocated port
type Session struct {
	ID         string    `json:"id"`
	Port       int       `json:"port"`
	EmployeeID string    `json:"employee_id"`
	AgentID    string    `json:"agent_id"`
	AgentName  string    `json:"agent_name"`
	Workspace  string    `json:"workspace"`
	StartTime  time.Time `json:"start_time"`
	LastActive time.Time `json:"last_active"`

	// Metrics
	RequestCount  int64 `json:"request_count"`
	BytesSent     int64 `json:"bytes_sent"`
	BytesReceived int64 `json:"bytes_received"`
}

// String returns a human-readable representation of the session
func (s *Session) String() string {
	return fmt.Sprintf("Session{id=%s, port=%d, agent=%s, workspace=%s}",
		s.ID, s.Port, s.AgentName, s.Workspace)
}

// RegisterSessionRequest contains the information needed to register a new session
type RegisterSessionRequest struct {
	SessionID  string `json:"session_id"`
	EmployeeID string `json:"employee_id"`
	AgentID    string `json:"agent_id"`
	AgentName  string `json:"agent_name"`
	Workspace  string `json:"workspace"`
}

// SessionManager manages dynamic port allocation for proxy sessions
type SessionManager struct {
	mu            sync.RWMutex
	sessions      map[string]*Session // sessionID -> Session
	portToSession map[int]string      // port -> sessionID
	minPort       int
	maxPort       int
	nextPort      int // For round-robin allocation
}

// NewSessionManager creates a new session manager with the specified port range
func NewSessionManager(minPort, maxPort int) *SessionManager {
	return &SessionManager{
		sessions:      make(map[string]*Session),
		portToSession: make(map[int]string),
		minPort:       minPort,
		maxPort:       maxPort,
		nextPort:      minPort,
	}
}

// Register allocates a port for a new session
func (sm *SessionManager) Register(req RegisterSessionRequest) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check if session already exists
	if existing, ok := sm.sessions[req.SessionID]; ok {
		return existing, nil
	}

	// Find an available port
	port, err := sm.findAvailablePort()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:         req.SessionID,
		Port:       port,
		EmployeeID: req.EmployeeID,
		AgentID:    req.AgentID,
		AgentName:  req.AgentName,
		Workspace:  req.Workspace,
		StartTime:  now,
		LastActive: now,
	}

	sm.sessions[req.SessionID] = session
	sm.portToSession[port] = req.SessionID

	return session, nil
}

// Unregister removes a session and frees its port
func (sm *SessionManager) Unregister(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, ok := sm.sessions[sessionID]
	if !ok {
		return nil // Already unregistered
	}

	delete(sm.portToSession, session.Port)
	delete(sm.sessions, sessionID)

	return nil
}

// GetByPort returns the session for a given port
func (sm *SessionManager) GetByPort(port int) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessionID, ok := sm.portToSession[port]
	if !ok {
		return nil
	}

	return sm.sessions[sessionID]
}

// GetByID returns session info by ID
func (sm *SessionManager) GetByID(sessionID string) *Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.sessions[sessionID]
}

// ListSessions returns all active sessions
func (sm *SessionManager) ListSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*Session, 0, len(sm.sessions))
	for _, s := range sm.sessions {
		sessions = append(sessions, s)
	}
	return sessions
}

// UpdateLastActive updates the last active timestamp for a session
func (sm *SessionManager) UpdateLastActive(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, ok := sm.sessions[sessionID]; ok {
		session.LastActive = time.Now()
	}
}

// IncrementMetrics updates request metrics for a session
func (sm *SessionManager) IncrementMetrics(sessionID string, bytesSent, bytesReceived int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, ok := sm.sessions[sessionID]; ok {
		session.RequestCount++
		session.BytesSent += bytesSent
		session.BytesReceived += bytesReceived
		session.LastActive = time.Now()
	}
}

// CleanupStale removes sessions that haven't been active within the timeout
func (sm *SessionManager) CleanupStale(timeout time.Duration) int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cutoff := time.Now().Add(-timeout)
	removed := 0

	for sessionID, session := range sm.sessions {
		if session.LastActive.Before(cutoff) {
			delete(sm.portToSession, session.Port)
			delete(sm.sessions, sessionID)
			removed++
		}
	}

	return removed
}

// PortRange returns the configured port range
func (sm *SessionManager) PortRange() (min, max int) {
	return sm.minPort, sm.maxPort
}

// ActiveCount returns the number of active sessions
func (sm *SessionManager) ActiveCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// findAvailablePort finds an available port using round-robin allocation
func (sm *SessionManager) findAvailablePort() (int, error) {
	totalPorts := sm.maxPort - sm.minPort + 1

	// Try each port starting from nextPort
	for i := 0; i < totalPorts; i++ {
		port := sm.minPort + ((sm.nextPort - sm.minPort + i) % totalPorts)

		if _, allocated := sm.portToSession[port]; !allocated {
			// Update nextPort for round-robin
			sm.nextPort = sm.minPort + ((port - sm.minPort + 1) % totalPorts)
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d (all %d ports in use)",
		sm.minPort, sm.maxPort, totalPorts)
}
