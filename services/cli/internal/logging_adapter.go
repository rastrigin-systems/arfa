package cli

import (
	"context"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
)

// LoggingAPIClientAdapter adapts the CLI APIClient to the logging APIClient interface
type LoggingAPIClientAdapter struct {
	client *APIClient
}

// NewLoggingAPIClientAdapter creates a new adapter for the API client
func NewLoggingAPIClientAdapter(client *APIClient) logging.APIClient {
	return &LoggingAPIClientAdapter{client: client}
}

// CreateLog sends a single log entry to the API
func (a *LoggingAPIClientAdapter) CreateLog(ctx context.Context, entry logging.LogEntry) error {
	apiEntry := LogEntry{
		SessionID:     entry.SessionID,
		AgentID:       entry.AgentID,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       entry.Content,
		Payload:       entry.Payload,
	}

	return a.client.CreateLog(ctx, apiEntry)
}

// CreateLogBatch sends multiple log entries in a single request
func (a *LoggingAPIClientAdapter) CreateLogBatch(ctx context.Context, entries []logging.LogEntry) error {
	apiEntries := make([]LogEntry, len(entries))
	for i, entry := range entries {
		apiEntries[i] = LogEntry{
			SessionID:     entry.SessionID,
			AgentID:       entry.AgentID,
			EventType:     entry.EventType,
			EventCategory: entry.EventCategory,
			Content:       entry.Content,
			Payload:       entry.Payload,
		}
	}

	return a.client.CreateLogBatch(ctx, apiEntries)
}
