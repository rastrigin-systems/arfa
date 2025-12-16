package cli_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestAuthService_RequireAuth_WithMocks tests RequireAuth using gomock.
// This demonstrates the mock-based testing pattern for AuthService.
func TestAuthService_RequireAuth_WithMocks(t *testing.T) {
	t.Run("success - authenticated with valid token", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		expectedConfig := &cli.Config{
			PlatformURL: "https://api.example.com",
			Token:       "valid-token",
			EmployeeID:  "emp-123",
		}

		// Set expectations
		mockConfigManager.EXPECT().IsAuthenticated().Return(true, nil)
		mockConfigManager.EXPECT().IsTokenValid().Return(true, nil)
		mockConfigManager.EXPECT().Load().Return(expectedConfig, nil)
		mockPlatformClient.EXPECT().SetToken("valid-token")
		mockPlatformClient.EXPECT().SetBaseURL("https://api.example.com")

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		config, err := authService.RequireAuth()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedConfig.Token, config.Token)
		assert.Equal(t, expectedConfig.EmployeeID, config.EmployeeID)
	})

	t.Run("failure - not authenticated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations - not authenticated
		mockConfigManager.EXPECT().IsAuthenticated().Return(false, nil)

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		config, err := authService.RequireAuth()

		// Assert
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("failure - token expired", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations - authenticated but token expired
		mockConfigManager.EXPECT().IsAuthenticated().Return(true, nil)
		mockConfigManager.EXPECT().IsTokenValid().Return(false, nil)

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		config, err := authService.RequireAuth()

		// Assert
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "token has expired")
	})

	t.Run("failure - config manager error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations - error from config manager
		mockConfigManager.EXPECT().IsAuthenticated().Return(false, errors.New("config read error"))

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		config, err := authService.RequireAuth()

		// Assert
		require.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "failed to check authentication")
	})
}

// TestAuthService_Login_WithMocks tests Login using gomock.
func TestAuthService_Login_WithMocks(t *testing.T) {
	t.Run("success - login saves config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		ctx := context.Background()
		platformURL := "https://api.example.com"
		email := "test@example.com"
		password := "secret123"

		loginResponse := &cli.LoginResponse{
			Token:     "new-token",
			ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			Employee: cli.LoginEmployeeInfo{
				ID:    "emp-456",
				OrgID: "org-789",
				Email: email,
			},
		}

		// Set expectations
		mockPlatformClient.EXPECT().SetBaseURL(platformURL)
		mockPlatformClient.EXPECT().Login(ctx, email, password).Return(loginResponse, nil)
		mockConfigManager.EXPECT().Save(gomock.Any()).DoAndReturn(func(config *cli.Config) error {
			assert.Equal(t, platformURL, config.PlatformURL)
			assert.Equal(t, "new-token", config.Token)
			assert.Equal(t, "emp-456", config.EmployeeID)
			return nil
		})

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		err := authService.Login(ctx, platformURL, email, password)

		// Assert
		require.NoError(t, err)
	})

	t.Run("failure - authentication failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		ctx := context.Background()
		platformURL := "https://api.example.com"
		email := "test@example.com"
		password := "wrong-password"

		// Set expectations - login fails
		mockPlatformClient.EXPECT().SetBaseURL(platformURL)
		mockPlatformClient.EXPECT().Login(ctx, email, password).Return(nil, errors.New("invalid credentials"))

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		err := authService.Login(ctx, platformURL, email, password)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("failure - config save fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		ctx := context.Background()
		platformURL := "https://api.example.com"
		email := "test@example.com"
		password := "secret123"

		loginResponse := &cli.LoginResponse{
			Token:     "new-token",
			ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			Employee: cli.LoginEmployeeInfo{
				ID:    "emp-456",
				OrgID: "org-789",
				Email: email,
			},
		}

		// Set expectations - login succeeds but save fails
		mockPlatformClient.EXPECT().SetBaseURL(platformURL)
		mockPlatformClient.EXPECT().Login(ctx, email, password).Return(loginResponse, nil)
		mockConfigManager.EXPECT().Save(gomock.Any()).Return(errors.New("disk full"))

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		err := authService.Login(ctx, platformURL, email, password)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save config")
	})
}

// TestAuthService_Logout_WithMocks tests Logout using gomock.
func TestAuthService_Logout_WithMocks(t *testing.T) {
	t.Run("success - clears config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations
		mockConfigManager.EXPECT().Clear().Return(nil)

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		err := authService.Logout()

		// Assert
		require.NoError(t, err)
	})

	t.Run("failure - clear fails", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations
		mockConfigManager.EXPECT().Clear().Return(errors.New("permission denied"))

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		err := authService.Logout()

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to clear config")
	})
}

// TestAuthService_IsAuthenticated_WithMocks tests IsAuthenticated using gomock.
func TestAuthService_IsAuthenticated_WithMocks(t *testing.T) {
	t.Run("returns config manager result", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		// Set expectations
		mockConfigManager.EXPECT().IsAuthenticated().Return(true, nil)

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		authenticated, err := authService.IsAuthenticated()

		// Assert
		require.NoError(t, err)
		assert.True(t, authenticated)
	})
}

// TestAuthService_GetConfig_WithMocks tests GetConfig using gomock.
func TestAuthService_GetConfig_WithMocks(t *testing.T) {
	t.Run("returns config from config manager", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockPlatformClientInterface(ctrl)

		expectedConfig := &cli.Config{
			PlatformURL: "https://api.example.com",
			Token:       "test-token",
			EmployeeID:  "emp-123",
		}

		// Set expectations
		mockConfigManager.EXPECT().Load().Return(expectedConfig, nil)

		// Create service with mocks
		authService := cli.NewAuthServiceWithInterfaces(mockConfigManager, mockPlatformClient)

		// Execute
		config, err := authService.GetConfig()

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedConfig, config)
	})
}
