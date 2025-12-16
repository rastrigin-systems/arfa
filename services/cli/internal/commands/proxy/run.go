package proxy

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/container"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/httpproxy"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
	"github.com/spf13/cobra"
)

// NewRunCommand creates the run command with dependencies from the container.
func NewRunCommand(c *container.Container) *cobra.Command {
	var port int

	cmd := &cobra.Command{
		Use:    "run",
		Short:  "Run the proxy daemon (internal)",
		Long:   "Run the MITM proxy in the foreground. This is called internally by 'proxy start'.",
		Hidden: true, // Hide from help since it's internal
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize logger (optional - daemon mode)
			loggerConfig := &logging.Config{
				Enabled:       true,
				BatchSize:     100,
				BatchInterval: 5 * time.Second,
				MaxRetries:    5,
				RetryBackoff:  1 * time.Second,
			}

			configManager, err := c.ConfigManager()
			if err != nil {
				return fmt.Errorf("failed to get config manager: %w", err)
			}

			// Load config to get platform URL and auth token
			config, err := configManager.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
			}

			// Use platform URL from config, fall back to default
			platformURL := "https://api.ubik.io"
			if config != nil && config.PlatformURL != "" {
				platformURL = config.PlatformURL
			}

			platformClient, err := c.PlatformClient()
			if err != nil {
				return fmt.Errorf("failed to get platform client: %w", err)
			}

			// Set auth token if available
			if config != nil && config.Token != "" {
				platformClient.SetToken(config.Token)
				fmt.Printf("Proxy logging authenticated to: %s\n", platformURL)
			} else {
				fmt.Fprintf(os.Stderr, "Warning: no auth token - logs will not be sent to platform\n")
			}

			apiClient := cli.NewPlatformAPIClient(platformClient)
			logger, err := logging.NewLogger(loggerConfig, apiClient)
			if err != nil {
				// Continue without logging - log warning to stderr
				fmt.Fprintf(os.Stderr, "Warning: failed to initialize logging: %v\n", err)
			}

			// Create daemon manager
			daemon, err := httpproxy.NewProxyDaemon()
			if err != nil {
				return fmt.Errorf("failed to create daemon manager: %w", err)
			}

			// Run until interrupted
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Handle signals
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				<-sigChan
				fmt.Println("\nShutting down proxy daemon...")
				cancel()
			}()

			// Run the daemon with full config (this saves state and blocks)
			daemonConfig := httpproxy.RunDaemonConfig{
				Port:        port,
				PlatformURL: platformURL,
			}
			if err := daemon.RunDaemonWithConfig(ctx, daemonConfig, logger); err != nil {
				return fmt.Errorf("daemon error: %w", err)
			}

			// Cleanup
			if logger != nil {
				logger.Close()
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&port, "port", httpproxy.DefaultProxyPort, "Port for the proxy to listen on")

	return cmd
}
