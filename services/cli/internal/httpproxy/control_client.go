package httpproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ControlClient provides a client for the Unix socket control API
type ControlClient struct {
	socketPath string
	timeout    time.Duration
	httpClient *http.Client
}

// ControlSessionResponse is the response from session registration
type ControlSessionResponse struct {
	SessionID string `json:"session_id"`
	Port      int    `json:"port"`
	ProxyAddr string `json:"proxy_addr"`
	CertPath  string `json:"cert_path,omitempty"`
}

// ControlHealthResponse is the response from health check
type ControlHealthResponse struct {
	Status          string `json:"status"`
	ActiveSessions  int    `json:"active_sessions"`
	PlatformHealthy bool   `json:"platform_healthy"`
	Uptime          string `json:"uptime"`
}

// NewControlClient creates a new control client
func NewControlClient(socketPath string) *ControlClient {
	client := &ControlClient{
		socketPath: socketPath,
		timeout:    5 * time.Second,
	}
	client.initHTTPClient()
	return client
}

// NewDefaultControlClient creates a control client with the default socket path
func NewDefaultControlClient() (*ControlClient, error) {
	socketPath, err := GetDefaultSocketPath()
	if err != nil {
		return nil, err
	}
	return NewControlClient(socketPath), nil
}

// initHTTPClient creates the HTTP client with Unix socket transport
func (c *ControlClient) initHTTPClient() {
	c.httpClient = &http.Client{
		Timeout: c.timeout,
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.DialTimeout("unix", c.socketPath, c.timeout)
			},
		},
	}
}

// SocketPath returns the socket path
func (c *ControlClient) SocketPath() string {
	return c.socketPath
}

// SetTimeout sets the request timeout
func (c *ControlClient) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
	c.initHTTPClient()
}

// RegisterSession registers a new session with the proxy daemon
func (c *ControlClient) RegisterSession(req RegisterSessionRequest) (*ControlSessionResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post("http://unix/sessions", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		var errMsg string
		json.NewDecoder(resp.Body).Decode(&errMsg)
		return nil, fmt.Errorf("registration failed (status %d): %s", resp.StatusCode, errMsg)
	}

	var result ControlSessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// UnregisterSession removes a session from the proxy daemon
func (c *ControlClient) UnregisterSession(sessionID string) error {
	req, err := http.NewRequest(http.MethodDelete, "http://unix/sessions/"+sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to control server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unregistration failed (status %d)", resp.StatusCode)
	}

	return nil
}

// ListSessions returns all active sessions
func (c *ControlClient) ListSessions() ([]*Session, error) {
	resp, err := c.httpClient.Get("http://unix/sessions")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list sessions failed (status %d)", resp.StatusCode)
	}

	var sessions []*Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return sessions, nil
}

// Health checks the daemon health status
func (c *ControlClient) Health() (*ControlHealthResponse, error) {
	resp, err := c.httpClient.Get("http://unix/health")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to control server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed (status %d)", resp.StatusCode)
	}

	var health ControlHealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &health, nil
}

// GetDefaultSocketPath returns the default socket path for the control API
func GetDefaultSocketPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".ubik", "proxy.sock"), nil
}
