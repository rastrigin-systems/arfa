package sync

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewSyncCommand creates the sync command with dependencies from the container.
func NewSyncCommand(c *container.Container) *cobra.Command {
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
			syncService, err := c.SyncService()
			if err != nil {
				return fmt.Errorf("failed to get sync service: %w", err)
			}

			ctx := context.Background()

			// Sync configs
			result, err := syncService.Sync(ctx)
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
				dockerClient, err := c.DockerClient()
				if err != nil {
					return fmt.Errorf("failed to get Docker client: %w", err)
				}

				syncService.SetDockerClient(dockerClient)

				// Start containers
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
