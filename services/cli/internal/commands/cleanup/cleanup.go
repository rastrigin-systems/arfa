package cleanup

import (
	"fmt"
	"os"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewCleanupCommand() *cobra.Command {
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
