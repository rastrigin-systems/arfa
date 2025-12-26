package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewUpdateCommand creates the policies update command.
func NewUpdateCommand(c *container.Container) *cobra.Command {
	var toolName, action, reason string
	var conditions []string
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "update <policy-id>",
		Short: "Update a tool policy",
		Long: `Update an existing tool policy.

Only the specified fields will be updated. The scope (org/team/employee)
cannot be changed after creation.

Examples:
  # Change action to audit
  arfa policies update abc123 --action audit

  # Update reason
  arfa policies update abc123 --reason "Updated policy reason"

  # Change tool pattern
  arfa policies update abc123 --tool "mcp__gcloud__*"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()
			policyID := args[0]

			// Check that at least one field is being updated
			if toolName == "" && action == "" && reason == "" && len(conditions) == 0 {
				return fmt.Errorf("at least one field must be specified for update")
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

			// Build request with only specified fields
			req := api.UpdateToolPolicyRequest{}

			if toolName != "" {
				req.ToolName = &toolName
			}
			if action != "" {
				if action != "deny" && action != "audit" {
					return fmt.Errorf("--action must be 'deny' or 'audit'")
				}
				a := api.ToolPolicyAction(action)
				req.Action = &a
			}
			if reason != "" {
				req.Reason = &reason
			}
			if len(conditions) > 0 {
				req.Conditions = parseConditions(conditions)
			}

			// Update policy
			policy, err := client.UpdatePolicy(ctx, policyID, req)
			if err != nil {
				return fmt.Errorf("failed to update policy: %w", err)
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(policy, "", "  ")
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}

			// Display success message
			_, _ = fmt.Fprintln(out, "Policy updated successfully!")
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintf(out, "  ID:     %s\n", policy.ID)
			_, _ = fmt.Fprintf(out, "  Tool:   %s\n", policy.ToolName)
			_, _ = fmt.Fprintf(out, "  Action: %s\n", strings.ToUpper(string(policy.Action)))
			_, _ = fmt.Fprintf(out, "  Scope:  %s\n", policy.Scope)
			if policy.Reason != nil && *policy.Reason != "" {
				_, _ = fmt.Fprintf(out, "  Reason: %s\n", *policy.Reason)
			}
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Run 'arfa policies sync' to update local cache.")

			return nil
		},
	}

	cmd.Flags().StringVar(&toolName, "tool", "", "New tool name or glob pattern")
	cmd.Flags().StringVar(&action, "action", "", "New action: deny, audit")
	cmd.Flags().StringVar(&reason, "reason", "", "New reason for the policy")
	cmd.Flags().StringSliceVar(&conditions, "condition", nil, "New conditions (replaces existing)")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}
