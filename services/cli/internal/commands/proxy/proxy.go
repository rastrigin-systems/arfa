package proxy

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewProxyCommand creates the proxy command group with dependencies from the container.
func NewProxyCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage the MITM proxy daemon",
		Long:  "Control the shared MITM proxy daemon that intercepts and logs LLM API calls.",
	}

	cmd.AddCommand(NewStartCommand(c))
	cmd.AddCommand(NewStopCommand(c))
	cmd.AddCommand(NewStatusCommand(c))
	cmd.AddCommand(NewRunCommand(c))
	cmd.AddCommand(NewSessionsCommand(c))
	cmd.AddCommand(NewHealthCommand(c))

	return cmd
}
