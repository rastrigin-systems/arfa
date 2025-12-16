package auth

import (
	"fmt"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/spf13/cobra"
)

func NewLoginCommand() *cobra.Command {
	var (
		platformURL string
		email       string
		password    string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the platform",
		Long:  "Login to the platform and store authentication token locally.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configManager, err := cli.NewConfigManager()
			if err != nil {
				return fmt.Errorf("failed to create config manager: %w", err)
			}

			// If no URL provided via flag, check config for saved platform URL
			if platformURL == "" || platformURL == "https://api.ubik.io" {
				config, err := configManager.Load()
				if err == nil && config.PlatformURL != "" {
					platformURL = config.PlatformURL
				} else if platformURL == "" {
					platformURL = "https://api.ubik.io" // Final fallback
				}
			}

			platformClient := cli.NewPlatformClient(platformURL)
			authService := cli.NewAuthService(configManager, platformClient)

			// Use interactive login if credentials not provided via flags
			if email == "" || password == "" {
				return authService.LoginInteractive()
			}

			// Non-interactive login
			if err := authService.Login(platformURL, email, password); err != nil {
				return err
			}

			fmt.Println("✓ Authenticated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&platformURL, "url", "", "Platform URL (defaults to saved URL or https://api.ubik.io)")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")

	return cmd
}
