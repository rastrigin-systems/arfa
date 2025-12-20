package interactive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/control"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/docker"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/sync"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/ui"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/workspace"
	"github.com/spf13/cobra"
)

// RunInteractiveMode starts an interactive session with an agent
func RunInteractiveMode(cmd *cobra.Command, args []string) error {
	// Get flags from command
	workspaceFlag, _ := cmd.Flags().GetString("workspace")
	agentFlag, _ := cmd.Flags().GetString("agent")
	pickFlag, _ := cmd.Flags().GetBool("pick")
	setDefaultFlag, _ := cmd.Flags().GetBool("set-default")

	return runInteractiveMode(workspaceFlag, agentFlag, pickFlag, setDefaultFlag)
}

func runInteractiveMode(workspaceFlag, agentFlag string, pickFlag, setDefaultFlag bool) error {
	// Initialize services
	configManager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	apiClient := api.NewClient("")
	authService := auth.NewService(configManager, apiClient)
	syncService := sync.NewService(configManager, apiClient, authService)

	// Ensure authenticated
	_, err = authService.RequireAuth()
	if err != nil {
		return err
	}

	// Get local agent configs
	agentConfigs, err := syncService.GetLocalAgentConfigs()
	if err != nil {
		return fmt.Errorf("failed to get agent configs: %w", err)
	}

	if len(agentConfigs) == 0 {
		fmt.Println("No agent configs found. Run 'ubik sync' to fetch configs from the platform.")
		return nil
	}

	// Initialize agent picker
	picker := ui.NewAgentPicker(configManager)

	// Select agent based on flags and defaults
	var selectedAgent *api.AgentConfig

	// Case 1: Explicit --agent flag (one-time override, no save)
	if agentFlag != "" {
		selectedAgent, err = syncService.GetAgentConfig(agentFlag)
		if err != nil {
			return fmt.Errorf("agent '%s' not found: %w", agentFlag, err)
		}
	} else if pickFlag || setDefaultFlag {
		// Case 2: --pick or --set-default flag - always show picker (even with single agent)
		selectedAgent, err = picker.SelectAgent(agentConfigs, true, true) // save as default, force interactive
		if err != nil {
			return err
		}

		// If --set-default, just save and exit (don't start session)
		if setDefaultFlag {
			fmt.Printf("\n✓ Default agent set to: %s\n", selectedAgent.AgentName)
			return nil
		}
	} else {
		// Case 3: No flags - use default or show picker
		config, _ := configManager.Load()

		if config != nil && config.DefaultAgent != "" {
			// Try to find the default agent
			if agent, err := syncService.GetAgentConfig(config.DefaultAgent); err == nil && agent.IsEnabled {
				selectedAgent = agent
			}
		}

		// If no valid default, show picker (and save selection)
		if selectedAgent == nil {
			// Count enabled agents
			enabledCount := 0
			for _, ac := range agentConfigs {
				if ac.IsEnabled {
					enabledCount++
				}
			}

			if enabledCount == 0 {
				return fmt.Errorf("no enabled agents found. Run 'ubik sync' to fetch configs.")
			}

			// If only one enabled agent, use it directly and save as default
			if enabledCount == 1 {
				for i := range agentConfigs {
					if agentConfigs[i].IsEnabled {
						selectedAgent = &agentConfigs[i]
						// Save as default silently
						if cfg, _ := configManager.Load(); cfg != nil {
							cfg.DefaultAgent = selectedAgent.AgentID
							configManager.Save(cfg)
						}
						break
					}
				}
			} else {
				// Multiple agents, show picker
				selectedAgent, err = picker.SelectAgent(agentConfigs, true, false) // save as default, don't force interactive
				if err != nil {
					return err
				}
			}
		}
	}

	fmt.Printf("✓ Agent: %s (%s)\n", selectedAgent.AgentName, selectedAgent.AgentType)
	if len(selectedAgent.MCPServers) > 0 {
		fmt.Printf("✓ MCP Servers: %d\n", len(selectedAgent.MCPServers))
	}

	// Get employee ID from config (stored during login)
	var employeeID string
	if cfg, _ := configManager.Load(); cfg != nil {
		employeeID = cfg.EmployeeID
	}

	// Get queue directory for log storage
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	queueDir := filepath.Join(home, ".ubik", "log_queue")

	// Initialize Control Service with API uploader
	var controlSvc *control.Service
	var controlProxy *control.ControlledProxy
	var sessionID string

	// Check for opt-out via environment variable
	if os.Getenv("UBIK_NO_LOGGING") == "" {
		// Create API uploader for the Control Service
		cliAPIClient := control.NewCLIAPIClient(apiClient)
		uploader := control.NewAPIUploader(cliAPIClient, employeeID, "")

		// Create Control Service
		controlSvc, err = control.NewService(control.ServiceConfig{
			EmployeeID:    employeeID,
			OrgID:         "", // TODO: Add OrgID to config when available
			AgentID:       selectedAgent.AgentID,
			QueueDir:      queueDir,
			FlushInterval: 5 * time.Second,
			MaxBatchSize:  10,
			Uploader:      uploader,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to initialize control service: %v\n", err)
		}
	}

	// Start Control Service and Proxy if enabled
	if controlSvc != nil {
		sessionID = controlSvc.SessionID()
		fmt.Printf("✓ Session: %s\n", sessionID)

		// Start background worker for log uploads
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go controlSvc.Start(ctx)

		// Start controlled proxy for HTTPS interception
		controlProxy = control.NewControlledProxy(controlSvc)
		if err := controlProxy.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to start proxy: %v\n", err)
		} else {
			defer controlProxy.Stop()
			fmt.Printf("✓ Proxy: localhost:%d\n", controlProxy.GetPort())
		}
	}

	// Workspace selection
	workspaceService := workspace.NewService()
	var workspacePath string

	if workspaceFlag != "" {
		// Validate provided workspace
		workspacePath = workspaceFlag
		if err := workspaceService.ValidatePath(workspacePath); err != nil {
			return fmt.Errorf("invalid workspace: %w", err)
		}
	} else {
		// Try to use current directory as workspace
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Check if CWD is valid - if so, use it directly (no prompt)
		if err := workspaceService.ValidatePath(cwd); err == nil {
			workspacePath = cwd
		} else {
			// CWD is not valid, prompt for selection
			workspacePath, err = workspaceService.SelectWorkspace(cwd)
			if err != nil {
				return fmt.Errorf("workspace selection failed: %w", err)
			}
		}
	}

	// Get and display workspace info
	workspaceInfo, err := workspaceService.GetWorkspaceInfo(workspacePath)
	if err != nil {
		return fmt.Errorf("failed to get workspace info: %w", err)
	}
	workspaceService.DisplayWorkspaceInfo(workspaceInfo)

	fmt.Println()

	// Check if agent binary is installed
	binaryPath, err := docker.FindAgentBinary(selectedAgent.AgentType)
	if err != nil {
		return err
	}
	fmt.Printf("✓ Agent binary: %s\n", binaryPath)

	// Start MCP servers if configured (Docker containers for MCP only)
	if len(selectedAgent.MCPServers) > 0 {
		fmt.Printf("Starting %d MCP server(s)...\n", len(selectedAgent.MCPServers))
		// TODO: Start MCP containers via Docker (keep Docker for MCP servers)
		// For now, skip MCP servers in native mode
		fmt.Println("  ⚠ MCP servers not yet supported in native mode")
	}

	// Configure native runner
	var proxyPort int
	var certPath string
	if controlProxy != nil {
		proxyPort = controlProxy.GetPort()
		certPath = controlProxy.GetCertPath()
	}

	runnerConfig := docker.RunnerConfig{
		AgentType: selectedAgent.AgentType,
		AgentID:   selectedAgent.AgentID,
		AgentName: selectedAgent.AgentName,
		Workspace: workspacePath,
		ProxyPort: proxyPort,
		CertPath:  certPath,
		SessionID: sessionID,
	}

	// Start interactive session
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Starting %s (native)\n", selectedAgent.AgentName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Run agent natively
	runCtx := context.Background()
	runner := docker.NewRunner()
	startTime := time.Now()

	err = runner.Run(runCtx, runnerConfig, os.Stdin, os.Stdout, os.Stderr)

	// Flush pending logs before exit
	if controlSvc != nil {
		controlSvc.Stop()
	}

	// Display session summary
	duration := time.Since(startTime).Round(time.Second)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Session:\n")
	fmt.Printf("  Agent:     %s\n", selectedAgent.AgentName)
	fmt.Printf("  Directory: %s\n", workspacePath)
	fmt.Printf("  Duration:  %s\n", duration)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return err
}
