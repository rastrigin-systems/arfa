// Package logparser provides parsers for AI provider API logs
package logparser

import (
	"encoding/json"
	"fmt"

	"github.com/sergeirastrigin/ubik-enterprise/pkg/types"
)

// AnthropicParser parses Anthropic API request/response JSON
type AnthropicParser struct{}

// NewAnthropicParser creates a new Anthropic parser
func NewAnthropicParser() *AnthropicParser {
	return &AnthropicParser{}
}

// Provider returns the provider this parser handles
func (p *AnthropicParser) Provider() types.LogProvider {
	return types.LogProviderAnthropic
}

// anthropicRequest represents the structure of an Anthropic /v1/messages request
type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

// anthropicMessage represents a message in the Anthropic API
type anthropicMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"` // Can be string or []contentBlock
}

// anthropicResponse represents the structure of an Anthropic /v1/messages response
type anthropicResponse struct {
	ID      string                   `json:"id"`
	Type    string                   `json:"type"`
	Role    string                   `json:"role"`
	Model   string                   `json:"model"`
	Content []anthropicResponseBlock `json:"content"`
	Usage   anthropicUsage           `json:"usage"`
	Error   *anthropicError          `json:"error,omitempty"`
}

// anthropicResponseBlock represents a content block in the response
type anthropicResponseBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// anthropicUsage represents token usage
type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// anthropicError represents an API error
type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// contentBlock represents a content block in user messages
type contentBlock struct {
	Type      string `json:"type"`
	Text      string `json:"text,omitempty"`
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"` // For tool_result
}

// ParseRequest parses an Anthropic /v1/messages request body
func (p *AnthropicParser) ParseRequest(body []byte) ([]types.ClassifiedLogEntry, error) {
	var req anthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic request: %w", err)
	}

	var entries []types.ClassifiedLogEntry

	// Find the last user message (the new prompt) and any tool results
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]

		// Parse content which can be string or array
		content, blocks, err := parseMessageContent(msg.Content)
		if err != nil {
			continue
		}

		if msg.Role == "user" {
			// Check if it's a tool result message
			if len(blocks) > 0 {
				for _, block := range blocks {
					if block.Type == "tool_result" {
						entry := types.ClassifiedLogEntry{
							EntryType:  types.LogTypeToolResult,
							Provider:   types.LogProviderAnthropic,
							Model:      req.Model,
							ToolID:     block.ToolUseID,
							ToolOutput: block.Content,
						}
						entries = append(entries, entry)
					} else if block.Type == "text" && block.Text != "" {
						// Only capture the last user prompt
						if len(entries) == 0 || entries[len(entries)-1].EntryType != types.LogTypeUserPrompt {
							entry := types.ClassifiedLogEntry{
								EntryType: types.LogTypeUserPrompt,
								Provider:  types.LogProviderAnthropic,
								Model:     req.Model,
								Content:   block.Text,
							}
							entries = append(entries, entry)
						}
					}
				}
				// If we found tool results, we're done
				if len(entries) > 0 {
					break
				}
			} else if content != "" {
				// Simple string content - this is the user prompt
				entry := types.ClassifiedLogEntry{
					EntryType: types.LogTypeUserPrompt,
					Provider:  types.LogProviderAnthropic,
					Model:     req.Model,
					Content:   content,
				}
				entries = append(entries, entry)
				break // Only capture the last user prompt
			}
		}
	}

	return entries, nil
}

// ParseResponse parses an Anthropic /v1/messages response body
func (p *AnthropicParser) ParseResponse(body []byte) ([]types.ClassifiedLogEntry, error) {
	// First check if it's an error response
	var errorCheck struct {
		Type  string          `json:"type"`
		Error *anthropicError `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &errorCheck); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	if errorCheck.Type == "error" && errorCheck.Error != nil {
		entry := types.ClassifiedLogEntry{
			EntryType:    types.LogTypeError,
			Provider:     types.LogProviderAnthropic,
			ErrorCode:    errorCheck.Error.Type,
			ErrorMessage: errorCheck.Error.Message,
		}
		return []types.ClassifiedLogEntry{entry}, nil
	}

	// Parse as normal response
	var resp anthropicResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Anthropic response: %w", err)
	}

	var entries []types.ClassifiedLogEntry

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			entry := types.ClassifiedLogEntry{
				EntryType:    types.LogTypeAIText,
				Provider:     types.LogProviderAnthropic,
				Model:        resp.Model,
				Content:      block.Text,
				TokensInput:  resp.Usage.InputTokens,
				TokensOutput: resp.Usage.OutputTokens,
			}
			entries = append(entries, entry)

		case "tool_use":
			var input map[string]any
			if len(block.Input) > 0 {
				json.Unmarshal(block.Input, &input)
			}
			entry := types.ClassifiedLogEntry{
				EntryType:    types.LogTypeToolCall,
				Provider:     types.LogProviderAnthropic,
				Model:        resp.Model,
				ToolName:     block.Name,
				ToolID:       block.ID,
				ToolInput:    input,
				TokensInput:  resp.Usage.InputTokens,
				TokensOutput: resp.Usage.OutputTokens,
			}
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

// parseMessageContent handles the fact that content can be a string or array of blocks
func parseMessageContent(raw json.RawMessage) (string, []contentBlock, error) {
	// Try as string first
	var str string
	if err := json.Unmarshal(raw, &str); err == nil {
		return str, nil, nil
	}

	// Try as array of content blocks
	var blocks []contentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		return "", blocks, nil
	}

	return "", nil, fmt.Errorf("could not parse content as string or array")
}
