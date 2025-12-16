package agents

import (
	"context"
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "request <agent-id>",
		Short: "Request access to an agent",
		Long:  "Request access to an AI agent by creating an employee agent configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			config, err := authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)
			ctx := context.Background()

			// Request the agent
			if err := agentService.RequestAgent(ctx, config.EmployeeID, agentID); err != nil {
				return fmt.Errorf("failed to request agent: %w", err)
			}

			fmt.Printf("\n✓ Agent access requested successfully\n")
			fmt.Printf("\nNext steps:\n")
			fmt.Printf("  1. Run 'ubik sync' to pull the new agent configuration\n")
			fmt.Printf("  2. Run 'ubik agents list --local' to see your configured agents\n\n")

			return nil
		},
	}
}
