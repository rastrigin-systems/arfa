package config

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewConfigCommand creates the config command with dependencies from the container.
func NewConfigCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Manage local configuration",
		Long:  "View and manage local CLI configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
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
