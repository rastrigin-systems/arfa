package policies

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewSyncCommand creates the policies sync command.
func NewSyncCommand(c *container.Container) *cobra.Command {
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync policies from platform",
		Long: `Fetch tool policies from the platform and save to local cache.

The CLI uses cached policies to enforce tool blocking during proxy mode.
Run this command periodically to ensure policies are up to date.

The policies are cached at ~/.arfa/policies.json

Examples:
  arfa policies sync
  arfa policies sync --json`,
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

			// Fetch policies from API
			resp, err := client.GetMyToolPolicies(ctx)
			if err != nil {
				return fmt.Errorf("failed to fetch policies: %w", err)
			}

			// Save to cache file
			if err := savePoliciesCache(resp); err != nil {
				return fmt.Errorf("failed to save policies cache: %w", err)
			}

			// Count by action
			denyCount := 0
			auditCount := 0
			for _, p := range resp.Policies {
				if p.Action == api.ToolPolicyActionDeny {
					denyCount++
				} else {
					auditCount++
				}
			}

			// Output as JSON if requested
			if showJSON {
				data, _ := json.MarshalIndent(resp, "", "  ")
				_, _ = fmt.Fprintln(out, string(data))
				return nil
			}

			// Display summary
			_, _ = fmt.Fprintf(out, "Synced %d policies (%d deny, %d audit)\n",
				len(resp.Policies), denyCount, auditCount)
			_, _ = fmt.Fprintf(out, "Version: %d\n", resp.Version)
			_, _ = fmt.Fprintf(out, "Synced at: %s\n", resp.SyncedAt)
			_, _ = fmt.Fprintln(out)
			_, _ = fmt.Fprintln(out, "Run 'arfa policies list' to view synced policies.")

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

// savePoliciesCache saves policies to ~/.arfa/policies.json
func savePoliciesCache(resp *api.EmployeeToolPoliciesResponse) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	arfaDir := filepath.Join(homeDir, ".arfa")
	if err := os.MkdirAll(arfaDir, 0700); err != nil {
		return fmt.Errorf("failed to create .arfa directory: %w", err)
	}

	policiesPath := filepath.Join(arfaDir, "policies.json")

	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal policies: %w", err)
	}

	if err := os.WriteFile(policiesPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write policies file: %w", err)
	}

	return nil
}
