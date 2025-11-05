package logging

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Config holds logging configuration
type Config struct {
	// Enabled controls whether logging is active
	Enabled bool

	// BatchSize is the number of log entries to buffer before sending
	BatchSize int

	// BatchInterval is the maximum time to wait before sending buffered logs
	BatchInterval time.Duration

	// QueueDir is the directory for offline log queue
	QueueDir string

	// MaxRetries is the maximum number of retry attempts for failed sends
	MaxRetries int

	// RetryBackoff is the initial backoff duration for retries
	RetryBackoff time.Duration
}

// LogEntry represents a single log entry to send to the API
type LogEntry struct {
	// SessionID identifies the CLI session
	SessionID string `json:"session_id,omitempty"`

	// AgentID identifies the agent (optional)
	AgentID string `json:"agent_id,omitempty"`

	// EventType specifies the type of event
	EventType string `json:"event_type"`

	// EventCategory categorizes the event
	EventCategory string `json:"event_category"`

	// Content is the log message content
	Content string `json:"content,omitempty"`

	// Payload contains additional structured data
	Payload map[string]interface{} `json:"payload,omitempty"`

	// Timestamp when the event occurred
	Timestamp time.Time `json:"timestamp"`
}

// APIClient defines the interface for sending logs to the platform API
type APIClient interface {
	// CreateLog sends a single log entry to the API
	CreateLog(ctx context.Context, entry LogEntry) error

	// CreateLogBatch sends multiple log entries in a single request
	CreateLogBatch(ctx context.Context, entries []LogEntry) error
}

// Logger manages I/O capture and log transmission
type Logger interface {
	// StartSession begins a new logging session and returns the session ID
	StartSession() uuid.UUID

	// EndSession marks the end of the current session
	EndSession()

	// SetAgentID sets the agent ID for all subsequent log entries
	SetAgentID(agentID string)

	// InterceptStdout wraps stdout to capture output
	InterceptStdout(original io.Writer) io.Writer

	// InterceptStderr wraps stderr to capture errors
	InterceptStderr(original io.Writer) io.Writer

	// InterceptStdin wraps stdin to capture input
	InterceptStdin(original io.Reader) io.Reader

	// LogInput logs user input
	LogInput(content string, metadata map[string]interface{})

	// LogOutput logs agent output
	LogOutput(content string, metadata map[string]interface{})

	// LogError logs error output
	LogError(content string, metadata map[string]interface{})

	// LogEvent logs a custom event
	LogEvent(eventType, category, content string, metadata map[string]interface{})

	// Flush forces immediate sending of buffered logs
	Flush()

	// Close shuts down the logger and flushes remaining logs
	Close() error
}
