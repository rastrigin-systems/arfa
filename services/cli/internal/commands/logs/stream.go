package logs

import (
	"context"
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewStreamCommand() *cobra.Command {
	var (
		follow  bool
		jsonOut bool
		verbose bool
	)

	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Stream real-time activity logs",
		Long:  "Connects to the platform via WebSocket to stream activity logs in real-time.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")

			logStreamer := cli.NewLogStreamer(platformClient, configManager)
			logStreamer.SetJSONOutput(jsonOut)
			logStreamer.SetVerbose(verbose)

			return logStreamer.StreamLogs(context.Background())
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs in real-time") // For consistency, though it's always following
	cmd.Flags().BoolVarP(&jsonOut, "json", "j", false, "Output full JSON for each log entry")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show full request/response payloads")

	return cmd
}
