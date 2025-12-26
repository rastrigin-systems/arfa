package logs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/logging"
	"github.com/spf13/cobra"
)

// NewViewCommand creates the view command with dependencies from the container.
func NewViewCommand(c *container.Container) *cobra.Command {
	var (
		category string
		limit    int
		offset   int
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View activity logs as JSON",
		Long: `View activity logs as JSON, filtered by category.

Categories:
  - all: All logs regardless of category [default]
  - classified: Parsed logs (tool calls, user prompts, AI responses, errors)
  - proxy: Raw API requests/responses from the MITM proxy
  - session: Session lifecycle events (session_start, session_end)

Pagination:
  Default limit is 100 logs. Use --limit and --offset for pagination.
  Use --limit 0 to fetch all logs.

Examples:
  arfa logs view                           # View last 100 logs (all categories)
  arfa logs view --category=proxy          # View last 100 proxy logs
  arfa logs view --limit=50                # View last 50 logs
  arfa logs view --limit=100 --offset=100  # View next 100 logs (pagination)
  arfa logs view --limit=0                 # View all logs (no limit)
  arfa logs view | less                    # Scroll through logs
  arfa logs view | jq '.[0]'               # View first log with jq`,
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

			// Get logs from the API with pagination
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
	cmd.Flags().IntVarP(&limit, "limit", "n", 100, "Maximum number of logs to return (0 for all)")
	cmd.Flags().IntVar(&offset, "offset", 0, "Number of logs to skip (for pagination)")

	return cmd
}
