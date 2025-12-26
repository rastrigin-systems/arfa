package setup

import (
	"fmt"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewSetupCommand creates the setup command
func NewSetupCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Show shell configuration for AI agent proxying",
		Long: `Show the configuration to add to your shell profile.

This enables AI tools (Claude Code, Cursor, Windsurf) to route
traffic through the arfa proxy for policy enforcement.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetup()
		},
	}

	return cmd
}

func runSetup() error {
	fmt.Println(`Add the following to your shell configuration:

For bash (~/.bashrc) or zsh (~/.zshrc):

  eval "$(arfa proxy env)"

For fish (~/.config/fish/config.fish):

  arfa proxy env | source

Then restart your terminal or source the file.`)

	return nil
}
