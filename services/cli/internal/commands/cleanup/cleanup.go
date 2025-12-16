package cleanup

import (
	"context"
	"fmt"
	"os"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewCleanupCommand creates the cleanup command with dependencies from the container.
func NewCleanupCommand(c *container.Container) *cobra.Command {
	var removeContainers bool
	var removeConfig bool

	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up containers and local state",
		Long:  "Remove Docker containers and optionally reset local configuration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			if removeContainers {
				syncService, err := c.SyncService()
				if err != nil {
					return fmt.Errorf("failed to get sync service: %w", err)
				}

				// Setup Docker client
				dockerClient, err := c.DockerClient()
				if err != nil {
					return fmt.Errorf("failed to get Docker client: %w", err)
				}

				syncService.SetDockerClient(dockerClient)

				fmt.Println("Stopping and removing containers...")
				ctx := context.Background()
				if err := syncService.StopContainers(ctx); err != nil {
					fmt.Printf("Warning: failed to stop some containers: %v\n", err)
				}

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
