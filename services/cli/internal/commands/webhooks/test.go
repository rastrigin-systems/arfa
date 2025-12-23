package webhooks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewTestCommand creates the webhooks test command.
func NewTestCommand(c *container.Container) *cobra.Command {
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "test <webhook-id>",
		Short: "Test a webhook destination",
		Long: `Send a test event to verify the webhook destination is working.

This sends a test payload to the configured URL and reports the result.

Examples:
  ubik webhooks test abc123-def456
  ubik webhooks test abc123-def456 --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()
			webhookID := args[0]

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

			// Get webhook details first
			webhook, err := client.GetWebhook(ctx, webhookID)
			if err != nil {
				return fmt.Errorf("webhook not found: %w", err)
			}

			fmt.Fprintf(out, "Testing webhook '%s'...\n", webhook.Name)

			// Test webhook
			result, err := client.TestWebhook(ctx, webhookID)
			if err != nil {
				return fmt.Errorf("failed to test webhook: %w", err)
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(result, "", "  ")
				fmt.Fprintln(out, string(data))
				return nil
			}

			// Display result
			fmt.Fprintln(out)
			if result.Success {
				fmt.Fprintln(out, "Test successful!")
				fmt.Fprintf(out, "  Response Status: %d\n", result.ResponseStatus)
				fmt.Fprintf(out, "  Response Time:   %dms\n", result.ResponseTimeMs)
			} else {
				fmt.Fprintln(out, "Test failed!")
				fmt.Fprintf(out, "  Response Status: %d\n", result.ResponseStatus)
				if result.ErrorMessage != "" {
					fmt.Fprintf(out, "  Error:           %s\n", result.ErrorMessage)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}
