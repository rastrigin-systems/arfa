package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/pkg/types"
)

// APILogEntry represents a log entry from the API response
type APILogEntry struct {
	ID            string                 `json:"id"`
	SessionID     string                 `json:"session_id"`
	AgentID       string                 `json:"agent_id,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// APILogsResponse represents the paginated logs response from the API
type APILogsResponse struct {
	Logs       []APILogEntry `json:"logs"`
	TotalCount int           `json:"total_count"`
	Page       int           `json:"page"`
	PerPage    int           `json:"per_page"`
}

// GetClassifiedLogs retrieves classified logs from the API, optionally filtered by session ID
func GetClassifiedLogs(configManager *ConfigManager, sessionID string) ([]types.ClassifiedLogEntry, error) {
	config, err := configManager.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	if config.Token == "" {
		return nil, fmt.Errorf("not authenticated - please run 'ubik login' first")
	}

	// Build API URL with filters
	apiURL, err := url.Parse(config.PlatformURL + "/api/v1/logs")
	if err != nil {
		return nil, fmt.Errorf("invalid API URL: %w", err)
	}

	query := apiURL.Query()
	query.Set("event_category", "classified")
	// Note: employee_id is automatically derived from JWT token on the backend
	query.Set("per_page", "1000") // Get a large batch
	if sessionID != "" {
		query.Set("session_id", sessionID)
	}
	apiURL.RawQuery = query.Encode()

	// Make API request
	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResp APILogsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
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

// convertAPILogToClassified converts an API log entry to a ClassifiedLogEntry
func convertAPILogToClassified(log APILogEntry) types.ClassifiedLogEntry {
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
