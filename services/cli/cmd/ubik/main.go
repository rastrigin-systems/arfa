func newLogsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Manage and stream activity logs",
		Long:  "View activity logs from the platform, including real-time streaming.",
	}

	cmd.AddCommand(newLogsStreamCommand()) // Add the stream subcommand

	return cmd
}

func newLogsStreamCommand() *cobra.Command {
	var (
		follow bool
		// TODO: Add flags for session_id, agent_id, employee_id for filtering
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
			
			return logStreamer.StreamLogs(context.Background())
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow logs in real-time") // For consistency, though it's always following

	return cmd
}