package control

import (
	"bytes"
	"io"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockToolLoggerQueue captures log entries for testing
type mockToolLoggerQueue struct {
	mu      sync.Mutex
	entries []LogEntry
}

func (q *mockToolLoggerQueue) Enqueue(entry LogEntry) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = append(q.entries, entry)
	return nil
}

func (q *mockToolLoggerQueue) Entries() []LogEntry {
	q.mu.Lock()
	defer q.mu.Unlock()
	return append([]LogEntry{}, q.entries...)
}

func TestNewToolCallLoggerHandler(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)

	assert.NotNil(t, h)
	assert.Equal(t, "ToolCallLogger", h.Name())
	assert.Equal(t, 40, h.Priority())
}

func TestToolCallLoggerHandler_HandleRequest_Noop(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	result := h.HandleRequest(ctx, req)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Empty(t, queue.Entries())
}

func TestToolCallLoggerHandler_HandleResponse_NonSSE(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"content": "hello"}`))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Empty(t, queue.Entries())
}

func TestToolCallLoggerHandler_HandleResponse_SingleToolCall(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_123","name":"Read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"file_path\":"}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"/test.txt\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "tool_call", entry.EventType)
	assert.Equal(t, "classified", entry.EventCategory)
	assert.Equal(t, "emp-1", entry.EmployeeID)
	assert.Equal(t, "org-1", entry.OrgID)
	assert.Equal(t, "claude-code", entry.ClientName)
	assert.Equal(t, "1.0.25", entry.ClientVersion)

	assert.Equal(t, "Read", entry.Payload["tool_name"])
	assert.Equal(t, "toolu_123", entry.Payload["tool_id"])
	assert.Equal(t, false, entry.Payload["blocked"])

	toolInput := entry.Payload["tool_input"].(map[string]interface{})
	assert.Equal(t, "/test.txt", toolInput["file_path"])
}

func TestToolCallLoggerHandler_HandleResponse_MultipleToolCalls(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"path\":\"/a.txt\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_2","name":"Bash","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"command\":\"ls -la\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 2)

	// First tool call
	assert.Equal(t, "Read", entries[0].Payload["tool_name"])
	assert.Equal(t, "toolu_1", entries[0].Payload["tool_id"])

	// Second tool call
	assert.Equal(t, "Bash", entries[1].Payload["tool_name"])
	assert.Equal(t, "toolu_2", entries[1].Payload["tool_id"])
}

func TestToolCallLoggerHandler_HandleResponse_TextBlockIgnored(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello world"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Empty(t, queue.Entries()) // No tool calls logged
}

func TestToolCallLoggerHandler_HandleResponse_EmptyInput(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"GetTime","input":{}}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	assert.Equal(t, "GetTime", entries[0].Payload["tool_name"])
	// Empty input should result in nil tool_input
	assert.Nil(t, entries[0].Payload["tool_input"])
}

func TestToolCallLoggerHandler_HandleResponse_InvalidJSON(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"not valid json"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1)

	// Invalid JSON should be stored as raw string
	toolInput := entries[0].Payload["tool_input"].(map[string]interface{})
	assert.Equal(t, "not valid json", toolInput["_raw"])
}

func TestToolCallLoggerHandler_HandleResponse_BodyRestored(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	originalBody := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Read","input":{}}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(originalBody))),
	}

	h.HandleResponse(ctx, res)

	// Body should be restored for downstream handlers
	restoredBody, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	assert.Equal(t, originalBody, string(restoredBody))
}

func TestToolCallLoggerHandler_NilQueue(t *testing.T) {
	h := NewToolCallLoggerHandler(nil)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Read","input":{}}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	// Should not panic with nil queue
	result := h.HandleResponse(ctx, res)
	assert.Equal(t, ActionContinue, result.Action)
}

func TestCleanSSEData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean JSON",
			input:    `{"type":"test"}`,
			expected: `{"type":"test"}`,
		},
		{
			name:     "trailing whitespace and brace",
			input:    `{"type":"test"}       }`,
			expected: `{"type":"test"}`,
		},
		{
			name:     "multiple trailing braces",
			input:    `{"nested":{"value":1}}            }`,
			expected: `{"nested":{"value":1}}`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "no trailing garbage",
			input:    `{"content_block":{"type":"tool_use","id":"123"}}`,
			expected: `{"content_block":{"type":"tool_use","id":"123"}}`,
		},
		{
			name:     "real anthropic format",
			input:    `{"type":"content_block_start","index":4,"content_block":{"type":"tool_use","id":"toolu_01","name":"Glob","input":{},"caller":{"type":"direct"}}}       }`,
			expected: `{"type":"content_block_start","index":4,"content_block":{"type":"tool_use","id":"toolu_01","name":"Glob","input":{},"caller":{"type":"direct"}}}`,
		},
		{
			name:     "braces in string values",
			input:    `{"partial_json":"{\"key\":\"value\"}"}       }`,
			expected: `{"partial_json":"{\"key\":\"value\"}"}`,
		},
		{
			name:     "escaped quotes in string",
			input:    `{"text":"He said \"hello\""}   }`,
			expected: `{"text":"He said \"hello\""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSSEData(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToolCallLoggerHandler_HandleResponse_TrailingGarbage(t *testing.T) {
	queue := &mockToolLoggerQueue{}
	h := NewToolCallLoggerHandler(queue)
	ctx := &HandlerContext{EmployeeID: "emp-1", OrgID: "org-1", SessionID: "sess-1", ClientName: "claude-code", ClientVersion: "1.0.25"}

	// SSE stream with trailing garbage (real Anthropic format)
	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_123","name":"Glob","input":{}}}       }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"pattern\":"}}            }

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"\"**/go.mod\"}"}}       }

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(bytes.NewReader([]byte(sseStream))),
	}

	result := h.HandleResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)

	entries := queue.Entries()
	require.Len(t, entries, 1, "Should parse tool call despite trailing garbage")

	entry := entries[0]
	assert.Equal(t, "tool_call", entry.EventType)
	assert.Equal(t, "Glob", entry.Payload["tool_name"])
	assert.Equal(t, "toolu_123", entry.Payload["tool_id"])

	toolInput := entry.Payload["tool_input"].(map[string]interface{})
	assert.Equal(t, "**/go.mod", toolInput["pattern"])
}
