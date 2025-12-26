package policies

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewDeleteCommand creates the policies delete command.
func NewDeleteCommand(c *container.Container) *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete <policy-id>",
		Short: "Delete a tool policy",
		Long: `Delete an existing tool policy.

By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  arfa policies delete abc123
  arfa policies delete abc123 --force`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()
			policyID := args[0]

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

			// Fetch policy first to show details
			policy, err := client.GetPolicy(ctx, policyID)
			if err != nil {
				return fmt.Errorf("failed to get policy: %w", err)
			}

			// Confirm deletion if not forced
			if !force {
				_, _ = fmt.Fprintln(out, "Policy to delete:")
				_, _ = fmt.Fprintf(out, "  ID:     %s\n", policy.ID)
				_, _ = fmt.Fprintf(out, "  Tool:   %s\n", policy.ToolName)
				_, _ = fmt.Fprintf(out, "  Action: %s\n", strings.ToUpper(string(policy.Action)))
				_, _ = fmt.Fprintf(out, "  Scope:  %s\n", policy.Scope)
				_, _ = fmt.Fprintln(out)

				reader := bufio.NewReader(cmd.InOrStdin())
				_, _ = fmt.Fprint(out, "Are you sure you want to delete this policy? [y/N]: ")
				answer, _ := reader.ReadString('\n')
				answer = strings.TrimSpace(strings.ToLower(answer))
				if answer != "y" && answer != "yes" {
					_, _ = fmt.Fprintln(out, "Deletion cancelled.")
					return nil
				}
			}

			// Delete policy
			if err := client.DeletePolicy(ctx, policyID); err != nil {
				return fmt.Errorf("failed to delete policy: %w", err)
			}

			_, _ = fmt.Fprintln(out, "Policy deleted successfully!")
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Run 'arfa policies sync' to update local cache.")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
