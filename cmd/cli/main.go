package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sergeirastrigin/ubik-enterprise/internal/cli"
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

	// Get local agent configs
	agentConfigs, err := syncService.GetLocalAgentConfigs()
	if err != nil {
		return fmt.Errorf("failed to get agent configs: %w", err)
	}

	if len(agentConfigs) == 0 {
		fmt.Println("No agent configs found. Run 'ubik sync' to fetch configs from the platform.")
		return nil
	}

	// Select agent (use flag if provided, otherwise use default)
	var selectedAgent *cli.AgentConfig
	if agentFlag != "" {
		selectedAgent, err = syncService.GetAgentConfig(agentFlag)
		if err != nil {
			return fmt.Errorf("agent '%s' not found: %w", agentFlag, err)
		}
	} else {
		// Use default agent (first enabled agent)
		for _, ac := range agentConfigs {
			if ac.IsEnabled {
				selectedAgent = &ac
				break
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
		// Interactive workspace selection
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		workspace, err = workspaceService.SelectWorkspace(cwd)
		if err != nil {
			return fmt.Errorf("workspace selection failed: %w", err)
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
	}

	// Start interactive session
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("✨ Interactive session started\n")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Attach to container and proxy I/O
	ctx := context.Background()
	session, err := proxyService.ExecuteInteractive(ctx, proxyOptions)

	// Display session summary
	if session != nil {
		fmt.Println()
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(session.String())
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	}

	return err
}
