package ui

import "time"

// AgentPickerItem represents an agent in the picker list
type AgentPickerItem struct {
	Name        string
	Type        string
	Provider    string
	DockerImage string
	ID          string
	IsDefault   bool
}

// StreamLogEntry represents a log entry received from the WebSocket
type StreamLogEntry struct {
	SessionID     string                 `json:"session_id,omitempty"`
	AgentID       string                 `json:"agent_id,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}
