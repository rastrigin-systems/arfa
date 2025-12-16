package interactive

import (
	"context"
	"fmt"
	"os"
	"time"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
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
	configManager, err := cli.NewConfigManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	platformClient := cli.NewPlatformClient("")
	authService := cli.NewAuthService(configManager, platformClient)
	syncService := cli.NewSyncService(configManager, platformClient, authService)

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
	picker := cli.NewAgentPicker(configManager)

	// Select agent based on flags and defaults
	var selectedAgent *cli.AgentConfig

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

	// Initialize logger (silently fails if disabled or opt-out)
	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		MaxRetries:    5,
		RetryBackoff:  1 * time.Second,
	}
	apiClient := cli.NewPlatformAPIClient(platformClient)
	logger, err := logging.NewLogger(loggerConfig, apiClient)
	if err != nil {
		// Log error but continue - logging is optional
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logging: %v\n", err)
	}

	// Ensure logger is closed on exit
	if logger != nil {
		defer logger.Close()
	}

	// Ensure proxy daemon is running (auto-start if needed)
	proxyDaemon, err := httpproxy.NewProxyDaemon()
	if err != nil {
		return fmt.Errorf("failed to create proxy daemon: %w", err)
	}

	proxyState, err := proxyDaemon.EnsureRunning(httpproxy.DefaultProxyPort)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to start proxy daemon: %v\n", err)
		// Continue without proxy - logging will still work for CLI I/O
	} else {
		fmt.Printf("✓ Proxy: localhost:%d\n", proxyState.Port)
	}

	// Start logging session
	var sessionID string
	if logger != nil {
		logger.SetAgentID(selectedAgent.AgentID)
		sid := logger.StartSession()
		sessionID = sid.String()
		fmt.Printf("✓ Session: %s\n", sessionID)
	}

	// Workspace selection
	workspaceService := cli.NewWorkspaceService()
	var workspace string

	if workspaceFlag != "" {
		// Validate provided workspace
		workspace = workspaceFlag
		if err := workspaceService.ValidatePath(workspace); err != nil {
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
			workspace = cwd
		} else {
			// CWD is not valid, prompt for selection
			workspace, err = workspaceService.SelectWorkspace(cwd)
			if err != nil {
				return fmt.Errorf("workspace selection failed: %w", err)
			}
		}
	}

	// Get and display workspace info
	workspaceInfo, err := workspaceService.GetWorkspaceInfo(workspace)
	if err != nil {
		return fmt.Errorf("failed to get workspace info: %w", err)
	}
	workspaceService.DisplayWorkspaceInfo(workspaceInfo)

	fmt.Println()

	// Check if agent binary is installed
	binaryPath, err := cli.FindAgentBinary(selectedAgent.AgentType)
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
	if proxyState != nil {
		proxyPort = proxyState.Port
		certPath = proxyState.CertPath
	}

	// Get config for JWT token (needed for proxy session registration)
	cliConfig, _ := configManager.Load()
	var jwtToken string
	if cliConfig != nil {
		jwtToken = cliConfig.Token
	}

	runnerConfig := cli.NativeRunnerConfig{
		AgentType: selectedAgent.AgentType,
		AgentID:   selectedAgent.AgentID,
		AgentName: selectedAgent.AgentName,
		Workspace: workspace,
		ProxyPort: proxyPort,
		CertPath:  certPath,
		SessionID: sessionID,
		Token:     jwtToken, // Pass JWT token for proxy authentication
	}

	// Start interactive session
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Starting %s (native)\n", selectedAgent.AgentName)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Run agent natively
	ctx := context.Background()
	runner := cli.NewNativeRunner()
	startTime := time.Now()

	err = runner.Run(ctx, runnerConfig, os.Stdin, os.Stdout, os.Stderr)

	// End logging session
	if logger != nil {
		logger.EndSession()
		logger.Flush()
	}

	// Display session summary
	duration := time.Since(startTime).Round(time.Second)
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Session:\n")
	fmt.Printf("  Agent:     %s\n", selectedAgent.AgentName)
	fmt.Printf("  Directory: %s\n", workspace)
	fmt.Printf("  Duration:  %s\n", duration)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return err
}
