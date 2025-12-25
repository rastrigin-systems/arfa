package logs

import (
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates the logs command group with dependencies from the container.
func NewLogsCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage and stream activity logs",
		Long:  "View activity logs from the platform, including real-time streaming.",
	}

	cmd.AddCommand(NewStreamCommand(c))
	cmd.AddCommand(NewViewCommand(c))

	return cmd
}
