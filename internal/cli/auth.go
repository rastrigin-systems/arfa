package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// AuthService handles authentication operations
type AuthService struct {
	configManager  *ConfigManager
	platformClient *PlatformClient
}

// NewAuthService creates a new AuthService
func NewAuthService(configManager *ConfigManager, platformClient *PlatformClient) *AuthService {
	return &AuthService{
		configManager:  configManager,
		platformClient: platformClient,
	}
}

// LoginInteractive performs interactive login
func (as *AuthService) LoginInteractive() error {
	reader := bufio.NewReader(os.Stdin)

	// Get platform URL
	fmt.Print("Platform URL [https://api.ubik.io]: ")
	platformURL, _ := reader.ReadString('\n')
	platformURL = strings.TrimSpace(platformURL)
	if platformURL == "" {
		platformURL = "https://api.ubik.io"
	}

	// Get email
	fmt.Print("Email: ")
	email, _ := reader.ReadString('\n')
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("email is required")
	}

	// Get password (hidden input)
	fmt.Print("Password: ")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print newline after password input
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	password := string(passwordBytes)
	if password == "" {
		return fmt.Errorf("password is required")
	}

	// Update platform client URL
	as.platformClient.baseURL = platformURL

	// Perform login
	fmt.Println("\nAuthenticating...")
	loginResp, err := as.platformClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save config
	config := &Config{
		PlatformURL: platformURL,
		Token:       loginResp.Token,
		EmployeeID:  loginResp.EmployeeID,
	}

	if err := as.configManager.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("✓ Authenticated successfully")
	fmt.Printf("✓ Employee ID: %s\n", loginResp.EmployeeID)

	return nil
}

// Login performs non-interactive login
func (as *AuthService) Login(platformURL, email, password string) error {
	// Update platform client URL
	as.platformClient.baseURL = platformURL

	// Perform login
	loginResp, err := as.platformClient.Login(email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save config
	config := &Config{
		PlatformURL: platformURL,
		Token:       loginResp.Token,
		EmployeeID:  loginResp.EmployeeID,
	}

	if err := as.configManager.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Logout removes stored credentials
func (as *AuthService) Logout() error {
	if err := as.configManager.Clear(); err != nil {
		return fmt.Errorf("failed to clear config: %w", err)
	}
	fmt.Println("✓ Logged out successfully")
	return nil
}

// IsAuthenticated checks if user is authenticated
func (as *AuthService) IsAuthenticated() (bool, error) {
	return as.configManager.IsAuthenticated()
}

// GetConfig returns the current config
func (as *AuthService) GetConfig() (*Config, error) {
	return as.configManager.Load()
}

// RequireAuth ensures user is authenticated, returns error if not
func (as *AuthService) RequireAuth() (*Config, error) {
	authenticated, err := as.IsAuthenticated()
	if err != nil {
		return nil, fmt.Errorf("failed to check authentication: %w", err)
	}

	if !authenticated {
		return nil, fmt.Errorf("not authenticated. Please run 'ubik login' first")
	}

	config, err := as.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set token on platform client
	as.platformClient.SetToken(config.Token)
	as.platformClient.baseURL = config.PlatformURL

	return config, nil
}
