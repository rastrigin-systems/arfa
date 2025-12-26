package ui

import "time"

// StreamLogEntry represents a log entry received from the WebSocket
type StreamLogEntry struct {
	AgentID       string                 `json:"agent_id,omitempty"`
	ClientName    string                 `json:"client_name,omitempty"`
	ClientVersion string                 `json:"client_version,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}
