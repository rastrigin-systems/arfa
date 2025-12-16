package auth

import (
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewLogoutCommand creates the logout command with dependencies from the container.
func NewLogoutCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout and clear credentials",
		Long:  "Remove stored authentication token and logout from the platform.",
		RunE: func(cmd *cobra.Command, args []string) error {
			authService, err := c.AuthService()
			if err != nil {
				return fmt.Errorf("failed to get auth service: %w", err)
			}

			return authService.Logout()
		},
	}
}
