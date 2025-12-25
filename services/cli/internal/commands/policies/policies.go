package policies

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// policyCacheFile represents the cached policies structure.
type policyCacheFile struct {
	Policies []api.ToolPolicy `json:"policies"`
	Version  int              `json:"version"`
	SyncedAt string           `json:"synced_at"`
}

// NewPoliciesCommand creates the policies command group.
func NewPoliciesCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Manage tool policies",
		Long:  "View and manage tool policies that control LLM tool access.",
	}

	cmd.AddCommand(NewListCommand(c))

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
organization administrator. Policies are synced from the platform when
you run 'arfa sync'.

Examples:
  arfa policies list           # Show deny policies (blocked tools)
  arfa policies list --all     # Show all policies including audit
  arfa policies list --json    # Output as JSON`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()

			cache, err := loadPoliciesCache()
			if err != nil {
				fmt.Fprintln(out, "No tool policies found.")
				fmt.Fprintln(out, "\nRun 'arfa sync' to fetch policies from the platform.")
				return nil
			}

			if len(cache.Policies) == 0 {
				fmt.Fprintln(out, "No tool policies are configured for your account.")
				fmt.Fprintln(out, "\nThis means all LLM tools are allowed.")
				return nil
			}

			// Filter policies if not showing all
			policies := cache.Policies
			if !showAll {
				policies = filterDenyPolicies(policies)
			}

			if len(policies) == 0 {
				fmt.Fprintln(out, "No blocking policies found.")
				fmt.Fprintln(out, "\nUse --all to see audit-only policies.")
				return nil
			}

			// Output as JSON if requested
			if showJSON {
				output := struct {
					Policies []api.ToolPolicy `json:"policies"`
					Version  int              `json:"version"`
					SyncedAt string           `json:"synced_at"`
				}{
					Policies: policies,
					Version:  cache.Version,
					SyncedAt: cache.SyncedAt,
				}
				data, _ := json.MarshalIndent(output, "", "  ")
				fmt.Fprintln(out, string(data))
				return nil
			}

			// Display table
			title := "Tool Policies (Blocked)"
			if showAll {
				title = "Tool Policies (All)"
			}
			fmt.Fprintf(out, "\n%s (%d):\n\n", title, len(policies))

			w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "TOOL\tACTION\tSCOPE\tREASON")
			fmt.Fprintln(w, "────\t──────\t─────\t──────")

			for _, policy := range policies {
				action := string(policy.Action)
				if policy.Action == api.ToolPolicyActionDeny {
					action = "DENY"
				} else {
					action = "audit"
				}

				scope := "-"
				if policy.Scope != nil {
					scope = string(*policy.Scope)
				}

				reason := "-"
				if policy.Reason != nil && *policy.Reason != "" {
					reason = *policy.Reason
					if len(reason) > 40 {
						reason = reason[:37] + "..."
					}
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", policy.ToolName, action, scope, reason)
			}

			w.Flush()

			fmt.Fprintf(out, "\nSynced at: %s (version %d)\n", cache.SyncedAt, cache.Version)
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Run 'arfa sync' to refresh policies from the platform.")
			fmt.Fprintln(out)

			return nil
		},
	}

	cmd.Flags().BoolVar(&showAll, "all", false, "Show all policies including audit-only")
	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

// loadPoliciesCache loads the policies from ~/.arfa/policies.json
func loadPoliciesCache() (*policyCacheFile, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	policiesPath := filepath.Join(homeDir, ".arfa", "policies.json")
	data, err := os.ReadFile(policiesPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read policies file: %w", err)
	}

	var cache policyCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to parse policies file: %w", err)
	}

	return &cache, nil
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
