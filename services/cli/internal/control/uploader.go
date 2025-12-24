package control

import (
	"context"
	"fmt"
)

// APIClient defines the interface for sending logs to the platform API.
type APIClient interface {
	// CreateLog sends a single log entry to the API.
	CreateLog(ctx context.Context, entry APILogEntry) error
}

// APILogEntry represents a log entry for the API.
// This matches the format expected by the platform API.
type APILogEntry struct {
	SessionID     string                 `json:"session_id,omitempty"`
	ClientName    string                 `json:"client_name,omitempty"`
	ClientVersion string                 `json:"client_version,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category,omitempty"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
}

// APIUploader implements the Uploader interface by sending logs to the platform API.
type APIUploader struct {
	client     APIClient
	employeeID string
	orgID      string
}

// NewAPIUploader creates a new API uploader.
func NewAPIUploader(client APIClient, employeeID, orgID string) *APIUploader {
	return &APIUploader{
		client:     client,
		employeeID: employeeID,
		orgID:      orgID,
	}
}

// Upload sends a batch of log entries to the API.
func (u *APIUploader) Upload(entries []LogEntry) error {
	if u.client == nil {
		return nil // Silently skip if no client
	}

	ctx := context.Background()

	for _, entry := range entries {
		// Convert control.LogEntry to APILogEntry
		apiEntry := APILogEntry{
			SessionID:     entry.SessionID,
			ClientName:    entry.ClientName,
			ClientVersion: entry.ClientVersion,
			EventType:     entry.EventType,
			EventCategory: entry.EventCategory,
			Payload:       entry.Payload,
		}

		// Include employee_id and org_id in payload
		if apiEntry.Payload == nil {
			apiEntry.Payload = make(map[string]interface{})
		}
		if entry.EmployeeID != "" {
			apiEntry.Payload["employee_id"] = entry.EmployeeID
		}
		if entry.OrgID != "" {
			apiEntry.Payload["org_id"] = entry.OrgID
		}

		if err := u.client.CreateLog(ctx, apiEntry); err != nil {
			return fmt.Errorf("failed to upload log: %w", err)
		}
	}

	return nil
}
