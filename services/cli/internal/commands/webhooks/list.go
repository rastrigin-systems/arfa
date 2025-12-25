package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewListCommand creates the webhooks list command.
func NewListCommand(c *container.Container) *cobra.Command {
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhook destinations",
		Long: `Display all configured webhook destinations for your organization.

Examples:
  arfa webhooks list           # List all webhooks
  arfa webhooks list --json    # Output as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()

			// Get auth service and require authentication
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			config, err := authService.RequireAuth()
			if err != nil {
				return fmt.Errorf("authentication required: %w", err)
			}

			// Create API client
			client := api.NewClient(config.PlatformURL)
			client.SetToken(config.Token)

			// Fetch webhooks
			resp, err := client.ListWebhooks(ctx)
			if err != nil {
				return fmt.Errorf("failed to list webhooks: %w", err)
			}

			if len(resp.Destinations) == 0 {
				fmt.Fprintln(out, "No webhook destinations configured.")
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Create one with: arfa webhooks create --name <name> --url <url>")
				return nil
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(resp, "", "  ")
				fmt.Fprintln(out, string(data))
				return nil
			}

			// Display table
			fmt.Fprintf(out, "\nWebhook Destinations (%d):\n\n", len(resp.Destinations))

			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tURL\tSTATUS\tEVENT TYPES")
			fmt.Fprintln(w, "────\t───\t──────\t───────────")

			for _, webhook := range resp.Destinations {
				status := "enabled"
				if !webhook.Enabled {
					status = "disabled"
				}

				eventTypes := "*"
				if len(webhook.EventTypes) > 0 {
					eventTypes = strings.Join(webhook.EventTypes, ", ")
					if len(eventTypes) > 30 {
						eventTypes = eventTypes[:27] + "..."
					}
				}

				url := webhook.URL
				if len(url) > 40 {
					url = url[:37] + "..."
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", webhook.Name, url, status, eventTypes)
			}

			w.Flush()
			fmt.Fprintln(out)

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}
