package control

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAPIClient captures API calls for testing
type mockAPIClient struct {
	mu       sync.Mutex
	entries  []APILogEntry
	failWith error
}

func (m *mockAPIClient) CreateLog(ctx context.Context, entry APILogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.failWith != nil {
		return m.failWith
	}

	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockAPIClient) Entries() []APILogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]APILogEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

func TestNewAPIUploader(t *testing.T) {
	client := &mockAPIClient{}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	require.NotNil(t, uploader)
	assert.Equal(t, client, uploader.client)
	assert.Equal(t, "emp-123", uploader.employeeID)
	assert.Equal(t, "org-456", uploader.orgID)
}

func TestAPIUploader_Upload_SingleEntry(t *testing.T) {
	client := &mockAPIClient{}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	entries := []LogEntry{
		{
			EmployeeID:    "emp-123",
			OrgID:         "org-456",
			ClientName:    "claude-code",
			ClientVersion: "1.0.25",
			EventType:     "api_request",
			EventCategory: "proxy",
			Payload:       map[string]interface{}{"method": "POST"},
		},
	}

	err := uploader.Upload(entries)

	require.NoError(t, err)
	assert.Len(t, client.Entries(), 1)

	uploaded := client.Entries()[0]
	assert.Equal(t, "claude-code", uploaded.ClientName)
	assert.Equal(t, "1.0.25", uploaded.ClientVersion)
	assert.Equal(t, "api_request", uploaded.EventType)
	assert.Equal(t, "proxy", uploaded.EventCategory)
	assert.Equal(t, "POST", uploaded.Payload["method"])
	assert.Equal(t, "emp-123", uploaded.Payload["employee_id"])
	assert.Equal(t, "org-456", uploaded.Payload["org_id"])
}

func TestAPIUploader_Upload_MultipleEntries(t *testing.T) {
	client := &mockAPIClient{}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	entries := []LogEntry{
		{EventType: "api_request"},
		{EventType: "api_response"},
		{EventType: "api_request"},
	}

	err := uploader.Upload(entries)

	require.NoError(t, err)
	assert.Len(t, client.Entries(), 3)
}

func TestAPIUploader_Upload_Error(t *testing.T) {
	client := &mockAPIClient{failWith: errors.New("API error")}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	entries := []LogEntry{
		{EventType: "api_request"},
	}

	err := uploader.Upload(entries)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "API error")
}

func TestAPIUploader_Upload_NilClient(t *testing.T) {
	uploader := NewAPIUploader(nil, "emp-123", "org-456")

	entries := []LogEntry{
		{EventType: "api_request"},
	}

	// Should not panic with nil client
	err := uploader.Upload(entries)
	require.NoError(t, err)
}

func TestAPIUploader_Upload_EmptyEntries(t *testing.T) {
	client := &mockAPIClient{}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	err := uploader.Upload([]LogEntry{})

	require.NoError(t, err)
	assert.Len(t, client.Entries(), 0)
}

func TestAPIUploader_Upload_PreservesPayload(t *testing.T) {
	client := &mockAPIClient{}
	uploader := NewAPIUploader(client, "emp-123", "org-456")

	entries := []LogEntry{
		{
			EventType: "api_request",
			Payload: map[string]interface{}{
				"method":  "POST",
				"url":     "https://api.anthropic.com/v1/messages",
				"headers": map[string]string{"Content-Type": "application/json"},
			},
		},
	}

	err := uploader.Upload(entries)

	require.NoError(t, err)
	uploaded := client.Entries()[0]
	assert.Equal(t, "POST", uploaded.Payload["method"])
	assert.Equal(t, "https://api.anthropic.com/v1/messages", uploaded.Payload["url"])
	assert.NotNil(t, uploaded.Payload["headers"])
}
