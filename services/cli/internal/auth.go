package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

// AuthService handles authentication operations
type AuthService struct {
	configManager  ConfigManagerInterface
	platformClient PlatformClientInterface
}

// NewAuthService creates a new AuthService with concrete types.
// This is the primary constructor for production use.
func NewAuthService(configManager *ConfigManager, platformClient *PlatformClient) *AuthService {
	return &AuthService{
		configManager:  configManager,
		platformClient: platformClient,
	}
}

// NewAuthServiceWithInterfaces creates a new AuthService with interface types.
// This constructor enables dependency injection for testing with mocks.
func NewAuthServiceWithInterfaces(configManager ConfigManagerInterface, platformClient PlatformClientInterface) *AuthService {
	return &AuthService{
		configManager:  configManager,
		platformClient: platformClient,
	}
}

// LoginInteractive performs interactive login
func (as *AuthService) LoginInteractive(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config to get saved platform URL
	savedConfig, _ := as.configManager.Load()
	defaultURL := "http://localhost:8080"
	if savedConfig != nil && savedConfig.PlatformURL != "" {
		defaultURL = savedConfig.PlatformURL
	}

	// Get platform URL
	fmt.Printf("Platform URL [%s]: ", defaultURL)
	platformURL, _ := reader.ReadString('\n')
	platformURL = strings.TrimSpace(platformURL)
	if platformURL == "" {
		platformURL = defaultURL
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
	as.platformClient.SetBaseURL(platformURL)

	// Perform login
	fmt.Println("\nAuthenticating...")
	loginResp, err := as.platformClient.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Parse expiration time
	expiresAt, err := parseExpiresAt(loginResp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse token expiration: %w", err)
	}

	// Save config
	config := &Config{
		PlatformURL:  platformURL,
		Token:        loginResp.Token,
		TokenExpires: expiresAt,
		EmployeeID:   loginResp.Employee.ID,
	}

	if err := as.configManager.Save(config); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("✓ Authenticated successfully")
	fmt.Printf("✓ Employee ID: %s\n", loginResp.Employee.ID)

	return nil
}

// Login performs non-interactive login
func (as *AuthService) Login(ctx context.Context, platformURL, email, password string) error {
	// Update platform client URL
	as.platformClient.SetBaseURL(platformURL)

	// Perform login
	loginResp, err := as.platformClient.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Parse expiration time
	expiresAt, err := parseExpiresAt(loginResp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse token expiration: %w", err)
	}

	// Save config
	config := &Config{
		PlatformURL:  platformURL,
		Token:        loginResp.Token,
		TokenExpires: expiresAt,
		EmployeeID:   loginResp.Employee.ID,
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

	// Check if token is still valid (not expired)
	tokenValid, err := as.configManager.IsTokenValid()
	if err != nil {
		return nil, fmt.Errorf("failed to check token validity: %w", err)
	}

	if !tokenValid {
		return nil, fmt.Errorf("authentication token has expired. Please run 'ubik login' again")
	}

	config, err := as.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set token on platform client
	as.platformClient.SetToken(config.Token)
	as.platformClient.SetBaseURL(config.PlatformURL)

	return config, nil
}

// parseExpiresAt parses the expiration timestamp from the API
func parseExpiresAt(expiresAt string) (time.Time, error) {
	if expiresAt == "" {
		// If no expiration provided, default to 24 hours from now
		return time.Now().Add(24 * time.Hour), nil
	}

	// Try RFC3339 format first
	t, err := time.Parse(time.RFC3339, expiresAt)
	if err == nil {
		return t, nil
	}

	// Try RFC3339Nano format
	t, err = time.Parse(time.RFC3339Nano, expiresAt)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid timestamp format: %s", expiresAt)
}
