package status

import (
	"fmt"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command with dependencies from the container.
func NewStatusCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long:  "Display current authentication status and proxy configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			// Check authentication
			authenticated, err := authService.IsAuthenticated()
			if err != nil {
				return fmt.Errorf("failed to check authentication: %w", err)
			}

			if !authenticated {
				fmt.Println("Status: Not authenticated")
				fmt.Println("\nRun 'arfa login' to get started.")
				return nil
			}

			config, err := configManager.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			fmt.Println("Status: Authenticated")
			fmt.Printf("Platform:       %s\n", config.PlatformURL)

			// Get claims from JWT
			if claims, err := config.GetClaims(); err == nil {
				fmt.Printf("Employee ID:    %s\n", claims.EmployeeID)
			}

			fmt.Println("\nRun 'arfa proxy start' to start the security proxy.")

			return nil
		},
	}
}
