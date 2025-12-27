package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/control"
	"github.com/spf13/cobra"
)

// NewStartCommand creates the start command.
func NewStartCommand(c *container.Container) *cobra.Command {
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

	// Check if proxy is already running
	pidFile, err := control.NewPIDFile()
	if err != nil {
		return fmt.Errorf("failed to create PID file manager: %w", err)
	}

	status, err := pidFile.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to check proxy status: %w", err)
	}

	if status.Running {
		return fmt.Errorf("proxy is already running (PID %d on port %d). Use 'arfa stop' to stop it first",
			status.Info.PID, status.Info.Port)
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
	defer func() { _ = controlProxy.Stop() }()

	port := controlProxy.GetPort()
	certPath := controlProxy.GetCertPath()

	// Write PID file
	if err := pidFile.Write(control.ProxyInfo{
		PID:       os.Getpid(),
		Port:      port,
		StartedAt: time.Now(),
		CertPath:  certPath,
	}); err != nil {
		fmt.Printf("Warning: Failed to write PID file: %v\n", err)
	}
	defer func() { _ = pidFile.Remove() }()

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
