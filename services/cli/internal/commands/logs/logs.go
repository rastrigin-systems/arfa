package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/logging"
	"github.com/rastrigin-systems/arfa/services/cli/internal/ui"
	"github.com/spf13/cobra"
)

// NewLogsCommand creates the unified logs command with dependencies from the container.
func NewLogsCommand(c *container.Container) *cobra.Command {
	var (
		category string
		limit    int
		offset   int
		follow   bool
	)

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "View or stream activity logs",
		Long: `View historical logs or stream real-time logs as JSON.

Categories:
  - all: All logs regardless of category [default]
  - classified: Parsed logs (tool calls, user prompts, AI responses, errors)
  - proxy: Raw API requests/responses from the MITM proxy
  - session: Session lifecycle events (session_start, session_end)

Pagination (historical mode only):
  Default limit is 100 logs. Use --limit and --offset for pagination.
  Use --limit 0 to fetch all logs.

Examples:
  # Historical logs (default)
  arfa logs                           # View last 100 logs (all categories)
  arfa logs -c proxy                  # View last 100 proxy logs
  arfa logs -n 50                     # View last 50 logs
  arfa logs -n 100 --offset 100       # View next 100 logs (pagination)
  arfa logs -n 0                      # View all logs (no limit)
  arfa logs | less                    # Scroll through logs
  arfa logs | jq '.[0]'               # View first log with jq
  arfa logs -c classified | jq '.[] | select(.tool_name == "Bash")'  # Filter with jq

  # Real-time streaming
  arfa logs -f                        # Stream all logs in real-time
  arfa logs -f -c proxy               # Stream only proxy logs
  arfa logs -f | jq -c '.'            # Stream logs, one JSON per line`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			// Default to "all" if no category specified
			if category == "" {
				category = "all"
			}

			// Validate category
			validCategories := map[string]bool{
				"classified": true,
				"proxy":      true,
				"session":    true,
				"all":        true,
			}
			if !validCategories[category] {
				return fmt.Errorf("invalid category: %s (must be: classified, proxy, session, or all)", category)
			}

			// Streaming mode
			if follow {
				return streamLogs(c, category)
			}

			// Historical mode
			logs, err := logging.GetLogsWithConfig(configManager, category, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			if len(logs) == 0 {
				fmt.Fprintf(os.Stderr, "No logs found for category: %s\n", category)
				fmt.Fprintln(os.Stderr, "Tip: Ensure the proxy is running and capturing traffic.")
				return nil
			}

			// Output as JSON
			data, err := json.MarshalIndent(logs, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal logs: %w", err)
			}
			fmt.Println(string(data))

			return nil
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "all", "Filter by category: all, classified, proxy, session")
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Maximum number of logs to return (0 for all, historical mode only)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of logs to skip (for pagination, historical mode only)")
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Stream logs in real-time (ignores limit/offset)")

	return cmd
}

// streamLogs handles real-time log streaming
func streamLogs(c *container.Container, category string) error {
	configManager, err := c.ConfigManager()
	if err != nil {
		return fmt.Errorf("failed to get config manager: %w", err)
	}

	platformClient, err := c.APIClient()
	if err != nil {
		return fmt.Errorf("failed to get platform client: %w", err)
	}

	logStreamer := ui.NewLogStreamer(platformClient, configManager)
	logStreamer.SetJSONOutput(true) // Always use JSON for consistency

	// TODO: Add category filtering to LogStreamer if needed
	// For now, stream all logs - can be filtered with jq

	return logStreamer.StreamLogs(context.Background())
}
