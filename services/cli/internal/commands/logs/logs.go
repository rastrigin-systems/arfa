package logs

import (
	"github.com/spf13/cobra"
)

func NewLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage and stream activity logs",
		Long:  "View activity logs from the platform, including real-time streaming.",
	}

	cmd.AddCommand(NewStreamCommand())
	cmd.AddCommand(NewViewCommand())

	return cmd
}
