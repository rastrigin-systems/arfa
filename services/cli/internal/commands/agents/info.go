package agents

import (
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info <agent-id>",
		Short: "Get agent details",
		Long:  "Display detailed information about a specific AI agent.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := args[0]

			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			_, err = authService.RequireAuth()
			if err != nil {
				return err
			}

			agentService := cli.NewAgentService(platformClient, configManager)
			agent, err := agentService.GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent info: %w", err)
			}

			fmt.Printf("\nAgent: %s\n", agent.Name)
			fmt.Printf("Provider: %s\n", agent.Provider)
			fmt.Printf("Description: %s\n", agent.Description)
			fmt.Printf("Pricing: %s\n", agent.PricingTier)
			fmt.Printf("ID: %s\n", agent.ID)

			if len(agent.SupportedPlatforms) > 0 {
				fmt.Printf("Platforms: ")
				for i, platform := range agent.SupportedPlatforms {
					if i > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", platform)
				}
				fmt.Println()
			}

			fmt.Printf("\nCreated: %s\n", agent.CreatedAt.Format("2006-01-02"))
			fmt.Printf("Updated: %s\n", agent.UpdatedAt.Format("2006-01-02"))
			fmt.Println()

			return nil
		},
	}
}
