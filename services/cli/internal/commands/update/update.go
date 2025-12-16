package update

import (
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewUpdateCommand() *cobra.Command {
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
