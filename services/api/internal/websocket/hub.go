package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogMessage represents a log entry to be broadcast to WebSocket clients
type LogMessage struct {
	ID            uuid.UUID              `json:"id"`
	OrgID         uuid.UUID              `json:"org_id"`
	EmployeeID    uuid.UUID              `json:"employee_id,omitempty"`
	SessionID     uuid.UUID              `json:"session_id,omitempty"`
	ClientName    string                 `json:"client_name,omitempty"`
	ClientVersion string                 `json:"client_version,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// ClientFilters defines subscription filters for a WebSocket client
type ClientFilters struct {
	SessionID  uuid.UUID `json:"session_id,omitempty"`
	EmployeeID uuid.UUID `json:"employee_id,omitempty"`
	ClientName string    `json:"client_name,omitempty"`
}

// Client represents a WebSocket client connection
type Client struct {
	orgID   uuid.UUID
	filters ClientFilters
	send    chan []byte
	conn    interface{} // Will be *websocket.Conn in real implementation
}

// Hub manages WebSocket client connections and broadcasts
type Hub struct {
	// Registered clients by organization
	clients map[*Client]bool

	// Inbound log messages from the application
	broadcast chan LogMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Stop signal
	stop chan struct{}

	// Mutex for thread-safe access to clients map
	mu sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan LogMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Run starts the hub's main event loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)

		case <-h.stop:
			return
		}
	}
}

// Stop signals the hub to stop running
func (h *Hub) Stop() {
	close(h.stop)
}

// registerClient adds a client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client] = true
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

// broadcastMessage sends a log message to all matching clients
func (h *Hub) broadcastMessage(message LogMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Serialize message once
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return
	}

	// Send to all clients that match filters
	for client := range h.clients {
		if h.shouldSendToClient(client, message) {
			select {
			case client.send <- messageBytes:
				// Message sent successfully
			default:
				// Client's send channel is full, skip
			}
		}
	}
}

// shouldSendToClient checks if a message should be sent to a specific client
func (h *Hub) shouldSendToClient(client *Client, message LogMessage) bool {
	// Multi-tenancy check - must be same organization
	if client.orgID != message.OrgID {
		return false
	}

	// Apply subscription filters
	filters := client.filters

	// If session filter is set, check if it matches
	if filters.SessionID != uuid.Nil && filters.SessionID != message.SessionID {
		return false
	}

	// If employee filter is set, check if it matches
	if filters.EmployeeID != uuid.Nil && filters.EmployeeID != message.EmployeeID {
		return false
	}

	// If client name filter is set, check if it matches
	if filters.ClientName != "" && filters.ClientName != message.ClientName {
		return false
	}

	return true
}

// IsClientRegistered checks if a client is currently registered (for testing)
func (h *Hub) IsClientRegistered(client *Client) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[client]
}

// Broadcast sends a log message to the hub for broadcasting
func (h *Hub) Broadcast(message LogMessage) {
	select {
	case h.broadcast <- message:
		// Message queued for broadcast
	default:
		// Broadcast channel is full, drop message
	}
}
