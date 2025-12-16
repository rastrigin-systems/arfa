package cli

import (
	"context"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
)

// PlatformAPIClient adapts the CLI PlatformClient to the logging APIClient interface
type PlatformAPIClient struct {
	client *PlatformClient
}

// NewPlatformAPIClient creates a new adapter for the platform client
func NewPlatformAPIClient(client *PlatformClient) logging.APIClient {
	return &PlatformAPIClient{client: client}
}

// CreateLog sends a single log entry to the API
func (p *PlatformAPIClient) CreateLog(ctx context.Context, entry logging.LogEntry) error {
	platformEntry := LogEntry{
		SessionID:     entry.SessionID,
		AgentID:       entry.AgentID,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       entry.Content,
		Payload:       entry.Payload,
	}

	return p.client.CreateLog(ctx, platformEntry)
}

// CreateLogBatch sends multiple log entries in a single request
func (p *PlatformAPIClient) CreateLogBatch(ctx context.Context, entries []logging.LogEntry) error {
	platformEntries := make([]LogEntry, len(entries))
	for i, entry := range entries {
		platformEntries[i] = LogEntry{
			SessionID:     entry.SessionID,
			AgentID:       entry.AgentID,
			EventType:     entry.EventType,
			EventCategory: entry.EventCategory,
			Content:       entry.Content,
			Payload:       entry.Payload,
		}
	}

	return p.client.CreateLogBatch(ctx, platformEntries)
}
