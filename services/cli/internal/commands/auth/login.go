package auth

import (
	"context"
	"fmt"

	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewLoginCommand creates the login command with dependencies from the container.
func NewLoginCommand(c *container.Container) *cobra.Command {
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
			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			// If no URL provided via flag, check config for saved platform URL
			if platformURL == "" || platformURL == config.DefaultPlatformURL() {
				cfg, err := configManager.Load()
				if err == nil && cfg.PlatformURL != "" {
					platformURL = cfg.PlatformURL
				} else if platformURL == "" {
					platformURL = config.DefaultPlatformURL()
				}
			}

			platformClient, err := c.APIClient()
			if err != nil {
				return fmt.Errorf("failed to get platform client: %w", err)
			}
			platformClient.SetBaseURL(platformURL)

			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			ctx := context.Background()

			// Use interactive login if credentials not provided via flags
			if email == "" || password == "" {
				return authService.LoginInteractive(ctx)
			}

			// Non-interactive login
			if err := authService.Login(ctx, platformURL, email, password); err != nil {
				return err
			}

			fmt.Println("✓ Authenticated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&platformURL, "url", "", "Platform URL (defaults to saved URL or ARFA_API_URL env var)")
	cmd.Flags().StringVar(&email, "email", "", "Email address")
	cmd.Flags().StringVar(&password, "password", "", "Password")

	return cmd
}
