package commands

import (
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/auth"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/logs"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/policies"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/setup"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/status"
	"github.com/rastrigin-systems/arfa/services/cli/internal/commands/webhooks"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and returns the root command with all subcommands registered.
// The container parameter provides dependency injection for all subcommands.
func NewRootCommand(version string, c *container.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "arfa",
		Short: "arfa - AI Agent Security Proxy",
		Long: `arfa CLI provides security monitoring and control for AI coding agents.
It runs as an HTTPS proxy that intercepts LLM API traffic, enabling:
- Real-time visibility into AI agent tool usage
- Policy enforcement (block/audit dangerous operations)
- Activity logging for compliance and debugging
- Webhook integration with SIEM systems

Commands:
  arfa start             Start the security proxy
  arfa stop              Stop the security proxy
  arfa status            Show status of all components
  arfa login             Authenticate with the platform
  arfa logs stream       Monitor AI agent activity
  arfa policies list     View active security policies`,
		Version: version,
		// No default action - just print help
	}

	// Register core commands (top level)
	rootCmd.AddCommand(NewStartCommand(c))
	rootCmd.AddCommand(NewStopCommand(c))
	rootCmd.AddCommand(NewEnvCommand(c))

	// Register auth commands
	rootCmd.AddCommand(auth.NewLoginCommand(c))
	rootCmd.AddCommand(auth.NewLogoutCommand(c))

	// Register status command
	rootCmd.AddCommand(status.NewStatusCommand(c))

	// Register config commands
	rootCmd.AddCommand(config.NewConfigCommand(c))

	// Register monitoring commands
	rootCmd.AddCommand(logs.NewLogsCommand(c))
	rootCmd.AddCommand(policies.NewPoliciesCommand(c))
	rootCmd.AddCommand(webhooks.NewWebhooksCommand(c))

	// Register setup commands
	rootCmd.AddCommand(setup.NewSetupCommand(c))

	return rootCmd
}
