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

// TestSyncService_Sync_WithMocks tests the Sync function using gomock.
func TestSyncService_Sync_WithMocks(t *testing.T) {
	t.Run("success - syncs agent configs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)

		ctx := context.Background()

		config := &cli.Config{
			PlatformURL: "https://api.example.com",
			Token:       "test-token",
			EmployeeID:  "emp-123",
		}

		agentConfigs := []cli.AgentConfig{
			{
				AgentID:   "agent-1",
				AgentName: "Claude Code",
				AgentType: "claude-code",
				Provider:  "anthropic",
				IsEnabled: true,
			},
		}

		// Set expectations
		mockAuthService.EXPECT().RequireAuth().Return(config, nil)
		mockPlatformClient.EXPECT().GetMyResolvedAgentConfigs(ctx).Return(agentConfigs, nil)
		// Save is called twice: once for agent configs, once for updating last sync
		mockConfigManager.EXPECT().GetConfigPath().Return("/tmp/test/config.json").AnyTimes()
		mockConfigManager.EXPECT().Save(gomock.Any()).Return(nil)

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

		// Execute
		result, err := syncService.Sync(ctx)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.AgentConfigs, 1)
		assert.Equal(t, "Claude Code", result.AgentConfigs[0].AgentName)
	})

	t.Run("failure - not authenticated", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)

		ctx := context.Background()

		// Set expectations - auth fails
		mockAuthService.EXPECT().RequireAuth().Return(nil, errors.New("not authenticated"))

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

		// Execute
		result, err := syncService.Sync(ctx)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("failure - platform API error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)

		ctx := context.Background()

		config := &cli.Config{
			PlatformURL: "https://api.example.com",
			Token:       "test-token",
			EmployeeID:  "emp-123",
		}

		// Set expectations - API call fails
		mockAuthService.EXPECT().RequireAuth().Return(config, nil)
		mockPlatformClient.EXPECT().GetMyResolvedAgentConfigs(ctx).Return(nil, errors.New("network error"))

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

		// Execute
		result, err := syncService.Sync(ctx)

		// Assert
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to fetch agent configs")
	})

	t.Run("success - empty configs", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)

		ctx := context.Background()

		config := &cli.Config{
			PlatformURL: "https://api.example.com",
			Token:       "test-token",
			EmployeeID:  "emp-123",
		}

		// Set expectations - no configs returned
		mockAuthService.EXPECT().RequireAuth().Return(config, nil)
		mockPlatformClient.EXPECT().GetMyResolvedAgentConfigs(ctx).Return([]cli.AgentConfig{}, nil)

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

		// Execute
		result, err := syncService.Sync(ctx)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.AgentConfigs, 0)
	})
}

// Note: GetLocalAgentConfigs and GetAgentConfig tests are omitted because they
// depend on file system operations that use a cached agent configs list rather
// than calling GetConfigPath() directly. These methods are better tested with
// integration tests that set up actual files.

// TestSyncService_SetContainerManager tests the SetContainerManager method.
func TestSyncService_SetContainerManager(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
	mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
	mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)
	mockContainerManager := mocks.NewMockContainerManagerInterface(ctrl)

	// Create service with mocks
	syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

	// Set container manager
	syncService.SetContainerManager(mockContainerManager)

	// The container manager should be set (we can't directly verify this without getter,
	// but we can verify no panic occurs)
}

// TestSyncService_StopContainers_WithMocks tests StopContainers using gomock.
func TestSyncService_StopContainers_WithMocks(t *testing.T) {
	t.Run("success - stops containers", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)
		mockContainerManager := mocks.NewMockContainerManagerInterface(ctrl)

		ctx := context.Background()

		// Set expectations
		mockContainerManager.EXPECT().StopContainers(ctx).Return(nil)

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)
		syncService.SetContainerManager(mockContainerManager)

		// Execute
		err := syncService.StopContainers(ctx)

		// Assert
		require.NoError(t, err)
	})

	t.Run("failure - container manager not set", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)

		ctx := context.Background()

		// Create service with mocks - no container manager set
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)

		// Execute
		err := syncService.StopContainers(ctx)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "container manager not configured")
	})
}

// TestSyncService_GetContainerStatus_WithMocks tests GetContainerStatus using gomock.
func TestSyncService_GetContainerStatus_WithMocks(t *testing.T) {
	t.Run("success - returns container status", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConfigManager := mocks.NewMockConfigManagerInterface(ctrl)
		mockPlatformClient := mocks.NewMockAPIClientInterface(ctrl)
		mockAuthService := mocks.NewMockAuthServiceInterface(ctrl)
		mockContainerManager := mocks.NewMockContainerManagerInterface(ctrl)

		ctx := context.Background()

		expectedStatus := []cli.ContainerInfo{
			{
				ID:      "container-1",
				Name:    "test-container",
				Image:   "test-image:latest",
				State:   "running",
				Status:  "Up 5 minutes",
				Created: time.Now().Unix(),
			},
		}

		// Set expectations
		mockContainerManager.EXPECT().GetContainerStatus(ctx).Return(expectedStatus, nil)

		// Create service with mocks
		syncService := cli.NewSyncServiceWithInterfaces(mockConfigManager, mockPlatformClient, mockAuthService)
		syncService.SetContainerManager(mockContainerManager)

		// Execute
		status, err := syncService.GetContainerStatus(ctx)

		// Assert
		require.NoError(t, err)
		assert.Len(t, status, 1)
		assert.Equal(t, "test-container", status[0].Name)
	})
}
