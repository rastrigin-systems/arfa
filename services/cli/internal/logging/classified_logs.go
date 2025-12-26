package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
	"github.com/rastrigin-systems/arfa/services/cli/internal/types"
)

// GetClassifiedLogs retrieves classified logs from the API using the provided api.Client.
// The api.Client must have a valid token set.
func GetClassifiedLogs(ctx context.Context, apiClient *api.Client) ([]types.ClassifiedLogEntry, error) {
	if apiClient == nil {
		return nil, fmt.Errorf("API client is required")
	}

	// Fetch logs using the api.Client
	params := api.GetLogsParams{
		EventCategory: "classified",
		PerPage:       1000,
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
func GetClassifiedLogsWithConfig(configManager *config.Manager) ([]types.ClassifiedLogEntry, error) {
	return GetLogsWithConfig(configManager, "classified", 100, 0)
}

// GetLogsWithConfig retrieves logs filtered by category using config for authentication.
// Category can be: "classified", "proxy", "session", or "all"
// limit: maximum number of logs to return (0 for all, default 100)
// offset: number of logs to skip for pagination (default 0)
func GetLogsWithConfig(configManager *config.Manager, category string, limit, offset int) ([]types.ClassifiedLogEntry, error) {
	cfg, err := configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		return nil, fmt.Errorf("not authenticated - please run 'arfa login' first")
	}

	// Create api.Client with config
	apiClient := api.NewClient(cfg.PlatformURL)
	apiClient.SetToken(cfg.Token)

	// Fetch logs using the api.Client
	params := api.GetLogsParams{
		PerPage: 10000, // Fetch a large batch, we'll do client-side pagination
	}

	// Only filter by category if not "all"
	if category != "all" {
		params.EventCategory = category
	}

	apiResp, err := apiClient.GetLogs(context.Background(), params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	// Convert API logs to ClassifiedLogEntry
	var classifiedLogs []types.ClassifiedLogEntry
	for _, log := range apiResp.Logs {
		entry := convertAPILogToClassified(log)
		classifiedLogs = append(classifiedLogs, entry)
	}

	// Sort by timestamp (descending - newest first)
	sort.Slice(classifiedLogs, func(i, j int) bool {
		return classifiedLogs[i].Timestamp.After(classifiedLogs[j].Timestamp)
	})

	// Apply client-side pagination
	start := offset
	if start >= len(classifiedLogs) {
		return []types.ClassifiedLogEntry{}, nil
	}

	end := start + limit
	if limit == 0 || end > len(classifiedLogs) {
		end = len(classifiedLogs)
	}

	return classifiedLogs[start:end], nil
}

// convertAPILogToClassified converts an API log entry to a ClassifiedLogEntry
func convertAPILogToClassified(log api.LogEntryResponse) types.ClassifiedLogEntry {
	entry := types.ClassifiedLogEntry{
		ID:        log.ID,
		Timestamp: log.CreatedAt,
		EntryType: types.LogEntryType(log.EventType),
		Content:   log.Content,
	}

	// For proxy logs (api_request, api_response), include the full payload as JSON
	if log.Payload != nil && (log.EventType == "api_request" || log.EventType == "api_response") {
		if payloadJSON, err := json.Marshal(log.Payload); err == nil {
			entry.Content = string(payloadJSON)
		}
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
