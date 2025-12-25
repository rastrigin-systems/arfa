// Package auth handles authentication operations for the CLI.
package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

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

	// Save config (only platform_url and token - claims are in JWT)
	cfg := &config.Config{
		PlatformURL: platformURL,
		Token:       loginResp.Token,
	}

	if err := s.configManager.Save(cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Parse claims from token to display employee info
	claims, _ := cfg.GetClaims()
	fmt.Println("✓ Authenticated successfully")
	if claims != nil {
		fmt.Printf("✓ Employee ID: %s\n", claims.EmployeeID)
	}

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

	// Save config (only platform_url and token - claims are in JWT)
	cfg := &config.Config{
		PlatformURL: platformURL,
		Token:       loginResp.Token,
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
