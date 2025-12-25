// Package auth handles authentication operations for the CLI.
package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"golang.org/x/term"
)

// Service handles authentication operations.
type Service struct {
	configManager ConfigManagerInterface
	apiClient     APIClientInterface
}

// NewService creates a new Service with concrete types.
// This is the primary constructor for production use.
func NewService(configManager *config.Manager, apiClient *api.Client) *Service {
	return &Service{
		configManager: configManager,
		apiClient:     apiClient,
	}
}

// NewServiceWithInterfaces creates a new Service with interface types.
// This constructor enables dependency injection for testing with mocks.
func NewServiceWithInterfaces(configManager ConfigManagerInterface, apiClient APIClientInterface) *Service {
	return &Service{
		configManager: configManager,
		apiClient:     apiClient,
	}
}

// LoginInteractive performs interactive login.
func (s *Service) LoginInteractive(ctx context.Context) error {
	reader := bufio.NewReader(os.Stdin)

	// Load existing config to get saved platform URL
	savedConfig, _ := s.configManager.Load()
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
	s.apiClient.SetBaseURL(platformURL)

	// Perform login
	fmt.Println("\nAuthenticating...")
	loginResp, err := s.apiClient.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Parse expiration time
	expiresAt, err := parseExpiresAt(loginResp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse token expiration: %w", err)
	}

	// Save config
	cfg := &config.Config{
		PlatformURL:  platformURL,
		Token:        loginResp.Token,
		TokenExpires: expiresAt,
		EmployeeID:   loginResp.Employee.ID,
		OrgID:        loginResp.Employee.OrgID,
	}

	if err := s.configManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Println("✓ Authenticated successfully")
	fmt.Printf("✓ Employee ID: %s\n", loginResp.Employee.ID)

	return nil
}

// Login performs non-interactive login.
func (s *Service) Login(ctx context.Context, platformURL, email, password string) error {
	// Update platform client URL
	s.apiClient.SetBaseURL(platformURL)

	// Perform login
	loginResp, err := s.apiClient.Login(ctx, email, password)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Parse expiration time
	expiresAt, err := parseExpiresAt(loginResp.ExpiresAt)
	if err != nil {
		return fmt.Errorf("failed to parse token expiration: %w", err)
	}

	// Save config
	cfg := &config.Config{
		PlatformURL:  platformURL,
		Token:        loginResp.Token,
		TokenExpires: expiresAt,
		EmployeeID:   loginResp.Employee.ID,
		OrgID:        loginResp.Employee.OrgID,
	}

	if err := s.configManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// Logout removes stored credentials.
func (s *Service) Logout() error {
	if err := s.configManager.Clear(); err != nil {
		return fmt.Errorf("failed to clear config: %w", err)
	}
	fmt.Println("✓ Logged out successfully")
	return nil
}

// IsAuthenticated checks if user is authenticated.
func (s *Service) IsAuthenticated() (bool, error) {
	return s.configManager.IsAuthenticated()
}

// GetConfig returns the current config.
func (s *Service) GetConfig() (*config.Config, error) {
	return s.configManager.Load()
}

// RequireAuth ensures user is authenticated, returns error if not.
func (s *Service) RequireAuth() (*config.Config, error) {
	authenticated, err := s.IsAuthenticated()
	if err != nil {
		return nil, fmt.Errorf("failed to check authentication: %w", err)
	}

	if !authenticated {
		return nil, fmt.Errorf("not authenticated. Please run 'arfa login' first")
	}

	// Check if token is still valid (not expired)
	tokenValid, err := s.configManager.IsTokenValid()
	if err != nil {
		return nil, fmt.Errorf("failed to check token validity: %w", err)
	}

	if !tokenValid {
		return nil, fmt.Errorf("authentication token has expired. Please run 'arfa login' again")
	}

	cfg, err := s.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set token on platform client
	s.apiClient.SetToken(cfg.Token)
	s.apiClient.SetBaseURL(cfg.PlatformURL)

	return cfg, nil
}

// parseExpiresAt parses the expiration timestamp from the API.
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
