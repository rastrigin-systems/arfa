package httpproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"
)

// SetSessionRequest is the request body for setting the active session
type SetSessionRequest struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
}

// ControlServer provides a Unix socket HTTP server for CLI → Daemon IPC
type ControlServer struct {
	sockPath    string
	proxyServer *ProxyServer
	listener    net.Listener
	server      *http.Server
}

// NewControlServer creates a new control server
func NewControlServer(sockPath string, proxyServer *ProxyServer) *ControlServer {
	return &ControlServer{
		sockPath:    sockPath,
		proxyServer: proxyServer,
	}
}

// Start starts the control server on the Unix socket
func (c *ControlServer) Start() error {
	// Remove existing socket file if present
	if err := os.Remove(c.sockPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	listener, err := net.Listen("unix", c.sockPath)
	if err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}
	c.listener = listener

	// Set socket permissions
	if err := os.Chmod(c.sockPath, 0600); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/session", c.handleSession)
	mux.HandleFunc("/health", c.handleHealth)

	c.server = &http.Server{Handler: mux}

	go func() {
		if err := c.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Control server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the control server
func (c *ControlServer) Stop() error {
	if c.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := c.server.Shutdown(ctx); err != nil {
			return err
		}
	}
	if c.listener != nil {
		c.listener.Close()
	}
	os.Remove(c.sockPath)
	return nil
}

// handleSession handles POST /session to set the active session
func (c *ControlServer) handleSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		http.Error(w, "session_id is required", http.StatusBadRequest)
		return
	}

	c.proxyServer.SetSession(req.SessionID, req.AgentID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// handleHealth handles GET /health for health checks
func (c *ControlServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// RegisterSession sends a session registration request to the daemon via Unix socket
func RegisterSession(sockPath, sessionID, agentID string) error {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", sockPath)
			},
		},
		Timeout: 5 * time.Second,
	}

	req := SetSessionRequest{
		SessionID: sessionID,
		AgentID:   agentID,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := client.Post("http://unix/session", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("daemon returned status %d", resp.StatusCode)
	}

	return nil
}
