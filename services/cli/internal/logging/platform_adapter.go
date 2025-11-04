package logging

import (
	"context"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
)

// PlatformAPIClient adapts the CLI PlatformClient to the logging APIClient interface
type PlatformAPIClient struct {
	client *cli.PlatformClient
}

// NewPlatformAPIClient creates a new adapter for the platform client
func NewPlatformAPIClient(client *cli.PlatformClient) APIClient {
	return &PlatformAPIClient{client: client}
}

// CreateLog sends a single log entry to the API
func (p *PlatformAPIClient) CreateLog(ctx context.Context, entry LogEntry) error {
	platformEntry := cli.LogEntry{
		SessionID:     entry.SessionID,
		AgentID:       entry.AgentID,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       entry.Content,
		Payload:       entry.Payload,
	}

	return p.client.CreateLog(platformEntry)
}

// CreateLogBatch sends multiple log entries in a single request
func (p *PlatformAPIClient) CreateLogBatch(ctx context.Context, entries []LogEntry) error {
	platformEntries := make([]cli.LogEntry, len(entries))
	for i, entry := range entries {
		platformEntries[i] = cli.LogEntry{
			SessionID:     entry.SessionID,
			AgentID:       entry.AgentID,
			EventType:     entry.EventType,
			EventCategory: entry.EventCategory,
			Content:       entry.Content,
			Payload:       entry.Payload,
		}
	}

	return p.client.CreateLogBatch(platformEntries)
}
