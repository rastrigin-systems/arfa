package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"golang.org/x/term"
)

// ProxyService handles I/O proxying between CLI and container
type ProxyService struct {
	dockerClient *DockerClient
}

// SessionInfo tracks information about an interactive session
type SessionInfo struct {
	ContainerID string
	AgentName   string
	WorkingDir  string
	StartTime   time.Time
	EndTime     time.Time
}

// ProxyOptions configures the proxy behavior
type ProxyOptions struct {
	ContainerID string
	AgentName   string
	WorkingDir  string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

// NewProxyService creates a new proxy service
func NewProxyService() *ProxyService {
	return &ProxyService{}
}

// SetDockerClient sets the Docker client for the proxy service
func (ps *ProxyService) SetDockerClient(client *DockerClient) {
	ps.dockerClient = client
}

// StartSession initializes a new session
func (ps *ProxyService) StartSession(containerID, agentName, workingDir string) *SessionInfo {
	return &SessionInfo{
		ContainerID: containerID,
		AgentName:   agentName,
		WorkingDir:  workingDir,
		StartTime:   time.Now(),
	}
}

// EndSession marks a session as ended
func (ps *ProxyService) EndSession(session *SessionInfo) {
	session.EndTime = time.Now()
}

// Duration returns the duration of the session
func (si *SessionInfo) Duration() time.Duration {
	if si.EndTime.IsZero() {
		return time.Since(si.StartTime)
	}
	return si.EndTime.Sub(si.StartTime)
}

// String returns a formatted string representation of the session
func (si *SessionInfo) String() string {
	duration := si.Duration()
	return fmt.Sprintf(
		"Session:\n"+
			"  Container: %s\n"+
			"  Agent:     %s\n"+
			"  Directory: %s\n"+
			"  Duration:  %s\n",
		si.ContainerID[:12], // Show first 12 chars of container ID
		si.AgentName,
		si.WorkingDir,
		duration.Round(time.Second),
	)
}

// Validate checks if proxy options are valid
func (po *ProxyOptions) Validate() error {
	if po.ContainerID == "" {
		return fmt.Errorf("container ID is required")
	}
	if po.AgentName == "" {
		return fmt.Errorf("agent name is required")
	}
	return nil
}

// AttachToContainer attaches to a running container and proxies I/O
func (ps *ProxyService) AttachToContainer(ctx context.Context, options ProxyOptions) error {
	if ps.dockerClient == nil {
		return fmt.Errorf("Docker client not configured")
	}

	if err := options.Validate(); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	// Use default streams if not provided
	if options.Stdin == nil {
		options.Stdin = os.Stdin
	}
	if options.Stdout == nil {
		options.Stdout = os.Stdout
	}
	if options.Stderr == nil {
		options.Stderr = os.Stderr
	}

	// Check if stdin is a terminal
	stdinFd := int(os.Stdin.Fd())
	isTerminal := term.IsTerminal(stdinFd)

	// If stdin is a terminal, set it to raw mode for proper TTY behavior
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

	// Create attach options
	attachOpts := container.AttachOptions{
		Stream: true,
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		Logs:   false,
	}

	// Attach to container
	resp, err := ps.dockerClient.cli.ContainerAttach(ctx, options.ContainerID, attachOpts)
	if err != nil {
		return fmt.Errorf("failed to attach to container: %w", err)
	}
	defer resp.Close()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	// Create error channel for streaming errors
	errChan := make(chan error, 2)

	// Start goroutine to copy stdin to container
	go func() {
		_, err := io.Copy(resp.Conn, options.Stdin)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("stdin copy error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Start goroutine to copy container output to stdout
	go func() {
		// For TTY containers, output is not multiplexed - just copy directly
		_, err := io.Copy(options.Stdout, resp.Reader)
		if err != nil && err != io.EOF {
			errChan <- fmt.Errorf("output copy error: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for either:
	// - Signal (Ctrl+C)
	// - Context cancellation
	// - Streaming error
	select {
	case sig := <-sigChan:
		// Restore terminal before printing
		if isTerminal && oldState != nil {
			_ = term.Restore(stdinFd, oldState)
		}
		fmt.Fprintf(os.Stderr, "\n\nReceived signal %v, detaching...\n", sig)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

// ExecuteInteractive starts a container in interactive mode and attaches to it
func (ps *ProxyService) ExecuteInteractive(ctx context.Context, options ProxyOptions) (*SessionInfo, error) {
	if ps.dockerClient == nil {
		return nil, fmt.Errorf("Docker client not configured")
	}

	if err := options.Validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// Start session
	session := ps.StartSession(options.ContainerID, options.AgentName, options.WorkingDir)

	// Display session info
	fmt.Printf("✓ Starting interactive session with %s\n", options.AgentName)
	fmt.Printf("✓ Container: %s\n", options.ContainerID[:12])
	fmt.Printf("✓ Working directory: %s\n", options.WorkingDir)
	fmt.Println()

	// Attach to container and proxy I/O
	err := ps.AttachToContainer(ctx, options)

	// End session
	ps.EndSession(session)

	if err != nil && err != context.Canceled {
		return session, fmt.Errorf("proxy error: %w", err)
	}

	// Display session summary
	fmt.Println()
	fmt.Printf("✓ Session ended\n")
	fmt.Printf("✓ Duration: %s\n", session.Duration().Round(time.Second))

	return session, nil
}
