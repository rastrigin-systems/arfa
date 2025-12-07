package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
	"github.com/spf13/cobra"
)

var version = "v0.2.0-dev"

func main() {
	if err := newRootCommand().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCommand() *cobra.Command {
	var (
		workspace string
		agentName string
	)

	rootCmd := &cobra.Command{
		Use:   "ubik",
		Short: "ubik - Container-orchestrated AI agent management",
		Long: `ubik CLI enables employees to use AI coding agents with centrally-managed
configurations from the platform. It manages Docker containers that run
Claude Code and MCP servers with injected configs.`,
		Version: version,
		// Run function executes when no subcommand is provided (interactive mode)
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractiveMode(workspace, agentName)
		},
	}

	// Add flags for interactive mode
	rootCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace directory (interactive prompt if not provided)")
	rootCmd.Flags().StringVar(&agentName, "agent", "", "Agent to use (uses default if not specified)")

	// Add subcommands
	rootCmd.AddCommand(newLoginCommand())
	rootCmd.AddCommand(newLogoutCommand())
	rootCmd.AddCommand(newSyncCommand())
	rootCmd.AddCommand(newConfigCommand())
	rootCmd.AddCommand(newStatusCommand())
	rootCmd.AddCommand(newStartCommand())
	rootCmd.AddCommand(newStopCommand())
	rootCmd.AddCommand(newAgentsCommand())
	rootCmd.AddCommand(newSkillsCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newCleanupCommand())

	return rootCmd
}

func newLoginCommand() *cobra.Command {
	var (
		platformURL string
		email       string
		password    string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the platform",
		Long:  "Login to the platform and store authentication token locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient(platformURL)
			authService := cli.NewAuthService(configManager, platformClient)

			// Use interactive login if credentials not provided via flags
			if email == "" || password == "" {
				return authService.LoginInteractive()
			}

			// Non-interactive login
			if err := authService.Login(platformURL, email, password); err != nil {
				return err
			}

			fmt.Println("✓ Authenticated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&platformURL, "url", "https://api.ubik.io", "Platform URL")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")

	return cmd
}

func newLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear credentials",
		Long:  "Remove stored authentication token and logout from the platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			return authService.Logout()
		},
	}
}

func newSyncCommand() *cobra.Command {
	var (
		startContainers bool
		workspace       string
		apiKey          string
	)

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync configs from platform",
		Long: `Fetches resolved configs from the platform and stores them locally.
Optionally starts Docker containers for agents and MCP servers.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)
			syncService := cli.NewSyncService(configManager, platformClient, authService)

			// Sync configs
			result, err := syncService.Sync()
			if err != nil {
				return err
			}

			fmt.Printf("\n✓ Sync completed at %s\n", result.UpdatedAt.Format("2006-01-02 15:04:05"))

			// Start containers if requested
			if startContainers {
				// Set default workspace if not provided
				if workspace == "" {
					workspace = "."
				}

				// Convert to absolute path
				absWorkspace, err := filepath.Abs(workspace)
				if err != nil {
					return fmt.Errorf("failed to resolve workspace path: %w", err)
				}
				workspace = absWorkspace

				// Setup Docker client
				dockerClient, err := cli.NewDockerClient()
				if err != nil {
					return fmt.Errorf("failed to create Docker client: %w", err)
				}
				defer dockerClient.Close()

				syncService.SetDockerClient(dockerClient)

				// Start containers
				if err := syncService.StartContainers(workspace, apiKey); err != nil {
					return fmt.Errorf("failed to start containers: %w", err)
				}

				fmt.Println("\n✓ Containers started successfully")
				fmt.Println("\nNext steps:")
				fmt.Println("  1. Run 'ubik status' to see container status")
				fmt.Println("  2. Run 'ubik stop' to stop containers")
			} else {
				fmt.Println("\nNext steps:")
				fmt.Println("  1. Run 'ubik start' to start containers")
				fmt.Println("  2. Run 'ubik status' to see container status")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&startContainers, "start-containers", false, "Start Docker containers after sync")
	cmd.Flags().StringVar(&workspace, "workspace", ".", "Workspace directory to mount in containers")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Anthropic API key for agents")

	return cmd
}

func newConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage local configuration",
		Long:  "View and manage local CLI configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			config, err := configManager.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if config.Token == "" {
				fmt.Println("Not authenticated. Run 'ubik login' first.")
				return nil
			}

			fmt.Printf("Platform URL:   %s\n", config.PlatformURL)
			fmt.Printf("Employee ID:    %s\n", config.EmployeeID)
			fmt.Printf("Default Agent:  %s\n", config.DefaultAgent)
			if !config.LastSync.IsZero() {
				fmt.Printf("Last Sync:      %s\n", config.LastSync.Format("2006-01-02 15:04:05"))
			}
			fmt.Printf("\nConfig Path:    %s\n", configManager.GetConfigPath())

			return nil
		},
	}
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long:  "Display current authentication status, agent configs, and running containers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)
			syncService := cli.NewSyncService(configManager, platformClient, authService)

			// Check authentication
			authenticated, err := authService.IsAuthenticated()
			if err != nil {
				return fmt.Errorf("failed to check authentication: %w", err)
			}

			if !authenticated {
				fmt.Println("Status: Not authenticated")
				fmt.Println("\nRun 'ubik login' to get started.")
				return nil
			}

			config, err := configManager.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Status: Authenticated")
			fmt.Printf("Platform:       %s\n", config.PlatformURL)
			fmt.Printf("Employee ID:    %s\n", config.EmployeeID)

			// Show local agent configs
			agentConfigs, err := syncService.GetLocalAgentConfigs()
			if err != nil {
				return fmt.Errorf("failed to get agent configs: %w", err)
			}

			if len(agentConfigs) == 0 {
				fmt.Println("\nNo agent configs found. Run 'ubik sync' to fetch configs.")
			} else {
				fmt.Printf("\nAgent Configs:  %d\n", len(agentConfigs))
				for _, ac := range agentConfigs {
					status := "disabled"
					if ac.IsEnabled {
						status = "enabled"
					}
					fmt.Printf("  • %s (%s) - %s\n", ac.AgentName, ac.AgentType, status)
					if len(ac.MCPServers) > 0 {
						fmt.Printf("    MCP Servers: %d\n", len(ac.MCPServers))
					}
				}
			}

			// Show container status
			fmt.Println()
			dockerClient, err := cli.NewDockerClient()
			if err != nil {
				fmt.Println("Docker Containers: (Docker not available)")
				return nil
			}
			defer dockerClient.Close()

			syncService.SetDockerClient(dockerClient)
			containers, err := syncService.GetContainerStatus()
			if err != nil {
				fmt.Printf("Docker Containers: (failed to get status: %v)\n", err)
				return nil
			}

			if len(containers) == 0 {
				fmt.Println("Docker Containers: None running")
				fmt.Println("\nRun 'ubik start' to start containers")
			} else {
				fmt.Printf("Docker Containers: %d\n", len(containers))
				for _, c := range containers {
					status := "●"
					if c.State == "running" {
						status = "🟢"
					} else {
						status = "⚪"
					}
					fmt.Printf("  %s %s (%s) - %s\n", status, c.Name, c.Image, c.Status)
				}
			}

			return nil
		},
	}
}

func newStartCommand() *cobra.Command {
	var (
		workspace string
		apiKey    string
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start Docker containers",
		Long:  "Start Docker containers for synced agent configs and MCP servers.",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			// Setup Docker client
			dockerClient, err := cli.NewDockerClient()
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			defer dockerClient.Close()

			syncService.SetDockerClient(dockerClient)

			// Set default workspace if not provided
			if workspace == "" {
				workspace = "."
			}

			// Convert to absolute path
			absWorkspace, err := filepath.Abs(workspace)
			if err != nil {
				return fmt.Errorf("failed to resolve workspace path: %w", err)
			}
			workspace = absWorkspace

			// Start containers
			if err := syncService.StartContainers(workspace, apiKey); err != nil {
				return fmt.Errorf("failed to start containers: %w", err)
			}

			fmt.Println("\n✓ Containers started successfully")
			fmt.Println("\nRun 'ubik status' to see container status")

			return nil
		},
	}

	cmd.Flags().StringVar(&workspace, "workspace", ".", "Workspace directory to mount in containers")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Anthropic API key for agents")

	return cmd
}

func newStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop Docker containers",
		Long:  "Stop all running ubik-managed Docker containers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)
			syncService := cli.NewSyncService(configManager, platformClient, authService)

			// Setup Docker client
			dockerClient, err := cli.NewDockerClient()
			if err != nil {
				return fmt.Errorf("failed to create Docker client: %w", err)
			}
			defer dockerClient.Close()

			syncService.SetDockerClient(dockerClient)

			// Stop containers
			if err := syncService.StopContainers(); err != nil {
				return fmt.Errorf("failed to stop containers: %w", err)
			}

			fmt.Println("\n✓ All containers stopped")

			return nil
		},
	}
}

func newAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage AI agents",
		Long:  "View available AI agents and manage agent access.",
	}

	cmd.AddCommand(newAgentsListCommand())
	cmd.AddCommand(newAgentsInfoCommand())
	cmd.AddCommand(newAgentsRequestCommand())
	cmd.AddCommand(commands.NewAgentsShowCommand())

	return cmd
}

func newSkillsCommand() *cobra.Command {
	configManager, err := cli.NewConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create config manager: %v\n", err)
		os.Exit(1)
	}

	platformClient := cli.NewPlatformClient("")
	authService := cli.NewAuthService(configManager, platformClient)

	return commands.NewSkillsCommand(configManager, platformClient, authService)
}

func newAgentsListCommand() *cobra.Command {
	var showLocal bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available agents",
		Long:  "Display all available AI agents from the platform catalog or locally configured agents.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// If showing local agents, no need to authenticate
			if showLocal {
				agentService := cli.NewAgentService(nil, configManager)
				agents, err := agentService.GetLocalAgents()
				if err != nil {
					return fmt.Errorf("failed to get local agents: %w", err)
				}

				if len(agents) == 0 {
					fmt.Println("No local agents configured. Run 'ubik sync' to fetch configs from the platform.")
					return nil
				}

				fmt.Printf("\nConfigured Agents (%d):\n\n", len(agents))

				// Create table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tID")
				fmt.Fprintln(w, "────\t────\t──────\t──")

				for _, agent := range agents {
					enabledStatus := "✓ enabled"
					if !agent.IsEnabled {
						enabledStatus = "✗ disabled"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.AgentName, agent.AgentType, enabledStatus, agent.AgentID)
				}

				w.Flush()
				fmt.Println()
				fmt.Println("💡 Tip: Use 'ubik agents show <name>' to see configuration for your agents")
				fmt.Println()

				return nil

			}

			// For platform agents, require authentication
			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)
			agents, err := agentService.ListAgents()
			if err != nil {
				return fmt.Errorf("failed to list agents: %w", err)
			}

			if len(agents) == 0 {
				fmt.Println("No agents available in the platform catalog.")
				return nil
			}

			fmt.Printf("\nAvailable Agents (%d):\n\n", len(agents))

			// Create table writer
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPROVIDER\tID\tDESCRIPTION")
			fmt.Fprintln(w, "────\t────────\t──\t───────────")

			for _, agent := range agents {
				description := agent.Description
				if len(description) > 60 {
					description = description[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.Name, agent.Provider, agent.ID, description)
			}

			w.Flush()
			fmt.Println()
			fmt.Println("💡 Tip: Use 'ubik agents info <id>' to see agent details")
			fmt.Println("        Use 'ubik agents show <name>' to see configuration for assigned agents")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&showLocal, "local", false, "Show locally configured agents only")

	return cmd
}

func newAgentsInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <agent-id>",
		Short: "Get agent details",
		Long:  "Display detailed information about a specific AI agent.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)
			agent, err := agentService.GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent info: %w", err)
			}

			fmt.Printf("\nAgent: %s\n", agent.Name)
			fmt.Printf("Provider: %s\n", agent.Provider)
			fmt.Printf("Description: %s\n", agent.Description)
			fmt.Printf("Pricing: %s\n", agent.PricingTier)
			fmt.Printf("ID: %s\n", agent.ID)

			if len(agent.SupportedPlatforms) > 0 {
				fmt.Printf("Platforms: ")
				for i, platform := range agent.SupportedPlatforms {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", platform)
				}
				fmt.Println()
			}

			fmt.Printf("\nCreated: %s\n", agent.CreatedAt.Format("2006-01-02"))
			fmt.Printf("Updated: %s\n", agent.UpdatedAt.Format("2006-01-02"))
			fmt.Println()

			return nil
		},
	}
}

func newAgentsRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "request <agent-id>",
		Short: "Request access to an agent",
		Long:  "Request access to an AI agent by creating an employee agent configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			config, err := authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)

			// Request the agent
			if err := agentService.RequestAgent(config.EmployeeID, agentID); err != nil {
				return fmt.Errorf("failed to request agent: %w", err)
			}

			fmt.Printf("\n✓ Agent access requested successfully\n")
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  1. Run 'ubik sync' to pull the new agent configuration\n")
			fmt.Printf("  2. Run 'ubik agents list --local' to see your configured agents\n\n")

			return nil
		},
	}
}

func newUpdateCommand() *cobra.Command {
	var autoSync bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for configuration updates",
		Long:  "Check if there are configuration updates available from the platform and optionally sync them.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			config, err := authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)
			syncService := cli.NewSyncService(configManager, platformClient, authService)

			fmt.Println("Checking for updates...")

			hasUpdates, err := agentService.CheckForUpdates(config.EmployeeID)
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !hasUpdates {
				fmt.Println("\n✓ Your configuration is up to date")
				return nil
			}

			fmt.Println("\n⚠ Updates available!")

			if autoSync {
				fmt.Println("\nSyncing updates...")
				if _, err := syncService.Sync(); err != nil {
					return fmt.Errorf("failed to sync: %w", err)
				}
				fmt.Println("\n✓ Configuration updated successfully")
			} else {
				fmt.Println("\nRun 'ubik sync' to apply updates")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&autoSync, "sync", false, "Automatically sync updates if available")

	return cmd
}

func newCleanupCommand() *cobra.Command {
	var removeContainers bool
	var removeConfig bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up containers and local state",
		Long:  "Remove Docker containers and optionally reset local configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			if removeContainers {
				platformClient := cli.NewPlatformClient("")
				authService := cli.NewAuthService(configManager, platformClient)
				syncService := cli.NewSyncService(configManager, platformClient, authService)

				// Setup Docker client
				dockerClient, err := cli.NewDockerClient()
				if err != nil {
					return fmt.Errorf("failed to create Docker client: %w", err)
				}
				defer dockerClient.Close()

				syncService.SetDockerClient(dockerClient)

				fmt.Println("Stopping and removing containers...")
				if err := syncService.StopContainers(); err != nil {
					fmt.Printf("Warning: failed to stop some containers: %v\n", err)
				}

				// TODO: Also remove containers, not just stop them
				// For now, stopping is sufficient

				fmt.Println("✓ Containers stopped")
			}

			if removeConfig {
				configPath := configManager.GetConfigPath()
				if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove config: %w", err)
				}
				fmt.Println("✓ Local configuration removed")
			}

			if !removeContainers && !removeConfig {
				fmt.Println("Nothing to clean up. Use --remove-containers or --remove-config")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&removeContainers, "remove-containers", false, "Stop and remove all Docker containers")
	cmd.Flags().BoolVar(&removeConfig, "remove-config", false, "Remove local configuration file")

	return cmd
}

// runInteractiveMode starts an interactive session with an agent
func runInteractiveMode(workspaceFlag, agentFlag string) error {
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

	// Initialize logger (silently fails if disabled or opt-out)
	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		MaxRetries:    5,
		RetryBackoff:  1 * time.Second,
	}
	apiClient := logging.NewPlatformAPIClient(platformClient)
	logger, err := logging.NewLogger(loggerConfig, apiClient)
	if err != nil {
		// Log error but continue - logging is optional
		fmt.Fprintf(os.Stderr, "Warning: failed to initialize logging: %v\n", err)
	}

	// Ensure logger is closed on exit
	if logger != nil {
		defer logger.Close()
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

	// Select agent (use flag if provided, otherwise use default from config or first available)
	var selectedAgent *cli.AgentConfig
	
	// 1. Priority: Command line flag
	if agentFlag != "" {
		selectedAgent, err = syncService.GetAgentConfig(agentFlag)
		if err != nil {
			return fmt.Errorf("agent '%s' not found: %w", agentFlag, err)
		}
	} else {
		// 2. Priority: Default agent from config
		config, err := configManager.Load()
		if err == nil && config.DefaultAgent != "" {
			// Try to find the default agent
			// We don't error if not found, just fall back to first enabled
			if agent, err := syncService.GetAgentConfig(config.DefaultAgent); err == nil && agent.IsEnabled {
				selectedAgent = agent
			}
		}

		// 3. Priority: First enabled agent
		if selectedAgent == nil {
			for _, ac := range agentConfigs {
				if ac.IsEnabled {
					selectedAgent = &ac
					break
				}
			}
		}
		
		if selectedAgent == nil {
			return fmt.Errorf("no enabled agents found. Run 'ubik sync' to fetch configs.")
		}
	}

	fmt.Printf("✓ Agent: %s (%s)\n", selectedAgent.AgentName, selectedAgent.AgentType)
	if len(selectedAgent.MCPServers) > 0 {
		fmt.Printf("✓ MCP Servers: %d\n", len(selectedAgent.MCPServers))
	}

	// Start logging session
	if logger != nil {
		logger.SetAgentID(selectedAgent.AgentID)
		sessionID := logger.StartSession()
		fmt.Printf("✓ Session ID: %s\n", sessionID.String())
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

	// Setup Docker client
	dockerClient, err := cli.NewDockerClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer dockerClient.Close()

	// Check if Docker is running
	if err := dockerClient.Ping(); err != nil {
		return fmt.Errorf("Docker is not running. Please start Docker and try again.")
	}

	syncService.SetDockerClient(dockerClient)

	// Check if containers are running
	containerStatus, err := syncService.GetContainerStatus()
	if err != nil {
		return fmt.Errorf("failed to get container status: %w", err)
	}

	// Find agent container by name pattern
	var agentContainerID string
	expectedContainerName := fmt.Sprintf("ubik-agent-%s", selectedAgent.AgentID)
	for _, c := range containerStatus {
		// Check if container name matches (removing leading "/" if present)
		containerName := c.Name
		if len(containerName) > 0 && containerName[0] == '/' {
			containerName = containerName[1:]
		}
		if containerName == expectedContainerName && c.State == "running" {
			agentContainerID = c.ID
			break
		}
	}

	// If container not running, start it
	if agentContainerID == "" {
		fmt.Println("🚀 Starting containers...")

		// Start containers (will start MCP servers and agent)
		apiKey := os.Getenv("ANTHROPIC_API_KEY")
		if err := syncService.StartContainers(workspace, apiKey); err != nil {
			return fmt.Errorf("failed to start containers: %w", err)
		}

		// Get container ID after starting
		containerStatus, err = syncService.GetContainerStatus()
		if err != nil {
			return fmt.Errorf("failed to get container status after start: %w", err)
		}

		for _, c := range containerStatus {
			containerName := c.Name
			if len(containerName) > 0 && containerName[0] == '/' {
				containerName = containerName[1:]
			}
			if containerName == expectedContainerName && c.State == "running" {
				agentContainerID = c.ID
				break
			}
		}

		if agentContainerID == "" {
			return fmt.Errorf("failed to start agent container")
		}
	}

	// Setup proxy service
	proxyService := cli.NewProxyService()
	proxyService.SetDockerClient(dockerClient)

	// Create proxy options
	proxyOptions := cli.ProxyOptions{
		ContainerID: agentContainerID,
		AgentName:   selectedAgent.AgentName,
		WorkingDir:  workspace,
		Logger:      logger,
	}

	// Start interactive session
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Interactive session started\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Attach to container and proxy I/O
	ctx := context.Background()
	session, err := proxyService.ExecuteInteractive(ctx, proxyOptions)

	// End logging session
	if logger != nil {
		logger.EndSession()
		logger.Flush()
	}

	// Display session summary
	if session != nil {
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(session.String())
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	return err
}
