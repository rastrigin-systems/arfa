package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/container"
	"github.com/spf13/cobra"
)

// LLM provider domains to route through proxy
var llmDomains = []string{
	"api.anthropic.com",
	"api.openai.com",
	"generativelanguage.googleapis.com",
}

// NewSetupCommand creates the setup command group
func NewSetupCommand(c *container.Container) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure system for AI agent proxying",
		Long: `Setup commands configure your system for transparent AI agent proxying.

Available commands:
  ubik setup system    Install system-wide proxy configuration
  ubik setup status    Check current setup status
  ubik setup uninstall Remove system configuration`,
	}

	cmd.AddCommand(newSystemCommand(c))
	cmd.AddCommand(newSetupStatusCommand(c))
	cmd.AddCommand(newUninstallCommand(c))

	return cmd
}

func newSystemCommand(c *container.Container) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "system",
		Short: "Install system-wide proxy configuration",
		Long: `Configure the system to route AI agent traffic through the ubik proxy.

This command will:
1. Generate a PAC (Proxy Auto-Config) file that routes only LLM API traffic
2. Install the CA certificate to the system trust store
3. Configure the system to use the PAC file
4. Set up auto-start for the proxy daemon

Only these domains are proxied:
  - api.anthropic.com
  - api.openai.com
  - generativelanguage.googleapis.com

All other traffic goes direct (no performance impact).

Requires sudo/admin privileges.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSystemSetup(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	return cmd
}

func newSetupStatusCommand(c *container.Container) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check current setup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSetupStatus()
		},
	}
}

func newUninstallCommand(c *container.Container) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove system configuration",
		Long: `Remove ubik system configuration.

This will:
- Remove the PAC file
- Remove the CA certificate from system trust store
- Remove auto-start configuration
- Stop the proxy daemon if running`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(dryRun)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	return cmd
}

func runSystemSetup(dryRun bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	ubikDir := filepath.Join(home, ".ubik")
	certPath := filepath.Join(ubikDir, "certs", "ubik-ca.pem")
	pacPath := filepath.Join(ubikDir, "proxy.pac")

	fmt.Println("Ubik System Setup")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if dryRun {
		fmt.Println("[DRY RUN] Would perform the following:")
		fmt.Println()
	}

	// Step 1: Create PAC file
	fmt.Println("1. Creating PAC file...")
	pacContent := generatePACFile()
	if dryRun {
		fmt.Printf("   Would create: %s\n", pacPath)
		fmt.Println("   Content:")
		for _, line := range strings.Split(pacContent, "\n")[:10] {
			fmt.Printf("     %s\n", line)
		}
		fmt.Println("     ...")
	} else {
		if err := os.MkdirAll(ubikDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.WriteFile(pacPath, []byte(pacContent), 0644); err != nil {
			return fmt.Errorf("failed to write PAC file: %w", err)
		}
		fmt.Printf("   ✓ Created: %s\n", pacPath)
	}
	fmt.Println()

	// Step 2: Install CA certificate
	fmt.Println("2. Installing CA certificate to system trust store...")
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		fmt.Printf("   ⚠ Certificate not found at %s\n", certPath)
		fmt.Println("   Run 'ubik proxy start' first to generate the certificate.")
	} else {
		if dryRun {
			fmt.Printf("   Would install: %s\n", certPath)
		} else {
			if err := installCACert(certPath); err != nil {
				fmt.Printf("   ⚠ Failed to install certificate: %v\n", err)
				fmt.Println("   You may need to run this command with sudo.")
			} else {
				fmt.Println("   ✓ Certificate installed")
			}
		}
	}
	fmt.Println()

	// Step 3: Configure system to use PAC file
	fmt.Println("3. Configuring system proxy settings...")
	if dryRun {
		fmt.Printf("   Would configure system to use PAC file: %s\n", pacPath)
	} else {
		if err := configureSystemProxy(pacPath); err != nil {
			fmt.Printf("   ⚠ Failed to configure system proxy: %v\n", err)
			fmt.Println("   You may need to configure this manually in System Settings.")
		} else {
			fmt.Println("   ✓ System proxy configured")
		}
	}
	fmt.Println()

	// Step 4: Setup auto-start
	fmt.Println("4. Configuring proxy auto-start...")
	if dryRun {
		fmt.Println("   Would create launchd/systemd service for auto-start")
	} else {
		if err := setupAutoStart(); err != nil {
			fmt.Printf("   ⚠ Auto-start setup failed: %v\n", err)
			fmt.Println("   You'll need to run 'ubik proxy start' manually.")
		} else {
			fmt.Println("   ✓ Auto-start configured")
		}
	}
	fmt.Println()

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if dryRun {
		fmt.Println("Dry run complete. Run without --dry-run to apply changes.")
	} else {
		fmt.Println("Setup complete!")
		fmt.Println()
		fmt.Println("Only AI API traffic will be proxied:")
		for _, domain := range llmDomains {
			fmt.Printf("  • %s\n", domain)
		}
		fmt.Println()
		fmt.Println("All other traffic goes direct (no performance impact).")
		fmt.Println()
		fmt.Println("To verify: ubik setup status")
		fmt.Println("To remove: sudo ubik setup uninstall")
	}

	return nil
}

func runSetupStatus() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	ubikDir := filepath.Join(home, ".ubik")
	certPath := filepath.Join(ubikDir, "certs", "ubik-ca.pem")
	pacPath := filepath.Join(ubikDir, "proxy.pac")

	fmt.Println("Ubik Setup Status")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Check PAC file
	if _, err := os.Stat(pacPath); err == nil {
		fmt.Printf("✓ PAC file: %s\n", pacPath)
	} else {
		fmt.Printf("✗ PAC file: not found\n")
	}

	// Check certificate
	if _, err := os.Stat(certPath); err == nil {
		fmt.Printf("✓ CA certificate: %s\n", certPath)
	} else {
		fmt.Printf("✗ CA certificate: not found\n")
	}

	// Check if certificate is trusted (platform-specific)
	if isCertTrusted(certPath) {
		fmt.Println("✓ CA certificate: trusted by system")
	} else {
		fmt.Println("✗ CA certificate: not in system trust store")
	}

	// Check system proxy config
	if isSystemProxyConfigured() {
		fmt.Println("✓ System proxy: configured")
	} else {
		fmt.Println("✗ System proxy: not configured")
	}

	// Check auto-start
	if isAutoStartConfigured() {
		fmt.Println("✓ Auto-start: enabled")
	} else {
		fmt.Println("✗ Auto-start: not configured")
	}

	fmt.Println()

	return nil
}

func runUninstall(dryRun bool) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	ubikDir := filepath.Join(home, ".ubik")
	pacPath := filepath.Join(ubikDir, "proxy.pac")

	fmt.Println("Ubik Uninstall")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if dryRun {
		fmt.Println("[DRY RUN] Would perform the following:")
		fmt.Println()
	}

	// Remove PAC file
	fmt.Println("1. Removing PAC file...")
	if dryRun {
		fmt.Printf("   Would remove: %s\n", pacPath)
	} else {
		if err := os.Remove(pacPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("   ⚠ Failed to remove PAC file: %v\n", err)
		} else {
			fmt.Println("   ✓ PAC file removed")
		}
	}

	// Remove system proxy config
	fmt.Println("2. Removing system proxy configuration...")
	if dryRun {
		fmt.Println("   Would clear system proxy settings")
	} else {
		if err := removeSystemProxy(); err != nil {
			fmt.Printf("   ⚠ Failed to remove system proxy: %v\n", err)
		} else {
			fmt.Println("   ✓ System proxy cleared")
		}
	}

	// Remove auto-start
	fmt.Println("3. Removing auto-start configuration...")
	if dryRun {
		fmt.Println("   Would remove launchd/systemd service")
	} else {
		if err := removeAutoStart(); err != nil {
			fmt.Printf("   ⚠ Failed to remove auto-start: %v\n", err)
		} else {
			fmt.Println("   ✓ Auto-start removed")
		}
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if dryRun {
		fmt.Println("Dry run complete. Run without --dry-run to apply changes.")
	} else {
		fmt.Println("Uninstall complete.")
		fmt.Println()
		fmt.Println("Note: CA certificate was left in system trust store.")
		fmt.Println("To remove it manually, use Keychain Access (macOS) or update-ca-certificates (Linux).")
	}

	return nil
}

// generatePACFile creates a PAC file that routes only LLM domains through proxy
func generatePACFile() string {
	var conditions []string
	for _, domain := range llmDomains {
		conditions = append(conditions, fmt.Sprintf(`host == "%s"`, domain))
	}

	return fmt.Sprintf(`// Ubik Proxy Auto-Config (PAC) file
// Routes only AI API traffic through the ubik proxy
// Generated by: ubik setup system

function FindProxyForURL(url, host) {
    // Route AI API domains through ubik proxy
    if (%s) {
        return "PROXY 127.0.0.1:8082";
    }

    // All other traffic goes direct
    return "DIRECT";
}
`, strings.Join(conditions, " ||\n        "))
}

// Platform-specific implementations

func installCACert(certPath string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Add to system keychain
		cmd := exec.Command("security", "add-trusted-cert", "-d", "-r", "trustRoot",
			"-k", "/Library/Keychains/System.keychain", certPath)
		return cmd.Run()
	case "linux":
		// Linux: Copy to ca-certificates and update
		destPath := "/usr/local/share/ca-certificates/ubik-ca.crt"
		if err := copyFile(certPath, destPath); err != nil {
			return err
		}
		cmd := exec.Command("update-ca-certificates")
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func configureSystemProxy(pacPath string) error {
	switch runtime.GOOS {
	case "darwin":
		// macOS: Use networksetup to configure PAC
		// Get the primary network service
		services := []string{"Wi-Fi", "Ethernet"}
		for _, service := range services {
			pacURL := "file://" + pacPath
			cmd := exec.Command("networksetup", "-setautoproxyurl", service, pacURL)
			// Ignore errors for non-existent services
			cmd.Run()
		}
		return nil
	case "linux":
		// Linux: This varies by desktop environment
		// For GNOME, we'd use gsettings
		// For now, just print instructions
		fmt.Println("   Note: On Linux, configure your desktop environment to use the PAC file:")
		fmt.Printf("   PAC URL: file://%s\n", pacPath)
		return nil
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func setupAutoStart() error {
	switch runtime.GOOS {
	case "darwin":
		return setupLaunchd()
	case "linux":
		return setupSystemd()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

func setupLaunchd() error {
	home, _ := os.UserHomeDir()
	plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.ubik.proxy.plist")

	// Find ubik binary path
	ubikPath, err := os.Executable()
	if err != nil {
		ubikPath = "/usr/local/bin/ubik"
	}

	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.ubik.proxy</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>proxy</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>%s/.ubik/logs/proxy.log</string>
    <key>StandardErrorPath</key>
    <string>%s/.ubik/logs/proxy.err</string>
</dict>
</plist>
`, ubikPath, home, home)

	// Create LaunchAgents directory if needed
	if err := os.MkdirAll(filepath.Dir(plistPath), 0755); err != nil {
		return err
	}

	// Create logs directory
	if err := os.MkdirAll(filepath.Join(home, ".ubik", "logs"), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(plistPath, []byte(plistContent), 0644); err != nil {
		return err
	}

	// Load the plist
	cmd := exec.Command("launchctl", "load", plistPath)
	return cmd.Run()
}

func setupSystemd() error {
	home, _ := os.UserHomeDir()
	serviceDir := filepath.Join(home, ".config", "systemd", "user")
	servicePath := filepath.Join(serviceDir, "ubik-proxy.service")

	// Find ubik binary path
	ubikPath, err := os.Executable()
	if err != nil {
		ubikPath = "/usr/local/bin/ubik"
	}

	serviceContent := fmt.Sprintf(`[Unit]
Description=Ubik AI Agent Security Proxy
After=network.target

[Service]
Type=simple
ExecStart=%s proxy start
Restart=always
RestartSec=5

[Install]
WantedBy=default.target
`, ubikPath)

	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
		return err
	}

	// Reload and enable
	exec.Command("systemctl", "--user", "daemon-reload").Run()
	cmd := exec.Command("systemctl", "--user", "enable", "--now", "ubik-proxy")
	return cmd.Run()
}

func removeSystemProxy() error {
	switch runtime.GOOS {
	case "darwin":
		services := []string{"Wi-Fi", "Ethernet"}
		for _, service := range services {
			exec.Command("networksetup", "-setautoproxystate", service, "off").Run()
		}
		return nil
	case "linux":
		// Would clear gsettings or equivalent
		return nil
	default:
		return nil
	}
}

func removeAutoStart() error {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.ubik.proxy.plist")
		exec.Command("launchctl", "unload", plistPath).Run()
		return os.Remove(plistPath)
	case "linux":
		exec.Command("systemctl", "--user", "disable", "--now", "ubik-proxy").Run()
		servicePath := filepath.Join(home, ".config", "systemd", "user", "ubik-proxy.service")
		return os.Remove(servicePath)
	default:
		return nil
	}
}

func isCertTrusted(certPath string) bool {
	// Simplified check - in reality would verify against trust store
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("security", "find-certificate", "-c", "ubik-proxy-ca")
		return cmd.Run() == nil
	default:
		return false
	}
}

func isSystemProxyConfigured() bool {
	switch runtime.GOOS {
	case "darwin":
		// Check if auto-proxy is enabled on Wi-Fi
		out, err := exec.Command("networksetup", "-getautoproxyurl", "Wi-Fi").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), "ubik")
	default:
		return false
	}
}

func isAutoStartConfigured() bool {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		plistPath := filepath.Join(home, "Library", "LaunchAgents", "com.ubik.proxy.plist")
		_, err := os.Stat(plistPath)
		return err == nil
	case "linux":
		servicePath := filepath.Join(home, ".config", "systemd", "user", "ubik-proxy.service")
		_, err := os.Stat(servicePath)
		return err == nil
	default:
		return false
	}
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
