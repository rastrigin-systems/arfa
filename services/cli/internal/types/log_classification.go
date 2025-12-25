// Package types contains shared types used across the ubik-enterprise platform
package types

import (
	"time"
)

// LogEntryType represents the classification of a log entry
type LogEntryType string

const (
	// Core entry types (Phase 1)
	LogTypeUserPrompt   LogEntryType = "user_prompt"   // User input to AI
	LogTypeAIText       LogEntryType = "ai_text"       // AI text response
	LogTypeToolCall     LogEntryType = "tool_call"     // AI invoking a tool
	LogTypeToolResult   LogEntryType = "tool_result"   // Result from tool execution
	LogTypeError        LogEntryType = "error"         // Error event
	LogTypeSessionStart LogEntryType = "session_start" // Session began
	LogTypeSessionEnd   LogEntryType = "session_end"   // Session ended

	// System entry types
	LogTypeAPIRequest  LogEntryType = "api_request"  // Raw API request (for debugging)
	LogTypeAPIResponse LogEntryType = "api_response" // Raw API response (for debugging)
)

// LogProvider represents the AI provider
type LogProvider string

const (
	LogProviderAnthropic LogProvider = "anthropic"
	LogProviderOpenAI    LogProvider = "openai"
	LogProviderGoogle    LogProvider = "google"
	LogProviderUnknown   LogProvider = "unknown"
)

// ClassifiedLogEntry represents a parsed and classified log entry
type ClassifiedLogEntry struct {
	// Identity
	ID            string    `json:"id"`
	SessionID     string    `json:"session_id"`
	ClientName    string    `json:"client_name,omitempty"`
	ClientVersion string    `json:"client_version,omitempty"`
	Timestamp     time.Time `json:"timestamp"`

	// Classification
	EntryType LogEntryType `json:"entry_type"`
	Provider  LogProvider  `json:"provider,omitempty"`

	// Content (varies by entry type)
	Content      string         `json:"content,omitempty"`       // For text entries (prompts, AI responses)
	ToolName     string         `json:"tool_name,omitempty"`     // For tool_call and tool_result
	ToolID       string         `json:"tool_id,omitempty"`       // Tool use ID for correlation
	ToolInput    map[string]any `json:"tool_input,omitempty"`    // For tool_call
	ToolOutput   string         `json:"tool_output,omitempty"`   // For tool_result
	ErrorMessage string         `json:"error_message,omitempty"` // For errors
	ErrorCode    string         `json:"error_code,omitempty"`    // For errors

	// Metrics
	Model        string `json:"model,omitempty"`
	TokensInput  int    `json:"tokens_input,omitempty"`
	TokensOutput int    `json:"tokens_output,omitempty"`

	// Future extensibility (Phase 2+) - optional fields
	PIIDetected    *bool    `json:"pii_detected,omitempty"`    // Phase 2: PII was found
	PIITypes       []string `json:"pii_types,omitempty"`       // Phase 2: Types of PII found
	RedactionCount *int     `json:"redaction_count,omitempty"` // Phase 2: Number of redactions
	Sensitivity    string   `json:"sensitivity,omitempty"`     // Phase 4: low|medium|high|critical
	Intent         string   `json:"intent,omitempty"`          // Phase 5: coding|research|writing|etc
}

// SessionSummary provides aggregate statistics for a session
type SessionSummary struct {
	SessionID    string        `json:"session_id"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      *time.Time    `json:"end_time,omitempty"`
	Duration     time.Duration `json:"duration,omitempty"`
	Provider     LogProvider   `json:"provider"`
	Model        string        `json:"model,omitempty"`
	TokensInput  int           `json:"tokens_input"`
	TokensOutput int           `json:"tokens_output"`
	ToolCalls    int           `json:"tool_calls"`
	ToolsByName  map[string]int `json:"tools_by_name,omitempty"` // Tool name -> count
	Errors       int           `json:"errors"`
	CostEstimate float64       `json:"cost_estimate,omitempty"` // USD
}

// LogParser defines the interface for parsing provider-specific API logs
type LogParser interface {
	// ParseRequest parses an API request body into classified entries
	ParseRequest(body []byte) ([]ClassifiedLogEntry, error)

	// ParseResponse parses an API response body into classified entries
	ParseResponse(body []byte) ([]ClassifiedLogEntry, error)

	// Provider returns the provider this parser handles
	Provider() LogProvider
}
