package update

import (
	"context"
	"fmt"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the update command with dependencies from the container.
func NewUpdateCommand(c *container.Container) *cobra.Command {
	var autoSync bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for configuration updates",
		Long:  "Check if there are configuration updates available from the platform and optionally sync them.",
		RunE: func(cmd *cobra.Command, args []string) error {
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			config, err := authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService, err := c.AgentService()
			if err != nil {
				return fmt.Errorf("failed to get agent service: %w", err)
			}

			syncService, err := c.SyncService()
			if err != nil {
				return fmt.Errorf("failed to get sync service: %w", err)
			}

			ctx := context.Background()

			fmt.Println("Checking for updates...")

			hasUpdates, err := agentService.CheckForUpdates(ctx, config.EmployeeID)
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
				if _, err := syncService.Sync(ctx); err != nil {
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
