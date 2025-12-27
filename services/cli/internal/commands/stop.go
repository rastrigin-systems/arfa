package commands

import (
	"fmt"
	"syscall"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/control"
	"github.com/spf13/cobra"
)

// NewStopCommand creates the stop command.
func NewStopCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy server",
		Long: `Stop the running proxy server.

This sends a graceful shutdown signal to the proxy process,
allowing it to flush pending logs before exiting.`,
		RunE: runStop,
	}
}

func runStop(cmd *cobra.Command, args []string) error {
	pidFile, err := control.NewPIDFile()
	if err != nil {
		return fmt.Errorf("failed to create PID file manager: %w", err)
	}

	status, err := pidFile.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to check proxy status: %w", err)
	}

	if !status.Running {
		fmt.Println("Proxy is not running")
		return nil
	}

	fmt.Printf("Stopping proxy (PID %d)...\n", status.Info.PID)

	// Send SIGTERM to the proxy process
	if err := syscall.Kill(status.Info.PID, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to stop proxy: %w", err)
	}

	// Remove PID file
	if err := pidFile.Remove(); err != nil {
		fmt.Printf("Warning: Failed to remove PID file: %v\n", err)
	}

	fmt.Println("✓ Proxy stopped")
	return nil
}
