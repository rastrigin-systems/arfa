package commands

import (
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/auth"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/config"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/logs"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/policies"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/proxy"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/setup"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/skills"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/status"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/commands/webhooks"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// NewRootCommand creates and returns the root command with all subcommands registered.
// The container parameter provides dependency injection for all subcommands.
func NewRootCommand(version string, c *container.Container) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "ubik",
		Short: "ubik - AI Agent Security Proxy",
		Long: `ubik CLI provides security monitoring and control for AI coding agents.
It runs as an HTTPS proxy that intercepts LLM API traffic, enabling:
- Real-time visibility into AI agent tool usage
- Policy enforcement (block/audit dangerous operations)
- Activity logging for compliance and debugging
- Webhook integration with SIEM systems

When run without subcommands, ubik starts the proxy server.

Examples:
  ubik                    Start the proxy (default)
  ubik login              Authenticate with the platform
  ubik logs stream        Monitor AI agent activity
  ubik policies list      View active security policies`,
		Version: version,
		// Default action: start proxy
		RunE: proxy.RunProxyStart,
	}

	// Register auth commands
	rootCmd.AddCommand(auth.NewLoginCommand(c))
	rootCmd.AddCommand(auth.NewLogoutCommand(c))

	// Register proxy commands
	rootCmd.AddCommand(proxy.NewProxyCommand(c))

	// Register config commands
	rootCmd.AddCommand(config.NewConfigCommand(c))
	rootCmd.AddCommand(status.NewStatusCommand(c))

	// Register monitoring commands
	rootCmd.AddCommand(skills.NewSkillsCommand(c))
	rootCmd.AddCommand(logs.NewLogsCommand(c))
	rootCmd.AddCommand(policies.NewPoliciesCommand(c))
	rootCmd.AddCommand(webhooks.NewWebhooksCommand(c))

	// Register setup commands
	rootCmd.AddCommand(setup.NewSetupCommand(c))

	return rootCmd
}
