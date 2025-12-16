package proxy

import (
	"fmt"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command with dependencies from the container.
func NewStatusCommand(_ *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show proxy daemon status",
		Long:  "Display the current status of the MITM proxy daemon.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if !daemon.IsRunning() {
				fmt.Println("Status: Not running")
				fmt.Println("\nRun 'ubik proxy start' to start the daemon")
				return nil
			}

			state, err := daemon.GetState()
			if err != nil {
				return fmt.Errorf("failed to get daemon state: %w", err)
			}

			fmt.Println("Status: Running")
			fmt.Printf("PID:    %d\n", state.PID)
			fmt.Printf("Port:   %d\n", state.Port)
			fmt.Printf("Uptime: %s\n", time.Since(state.StartTime).Round(time.Second))
			fmt.Printf("Cert:   %s\n", state.CertPath)

			return nil
		},
	}
}
