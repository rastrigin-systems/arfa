package agents

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewListCommand creates the list command with dependencies from the container.
func NewListCommand(c *container.Container) *cobra.Command {
	var showLocal bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available agents",
		Long:  "Display all available AI agents from the platform catalog or locally configured agents.",
		RunE: func(cmd *cobra.Command, args []string) error {
			agentService, err := c.AgentService()
			if err != nil {
				return fmt.Errorf("failed to get agent service: %w", err)
			}

			// If showing local agents, no need to authenticate
			if showLocal {
				agents, err := agentService.GetLocalAgents()
				if err != nil {
					return fmt.Errorf("failed to get local agents: %w", err)
				}

				if len(agents) == 0 {
					fmt.Println("No local agents configured. Run 'ubik sync' to fetch configs from the platform.")
					return nil
				}

				fmt.Printf("\nConfigured Agents (%d):\n\n", len(agents))

				// Create table writer
				w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
				fmt.Fprintln(w, "NAME\tTYPE\tSTATUS\tID")
				fmt.Fprintln(w, "────\t────\t──────\t──")

				for _, agent := range agents {
					enabledStatus := "✓ enabled"
					if !agent.IsEnabled {
						enabledStatus = "✗ disabled"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.AgentName, agent.AgentType, enabledStatus, agent.AgentID)
				}

				w.Flush()
				fmt.Println()
				fmt.Println("💡 Tip: Use 'ubik agents show <name>' to see configuration for your agents")
				fmt.Println()

				return nil
			}

			// For platform agents, require authentication
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			ctx := context.Background()
			agents, err := agentService.ListAgents(ctx)
			if err != nil {
				return fmt.Errorf("failed to list agents: %w", err)
			}

			if len(agents) == 0 {
				fmt.Println("No agents available in the platform catalog.")
				return nil
			}

			// Get user's enabled agents to show status
			syncService, err := c.SyncService()
			if err != nil {
				return fmt.Errorf("failed to get sync service: %w", err)
			}
			enabledAgents, _ := syncService.GetLocalAgentConfigs()

			// Build map of enabled agent IDs
			enabledMap := make(map[string]bool)
			for _, ea := range enabledAgents {
				if ea.IsEnabled {
					enabledMap[ea.AgentID] = true
				}
			}

			fmt.Printf("\nAvailable Agents (%d):\n\n", len(agents))

			// Create table writer
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tPROVIDER\tSTATUS\tDESCRIPTION")
			fmt.Fprintln(w, "────\t────────\t──────\t───────────")

			for _, agent := range agents {
				description := agent.Description
				if len(description) > 50 {
					description = description[:47] + "..."
				}
				status := "- not assigned"
				if enabledMap[agent.ID] {
					status = "✓ enabled"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", agent.Name, agent.Provider, status, description)
			}

			w.Flush()
			fmt.Println()
			fmt.Println("💡 Tip: Use 'ubik agents info <id>' to see agent details")
			fmt.Println("        Use 'ubik' or 'ubik --pick' to start an agent")
			fmt.Println()

			return nil
		},
	}

	cmd.Flags().BoolVar(&showLocal, "local", false, "Show locally configured agents only")

	return cmd
}
