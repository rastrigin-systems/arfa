package control

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicyHandler_Name(t *testing.T) {
	h := NewPolicyHandler()
	assert.Equal(t, "PolicyHandler", h.Name())
}

func TestPolicyHandler_Priority(t *testing.T) {
	h := NewPolicyHandler()
	assert.Equal(t, 110, h.Priority())
}

func TestPolicyHandler_HandleRequest(t *testing.T) {
	h := NewPolicyHandler()
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := h.HandleRequest(ctx, req)

	assert.True(t, result.ShouldContinue())
	assert.Nil(t, result.ModifiedRequest)
}

func TestPolicyHandler_HandleResponse_NonSSE(t *testing.T) {
	h := NewPolicyHandlerWithDenyList(map[string]string{"Bash": "blocked"})
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")

	body := `{"type":"message","content":[{"type":"text","text":"Hello"}]}`
	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	result := h.HandleResponse(ctx, res)

	assert.True(t, result.ShouldContinue())
	// Non-SSE responses are not modified
}

func TestPolicyHandler_HandleResponse_AllowedTool(t *testing.T) {
	// Block Bash, but allow Read
	h := NewPolicyHandlerWithDenyList(map[string]string{"Bash": "blocked"})
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")

	// SSE stream with Read tool (not blocked)
	sseStream := `event: message_start
data: {"type":"message_start","message":{"id":"msg_1"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"file_path\":\"/test.txt\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(sseStream)),
	}

	result := h.HandleResponse(ctx, res)

	assert.True(t, result.ShouldContinue())
	require.NotNil(t, result.ModifiedResponse)

	// Body should contain the tool_use (not blocked)
	modifiedBody, _ := io.ReadAll(result.ModifiedResponse.Body)
	assert.Contains(t, string(modifiedBody), `"name":"Read"`)
	assert.Contains(t, string(modifiedBody), `"type":"tool_use"`)
}

func TestPolicyHandler_HandleResponse_BlockedTool(t *testing.T) {
	h := NewPolicyHandlerWithDenyList(map[string]string{
		"Bash": "Shell commands are blocked by organization policy",
	})
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")

	// SSE stream with Bash tool (blocked)
	sseStream := `event: message_start
data: {"type":"message_start","message":{"id":"msg_1"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{\"command\":\"ls -la\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_stop
data: {"type":"message_stop"}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(sseStream)),
	}

	result := h.HandleResponse(ctx, res)

	assert.True(t, result.ShouldContinue())
	require.NotNil(t, result.ModifiedResponse)

	// Body should NOT contain tool_use - should be replaced with text block
	modifiedBody, _ := io.ReadAll(result.ModifiedResponse.Body)
	bodyStr := string(modifiedBody)

	// Should not have tool_use
	assert.NotContains(t, bodyStr, `"type":"tool_use"`)

	// Should have replacement text block with error message
	assert.Contains(t, bodyStr, `"type":"text"`)
	assert.Contains(t, bodyStr, `"type":"text_delta"`)
	assert.Contains(t, bodyStr, "TOOL BLOCKED BY ORGANIZATION POLICY")
	assert.Contains(t, bodyStr, "Bash")
	assert.Contains(t, bodyStr, "Shell commands are blocked by organization policy")
}

func TestPolicyHandler_HandleResponse_MixedBlocks(t *testing.T) {
	h := NewPolicyHandlerWithDenyList(map[string]string{
		"Bash": "blocked",
	})
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")

	// SSE stream with text block, then Bash (blocked), then Read (allowed)
	sseStream := `event: message_start
data: {"type":"message_start","message":{"id":"msg_1"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Let me help"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: content_block_start
data: {"type":"content_block_start","index":1,"content_block":{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":1,"delta":{"type":"input_json_delta","partial_json":"{\"command\":\"rm -rf /\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":1}

event: content_block_start
data: {"type":"content_block_start","index":2,"content_block":{"type":"tool_use","id":"toolu_2","name":"Read","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":2,"delta":{"type":"input_json_delta","partial_json":"{\"file_path\":\"/etc/passwd\"}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":2}

event: message_stop
data: {"type":"message_stop"}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(sseStream)),
	}

	result := h.HandleResponse(ctx, res)

	require.NotNil(t, result.ModifiedResponse)
	modifiedBody, _ := io.ReadAll(result.ModifiedResponse.Body)
	bodyStr := string(modifiedBody)

	// Text block should remain
	assert.Contains(t, bodyStr, "Let me help")

	// Bash should be blocked
	assert.Contains(t, bodyStr, "TOOL BLOCKED")

	// Read should remain (allowed)
	assert.Contains(t, bodyStr, `"name":"Read"`)
}

func TestPolicyHandler_EmptyDenyList(t *testing.T) {
	// Override HOME to temp dir to ensure no policies are loaded from cache
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	h := NewPolicyHandler() // Should have empty deny list (no cache file)
	ctx := NewHandlerContext("emp-1", "org-1", "sess-1", "agent-1")

	sseStream := `event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"toolu_1","name":"Bash","input":{}}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`

	res := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(sseStream)),
	}

	result := h.HandleResponse(ctx, res)

	// With empty deny list, nothing should be modified
	require.NotNil(t, result.ModifiedResponse)
	modifiedBody, _ := io.ReadAll(result.ModifiedResponse.Body)

	// Original content preserved
	assert.Contains(t, string(modifiedBody), `"name":"Bash"`)
	assert.Contains(t, string(modifiedBody), `"type":"tool_use"`)
}

func TestProcessSSEStream_NoBlocking(t *testing.T) {
	h := &PolicyHandler{denyList: map[string]string{}, globPatterns: map[string]string{}}

	input := []byte(`event: test
data: {"foo":"bar"}

`)

	output, modified := h.processSSEStream(input)

	assert.False(t, modified)
	assert.Equal(t, input, output)
}

func TestProcessSSEStream_BlockBash(t *testing.T) {
	h := &PolicyHandler{denyList: map[string]string{"Bash": "no shell"}, globPatterns: map[string]string{}}

	input := []byte(`event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"tool_use","id":"t1","name":"Bash","input":{}}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{}"}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

`)

	output, modified := h.processSSEStream(input)

	assert.True(t, modified)

	// Should have text block instead of tool_use
	outputStr := string(output)
	assert.Contains(t, outputStr, `"type":"text"`)
	assert.Contains(t, outputStr, "TOOL BLOCKED")
	assert.NotContains(t, outputStr, `"type":"tool_use"`)
}

func TestFormatBlockError(t *testing.T) {
	h := &PolicyHandler{}

	msg := h.formatBlockError("Bash", "Shell access denied")

	assert.Contains(t, msg, "TOOL BLOCKED BY ORGANIZATION POLICY")
	assert.Contains(t, msg, "Tool: Bash")
	assert.Contains(t, msg, "Reason: Shell access denied")
	assert.Contains(t, msg, "ubik policies list")
}

func TestWriteBlockedEvent(t *testing.T) {
	h := &PolicyHandler{}
	var buf bytes.Buffer

	h.writeBlockedEvent(&buf, 5, "Bash", "blocked")

	output := buf.String()

	// Should have 3 events: content_block_start, content_block_delta, content_block_stop
	assert.Equal(t, 3, strings.Count(output, "event: content_block"))
	assert.Contains(t, output, `"index":5`)
	assert.Contains(t, output, `"type":"text"`)
	assert.Contains(t, output, `"type":"text_delta"`)
	assert.Contains(t, output, "TOOL BLOCKED")
}

func TestPolicyHandler_LoadFromCache(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a policies.json file
	ubikDir := tempDir + "/.ubik"
	os.MkdirAll(ubikDir, 0700)

	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked"},
			{"tool_name": "Write", "action": "deny", "reason": "Writes blocked"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(ubikDir+"/policies.json", []byte(cacheContent), 0600)

	// Create handler - should load from cache
	h := NewPolicyHandler()

	// Test that Bash is blocked
	reason, blocked := h.isBlocked("Bash")
	assert.True(t, blocked)
	assert.Equal(t, "Shell blocked", reason)

	// Test that Write is blocked
	reason, blocked = h.isBlocked("Write")
	assert.True(t, blocked)
	assert.Equal(t, "Writes blocked", reason)

	// Test that other tools are not blocked
	_, blocked = h.isBlocked("Read")
	assert.False(t, blocked)
}

func TestPolicyHandler_GlobPattern(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a policies.json file with a glob pattern
	ubikDir := tempDir + "/.ubik"
	os.MkdirAll(ubikDir, 0700)

	cacheContent := `{
		"policies": [
			{"tool_name": "mcp__gcloud__%", "action": "deny", "reason": "GCloud MCP blocked"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(ubikDir+"/policies.json", []byte(cacheContent), 0600)

	// Create handler - should load from cache
	h := NewPolicyHandler()

	// Test that mcp__gcloud__run_gcloud_command is blocked (matches pattern)
	reason, blocked := h.isBlocked("mcp__gcloud__run_gcloud_command")
	assert.True(t, blocked)
	assert.Equal(t, "GCloud MCP blocked", reason)

	// Test that mcp__gcloud__list_instances is also blocked
	reason, blocked = h.isBlocked("mcp__gcloud__list_instances")
	assert.True(t, blocked)

	// Test that other MCP tools are not blocked
	_, blocked = h.isBlocked("mcp__filesystem__read_file")
	assert.False(t, blocked)

	// Test that plain Bash is not blocked
	_, blocked = h.isBlocked("Bash")
	assert.False(t, blocked)
}

func TestPolicyHandler_SkipsAuditPolicies(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a policies.json file with both deny and audit policies
	ubikDir := tempDir + "/.ubik"
	os.MkdirAll(ubikDir, 0700)

	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked"},
			{"tool_name": "Write", "action": "audit", "reason": "Writes audited"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(ubikDir+"/policies.json", []byte(cacheContent), 0600)

	// Create handler - should only load deny policies
	h := NewPolicyHandler()

	// Bash should be blocked (deny)
	_, blocked := h.isBlocked("Bash")
	assert.True(t, blocked)

	// Write should NOT be blocked (audit only)
	_, blocked = h.isBlocked("Write")
	assert.False(t, blocked)
}

func TestPolicyHandler_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create a policies.json file
	ubikDir := tempDir + "/.ubik"
	os.MkdirAll(ubikDir, 0700)

	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(ubikDir+"/policies.json", []byte(cacheContent), 0600)

	h := NewPolicyHandler()

	// Test case variations
	_, blocked := h.isBlocked("Bash")
	assert.True(t, blocked, "exact case should be blocked")

	_, blocked = h.isBlocked("bash")
	assert.True(t, blocked, "lowercase should be blocked")

	_, blocked = h.isBlocked("BASH")
	assert.True(t, blocked, "uppercase should be blocked")
}
