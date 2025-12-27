package status

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/rastrigin-systems/arfa/services/cli/internal/control"
	"github.com/spf13/cobra"
)

// NewStatusCommand creates the status command with dependencies from the container.
func NewStatusCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current status",
		Long:  "Display comprehensive status of all Arfa components.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(c)
		},
	}
}

func runStatus(c *container.Container) error {
	fmt.Println("┌─────────────────────────────────────────────────────────────┐")
	fmt.Println("│                        ARFA STATUS                          │")
	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	// Authentication status
	printAuthStatus(c)

	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	// Proxy status
	printProxyStatus()

	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	// Log queue status
	printQueueStatus()

	fmt.Println("├─────────────────────────────────────────────────────────────┤")

	// API server status
	printAPIStatus(c)

	fmt.Println("└─────────────────────────────────────────────────────────────┘")

	return nil
}

func printAuthStatus(c *container.Container) {
	fmt.Println("│ Authentication                                              │")

	configManager, err := c.ConfigManager()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error loading config                      │")
		return
	}

	authService, err := c.AuthService()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error loading auth service                │")
		return
	}

	authenticated, err := authService.IsAuthenticated()
	if err != nil || !authenticated {
		fmt.Println("│   Status:       ✗ Not authenticated                         │")
		fmt.Println("│   Action:       Run 'arfa login' to authenticate            │")
		return
	}

	config, err := configManager.Load()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error loading config                      │")
		return
	}

	fmt.Println("│   Status:       ✓ Authenticated                             │")
	fmt.Printf("│   Platform:     %-43s │\n", truncate(config.PlatformURL, 43))

	// Get claims from JWT
	if claims, err := config.GetClaims(); err == nil {
		fmt.Printf("│   Employee:     %-43s │\n", truncate(claims.EmployeeID, 43))

		// Check token expiry
		if !claims.ExpiresAt.IsZero() {
			remaining := time.Until(claims.ExpiresAt)
			if remaining > 0 {
				fmt.Printf("│   Token Expiry: %-43s │\n", formatDuration(remaining)+" remaining")
			} else {
				fmt.Println("│   Token Expiry: ✗ Expired                                  │")
			}
		}
	}
}

func printProxyStatus() {
	fmt.Println("│ Proxy                                                       │")

	pidFile, err := control.NewPIDFile()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error checking status                     │")
		return
	}

	status, err := pidFile.GetStatus()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error checking status                     │")
		return
	}

	if !status.Running {
		fmt.Println("│   Status:       ○ Not running                               │")
		fmt.Println("│   Action:       Run 'arfa start' to start proxy             │")
		return
	}

	fmt.Printf("│   Status:       ✓ Running on localhost:%-21d │\n", status.Info.Port)
	fmt.Printf("│   PID:          %-43d │\n", status.Info.PID)
	fmt.Printf("│   Uptime:       %-43s │\n", formatDuration(status.Uptime))
	fmt.Printf("│   Certificate:  %-43s │\n", truncate(status.Info.CertPath, 43))
}

func printQueueStatus() {
	fmt.Println("│ Log Queue                                                   │")

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error checking queue                      │")
		return
	}

	queueDir := filepath.Join(home, ".arfa", "log_queue")

	// Count pending files
	files, err := filepath.Glob(filepath.Join(queueDir, "*.json"))
	if err != nil {
		fmt.Println("│   Status:       ✗ Error reading queue                       │")
		return
	}

	pendingCount := len(files)
	if pendingCount == 0 {
		fmt.Println("│   Pending:      ✓ 0 entries (all synced)                    │")
	} else {
		fmt.Printf("│   Pending:      ○ %d entries waiting to sync                 │\n", pendingCount)
	}
	fmt.Printf("│   Queue Dir:    %-43s │\n", truncate(queueDir, 43))
}

func printAPIStatus(c *container.Container) {
	fmt.Println("│ API Server                                                  │")

	configManager, err := c.ConfigManager()
	if err != nil {
		fmt.Println("│   Status:       ✗ Error loading config                      │")
		return
	}

	config, err := configManager.Load()
	if err != nil || config.PlatformURL == "" {
		fmt.Println("│   Status:       ○ Not configured                            │")
		return
	}

	// Health check with timeout
	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	resp, err := client.Get(config.PlatformURL + "/api/v1/health")
	latency := time.Since(start)

	if err != nil {
		fmt.Println("│   Status:       ✗ Unreachable                               │")
		fmt.Printf("│   Error:        %-43s │\n", truncate(err.Error(), 43))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("│   Status:       ✓ Reachable (%dms)                           │\n", latency.Milliseconds())
	} else {
		fmt.Printf("│   Status:       ○ Responded with %d                          │\n", resp.StatusCode)
	}
}

// truncate truncates a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
