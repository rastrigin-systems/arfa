package commands

import (
	"fmt"
	"os"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/agents"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/cleanup"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/config"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/interactive"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/logs"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/proxy"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/skills"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/status"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/sync"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/commands/update"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and returns the root command with all subcommands registered
func NewRootCommand(version string) *cobra.Command {
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
	rootCmd.AddCommand(auth.NewLoginCommand())
	rootCmd.AddCommand(auth.NewLogoutCommand())

	// Register config commands
	rootCmd.AddCommand(config.NewConfigCommand())
	rootCmd.AddCommand(sync.NewSyncCommand())
	rootCmd.AddCommand(status.NewStatusCommand())
	rootCmd.AddCommand(update.NewUpdateCommand())
	rootCmd.AddCommand(cleanup.NewCleanupCommand())

	// Register command groups
	rootCmd.AddCommand(agents.NewAgentsCommand())
	rootCmd.AddCommand(newSkillsCommand())
	rootCmd.AddCommand(logs.NewLogsCommand())
	rootCmd.AddCommand(proxy.NewProxyCommand())

	// Legacy top-level commands (aliases for proxy start/stop)
	rootCmd.AddCommand(newStartCommand())
	rootCmd.AddCommand(newStopCommand())

	return rootCmd
}

// newSkillsCommand creates the skills command with initialized dependencies
func newSkillsCommand() *cobra.Command {
	configManager, err := cli.NewConfigManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to create config manager: %v\n", err)
		os.Exit(1)
	}

	platformClient := cli.NewPlatformClient("")
	authService := cli.NewAuthService(configManager, platformClient)

	return skills.NewSkillsCommand(configManager, platformClient, authService)
}

// newStartCommand creates the legacy 'start' command (alias for 'proxy start')
func newStartCommand() *cobra.Command {
	return proxy.NewStartCommand()
}

// newStopCommand creates the legacy 'stop' command (alias for 'proxy stop')
func newStopCommand() *cobra.Command {
	return proxy.NewStopCommand()
}
