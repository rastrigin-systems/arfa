package webhooks

import (
	"context"
	"fmt"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the webhooks delete command.
func NewDeleteCommand(c *container.Container) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <webhook-id>",
		Short: "Delete a webhook destination",
		Long: `Delete a webhook destination by ID.

Examples:
  arfa webhooks delete abc123-def456
  arfa webhooks delete abc123-def456 --force   # Skip confirmation`,
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

			// Get webhook details first to show what we're deleting
			webhook, err := client.GetWebhook(ctx, webhookID)
			if err != nil {
				return fmt.Errorf("webhook not found: %w", err)
			}

			// Confirm deletion if not forced
			if !force {
				_, _ = fmt.Fprintf(out, "Delete webhook '%s' (%s)? [y/N] ", webhook.Name, webhook.ID)
				var confirm string
				_, _ = fmt.Scanln(&confirm)
				if confirm != "y" && confirm != "Y" {
					_, _ = fmt.Fprintln(out, "Cancelled.")
					return nil
				}
			}

			// Delete webhook
			if err := client.DeleteWebhook(ctx, webhookID); err != nil {
				return fmt.Errorf("failed to delete webhook: %w", err)
			}

			_, _ = fmt.Fprintf(out, "Webhook '%s' deleted.\n", webhook.Name)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
