package webhooks

import (
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewWebhooksCommand creates the webhooks command group with dependencies from the container.
func NewWebhooksCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhooks",
		Short: "Manage webhook destinations for log export",
		Long: `Configure webhook destinations to export activity logs to external systems
like SIEM tools (Splunk, Datadog, Elasticsearch) or custom endpoints.

Requires admin permissions.`,
	}

	cmd.AddCommand(NewListCommand(c))
	cmd.AddCommand(NewCreateCommand(c))
	cmd.AddCommand(NewDeleteCommand(c))
	cmd.AddCommand(NewTestCommand(c))

	return cmd
}
