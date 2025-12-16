package auth

import (
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewLogoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear credentials",
		Long:  "Remove stored authentication token and logout from the platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			platformClient := cli.NewPlatformClient("")
			authService := cli.NewAuthService(configManager, platformClient)

			return authService.Logout()
		},
	}
}
