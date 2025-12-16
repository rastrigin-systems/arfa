package proxy

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/spf13/cobra"
)

// NewSessionsCommand creates the sessions command with dependencies from the container.
func NewSessionsCommand(_ *container.Container) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "sessions",
		Short: "List active proxy sessions",
		Long:  "Display all active sessions connected through the security gateway.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := httpproxy.NewDefaultControlClient()
			if err != nil {
				return fmt.Errorf("failed to create control client: %w", err)
			}

			sessions, err := client.ListSessions()
			if err != nil {
				return fmt.Errorf("failed to list sessions (is proxy running?): %w", err)
			}

			if jsonOutput {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				return enc.Encode(sessions)
			}

			if len(sessions) == 0 {
				fmt.Println("No active sessions")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SESSION ID\tPORT\tAGENT\tWORKSPACE\tUPTIME")
			for _, s := range sessions {
				uptime := time.Since(s.StartTime).Round(time.Second)
				fmt.Fprintf(w, "%s\t%d\t%s\t%s\t%s\n",
					truncateStr(s.ID, 12),
					s.Port,
					s.AgentName,
					truncateStr(s.Workspace, 30),
					uptime,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

// truncateStr truncates a string to maxLen characters with "..." suffix
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
