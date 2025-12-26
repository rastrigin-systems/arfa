package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewPoliciesCommand creates the policies command group.
func NewPoliciesCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Manage tool policies",
		Long: `View and manage tool policies that control LLM tool access.

Tool policies allow administrators to block or audit specific tools
used by AI agents in your organization.

Commands:
  list    - List policies from the platform
  create  - Create a new policy (admin/manager)
  update  - Update an existing policy (admin/manager)
  delete  - Delete a policy (admin/manager)`,
	}

	cmd.AddCommand(NewListCommand(c))
	cmd.AddCommand(NewCreateCommand(c))
	cmd.AddCommand(NewUpdateCommand(c))
	cmd.AddCommand(NewDeleteCommand(c))

	return cmd
}

// NewListCommand creates the policies list command.
func NewListCommand(c *container.Container) *cobra.Command {
	var showAll bool
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List active tool policies",
		Long: `Display tool policies that are currently in effect for your account.

These policies control which LLM tools can be used and are set by your
organization administrator.

Examples:
  arfa policies list           # Show deny policies (blocked tools)
  arfa policies list --all     # Show all policies including audit
  arfa policies list --json    # Output as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			// Get API client
			client, err := c.APIClient()
			if err != nil {
				return fmt.Errorf("not logged in. Run 'arfa login' first: %w", err)
			}

			// Fetch policies from API
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			resp, err := client.GetMyToolPolicies(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch policies: %w", err)
			}

			if len(resp.Policies) == 0 {
				_, _ = fmt.Fprintln(out, "No tool policies are configured for your account.")
				_, _ = fmt.Fprintln(out, "\nThis means all LLM tools are allowed.")
				return nil
			}

			// Filter policies if not showing all
			policies := resp.Policies
			if !showAll {
				policies = filterDenyPolicies(policies)
			}

			if len(policies) == 0 {
				_, _ = fmt.Fprintln(out, "No blocking policies found.")
				_, _ = fmt.Fprintln(out, "\nUse --all to see audit-only policies.")
				return nil
			}

			// Output as JSON if requested
			if showJSON {
				output := struct {
					Policies []api.ToolPolicy `json:"policies"`
				}{
					Policies: policies,
				}
				data, _ := json.MarshalIndent(output, "", "  ")
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}

			// Display table
			title := "Tool Policies (Blocked)"
			if showAll {
				title = "Tool Policies (All)"
			}
			_, _ = fmt.Fprintf(out, "\n%s (%d):\n\n", title, len(policies))

			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tTOOL\tACTION\tSCOPE\tCONDITIONS\tREASON")
			_, _ = fmt.Fprintln(w, "────────\t────\t──────\t─────\t──────────\t──────")

			for _, policy := range policies {
				// Short ID (first 8 chars)
				id := policy.ID
				if len(id) > 8 {
					id = id[:8]
				}
				if id == "" {
					id = "-"
				}

				var action string
				if policy.Action == api.ToolPolicyActionDeny {
					action = "DENY"
				} else {
					action = "audit"
				}

				scope := "-"
				if policy.Scope != "" {
					scope = string(policy.Scope)
				}

				conditions := formatConditions(policy.Conditions)

				reason := "-"
				if policy.Reason != nil && *policy.Reason != "" {
					reason = *policy.Reason
				}

				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", id, policy.ToolName, action, scope, conditions, reason)
			}

			_ = w.Flush()
			_, _ = fmt.Fprintln(out)

			return nil
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all policies including audit-only")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

// filterDenyPolicies returns only policies with action="deny"
func filterDenyPolicies(policies []api.ToolPolicy) []api.ToolPolicy {
	var result []api.ToolPolicy
	for _, p := range policies {
		if p.Action == api.ToolPolicyActionDeny {
			result = append(result, p)
		}
	}
	return result
}

// formatConditions returns a human-readable summary of policy conditions
func formatConditions(conditions map[string]interface{}) string {
	if len(conditions) == 0 {
		return "-"
	}

	var parts []string
	for param, condition := range conditions {
		var condStr string
		switch v := condition.(type) {
		case string:
			// Regex pattern: param=~pattern
			condStr = fmt.Sprintf("%s=~%s", param, truncate(v, 15))
		case map[string]interface{}:
			// Operator-based: {contains: x} or {equals: x}
			for op, val := range v {
				valStr := fmt.Sprintf("%v", val)
				condStr = fmt.Sprintf("%s %s %s", param, op, truncate(valStr, 10))
				break // Only show first operator
			}
		default:
			condStr = fmt.Sprintf("%s=?", param)
		}
		parts = append(parts, condStr)
	}

	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}

	if len(result) > 25 {
		return result[:22] + "..."
	}
	return result
}

// truncate shortens a string to maxLen characters
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
