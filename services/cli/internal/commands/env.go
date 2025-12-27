package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/control"
	"github.com/spf13/cobra"
)

// NewEnvCommand creates the env command.
func NewEnvCommand(c *container.Container) *cobra.Command {
	var formatFlag string

	cmd := &cobra.Command{
		Use:   "env",
		Short: "Output environment variables for proxy configuration",
		Long: `Output shell export statements for configuring the proxy.

Use with eval to set environment variables in your shell:
  eval $(arfa env)

For CI pipelines:
  # GitHub Actions
  arfa env >> $GITHUB_ENV

  # GitLab CI / generic
  eval $(arfa env)

Formats:
  shell   - Shell export statements (default)
  github  - GitHub Actions format (KEY=VALUE)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnv(formatFlag)
		},
	}

	cmd.Flags().StringVar(&formatFlag, "format", "shell", "Output format: shell, github")

	return cmd
}

func runEnv(format string) error {
	// Get home directory for cert path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	// Check if proxy is running and get actual port
	pidFile, err := control.NewPIDFile()
	if err != nil {
		return fmt.Errorf("failed to create PID file manager: %w", err)
	}

	var proxyURL string
	var certPath string

	status, err := pidFile.GetStatus()
	if err == nil && status.Running && status.Info != nil {
		// Use actual running proxy port
		proxyURL = fmt.Sprintf("http://127.0.0.1:%d", status.Info.Port)
		certPath = status.Info.CertPath
	} else {
		// Fallback to defaults
		proxyURL = "http://127.0.0.1:8082"
		certPath = filepath.Join(home, ".arfa", "certs", "arfa-ca.pem")
	}

	switch format {
	case "github":
		// GitHub Actions format: KEY=VALUE (one per line, append to $GITHUB_ENV)
		fmt.Printf("HTTPS_PROXY=%s\n", proxyURL)
		fmt.Printf("SSL_CERT_FILE=%s\n", certPath)
		fmt.Printf("NODE_EXTRA_CA_CERTS=%s\n", certPath)
	case "shell":
		fallthrough
	default:
		// Shell export format
		fmt.Printf("export HTTPS_PROXY=\"%s\"\n", proxyURL)
		fmt.Printf("export SSL_CERT_FILE=\"%s\"\n", certPath)
		fmt.Printf("export NODE_EXTRA_CA_CERTS=\"%s\"\n", certPath)
	}

	return nil
}
