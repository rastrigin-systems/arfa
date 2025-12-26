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

// NewCreateCommand creates the policies create command.
func NewCreateCommand(c *container.Container) *cobra.Command {
	var toolName, action, reason string
	var teamID, employeeID string
	var conditions []string
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a tool policy",
		Long: `Create a new tool policy to control LLM tool access.

Policies can block or audit specific tools or patterns. Use glob patterns
to match multiple tools (e.g., "mcp__*" matches all MCP tools).

Actions:
  deny   - Block the tool (default)
  audit  - Allow but log usage

Scopes:
  Organization - No team or employee flags (default)
  Team         - Use --team flag
  Employee     - Use --employee flag

Examples:
  # Block Bash for entire organization
  arfa policies create --tool Bash --action deny --reason "Shell blocked"

  # Block all MCP tools for a specific team
  arfa policies create --tool "mcp__*" --action deny --team 123e4567-e89b-12d3-a456-426614174000

  # Audit dangerous commands (conditional policy)
  arfa policies create --tool Bash --action deny \
    --condition 'command=~rm\s+-rf' \
    --reason "Destructive commands blocked"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			ctx := context.Background()

			// Validate required flags
			if toolName == "" {
				return fmt.Errorf("--tool is required")
			}
			if action == "" {
				action = "deny"
			}
			if action != "deny" && action != "audit" {
				return fmt.Errorf("--action must be 'deny' or 'audit'")
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
			req := api.CreateToolPolicyRequest{
				ToolName: toolName,
				Action:   api.ToolPolicyAction(action),
			}

			if reason != "" {
				req.Reason = &reason
			}
			if teamID != "" {
				req.TeamID = &teamID
			}
			if employeeID != "" {
				req.EmployeeID = &employeeID
			}

			// Parse conditions
			if len(conditions) > 0 {
				req.Conditions = parseConditions(conditions)
			}

			// Create policy
			policy, err := client.CreatePolicy(ctx, req)
			if err != nil {
				return fmt.Errorf("failed to create policy: %w", err)
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(policy, "", "  ")
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}

			// Display success message
			_, _ = fmt.Fprintln(out, "Policy created successfully!")
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

	cmd.Flags().StringVar(&toolName, "tool", "", "Tool name or glob pattern (required)")
	cmd.Flags().StringVar(&action, "action", "deny", "Action to take: deny, audit")
	cmd.Flags().StringVar(&reason, "reason", "", "Human-readable reason for the policy")
	cmd.Flags().StringVar(&teamID, "team", "", "Apply policy to specific team ID")
	cmd.Flags().StringVar(&employeeID, "employee", "", "Apply policy to specific employee ID")
	cmd.Flags().StringSliceVar(&conditions, "condition", nil, "Condition in format 'param=~regex' (repeatable)")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	_ = cmd.MarkFlagRequired("tool")

	return cmd
}

// parseConditions converts condition strings to a conditions map.
// Format: "param=~regex" becomes {"any": [{"param_path": "param", "operator": "contains", "value": "regex"}]}
func parseConditions(conditions []string) map[string]interface{} {
	if len(conditions) == 0 {
		return nil
	}

	var conditionList []map[string]interface{}
	for _, cond := range conditions {
		parts := strings.SplitN(cond, "=~", 2)
		if len(parts) == 2 {
			conditionList = append(conditionList, map[string]interface{}{
				"param_path": parts[0],
				"operator":   "matches",
				"value":      parts[1],
			})
		}
	}

	if len(conditionList) == 0 {
		return nil
	}

	return map[string]interface{}{
		"any": conditionList,
	}
}
