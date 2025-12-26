package proxy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/control"
	"github.com/spf13/cobra"
)

// NewProxyCommand creates the proxy command group
func NewProxyCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage the AI agent security proxy",
		Long: `The proxy command manages the HTTPS proxy that intercepts
AI agent traffic for logging and policy enforcement.

Examples:
  arfa proxy start       Start the proxy server
  arfa proxy stop        Stop the proxy server
  arfa proxy status      Show proxy status
  arfa proxy health      Check proxy health`,
	}

	cmd.AddCommand(newStartCommand(c))
	cmd.AddCommand(newStopCommand(c))
	cmd.AddCommand(newStatusCommand(c))
	cmd.AddCommand(newHealthCommand(c))
	cmd.AddCommand(newEnvCommand(c))

	return cmd
}

// RunProxyStart is the default action when arfa is run without subcommands
func RunProxyStart(cmd *cobra.Command, args []string) error {
	return runStart(cmd, args)
}

func newStartCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the proxy server",
		Long: `Start the HTTPS proxy server for AI agent traffic interception.

The proxy will:
- Listen on localhost:8082 (default)
- Intercept HTTPS traffic from AI agents (Claude Code, Cursor, etc.)
- Log all tool usage to the platform
- Enforce security policies (block dangerous operations)

To use the proxy with AI agents:
  export HTTPS_PROXY=http://localhost:8082
  export NODE_EXTRA_CA_CERTS=~/.arfa/certs/ca.pem
  claude  # Now proxied

Or run 'arfa setup' to auto-configure your AI tools.`,
		RunE: runStart,
	}
}

func newStopCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop()
		},
	}
}

func runStop() error {
	pidFile := getPidFilePath()

	// Read PID file
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No proxy running (PID file not found)")
			return nil
		}
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		// Clean up invalid PID file
		_ = os.Remove(pidFile)
		return fmt.Errorf("invalid PID file: %w", err)
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		_ = os.Remove(pidFile)
		fmt.Println("No proxy running (process not found)")
		return nil
	}

	// Send SIGTERM for graceful shutdown
	fmt.Printf("Stopping proxy (PID %d)...\n", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// Process might already be dead
		_ = os.Remove(pidFile)
		fmt.Println("Proxy already stopped")
		return nil
	}

	// Wait briefly for graceful shutdown
	time.Sleep(500 * time.Millisecond)

	// Check if still running, force kill if needed
	if err := process.Signal(syscall.Signal(0)); err == nil {
		fmt.Println("Force stopping...")
		_ = process.Signal(syscall.SIGKILL)
	}

	_ = os.Remove(pidFile)
	fmt.Println("✓ Proxy stopped")
	return nil
}

func getPidFilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".arfa", "proxy.pid")
}

func newStatusCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show proxy status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Check if proxy is running via PID file
			fmt.Println("Proxy status: Not implemented yet")
			fmt.Println("Run 'arfa proxy start' to start the proxy")
			return nil
		},
	}
}

func newHealthCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check proxy health",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Ping proxy health endpoint
			fmt.Println("Proxy health check: Not implemented yet")
			return nil
		},
	}
}

func newEnvCommand(c *container.Container) *cobra.Command {
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Output environment variables for proxy configuration",
		Long: `Output shell export statements for configuring the proxy.

Use with eval to set environment variables in your shell:
  eval $(arfa proxy env)

For CI pipelines:
  # GitHub Actions
  arfa proxy env >> $GITHUB_ENV

  # GitLab CI / generic
  eval $(arfa proxy env)

Formats:
  shell   - Shell export statements (default)
  github  - GitHub Actions format (KEY=VALUE)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnv(formatFlag)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "shell", "Output format: shell, github")

	return cmd
}

func runEnv(format string) error {
	// Get home directory for cert path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	certPath := filepath.Join(home, ".arfa", "certs", "arfa-ca.pem")

	// Default proxy URL - assumes proxy is running on default port
	// TODO: Could check PID file or try to detect running proxy port
	proxyURL := "http://127.0.0.1:8082"

	switch format {
	case "github":
		// GitHub Actions format: KEY=VALUE (one per line, append to $GITHUB_ENV)
		fmt.Printf("HTTPS_PROXY=%s\n", proxyURL)
		fmt.Printf("SSL_CERT_FILE=%s\n", certPath)
		fmt.Printf("NODE_EXTRA_CA_CERTS=%s\n", certPath)
	case "shell":
		fallthrough
	default:
		// Shell export format
		fmt.Printf("export HTTPS_PROXY=\"%s\"\n", proxyURL)
		fmt.Printf("export SSL_CERT_FILE=\"%s\"\n", certPath)
		fmt.Printf("export NODE_EXTRA_CA_CERTS=\"%s\"\n", certPath)
	}

	return nil
}

func runStart(cmd *cobra.Command, args []string) error {
	// Initialize services
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	apiClient := api.NewClient("")
	authService := auth.NewService(configManager, apiClient)

	// Ensure authenticated
	_, err = authService.RequireAuth()
	if err != nil {
		return err
	}

	// Get employee ID and org ID from JWT claims
	var employeeID, orgID string
	if cfg, _ := configManager.Load(); cfg != nil {
		if claims, err := cfg.GetClaims(); err == nil {
			employeeID = claims.EmployeeID
			orgID = claims.OrgID
		}
	}

	// Get queue directory for log storage
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	queueDir := filepath.Join(home, ".arfa", "log_queue")

	// Create uploader for sending logs to API
	var uploader control.Uploader
	if os.Getenv("ARFA_NO_LOGGING") == "" {
		cliAPIClient := control.NewCLIAPIClient(apiClient)
		uploader = control.NewAPIUploader(cliAPIClient, employeeID, "")
	}

	// Create Control Service
	// Client is detected automatically from User-Agent headers
	controlSvc, err := control.NewService(control.ServiceConfig{
		EmployeeID:    employeeID,
		OrgID:         orgID,
		QueueDir:      queueDir,
		FlushInterval: 5 * time.Second,
		MaxBatchSize:  10,
		Uploader:      uploader,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize control service: %w", err)
	}

	// Start background worker for log uploads
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go controlSvc.Start(ctx)

	// Enable real-time policy updates via WebSocket
	cfg, _ := configManager.Load()
	if cfg != nil && cfg.Token != "" && cfg.PlatformURL != "" {
		fmt.Println("Connecting to policy server...")
		if err := controlSvc.EnableRealtimePolicies(ctx, cfg.PlatformURL, cfg.Token); err != nil {
			fmt.Printf("Warning: Failed to enable real-time policies: %v\n", err)
			fmt.Println("Falling back to cached policies")
		} else {
			// Wait for initial policies (with 10 second timeout)
			waitCtx, waitCancel := context.WithTimeout(ctx, 10*time.Second)
			if err := controlSvc.WaitForPolicies(waitCtx, 10*time.Second); err != nil {
				fmt.Printf("Warning: Timeout waiting for policies: %v\n", err)
				fmt.Println("Using cached policies, real-time updates will continue in background")
			} else {
				if pc := controlSvc.PolicyClient(); pc != nil {
					fmt.Printf("✓ Connected to policy server (%d policies loaded)\n", pc.PolicyCount())
				}
			}
			waitCancel()
		}
	}

	// Start controlled proxy
	controlProxy := control.NewControlledProxy(controlSvc)
	if err := controlProxy.Start(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// Write PID file for proxy stop command
	pidFile := getPidFilePath()
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		fmt.Printf("Warning: Failed to write PID file: %v\n", err)
	}

	defer func() {
		_ = controlProxy.Stop()
		_ = os.Remove(pidFile) // Clean up PID file on exit
	}()

	port := controlProxy.GetPort()
	certPath := controlProxy.GetCertPath()

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✓ Proxy started on localhost:%d\n", port)
	fmt.Printf("✓ Certificate: %s\n", certPath)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println("To use with AI agents:")
	fmt.Printf("  export HTTPS_PROXY=http://localhost:%d\n", port)
	fmt.Printf("  export NODE_EXTRA_CA_CERTS=%s\n", certPath)
	fmt.Println()
	fmt.Println("Or run 'arfa setup' to auto-configure your AI tools.")
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down...")

	// Flush pending logs
	controlSvc.Stop()

	fmt.Println("✓ Proxy stopped")
	return nil
}
