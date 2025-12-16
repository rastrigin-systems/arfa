package agents

import (
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewAgentsCommand creates the agents command group with dependencies from the container.
func NewAgentsCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Manage AI agents",
		Long:  "View available AI agents and manage agent access.",
	}

	cmd.AddCommand(NewListCommand(c))
	cmd.AddCommand(NewInfoCommand(c))
	cmd.AddCommand(NewRequestCommand(c))
	cmd.AddCommand(NewShowCommand(c))

	return cmd
}
