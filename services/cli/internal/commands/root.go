package commands

import (
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/agents"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/cleanup"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/config"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/interactive"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/logs"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/skills"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/status"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/sync"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/update"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and returns the root command with all subcommands registered.
// The container parameter provides dependency injection for all subcommands.
func NewRootCommand(version string, c *container.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ubik",
		Short: "ubik - Container-orchestrated AI agent management",
		Long: `ubik CLI enables employees to use AI coding agents with centrally-managed
configurations from the platform. It manages Docker containers that run
Claude Code and MCP servers with injected configs.

When run without subcommands, ubik starts an interactive session with your
default agent. If no default is set, you'll be prompted to select one.

Examples:
  ubik                    Run default agent (or pick if none set)
  ubik --pick             Always show agent picker
  ubik --agent "Claude"   Run specific agent (one-time, no save)
  ubik --set-default      Set default agent without starting`,
		Version: version,
		// Run function executes when no subcommand is provided (interactive mode)
		RunE: interactive.RunInteractiveMode,
	}

	// Add flags for interactive mode
	rootCmd.Flags().StringP("workspace", "w", "", "Workspace directory (interactive prompt if not provided)")
	rootCmd.Flags().StringP("agent", "a", "", "Agent to use (one-time override, doesn't change default)")
	rootCmd.Flags().BoolP("pick", "p", false, "Always show agent picker (saves selection as default)")
	rootCmd.Flags().Bool("set-default", false, "Set default agent without starting a session")

	// Register auth commands
	rootCmd.AddCommand(auth.NewLoginCommand(c))
	rootCmd.AddCommand(auth.NewLogoutCommand(c))

	// Register config commands
	rootCmd.AddCommand(config.NewConfigCommand(c))
	rootCmd.AddCommand(sync.NewSyncCommand(c))
	rootCmd.AddCommand(status.NewStatusCommand(c))
	rootCmd.AddCommand(update.NewUpdateCommand(c))
	rootCmd.AddCommand(cleanup.NewCleanupCommand(c))

	// Register command groups
	rootCmd.AddCommand(agents.NewAgentsCommand(c))
	rootCmd.AddCommand(skills.NewSkillsCommand(c))
	rootCmd.AddCommand(logs.NewLogsCommand(c))

	return rootCmd
}
