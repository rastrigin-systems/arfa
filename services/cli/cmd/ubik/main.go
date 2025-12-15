package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/pkg/types"
	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logparser"
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
		workspace  string
		agentName  string
		pick       bool
		setDefault bool
	)

	rootCmd := &cobra.Command{
		Use:   "ubik",
		Short: "ubik - Container-orchestrated AI agent management",
		Long: `ubik CLI enables employees to use AI coding agents with centrally-managed
configurations from the platform. It manages Docker containers that run
Claude Code and MCP servers with injected configs.

When run without subcommands, ubik starts an interactive session with your
default agent. If no default is set, you'll be prompted to select one.

Examples:
  ubik                    Run default agent (or pick if none set)
  ubik --pick             Always show agent picker
  ubik --agent "Claude"   Run specific agent (one-time, no save)
  ubik --set-default      Set default agent without starting`,
		Version: version,
		// Run function executes when no subcommand is provided (interactive mode)
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInteractiveMode(workspace, agentName, pick, setDefault)
		},
	}

	// Add flags for interactive mode
	rootCmd.Flags().StringVar(&workspace, "workspace", "", "Workspace directory (interactive prompt if not provided)")
	rootCmd.Flags().StringVar(&agentName, "agent", "", "Agent to use (one-time override, doesn't change default)")
	rootCmd.Flags().BoolVar(&pick, "pick", false, "Always show agent picker (saves selection as default)")
	rootCmd.Flags().BoolVar(&setDefault, "set-default", false, "Set default agent without starting a session")

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
	rootCmd.AddCommand(newLogsCommand())  // Add the new logs command
	rootCmd.AddCommand(newProxyCommand()) // Add proxy daemon management

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

			// If no URL provided via flag, check config for saved platform URL
			if platformURL == "" || platformURL == "https://api.ubik.io" {
				config, err := configManager.Load()
				if err == nil && config.PlatformURL != "" {
					platformURL = config.PlatformURL
				} else if platformURL == "" {
					platformURL = "https://api.ubik.io" // Final fallback
				}
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

	cmd.Flags().StringVar(&platformURL, "url", "", "Platform URL (defaults to saved URL or https://api.ubik.io)")
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
		Long:  "Display current authentication status, agent configs, and proxy daemon status.",
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

					// Check if agent binary is installed
					binaryStatus := ""
					if _, err := cli.FindAgentBinary(ac.AgentType); err != nil {
						binaryStatus = " (not installed)"
					}

					fmt.Printf("  • %s (%s) - %s%s\n", ac.AgentName, ac.AgentType, status, binaryStatus)
					if len(ac.MCPServers) > 0 {
						fmt.Printf("    MCP Servers: %d\n", len(ac.MCPServers))
					}
				}
			}

			// Show proxy daemon status
			fmt.Println()
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				fmt.Printf("Proxy Daemon: (failed to check: %v)\n", err)
				return nil
			}

			if !daemon.IsRunning() {
				fmt.Println("Proxy Daemon: Not running")
				fmt.Println("\nRun 'ubik' to start an interactive session (proxy auto-starts)")
			} else {
				state, err := daemon.GetState()
				if err != nil {
					fmt.Printf("Proxy Daemon: Running (failed to get details: %v)\n", err)
				} else {
					fmt.Println("Proxy Daemon: Running")
					fmt.Printf("  Port:   %d\n", state.Port)
					fmt.Printf("  PID:    %d\n", state.PID)
					fmt.Printf("  Uptime: %s\n", time.Since(state.StartTime).Round(time.Second))
				}
			}

			return nil
		},
	}
}

func newStartCommand() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy daemon",
		Long: `Start the MITM proxy daemon in the background.

The proxy daemon intercepts and logs all LLM API calls (Anthropic, OpenAI, Google).
It runs as a shared singleton - all CLI sessions use the same proxy.

To start an interactive agent session, simply run 'ubik' without arguments.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Start proxy daemon
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if err := daemon.Start(port); err != nil {
				return fmt.Errorf("failed to start proxy daemon: %w", err)
			}

			fmt.Println("\nNext steps:")
			fmt.Println("  1. Run 'ubik' to start an interactive agent session")
			fmt.Println("  2. Run 'ubik proxy status' to check proxy status")
			fmt.Println("  3. Run 'ubik proxy stop' to stop the proxy daemon")

			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", httpproxy.DefaultProxyPort, "Port for the proxy to listen on")

	return cmd
}

func newStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy daemon",
		Long:  "Stop the MITM proxy daemon if it's running.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if !daemon.IsRunning() {
				fmt.Println("Proxy daemon is not running")
				return nil
			}

			if err := daemon.Stop(); err != nil {
				return fmt.Errorf("failed to stop proxy daemon: %w", err)
			}

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

			// Get user's enabled agents to show status
			syncService := cli.NewSyncService(configManager, platformClient, authService)
			enabledAgents, _ := syncService.GetLocalAgentConfigs()

			// Build map of enabled agent IDs
			enabledMap := make(map[string]bool)
			for _, ea := range enabledAgents {
				if ea.IsEnabled {
					enabledMap[ea.AgentID] = true
				}
			}

			fmt.Printf("\nAvailable Agents (%d):\n\n", len(agents))

			// Create table writer
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPROVIDER\tSTATUS\tDESCRIPTION")
			fmt.Fprintln(w, "────\t────────\t──────\t───────────")

			for _, agent := range agents {
				description := agent.Description
				if len(description) > 50 {
					description = description[:47] + "..."
				}
				status := "- not assigned"
				if enabledMap[agent.ID] {
					status = "✓ enabled"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.Name, agent.Provider, status, description)
			}

			w.Flush()
			fmt.Println()
			fmt.Println("💡 Tip: Use 'ubik agents info <id>' to see agent details")
			fmt.Println("        Use 'ubik' or 'ubik --pick' to start an agent")
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

	// Register session with proxy daemon for API request logging
	if proxyState != nil && sessionID != "" {
		homeDir, _ := os.UserHomeDir()
		sockPath := filepath.Join(homeDir, ".ubik", "proxy.sock")
		if err := httpproxy.RegisterSession(sockPath, sessionID, selectedAgent.AgentID); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to register session with proxy: %v\n", err)
		}
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

	runnerConfig := cli.NativeRunnerConfig{
		AgentType: selectedAgent.AgentType,
		AgentID:   selectedAgent.AgentID,
		AgentName: selectedAgent.AgentName,
		Workspace: workspace,
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

func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage and stream activity logs",
		Long:  "View activity logs from the platform, including real-time streaming.",
	}

	cmd.AddCommand(newLogsStreamCommand()) // Add the stream subcommand
	cmd.AddCommand(newLogsViewCommand())   // Add the view subcommand

	return cmd
}

func newLogsViewCommand() *cobra.Command {
	var (
		format    string
		sessionID string
		noEmoji   bool
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View classified logs in human-readable format",
		Long: `View activity logs parsed and classified into human-readable entries.

Displays logs categorized as:
  - USER_PROMPT: User input to the AI
  - AI_TEXT: AI text responses
  - TOOL_CALL: Tools invoked by the AI
  - TOOL_RESULT: Results from tool execution
  - ERROR: Any errors that occurred

Examples:
  ubik logs view                    # View logs in pretty format
  ubik logs view --format=json      # View logs as JSON
  ubik logs view --no-emoji         # Disable emoji icons`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Get classified logs from the current session or storage
			logs, err := cli.GetClassifiedLogs(configManager, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			if len(logs) == 0 {
				fmt.Println("No classified logs found.")
				fmt.Println("Tip: Run 'ubik sync' to start a session and generate logs.")
				return nil
			}

			// Format and display
			formatter := logparser.DefaultFormatter()
			formatter.UseEmoji = !noEmoji

			switch format {
			case "json":
				// Output as JSON
				data, err := json.MarshalIndent(logs, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal logs: %w", err)
				}
				fmt.Println(string(data))

			case "pretty", "":
				// Output in human-readable format
				if sessionID != "" {
					fmt.Println(formatter.FormatSession(sessionID, logs))
				} else {
					// Group by session
					sessions := make(map[string][]types.ClassifiedLogEntry)
					for _, log := range logs {
						sessions[log.SessionID] = append(sessions[log.SessionID], log)
					}
					for sid, sessionLogs := range sessions {
						fmt.Println(formatter.FormatSession(sid, sessionLogs))
						fmt.Println()
					}
				}

			default:
				return fmt.Errorf("unknown format: %s (use 'pretty' or 'json')", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "pretty", "Output format: pretty, json")
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Filter by session ID")
	cmd.Flags().BoolVar(&noEmoji, "no-emoji", false, "Disable emoji icons in output")

	return cmd
}

func newLogsStreamCommand() *cobra.Command {
	var (
		follow  bool
		jsonOut bool
		verbose bool
		// TODO: Add flags for session_id, agent_id, employee_id for filtering
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream real-time activity logs",
		Long:  "Connects to the platform via WebSocket to stream activity logs in real-time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")

			logStreamer := cli.NewLogStreamer(platformClient, configManager)
			logStreamer.SetJSONOutput(jsonOut)
			logStreamer.SetVerbose(verbose)

			return logStreamer.StreamLogs(context.Background())
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs in real-time") // For consistency, though it's always following
	cmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output full JSON for each log entry")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full request/response payloads")

	return cmd
}

func newProxyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage the MITM proxy daemon",
		Long:  "Control the shared MITM proxy daemon that intercepts and logs LLM API calls.",
	}

	cmd.AddCommand(newProxyStartCommand())
	cmd.AddCommand(newProxyStopCommand())
	cmd.AddCommand(newProxyStatusCommand())
	cmd.AddCommand(newProxyRunCommand())

	return cmd
}

func newProxyStartCommand() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy daemon",
		Long:  "Start the MITM proxy daemon in the background. The daemon is shared across all CLI sessions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			return daemon.Start(port)
		},
	}

	cmd.Flags().IntVar(&port, "port", httpproxy.DefaultProxyPort, "Port for the proxy to listen on")

	return cmd
}

func newProxyStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy daemon",
		Long:  "Stop the MITM proxy daemon if it's running.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if !daemon.IsRunning() {
				fmt.Println("Proxy daemon is not running")
				return nil
			}

			return daemon.Stop()
		},
	}
}

func newProxyStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show proxy daemon status",
		Long:  "Display the current status of the MITM proxy daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if !daemon.IsRunning() {
				fmt.Println("Status: Not running")
				fmt.Println("\nRun 'ubik proxy start' to start the daemon")
				return nil
			}

			state, err := daemon.GetState()
			if err != nil {
				return fmt.Errorf("failed to get daemon state: %w", err)
			}

			fmt.Println("Status: Running")
			fmt.Printf("PID:    %d\n", state.PID)
			fmt.Printf("Port:   %d\n", state.Port)
			fmt.Printf("Uptime: %s\n", time.Since(state.StartTime).Round(time.Second))
			fmt.Printf("Cert:   %s\n", state.CertPath)

			return nil
		},
	}
}

func newProxyRunCommand() *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:    "run",
		Short:  "Run the proxy daemon (internal)",
		Long:   "Run the MITM proxy in the foreground. This is called internally by 'proxy start'.",
		Hidden: true, // Hide from help since it's internal
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logger (optional - daemon mode)
			loggerConfig := &logging.Config{
				Enabled:       true,
				BatchSize:     100,
				BatchInterval: 5 * time.Second,
				MaxRetries:    5,
				RetryBackoff:  1 * time.Second,
			}

			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			apiClient := cli.NewPlatformAPIClient(platformClient)
			logger, err := logging.NewLogger(loggerConfig, apiClient)
			if err != nil {
				// Continue without logging - log warning to stderr
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize logging: %v\n", err)
			}

			_ = configManager // Silence unused warning for now

			// Create daemon manager
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create daemon manager: %w", err)
			}

			// Run until interrupted
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down proxy daemon...")
				cancel()
			}()

			// Run the daemon (this saves state and blocks)
			if err := daemon.RunDaemon(ctx, port, logger); err != nil {
				return fmt.Errorf("daemon error: %w", err)
			}

			// Cleanup
			if logger != nil {
				logger.Close()
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", httpproxy.DefaultProxyPort, "Port for the proxy to listen on")

	return cmd
}
