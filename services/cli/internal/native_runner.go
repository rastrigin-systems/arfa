package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"
)

// NativeRunner manages native process execution for agents
type NativeRunner struct {
	cmd       *exec.Cmd
	pty       *os.File
	workspace string
	proxyPort int
	certPath  string
	sessionID string
	agentID   string
	stopped   bool
	mu        sync.Mutex
}

// NativeRunnerConfig contains configuration for starting an agent
type NativeRunnerConfig struct {
	AgentType   string
	AgentID     string
	AgentName   string
	Workspace   string
	APIKey      string
	ProxyPort   int
	CertPath    string
	SessionID   string
	Environment map[string]string // Additional env vars from agent config
}

// AgentBinary maps agent types to their binary names
var AgentBinaries = map[string][]string{
	"claude-code": {"claude"},
	"cursor":      {"cursor"},
	"windsurf":    {"windsurf"},
	"gemini":      {"gemini"},
	"aider":       {"aider"},
}

// AgentEnvVars maps agent types to their API key environment variable names
var AgentEnvVars = map[string]string{
	"claude-code": "ANTHROPIC_API_KEY",
	"cursor":      "ANTHROPIC_API_KEY",
	"windsurf":    "ANTHROPIC_API_KEY",
	"gemini":      "GEMINI_API_KEY",
	"aider":       "ANTHROPIC_API_KEY",
}

// NewNativeRunner creates a new native runner instance
func NewNativeRunner() *NativeRunner {
	return &NativeRunner{}
}

// FindAgentBinary locates the agent binary in the system PATH
func FindAgentBinary(agentType string) (string, error) {
	binaries, ok := AgentBinaries[agentType]
	if !ok {
		return "", fmt.Errorf("unknown agent type: %s", agentType)
	}

	for _, name := range binaries {
		path, err := exec.LookPath(name)
		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("agent binary not found for %s. Please install it first.\n\nInstallation instructions:\n%s",
		agentType, getInstallInstructions(agentType))
}

// getInstallInstructions returns installation instructions for an agent
func getInstallInstructions(agentType string) string {
	instructions := map[string]string{
		"claude-code": "  npm install -g @anthropic-ai/claude-code\n  # or: brew install claude-code",
		"cursor":      "  Download from https://cursor.sh",
		"windsurf":    "  Download from https://windsurf.dev",
		"aider":       "  pip install aider-chat\n  # or: brew install aider",
		"gemini":      "  Visit https://cloud.google.com/gemini for CLI access",
	}

	if inst, ok := instructions[agentType]; ok {
		return inst
	}
	return "  Visit the agent's website for installation instructions"
}

// Start launches the agent as a native process
func (r *NativeRunner) Start(ctx context.Context, config NativeRunnerConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Find the agent binary
	binaryPath, err := FindAgentBinary(config.AgentType)
	if err != nil {
		return err
	}

	r.workspace = config.Workspace
	r.proxyPort = config.ProxyPort
	r.certPath = config.CertPath
	r.sessionID = config.SessionID
	r.agentID = config.AgentID

	// Build command
	r.cmd = exec.CommandContext(ctx, binaryPath)
	r.cmd.Dir = config.Workspace

	// Build environment
	env := os.Environ()

	// Add API key
	if config.APIKey != "" {
		envVar := AgentEnvVars[config.AgentType]
		if envVar == "" {
			envVar = "ANTHROPIC_API_KEY" // Default
		}
		env = append(env, fmt.Sprintf("%s=%s", envVar, config.APIKey))
	}

	// Add proxy configuration
	if config.ProxyPort > 0 {
		proxyURL := fmt.Sprintf("http://localhost:%d", config.ProxyPort)
		env = append(env,
			fmt.Sprintf("HTTP_PROXY=%s", proxyURL),
			fmt.Sprintf("HTTPS_PROXY=%s", proxyURL),
			fmt.Sprintf("http_proxy=%s", proxyURL),
			fmt.Sprintf("https_proxy=%s", proxyURL),
		)

		// Add CA certificate paths for various runtimes
		if config.CertPath != "" {
			env = append(env,
				fmt.Sprintf("NODE_EXTRA_CA_CERTS=%s", config.CertPath),
				fmt.Sprintf("REQUESTS_CA_BUNDLE=%s", config.CertPath),
				fmt.Sprintf("SSL_CERT_FILE=%s", config.CertPath),
				fmt.Sprintf("CURL_CA_BUNDLE=%s", config.CertPath),
			)
		}
	}

	// Add session tracking headers for proxy
	if config.SessionID != "" {
		env = append(env, fmt.Sprintf("UBIK_SESSION_ID=%s", config.SessionID))
	}
	if config.AgentID != "" {
		env = append(env, fmt.Sprintf("UBIK_AGENT_ID=%s", config.AgentID))
	}

	// Add any additional environment variables from config
	for k, v := range config.Environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	r.cmd.Env = env

	// Start with PTY for interactive mode
	r.pty, err = pty.Start(r.cmd)
	if err != nil {
		return fmt.Errorf("failed to start agent with PTY: %w", err)
	}

	return nil
}

// Run executes the agent and handles I/O proxying
func (r *NativeRunner) Run(ctx context.Context, config NativeRunnerConfig, stdin io.Reader, stdout, stderr io.Writer) error {
	// Start the process
	if err := r.Start(ctx, config); err != nil {
		return err
	}

	// Get terminal state
	stdinFd := int(os.Stdin.Fd())
	isTerminal := term.IsTerminal(stdinFd)

	// Set terminal to raw mode for proper TTY behavior
	var oldState *term.State
	if isTerminal {
		var err error
		oldState, err = term.MakeRaw(stdinFd)
		if err != nil {
			return fmt.Errorf("failed to set terminal to raw mode: %w", err)
		}
		defer func() {
			_ = term.Restore(stdinFd, oldState)
		}()
	}

	// Handle terminal resize
	resizeChan := make(chan os.Signal, 1)
	signal.Notify(resizeChan, syscall.SIGWINCH)
	defer signal.Stop(resizeChan)

	go func() {
		for range resizeChan {
			r.resizePty()
		}
	}()

	// Initial resize
	r.resizePty()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Create error channels for I/O goroutines
	errChan := make(chan error, 2)
	doneChan := make(chan struct{})

	// Copy stdin to PTY
	go func() {
		_, err := io.Copy(r.pty, stdin)
		if err != nil && err != io.EOF {
			select {
			case errChan <- fmt.Errorf("stdin error: %w", err):
			case <-doneChan:
			}
		}
	}()

	// Copy PTY output to stdout
	go func() {
		_, err := io.Copy(stdout, r.pty)
		if err != nil && err != io.EOF {
			select {
			case errChan <- fmt.Errorf("stdout error: %w", err):
			case <-doneChan:
			}
		}
		// When PTY closes, signal we're done
		select {
		case errChan <- nil:
		case <-doneChan:
		}
	}()

	// Wait for completion
	select {
	case sig := <-sigChan:
		// Restore terminal before printing
		if isTerminal && oldState != nil {
			_ = term.Restore(stdinFd, oldState)
		}
		fmt.Fprintf(stderr, "\n\nReceived signal %v, stopping agent...\n", sig)
		r.Stop()
		return nil

	case <-ctx.Done():
		r.Stop()
		return ctx.Err()

	case err := <-errChan:
		close(doneChan)
		// Wait for process to exit
		r.cmd.Wait()
		return err
	}
}

// resizePty resizes the PTY to match the terminal size
func (r *NativeRunner) resizePty() {
	if r.pty == nil {
		return
	}

	// Get terminal size
	size, err := pty.GetsizeFull(os.Stdin)
	if err != nil {
		return
	}

	// Resize PTY
	pty.Setsize(r.pty, size)
}

// Stop terminates the agent process
func (r *NativeRunner) Stop() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.stopped {
		return nil
	}
	r.stopped = true

	if r.pty != nil {
		r.pty.Close()
	}

	if r.cmd != nil && r.cmd.Process != nil {
		// Try graceful shutdown first
		r.cmd.Process.Signal(syscall.SIGTERM)

		// Wait a bit for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- r.cmd.Wait()
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(3 * time.Second):
			// Force kill
			r.cmd.Process.Kill()
		}
	}

	return nil
}

// Wait waits for the process to complete
func (r *NativeRunner) Wait() error {
	if r.cmd == nil {
		return nil
	}
	return r.cmd.Wait()
}

// IsRunning returns true if the process is still running
func (r *NativeRunner) IsRunning() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}

	// Check if process is still running
	err := r.cmd.Process.Signal(syscall.Signal(0))
	return err == nil
}

// PID returns the process ID of the running agent
func (r *NativeRunner) PID() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cmd == nil || r.cmd.Process == nil {
		return 0
	}
	return r.cmd.Process.Pid
}

// ProcessInfo represents information about a running agent process
type ProcessInfo struct {
	PID       int
	AgentID   string
	AgentName string
	AgentType string
	Workspace string
	SessionID string
	StartTime time.Time
}

// ProcessManager tracks running agent processes
type ProcessManager struct {
	processes map[int]*ProcessInfo
	mu        sync.RWMutex
}

// NewProcessManager creates a new process manager
func NewProcessManager() *ProcessManager {
	return &ProcessManager{
		processes: make(map[int]*ProcessInfo),
	}
}

// Register adds a process to the manager
func (pm *ProcessManager) Register(info *ProcessInfo) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.processes[info.PID] = info
}

// Unregister removes a process from the manager
func (pm *ProcessManager) Unregister(pid int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	delete(pm.processes, pid)
}

// List returns all registered processes
func (pm *ProcessManager) List() []*ProcessInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*ProcessInfo, 0, len(pm.processes))
	for _, p := range pm.processes {
		result = append(result, p)
	}
	return result
}

// GetByPID returns a process by its PID
func (pm *ProcessManager) GetByPID(pid int) *ProcessInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.processes[pid]
}

// SaveToFile persists process info to disk
func (pm *ProcessManager) SaveToFile() error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	processFile := filepath.Join(homeDir, ".ubik", "processes.json")

	// Create directory if needed
	os.MkdirAll(filepath.Dir(processFile), 0700)

	// Write process info
	// TODO: Implement JSON marshaling
	return nil
}

// LoadFromFile loads process info from disk
func (pm *ProcessManager) LoadFromFile() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	processFile := filepath.Join(homeDir, ".ubik", "processes.json")

	// Check if file exists
	if _, err := os.Stat(processFile); os.IsNotExist(err) {
		return nil
	}

	// TODO: Implement JSON unmarshaling
	return nil
}
