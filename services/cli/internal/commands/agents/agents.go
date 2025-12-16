package agents

import (
	"github.com/spf13/cobra"
)

func NewAgentsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage AI agents",
		Long:  "View available AI agents and manage agent access.",
	}

	cmd.AddCommand(NewListCommand())
	cmd.AddCommand(NewInfoCommand())
	cmd.AddCommand(NewRequestCommand())
	cmd.AddCommand(NewShowCommand())

	return cmd
}
