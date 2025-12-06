package logging

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	cli "github.com/sergeirastrigin/ubik-enterprise/services/cli/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlatformAPIClient(t *testing.T) {
	// Create a mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/logs", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"123"}`))
	}))
	defer server.Close()

	// Create platform client pointing to mock server
	platformClient := cli.NewPlatformClient(server.URL)

	// Create adapter
	adapter := NewPlatformAPIClient(platformClient)
	require.NotNil(t, adapter)

	// Test CreateLog
	entry := LogEntry{
		SessionID:     "session-123",
		AgentID:       "agent-456",
		EventType:     "input",
		EventCategory: "cli",
		Content:       "test content",
		Payload:       map[string]interface{}{"key": "value"},
	}

	err := adapter.CreateLog(context.Background(), entry)
	assert.NoError(t, err)

	// Test CreateLogBatch
	entries := []LogEntry{entry, entry}
	err = adapter.CreateLogBatch(context.Background(), entries)
	assert.NoError(t, err)
}

func TestPlatformAPIClientError(t *testing.T) {
	// Create a mock HTTP server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal error"}`))
	}))
	defer server.Close()

	// Create platform client pointing to mock server
	platformClient := cli.NewPlatformClient(server.URL)

	// Create adapter
	adapter := NewPlatformAPIClient(platformClient)

	// Test CreateLog with error
	entry := LogEntry{
		EventType:     "input",
		EventCategory: "cli",
	}

	err := adapter.CreateLog(context.Background(), entry)
	assert.Error(t, err)

	// Test CreateLogBatch with error
	entries := []LogEntry{entry}
	err = adapter.CreateLogBatch(context.Background(), entries)
	assert.Error(t, err)
}
