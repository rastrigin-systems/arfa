package proxy

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/spf13/cobra"
)

// NewHealthCommand creates the health command with dependencies from the container.
func NewHealthCommand(_ *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Check security gateway health",
		Long:  "Check the health status of the security gateway including platform connectivity.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpproxy.NewDefaultControlClient()
			if err != nil {
				return fmt.Errorf("failed to create control client: %w", err)
			}

			health, err := client.Health()
			if err != nil {
				fmt.Println("Status: UNHEALTHY")
				fmt.Printf("Error:  %v\n", err)
				fmt.Println("\nThe security gateway is not responding.")
				fmt.Println("Run 'ubik proxy start' to start the gateway.")
				return nil
			}

			if health.Status == "ok" {
				fmt.Println("Status: HEALTHY")
			} else {
				fmt.Printf("Status: %s\n", health.Status)
			}
			fmt.Printf("Active Sessions: %d\n", health.ActiveSessions)
			fmt.Printf("Platform Connected: %v\n", health.PlatformHealthy)
			fmt.Printf("Uptime: %s\n", health.Uptime)

			if !health.PlatformHealthy {
				fmt.Println("\nWarning: Platform is not connected.")
				fmt.Println("Security policies cannot be synced. Fail-closed behavior may block requests.")
			}

			return nil
		},
	}
}
