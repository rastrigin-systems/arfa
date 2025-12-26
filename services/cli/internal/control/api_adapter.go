package control

import (
	"context"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
)

// CLIAPIClient wraps the CLI api.Client to implement the control.APIClient interface.
type CLIAPIClient struct {
	client *api.Client
}

// NewCLIAPIClient creates a new adapter for the CLI API client.
func NewCLIAPIClient(client *api.Client) *CLIAPIClient {
	return &CLIAPIClient{client: client}
}

// CreateLog sends a single log entry to the API.
func (c *CLIAPIClient) CreateLog(ctx context.Context, entry APILogEntry) error {
	apiEntry := api.LogEntry{
		ClientName:    entry.ClientName,
		ClientVersion: entry.ClientVersion,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       entry.Content,
		Payload:       entry.Payload,
	}

	return c.client.CreateLog(ctx, apiEntry)
}
