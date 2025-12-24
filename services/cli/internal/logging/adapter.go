package logging

import (
	"context"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
)

// APIClientAdapter adapts the CLI api.Client to the logging APIClient interface
type APIClientAdapter struct {
	client *api.Client
}

// NewAPIClientAdapter creates a new adapter for the API client
func NewAPIClientAdapter(client *api.Client) APIClient {
	return &APIClientAdapter{client: client}
}

// CreateLog sends a single log entry to the API
func (a *APIClientAdapter) CreateLog(ctx context.Context, entry LogEntry) error {
	apiEntry := api.LogEntry{
		SessionID:     entry.SessionID,
		ClientName:    entry.ClientName,
		ClientVersion: entry.ClientVersion,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       entry.Content,
		Payload:       entry.Payload,
	}

	return a.client.CreateLog(ctx, apiEntry)
}

// CreateLogBatch sends multiple log entries in a single request
func (a *APIClientAdapter) CreateLogBatch(ctx context.Context, entries []LogEntry) error {
	apiEntries := make([]api.LogEntry, len(entries))
	for i, entry := range entries {
		apiEntries[i] = api.LogEntry{
			SessionID:     entry.SessionID,
			ClientName:    entry.ClientName,
			ClientVersion: entry.ClientVersion,
			EventType:     entry.EventType,
			EventCategory: entry.EventCategory,
			Content:       entry.Content,
			Payload:       entry.Payload,
		}
	}

	return a.client.CreateLogBatch(ctx, apiEntries)
}
