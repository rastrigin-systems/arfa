package proxy

import (
	"github.com/spf13/cobra"
)

func NewProxyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage the MITM proxy daemon",
		Long:  "Control the shared MITM proxy daemon that intercepts and logs LLM API calls.",
	}

	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewStopCommand())
	cmd.AddCommand(NewStatusCommand())
	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewSessionsCommand())
	cmd.AddCommand(NewHealthCommand())

	return cmd
}
