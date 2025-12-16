package proxy

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/spf13/cobra"
)

// NewStartCommand creates the start command with dependencies from the container.
func NewStartCommand(_ *container.Container) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy daemon",
		Long:  "Start the MITM proxy daemon in the background. The daemon is shared across all CLI sessions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			return daemon.Start(port)
		},
	}

	cmd.Flags().IntVar(&port, "port", httpproxy.DefaultProxyPort, "Port for the proxy to listen on")

	return cmd
}
