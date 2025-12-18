package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"

	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/auth"
)

const (
	// Time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// Send pings to peer with this period (must be less than pongWait)
	pingPeriod = 30 * time.Second

	// Maximum message size allowed from peer
	maxMessageSize = 10 * 1024 // 10KB

	// Maximum messages per second per connection
	maxMessagesPerSecond = 1000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: Add proper origin checking in production
		return true
	},
}

// Handler handles WebSocket connections for log streaming
type Handler struct {
	hub *Hub
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub) *Handler {
	return &Handler{
		hub: hub,
	}
}

// ServeHTTP handles WebSocket upgrade requests
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a WebSocket upgrade request
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Expected WebSocket upgrade request", http.StatusBadRequest)
		return
	}

	// Extract JWT token from Authorization header
	tokenString := extractToken(r)
	if tokenString == "" {
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
		return
	}

	// Validate JWT token
	claims, err := auth.VerifyJWT(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
		return
	}

	// Parse org_id from claims
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid org_id in token", http.StatusUnauthorized)
		return
	}

	// Parse subscription filters from query parameters
	filters, err := parseSubscriptionFilters(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid filters: %v", err), http.StatusBadRequest)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// Create client
	client := &Client{
		orgID:   orgID,
		filters: filters,
		send:    make(chan []byte, 256),
		conn:    conn,
	}

	// Register client with hub
	h.hub.register <- client

	// Start client's read and write pumps
	go client.writePump()
	go client.readPump(h.hub)
}

// extractToken extracts JWT token from Authorization header or query parameter
// Priority: Authorization header > query parameter (for backward compatibility)
func extractToken(r *http.Request) string {
	// Try Authorization header first (standard method)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		// Expected format: "Bearer <token>"
		const bearerPrefix = "Bearer "
		if len(authHeader) >= len(bearerPrefix) && authHeader[:len(bearerPrefix)] == bearerPrefix {
			return authHeader[len(bearerPrefix):]
		}
		// Malformed header - return empty string
		return ""
	}

	// Fallback to query parameter for WebSocket connections from browser
	// (Browser WebSocket API cannot set custom headers in initial handshake)
	token := r.URL.Query().Get("token")
	return token
}

// parseSubscriptionFilters parses subscription filters from query parameters
func parseSubscriptionFilters(r *http.Request) (ClientFilters, error) {
	var filters ClientFilters

	// Parse session_id
	if sessionIDStr := r.URL.Query().Get("session_id"); sessionIDStr != "" {
		sessionID, err := uuid.Parse(sessionIDStr)
		if err != nil {
			return filters, fmt.Errorf("invalid session_id: %w", err)
		}
		filters.SessionID = sessionID
	}

	// Parse employee_id
	if employeeIDStr := r.URL.Query().Get("employee_id"); employeeIDStr != "" {
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			return filters, fmt.Errorf("invalid employee_id: %w", err)
		}
		filters.EmployeeID = employeeID
	}

	// Parse agent_id
	if agentIDStr := r.URL.Query().Get("agent_id"); agentIDStr != "" {
		agentID, err := uuid.Parse(agentIDStr)
		if err != nil {
			return filters, fmt.Errorf("invalid agent_id: %w", err)
		}
		filters.AgentID = agentID
	}

	return filters, nil
}

// readPump handles incoming messages from the WebSocket client
func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		if wsConn, ok := c.conn.(*websocket.Conn); ok {
			wsConn.Close()
		}
	}()

	// Configure connection
	if wsConn, ok := c.conn.(*websocket.Conn); ok {
		wsConn.SetReadLimit(maxMessageSize)
		wsConn.SetReadDeadline(time.Now().Add(pongWait))
		wsConn.SetPongHandler(func(string) error {
			wsConn.SetReadDeadline(time.Now().Add(pongWait))
			return nil
		})
	}

	// Rate limiter: 1000 messages per second
	limiter := rate.NewLimiter(maxMessagesPerSecond, maxMessagesPerSecond)

	// Read messages from client
	for {
		if wsConn, ok := c.conn.(*websocket.Conn); ok {
			_, message, err := wsConn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				break
			}

			// Check rate limit
			if !limiter.Allow() {
				log.Printf("Rate limit exceeded for client %v", c.orgID)
				continue
			}

			// Handle incoming messages (for future features like filter updates)
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Invalid message format: %v", err)
				continue
			}

			// Currently, we don't process client messages
			// In the future, we could support:
			// - Filter updates
			// - Subscription changes
			// - Heartbeat acknowledgments
			_ = msg // Use the variable to avoid compiler error
		}
	}
}

// writePump handles outgoing messages to the WebSocket client
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if wsConn, ok := c.conn.(*websocket.Conn); ok {
			wsConn.Close()
		}
	}()

	for {
		select {
		case message, channelOpen := <-c.send:
			wsConn, isWS := c.conn.(*websocket.Conn)
			if !isWS {
				return
			}

			wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !channelOpen {
				// Hub closed the channel
				wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Write message
			if err := wsConn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			wsConn, isWS := c.conn.(*websocket.Conn)
			if !isWS {
				return
			}

			wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := wsConn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
