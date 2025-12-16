package logs

import (
	"encoding/json"
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/pkg/types"
	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logparser"
	"github.com/spf13/cobra"
)

func NewViewCommand() *cobra.Command {
	var (
		format    string
		sessionID string
		noEmoji   bool
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View classified logs in human-readable format",
		Long: `View activity logs parsed and classified into human-readable entries.

Displays logs categorized as:
  - USER_PROMPT: User input to the AI
  - AI_TEXT: AI text responses
  - TOOL_CALL: Tools invoked by the AI
  - TOOL_RESULT: Results from tool execution
  - ERROR: Any errors that occurred

Examples:
  ubik logs view                    # View logs in pretty format
  ubik logs view --format=json      # View logs as JSON
  ubik logs view --no-emoji         # Disable emoji icons`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// Get classified logs from the current session or storage
			logs, err := cli.GetClassifiedLogs(configManager, sessionID)
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			if len(logs) == 0 {
				fmt.Println("No classified logs found.")
				fmt.Println("Tip: Run 'ubik sync' to start a session and generate logs.")
				return nil
			}

			// Format and display
			formatter := logparser.DefaultFormatter()
			formatter.UseEmoji = !noEmoji

			switch format {
			case "json":
				// Output as JSON
				data, err := json.MarshalIndent(logs, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal logs: %w", err)
				}
				fmt.Println(string(data))

			case "pretty", "":
				// Output in human-readable format
				if sessionID != "" {
					fmt.Println(formatter.FormatSession(sessionID, logs))
				} else {
					// Group by session
					sessions := make(map[string][]types.ClassifiedLogEntry)
					for _, log := range logs {
						sessions[log.SessionID] = append(sessions[log.SessionID], log)
					}
					for sid, sessionLogs := range sessions {
						fmt.Println(formatter.FormatSession(sid, sessionLogs))
						fmt.Println()
					}
				}

			default:
				return fmt.Errorf("unknown format: %s (use 'pretty' or 'json')", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "pretty", "Output format: pretty, json")
	cmd.Flags().StringVarP(&sessionID, "session", "s", "", "Filter by session ID")
	cmd.Flags().BoolVar(&noEmoji, "no-emoji", false, "Disable emoji icons in output")

	return cmd
}
