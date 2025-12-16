package proxy

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/spf13/cobra"
)

func NewStopCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy daemon",
		Long:  "Stop the MITM proxy daemon if it's running.",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create proxy daemon: %w", err)
			}

			if !daemon.IsRunning() {
				fmt.Println("Proxy daemon is not running")
				return nil
			}

			return daemon.Stop()
		},
	}
}
