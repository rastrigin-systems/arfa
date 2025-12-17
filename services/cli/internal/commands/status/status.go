package status

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/docker"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command with dependencies from the container.
func NewStatusCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long:  "Display current authentication status, agent configs, and proxy daemon status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			syncService, err := c.SyncService()
			if err != nil {
				return fmt.Errorf("failed to get sync service: %w", err)
			}

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
					if _, err := docker.FindAgentBinary(ac.AgentType); err != nil {
						binaryStatus = " (not installed)"
					}

					fmt.Printf("  • %s (%s) - %s%s\n", ac.AgentName, ac.AgentType, status, binaryStatus)
					if len(ac.MCPServers) > 0 {
						fmt.Printf("    MCP Servers: %d\n", len(ac.MCPServers))
					}
				}
			}

			return nil
		},
	}
}
