package sync

import (
	"context"
	"fmt"
	"path/filepath"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewSyncCommand() *cobra.Command {
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
				ctx := context.Background()
				if err := syncService.StartContainers(ctx, workspace, apiKey); err != nil {
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
