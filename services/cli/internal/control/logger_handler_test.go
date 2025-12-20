package control

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLoggerQueue captures enqueued entries for testing
type mockLoggerQueue struct {
	mu      sync.Mutex
	entries []LogEntry
}

func (m *mockLoggerQueue) Enqueue(entry LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries = append(m.entries, entry)
	return nil
}

func (m *mockLoggerQueue) Entries() []LogEntry {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]LogEntry, len(m.entries))
	copy(result, m.entries)
	return result
}

func TestNewLoggerHandler(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	require.NotNil(t, handler)
	assert.Equal(t, "logger", handler.Name())
	assert.Equal(t, 50, handler.Priority()) // Mid-priority
}

func TestLoggerHandler_ImplementsInterface(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	// Verify interface compliance
	var _ Handler = handler
}

func TestLoggerHandler_HandleRequest_LogsEntry(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := handler.HandleRequest(ctx, req)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "emp-123", entry.EmployeeID)
	assert.Equal(t, "org-456", entry.OrgID)
	assert.Equal(t, "sess-789", entry.SessionID)
	assert.Equal(t, "agent-abc", entry.AgentID)
	assert.Equal(t, "api_request", entry.EventType)
	assert.Equal(t, "proxy", entry.EventCategory)
}

func TestLoggerHandler_HandleRequest_CapturesMethod(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, "POST", entries[0].Payload["method"])
}

func TestLoggerHandler_HandleRequest_CapturesURL(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, "https://api.anthropic.com/v1/messages", entries[0].Payload["url"])
}

func TestLoggerHandler_HandleRequest_CapturesHost(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, "api.anthropic.com", entries[0].Payload["host"])
}

func TestLoggerHandler_HandleRequest_CapturesBody(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	body := `{"model": "claude-3", "messages": [{"role": "user", "content": "Hello"}]}`
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBufferString(body))

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, body, entries[0].Payload["body"])
}

func TestLoggerHandler_HandleRequest_PreservesRequestBody(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	body := `{"model": "claude-3"}`
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBufferString(body))

	handler.HandleRequest(ctx, req)

	// Original request body should still be readable
	bodyBytes, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(bodyBytes))
}

func TestLoggerHandler_HandleRequest_CapturesHeaders(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Custom", "value")

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	headers := entries[0].Payload["headers"].(map[string]string)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "value", headers["X-Custom"])
}

func TestLoggerHandler_HandleRequest_RedactsSensitiveHeaders(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	req.Header.Set("Authorization", "Bearer sk-secret-key")
	req.Header.Set("X-Api-Key", "secret-api-key")
	req.Header.Set("Cookie", "session=abc123")

	handler.HandleRequest(ctx, req)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	headers := entries[0].Payload["headers"].(map[string]string)
	assert.Equal(t, "[REDACTED]", headers["Authorization"])
	assert.Equal(t, "[REDACTED]", headers["X-Api-Key"])
	assert.Equal(t, "[REDACTED]", headers["Cookie"])
}

func TestLoggerHandler_HandleRequest_SetsTimestamp(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	before := time.Now()
	handler.HandleRequest(ctx, req)
	after := time.Now()

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.True(t, entries[0].Timestamp.After(before) || entries[0].Timestamp.Equal(before))
	assert.True(t, entries[0].Timestamp.Before(after) || entries[0].Timestamp.Equal(after))
}

func TestLoggerHandler_HandleResponse_LogsEntry(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(`{"content": "Hello!"}`)),
	}

	result := handler.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "emp-123", entry.EmployeeID)
	assert.Equal(t, "org-456", entry.OrgID)
	assert.Equal(t, "sess-789", entry.SessionID)
	assert.Equal(t, "agent-abc", entry.AgentID)
	assert.Equal(t, "api_response", entry.EventType)
	assert.Equal(t, "proxy", entry.EventCategory)
}

func TestLoggerHandler_HandleResponse_CapturesStatusCode(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	handler.HandleResponse(ctx, res)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, 200, entries[0].Payload["status_code"])
}

func TestLoggerHandler_HandleResponse_CapturesBody(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	body := `{"content": "Hello from Claude!"}`
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	handler.HandleResponse(ctx, res)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, body, entries[0].Payload["body"])
}

func TestLoggerHandler_HandleResponse_PreservesResponseBody(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	body := `{"content": "Hello!"}`
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	handler.HandleResponse(ctx, res)

	// Response body should still be readable
	bodyBytes, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(bodyBytes))
}

func TestLoggerHandler_HandleResponse_CapturesHeaders(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}
	res.Header.Set("Content-Type", "application/json")
	res.Header.Set("X-Request-Id", "req-123")

	handler.HandleResponse(ctx, res)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	headers := entries[0].Payload["headers"].(map[string]string)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "req-123", headers["X-Request-Id"])
}

func TestLoggerHandler_AlwaysContinues(t *testing.T) {
	queue := &mockLoggerQueue{}
	handler := NewLoggerHandler(queue)

	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	res := &http.Response{
		StatusCode: 500, // Even on error response
		Request:    req,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}

	reqResult := handler.HandleRequest(ctx, req)
	resResult := handler.HandleResponse(ctx, res)

	// Logger should never block
	assert.Equal(t, ActionContinue, reqResult.Action)
	assert.Equal(t, ActionContinue, resResult.Action)
}
