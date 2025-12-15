package httpproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
)

// DaemonState represents the state of the proxy daemon
type DaemonState struct {
	PID        int       `json:"pid"`
	Port       int       `json:"port"`
	StartTime  time.Time `json:"start_time"`
	CertPath   string    `json:"cert_path"`
	SocketPath string    `json:"socket_path"` // Control API socket
	PortRange  struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"port_range"` // Session port range
}

// ProxyDaemon manages the singleton proxy daemon lifecycle
type ProxyDaemon struct {
	stateFile      string
	sockFile       string
	mu             sync.Mutex
	sessionManager *SessionManager
	controlServer  *ControlServer
}

// NewProxyDaemon creates a new proxy daemon manager
func NewProxyDaemon() (*ProxyDaemon, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	ubikDir := filepath.Join(homeDir, ".ubik")
	if err := os.MkdirAll(ubikDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create ubik directory: %w", err)
	}

	return &ProxyDaemon{
		stateFile: filepath.Join(ubikDir, "proxy.json"),
		sockFile:  filepath.Join(ubikDir, "proxy.sock"),
	}, nil
}

// IsRunning checks if the proxy daemon is currently running
func (d *ProxyDaemon) IsRunning() bool {
	state, err := d.GetState()
	if err != nil || state == nil {
		return false
	}

	// Check if process is actually running
	process, err := os.FindProcess(state.PID)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0 to check
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process not running, clean up state file
		d.cleanupStateFile()
		return false
	}

	// Also verify the port is actually listening
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", state.Port), time.Second)
	if err != nil {
		// Process exists but not listening, might be zombie
		return false
	}
	conn.Close()

	return true
}

// GetState reads the daemon state from disk
func (d *ProxyDaemon) GetState() (*DaemonState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := os.ReadFile(d.stateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state DaemonState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state file: %w", err)
	}

	return &state, nil
}

// saveState writes the daemon state to disk
func (d *ProxyDaemon) saveState(state *DaemonState) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(d.stateFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// cleanupStateFile removes the state file
func (d *ProxyDaemon) cleanupStateFile() {
	os.Remove(d.stateFile)
	os.Remove(d.sockFile)
}

// Start starts the proxy daemon if not already running
func (d *ProxyDaemon) Start(port int) error {
	if d.IsRunning() {
		state, _ := d.GetState()
		if state != nil && state.Port == port {
			fmt.Printf("✓ Proxy daemon already running on port %d (PID: %d)\n", state.Port, state.PID)
			return nil
		}
		// Different port requested, stop existing
		d.Stop()
	}

	// Find the ubik binary path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start daemon process
	cmd := exec.Command(execPath, "proxy", "run", "--port", fmt.Sprintf("%d", port))
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Detach from parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start proxy daemon: %w", err)
	}

	// Wait for daemon to be ready
	for i := 0; i < 50; i++ { // 5 seconds max
		time.Sleep(100 * time.Millisecond)
		if d.IsRunning() {
			state, _ := d.GetState()
			if state != nil {
				fmt.Printf("✓ Proxy daemon started on port %d (PID: %d)\n", state.Port, state.PID)
				return nil
			}
		}
	}

	return fmt.Errorf("proxy daemon failed to start within timeout")
}

// Stop stops the proxy daemon
func (d *ProxyDaemon) Stop() error {
	state, err := d.GetState()
	if err != nil {
		return err
	}

	if state == nil {
		return nil
	}

	// Send SIGTERM to the process
	process, err := os.FindProcess(state.PID)
	if err != nil {
		d.cleanupStateFile()
		return nil
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		d.cleanupStateFile()
		return nil
	}

	// Wait for process to exit
	for i := 0; i < 30; i++ { // 3 seconds max
		time.Sleep(100 * time.Millisecond)
		if err := process.Signal(syscall.Signal(0)); err != nil {
			// Process exited
			d.cleanupStateFile()
			fmt.Println("✓ Proxy daemon stopped")
			return nil
		}
	}

	// Force kill
	process.Kill()
	d.cleanupStateFile()
	fmt.Println("✓ Proxy daemon killed")

	return nil
}

// EnsureRunning starts the daemon if not running, returns the state
func (d *ProxyDaemon) EnsureRunning(port int) (*DaemonState, error) {
	if !d.IsRunning() {
		if err := d.Start(port); err != nil {
			return nil, err
		}
	}

	return d.GetState()
}

// GetSessionManager returns the session manager (available after daemon starts)
func (d *ProxyDaemon) GetSessionManager() *SessionManager {
	return d.sessionManager
}

// GetControlSocketPath returns the control socket path
func (d *ProxyDaemon) GetControlSocketPath() string {
	return d.sockFile
}

// DefaultProxyPort is the default port for the proxy daemon
const DefaultProxyPort = 8082

// DefaultMinPort is the starting port for session proxy listeners
const DefaultMinPort = 8100

// DefaultMaxPort is the ending port for session proxy listeners
const DefaultMaxPort = 8109

// RunDaemonConfig holds configuration for RunDaemon
type RunDaemonConfig struct {
	Port        int
	PlatformURL string // Platform API URL for token validation
}

// RunDaemon is called when starting the proxy in daemon mode
// This should be called from `ubik proxy run`
func (d *ProxyDaemon) RunDaemon(ctx context.Context, port int, logger logging.Logger) error {
	return d.RunDaemonWithConfig(ctx, RunDaemonConfig{Port: port}, logger)
}

// RunDaemonWithConfig is called when starting the proxy in daemon mode with full config
func (d *ProxyDaemon) RunDaemonWithConfig(ctx context.Context, cfg RunDaemonConfig, logger logging.Logger) error {
	// Create session manager for multi-session support
	d.sessionManager = NewSessionManager(DefaultMinPort, DefaultMaxPort)

	// Create policy engine for security enforcement
	policyEngine := NewPolicyEngine()
	// Note: Policy sync will be enabled when platform client is configured
	// For now, we set it as healthy to allow traffic (can be changed to fail-closed)
	policyEngine.SetPlatformHealthy(true)
	policyEngine.EnablePIIDetection(true)

	// Create and start proxy server
	server := NewProxyServer(logger)
	server.SetPolicyEngine(policyEngine)
	server.SetSessionManager(d.sessionManager)

	if err := server.Start(cfg.Port); err != nil {
		return fmt.Errorf("failed to start proxy server: %w", err)
	}

	// Create and start control server (Unix socket API)
	d.controlServer = NewControlServer(d.sockFile, d.sessionManager)
	d.controlServer.SetCertPath(server.GetCAPath())
	d.controlServer.SetPolicyEngine(policyEngine)
	d.controlServer.SetRequireToken(true) // Enable JWT validation
	if cfg.PlatformURL != "" {
		d.controlServer.SetPlatformURL(cfg.PlatformURL)
	}

	if err := d.controlServer.Start(ctx); err != nil {
		server.Stop(ctx)
		return fmt.Errorf("failed to start control server: %w", err)
	}

	// Start cleanup worker for stale sessions
	d.sessionManager.StartCleanupWorker(ctx)

	// Save state
	state := &DaemonState{
		PID:        os.Getpid(),
		Port:       cfg.Port,
		StartTime:  time.Now(),
		CertPath:   server.GetCAPath(),
		SocketPath: d.sockFile,
	}
	state.PortRange.Min = DefaultMinPort
	state.PortRange.Max = DefaultMaxPort

	if err := d.saveState(state); err != nil {
		d.controlServer.Stop()
		server.Stop(ctx)
		return fmt.Errorf("failed to save daemon state: %w", err)
	}

	fmt.Printf("Proxy daemon running on port %d (PID: %d)\n", cfg.Port, os.Getpid())
	fmt.Printf("Control socket: %s\n", d.sockFile)
	fmt.Printf("Session ports: %d-%d\n", DefaultMinPort, DefaultMaxPort)
	fmt.Printf("Security: JWT validation enabled, PII detection enabled\n")
	fmt.Printf("Stale session timeout: %v\n", StaleSessionTimeout)

	// Wait for context cancellation (caller handles signals)
	<-ctx.Done()

	// Cleanup
	policyEngine.Stop()
	d.controlServer.Stop()
	server.Stop(ctx)
	d.cleanupStateFile()

	return nil
}

// GetCertPath returns the path to the CA certificate
func (d *ProxyDaemon) GetCertPath() (string, error) {
	state, err := d.GetState()
	if err != nil {
		return "", err
	}
	if state == nil {
		return "", fmt.Errorf("proxy daemon not running")
	}
	return state.CertPath, nil
}

// GetPort returns the port the daemon is running on
func (d *ProxyDaemon) GetPort() (int, error) {
	state, err := d.GetState()
	if err != nil {
		return 0, err
	}
	if state == nil {
		return 0, fmt.Errorf("proxy daemon not running")
	}
	return state.Port, nil
}
