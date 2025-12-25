package logging

import (
	"context"
	"fmt"
	"sort"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/types"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// GetClassifiedLogs retrieves classified logs from the API using the provided api.Client.
// The api.Client must have a valid token set.
func GetClassifiedLogs(ctx context.Context, apiClient *api.Client, sessionID string) ([]types.ClassifiedLogEntry, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("API client is required")
	}

	// Fetch logs using the api.Client
	params := api.GetLogsParams{
		EventCategory: "classified",
		PerPage:       1000,
	}
	if sessionID != "" {
		params.SessionID = sessionID
	}

	apiResp, err := apiClient.GetLogs(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	// Convert API logs to ClassifiedLogEntry
	var classifiedLogs []types.ClassifiedLogEntry
	for _, log := range apiResp.Logs {
		entry := convertAPILogToClassified(log)
		classifiedLogs = append(classifiedLogs, entry)
	}

	// Sort by timestamp
	sort.Slice(classifiedLogs, func(i, j int) bool {
		return classifiedLogs[i].Timestamp.Before(classifiedLogs[j].Timestamp)
	})

	return classifiedLogs, nil
}

// GetClassifiedLogsWithConfig retrieves classified logs using config for authentication.
// This is a convenience function that creates an api.Client from the config.
func GetClassifiedLogsWithConfig(configManager *config.Manager, sessionID string) ([]types.ClassifiedLogEntry, error) {
	cfg, err := configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("not authenticated - please run 'ubik login' first")
	}

	// Create api.Client with config
	apiClient := api.NewClient(cfg.PlatformURL)
	apiClient.SetToken(cfg.Token)

	return GetClassifiedLogs(context.Background(), apiClient, sessionID)
}

// convertAPILogToClassified converts an API log entry to a ClassifiedLogEntry
func convertAPILogToClassified(log api.LogEntryResponse) types.ClassifiedLogEntry {
	entry := types.ClassifiedLogEntry{
		ID:        log.ID,
		SessionID: log.SessionID,
		Timestamp: log.CreatedAt,
		EntryType: types.LogEntryType(log.EventType),
		Content:   log.Content,
	}

	// Extract fields from payload
	if log.Payload != nil {
		if provider, ok := log.Payload["provider"].(string); ok {
			entry.Provider = types.LogProvider(provider)
		}
		if model, ok := log.Payload["model"].(string); ok {
			entry.Model = model
		}
		if tokensInput, ok := log.Payload["tokens_input"].(float64); ok {
			entry.TokensInput = int(tokensInput)
		}
		if tokensOutput, ok := log.Payload["tokens_output"].(float64); ok {
			entry.TokensOutput = int(tokensOutput)
		}
		if toolName, ok := log.Payload["tool_name"].(string); ok {
			entry.ToolName = toolName
		}
		if toolID, ok := log.Payload["tool_id"].(string); ok {
			entry.ToolID = toolID
		}
		if toolInput, ok := log.Payload["tool_input"].(map[string]interface{}); ok {
			entry.ToolInput = toolInput
		}
		if errorCode, ok := log.Payload["error_code"].(string); ok {
			entry.ErrorCode = errorCode
		}
		if errorMessage, ok := log.Payload["error_message"].(string); ok {
			entry.ErrorMessage = errorMessage
		}
	}

	return entry
}
