package httpproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// ControlServer provides a Unix socket API for session management
type ControlServer struct {
	socketPath     string
	sessionManager *SessionManager
	listener       net.Listener
	server         *http.Server
	startTime      time.Time
	mu             sync.Mutex

	// Optional: policy engine reference (set later)
	policyEngine interface {
		IsPlatformHealthy() bool
	}

	// Optional: cert path for response
	certPath string
}

// NewControlServer creates a new control server
func NewControlServer(socketPath string, sessionManager *SessionManager) *ControlServer {
	return &ControlServer{
		socketPath:     socketPath,
		sessionManager: sessionManager,
	}
}

// SocketPath returns the socket path
func (cs *ControlServer) SocketPath() string {
	return cs.socketPath
}

// SetPolicyEngine sets the policy engine reference
func (cs *ControlServer) SetPolicyEngine(pe interface{ IsPlatformHealthy() bool }) {
	cs.policyEngine = pe
}

// SetCertPath sets the CA certificate path for responses
func (cs *ControlServer) SetCertPath(path string) {
	cs.certPath = path
}

// Start starts the control server
func (cs *ControlServer) Start(ctx context.Context) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Remove existing socket if present
	if err := os.Remove(cs.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create Unix socket listener
	listener, err := net.Listen("unix", cs.socketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	cs.listener = listener

	// Set socket permissions to 0600 for security
	if err := os.Chmod(cs.socketPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	// Create HTTP server with routes
	mux := http.NewServeMux()
	mux.HandleFunc("/sessions", cs.handleSessions)
	mux.HandleFunc("/sessions/", cs.handleSessionByID)
	mux.HandleFunc("/health", cs.handleHealth)

	cs.server = &http.Server{
		Handler: mux,
	}
	cs.startTime = time.Now()

	// Start serving in background
	go func() {
		if err := cs.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "Control server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the control server
func (cs *ControlServer) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cs.server.Shutdown(ctx)
	}

	if cs.listener != nil {
		cs.listener.Close()
	}

	// Remove socket file
	os.Remove(cs.socketPath)

	return nil
}

// handleSessions handles POST /sessions (register) and GET /sessions (list)
func (cs *ControlServer) handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		cs.handleRegister(w, r)
	case http.MethodGet:
		cs.handleList(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSessionByID handles DELETE /sessions/:id
func (cs *ControlServer) handleSessionByID(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from path
	path := strings.TrimPrefix(r.URL.Path, "/sessions/")
	if path == "" {
		http.Error(w, "Session ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodDelete:
		cs.handleUnregister(w, r, path)
	case http.MethodGet:
		cs.handleGetSession(w, r, path)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRegister handles POST /sessions
func (cs *ControlServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req RegisterSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	session, err := cs.sessionManager.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	resp := struct {
		SessionID string `json:"session_id"`
		Port      int    `json:"port"`
		ProxyAddr string `json:"proxy_addr"`
		CertPath  string `json:"cert_path,omitempty"`
	}{
		SessionID: session.ID,
		Port:      session.Port,
		ProxyAddr: fmt.Sprintf("localhost:%d", session.Port),
		CertPath:  cs.certPath,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleUnregister handles DELETE /sessions/:id
func (cs *ControlServer) handleUnregister(w http.ResponseWriter, r *http.Request, sessionID string) {
	if err := cs.sessionManager.Unregister(sessionID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleList handles GET /sessions
func (cs *ControlServer) handleList(w http.ResponseWriter, r *http.Request) {
	sessions := cs.sessionManager.ListSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

// handleGetSession handles GET /sessions/:id
func (cs *ControlServer) handleGetSession(w http.ResponseWriter, r *http.Request, sessionID string) {
	session := cs.sessionManager.GetByID(sessionID)
	if session == nil {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// handleHealth handles GET /health
func (cs *ControlServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	platformHealthy := true
	if cs.policyEngine != nil {
		platformHealthy = cs.policyEngine.IsPlatformHealthy()
	}

	uptime := time.Since(cs.startTime).Round(time.Second)

	resp := struct {
		Status          string `json:"status"`
		ActiveSessions  int    `json:"active_sessions"`
		PlatformHealthy bool   `json:"platform_healthy"`
		Uptime          string `json:"uptime"`
	}{
		Status:          "ok",
		ActiveSessions:  cs.sessionManager.ActiveCount(),
		PlatformHealthy: platformHealthy,
		Uptime:          uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
