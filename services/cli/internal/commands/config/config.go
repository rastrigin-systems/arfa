package config

import (
	"fmt"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
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
				fmt.Println("Not authenticated. Run 'arfa login' first.")
				return nil
			}

			fmt.Printf("Platform URL:   %s\n", config.PlatformURL)

			// Get claims from JWT
			if claims, err := config.GetClaims(); err == nil {
				fmt.Printf("Employee ID:    %s\n", claims.EmployeeID)
				fmt.Printf("Org ID:         %s\n", claims.OrgID)
				fmt.Printf("Token Expires:  %s\n", claims.ExpiresAt.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("\nConfig Path:    %s\n", configManager.GetConfigPath())

			return nil
		},
	}
}
