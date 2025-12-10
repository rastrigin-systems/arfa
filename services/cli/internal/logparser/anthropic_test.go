package logparser

import (
	"encoding/json"
	"testing"

	"github.com/sergeirastrigin/ubik-enterprise/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnthropicParser_Provider(t *testing.T) {
	parser := NewAnthropicParser()
	assert.Equal(t, types.LogProviderAnthropic, parser.Provider())
}

func TestAnthropicParser_ParseRequest_SimpleUserMessage(t *testing.T) {
	parser := NewAnthropicParser()

	requestBody := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 8096,
		"messages": [
			{"role": "user", "content": "Fix the bug in auth.go"}
		]
	}`

	entries, err := parser.ParseRequest([]byte(requestBody))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, types.LogTypeUserPrompt, entry.EntryType)
	assert.Equal(t, "Fix the bug in auth.go", entry.Content)
	assert.Equal(t, "claude-sonnet-4-20250514", entry.Model)
	assert.Equal(t, types.LogProviderAnthropic, entry.Provider)
}

func TestAnthropicParser_ParseRequest_ConversationHistory(t *testing.T) {
	parser := NewAnthropicParser()

	// Request with conversation history - we only want the LAST user message
	requestBody := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 8096,
		"messages": [
			{"role": "user", "content": "Fix the bug in auth.go"},
			{"role": "assistant", "content": "I'll help fix that bug."},
			{"role": "user", "content": "Thanks, now add tests"}
		]
	}`

	entries, err := parser.ParseRequest([]byte(requestBody))
	require.NoError(t, err)
	// Should only return the last user message (the new prompt)
	require.Len(t, entries, 1)

	assert.Equal(t, types.LogTypeUserPrompt, entries[0].EntryType)
	assert.Equal(t, "Thanks, now add tests", entries[0].Content)
}

func TestAnthropicParser_ParseRequest_ContentBlocksArray(t *testing.T) {
	parser := NewAnthropicParser()

	// User content can be an array of content blocks
	requestBody := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 8096,
		"messages": [
			{
				"role": "user",
				"content": [
					{"type": "text", "text": "What is in this image?"},
					{"type": "image", "source": {"type": "base64", "media_type": "image/png", "data": "..."}}
				]
			}
		]
	}`

	entries, err := parser.ParseRequest([]byte(requestBody))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	// Should extract text content
	assert.Equal(t, types.LogTypeUserPrompt, entries[0].EntryType)
	assert.Equal(t, "What is in this image?", entries[0].Content)
}

func TestAnthropicParser_ParseRequest_ToolResultMessage(t *testing.T) {
	parser := NewAnthropicParser()

	// Request containing tool results from previous turn
	requestBody := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 8096,
		"messages": [
			{"role": "user", "content": "Read auth.go"},
			{
				"role": "assistant",
				"content": [
					{"type": "tool_use", "id": "tool_1", "name": "Read", "input": {"file_path": "/app/auth.go"}}
				]
			},
			{
				"role": "user",
				"content": [
					{"type": "tool_result", "tool_use_id": "tool_1", "content": "package auth\n\nfunc Login() {}"}
				]
			}
		]
	}`

	entries, err := parser.ParseRequest([]byte(requestBody))
	require.NoError(t, err)
	// Should capture the tool result
	require.Len(t, entries, 1)

	assert.Equal(t, types.LogTypeToolResult, entries[0].EntryType)
	assert.Equal(t, "tool_1", entries[0].ToolID)
	assert.Contains(t, entries[0].ToolOutput, "package auth")
}

func TestAnthropicParser_ParseResponse_TextOnly(t *testing.T) {
	parser := NewAnthropicParser()

	responseBody := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "I'll help you fix the bug in auth.go."}
		],
		"model": "claude-sonnet-4-20250514",
		"usage": {
			"input_tokens": 150,
			"output_tokens": 25
		}
	}`

	entries, err := parser.ParseResponse([]byte(responseBody))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, types.LogTypeAIText, entry.EntryType)
	assert.Equal(t, "I'll help you fix the bug in auth.go.", entry.Content)
	assert.Equal(t, 150, entry.TokensInput)
	assert.Equal(t, 25, entry.TokensOutput)
	assert.Equal(t, "claude-sonnet-4-20250514", entry.Model)
}

func TestAnthropicParser_ParseResponse_WithToolUse(t *testing.T) {
	parser := NewAnthropicParser()

	responseBody := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "I'll read the file first."},
			{
				"type": "tool_use",
				"id": "tool_abc",
				"name": "Read",
				"input": {"file_path": "/app/auth.go"}
			}
		],
		"model": "claude-sonnet-4-20250514",
		"usage": {
			"input_tokens": 150,
			"output_tokens": 50
		}
	}`

	entries, err := parser.ParseResponse([]byte(responseBody))
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// First entry: AI text
	assert.Equal(t, types.LogTypeAIText, entries[0].EntryType)
	assert.Equal(t, "I'll read the file first.", entries[0].Content)

	// Second entry: Tool call
	assert.Equal(t, types.LogTypeToolCall, entries[1].EntryType)
	assert.Equal(t, "Read", entries[1].ToolName)
	assert.Equal(t, "tool_abc", entries[1].ToolID)
	assert.Equal(t, "/app/auth.go", entries[1].ToolInput["file_path"])
}

func TestAnthropicParser_ParseResponse_MultipleToolCalls(t *testing.T) {
	parser := NewAnthropicParser()

	responseBody := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [
			{"type": "text", "text": "I'll search for the files."},
			{
				"type": "tool_use",
				"id": "tool_1",
				"name": "Glob",
				"input": {"pattern": "**/*.go"}
			},
			{
				"type": "tool_use",
				"id": "tool_2",
				"name": "Grep",
				"input": {"pattern": "func Login", "path": "."}
			}
		],
		"model": "claude-sonnet-4-20250514",
		"usage": {"input_tokens": 100, "output_tokens": 80}
	}`

	entries, err := parser.ParseResponse([]byte(responseBody))
	require.NoError(t, err)
	require.Len(t, entries, 3)

	assert.Equal(t, types.LogTypeAIText, entries[0].EntryType)
	assert.Equal(t, types.LogTypeToolCall, entries[1].EntryType)
	assert.Equal(t, "Glob", entries[1].ToolName)
	assert.Equal(t, types.LogTypeToolCall, entries[2].EntryType)
	assert.Equal(t, "Grep", entries[2].ToolName)
}

func TestAnthropicParser_ParseResponse_Error(t *testing.T) {
	parser := NewAnthropicParser()

	responseBody := `{
		"type": "error",
		"error": {
			"type": "rate_limit_error",
			"message": "You have exceeded your rate limit."
		}
	}`

	entries, err := parser.ParseResponse([]byte(responseBody))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, types.LogTypeError, entry.EntryType)
	assert.Equal(t, "rate_limit_error", entry.ErrorCode)
	assert.Equal(t, "You have exceeded your rate limit.", entry.ErrorMessage)
}

func TestAnthropicParser_ParseRequest_InvalidJSON(t *testing.T) {
	parser := NewAnthropicParser()

	_, err := parser.ParseRequest([]byte("not valid json"))
	assert.Error(t, err)
}

func TestAnthropicParser_ParseResponse_InvalidJSON(t *testing.T) {
	parser := NewAnthropicParser()

	_, err := parser.ParseResponse([]byte("not valid json"))
	assert.Error(t, err)
}

func TestAnthropicParser_ParseRequest_EmptyMessages(t *testing.T) {
	parser := NewAnthropicParser()

	requestBody := `{
		"model": "claude-sonnet-4-20250514",
		"max_tokens": 8096,
		"messages": []
	}`

	entries, err := parser.ParseRequest([]byte(requestBody))
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestAnthropicParser_ParseResponse_EmptyContent(t *testing.T) {
	parser := NewAnthropicParser()

	responseBody := `{
		"id": "msg_123",
		"type": "message",
		"role": "assistant",
		"content": [],
		"model": "claude-sonnet-4-20250514",
		"usage": {"input_tokens": 10, "output_tokens": 0}
	}`

	entries, err := parser.ParseResponse([]byte(responseBody))
	require.NoError(t, err)
	assert.Empty(t, entries)
}
