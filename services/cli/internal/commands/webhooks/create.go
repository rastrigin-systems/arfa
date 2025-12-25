package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewCreateCommand creates the webhooks create command.
func NewCreateCommand(c *container.Container) *cobra.Command {
	var name, url, authType, bearerToken string
	var eventTypes []string
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook destination",
		Long: `Create a new webhook destination for exporting activity logs.

The webhook will receive POST requests with JSON payloads containing log events.
Each request includes an HMAC-SHA256 signature for verification.

Authentication options:
  --auth-type bearer    Use Bearer token authentication
  --auth-type header    Use custom header authentication
  --auth-type basic     Use HTTP Basic authentication

Examples:
  arfa webhooks create --name "SIEM Export" --url https://siem.example.com/events
  arfa webhooks create --name "Splunk" --url https://splunk.example.com/events \
    --auth-type bearer --bearer-token "sk-xxx" \
    --event-types tool_call,permission_denied`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()

			// Validate required flags
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			if url == "" {
				return fmt.Errorf("--url is required")
			}

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

			// Build request
			req := api.CreateWebhookRequest{
				Name:       name,
				URL:        url,
				AuthType:   authType,
				EventTypes: eventTypes,
			}

			// Add auth config if bearer token is provided
			if authType == "bearer" && bearerToken != "" {
				req.AuthConfig = map[string]string{
					"token": bearerToken,
				}
			}

			// Create webhook
			webhook, err := client.CreateWebhook(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create webhook: %w", err)
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(webhook, "", "  ")
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}

			// Display success message
			_, _ = fmt.Fprintln(out, "Webhook created successfully!")
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "  Name:        %s\n", webhook.Name)
			_, _ = fmt.Fprintf(out, "  ID:          %s\n", webhook.ID)
			_, _ = fmt.Fprintf(out, "  URL:         %s\n", webhook.URL)
			_, _ = fmt.Fprintf(out, "  Status:      %s\n", statusString(webhook.Enabled))
			if len(webhook.EventTypes) > 0 {
				_, _ = fmt.Fprintf(out, "  Event Types: %s\n", strings.Join(webhook.EventTypes, ", "))
			} else {
				_, _ = fmt.Fprintf(out, "  Event Types: all\n")
			}
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Test the webhook with: arfa webhooks test "+webhook.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the webhook destination (required)")
	cmd.Flags().StringVar(&url, "url", "", "URL to send events to (required)")
	cmd.Flags().StringVar(&authType, "auth-type", "none", "Authentication type: none, bearer, header, basic")
	cmd.Flags().StringVar(&bearerToken, "bearer-token", "", "Bearer token for authentication")
	cmd.Flags().StringSliceVar(&eventTypes, "event-types", nil, "Event types to forward (default: all)")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("url")

	return cmd
}

func statusString(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}
