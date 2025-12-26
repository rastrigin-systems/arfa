package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
)

// PolicyHandler handles WebSocket connections for policy streaming to proxies
type PolicyHandler struct {
	hub     *PolicyHub
	queries db.Querier
}

// NewPolicyHandler creates a new policy WebSocket handler
func NewPolicyHandler(hub *PolicyHub, queries db.Querier) *PolicyHandler {
	return &PolicyHandler{
		hub:     hub,
		queries: queries,
	}
}

// ServeHTTP handles WebSocket upgrade requests for policy streaming
func (h *PolicyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check if this is a WebSocket upgrade request
	if r.Header.Get("Upgrade") != "websocket" {
		http.Error(w, "Expected WebSocket upgrade request", http.StatusBadRequest)
		return
	}

	// Extract JWT token
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

	// Parse IDs from claims
	employeeID, err := uuid.Parse(claims.EmployeeID)
	if err != nil {
		http.Error(w, "Invalid employee_id in token", http.StatusUnauthorized)
		return
	}

	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid org_id in token", http.StatusUnauthorized)
		return
	}

	// Verify employee is active
	employee, err := h.queries.GetEmployee(r.Context(), employeeID)
	if err != nil {
		http.Error(w, "Employee not found", http.StatusUnauthorized)
		return
	}

	if employee.Status != "active" {
		http.Error(w, "Employee account is not active", http.StatusForbidden)
		return
	}

	// Get team_id (may be null)
	var teamID *uuid.UUID
	if employee.TeamID.Valid {
		tid := uuid.UUID(employee.TeamID.Bytes)
		teamID = &tid
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Policy WebSocket upgrade failed: %v", err)
		return
	}

	// Create policy connection
	connID := uuid.New().String()
	policyConn := &PolicyConn{
		ID:          connID,
		OrgID:       orgID,
		EmployeeID:  employeeID,
		TeamID:      teamID,
		ConnectedAt: time.Now(),
		send:        make(chan []byte, 256),
		conn:        conn,
	}

	// Register connection with hub
	h.hub.Register(policyConn)

	log.Printf("Policy WebSocket connected: employee=%s org=%s", employeeID, orgID)

	// Fetch and send initial policies
	go h.sendInitialPolicies(r.Context(), policyConn)

	// Start read/write pumps
	go h.writePump(policyConn)
	go h.readPump(policyConn)
}

// sendInitialPolicies fetches and sends all applicable policies to a new connection
func (h *PolicyHandler) sendInitialPolicies(ctx context.Context, conn *PolicyConn) {
	// Build query params
	params := db.GetToolPoliciesForEmployeeParams{
		OrgID:      conn.OrgID,
		EmployeeID: pgtype.UUID{Bytes: conn.EmployeeID, Valid: true},
	}

	if conn.TeamID != nil {
		params.TeamID = pgtype.UUID{Bytes: *conn.TeamID, Valid: true}
	}

	// Fetch policies from database
	dbPolicies, err := h.queries.GetToolPoliciesForEmployee(ctx, params)
	if err != nil {
		log.Printf("Failed to fetch initial policies for connection %s: %v", conn.ID, err)
		return
	}

	// Convert to PolicyData format
	policies := make([]PolicyData, len(dbPolicies))
	for i, p := range dbPolicies {
		policies[i] = dbPolicyToPolicyData(p)
	}

	// Send init message
	if err := h.hub.SendInitMessage(conn, policies); err != nil {
		log.Printf("Failed to send init message to connection %s: %v", conn.ID, err)
	}
}

// readPump handles incoming messages from the proxy
func (h *PolicyHandler) readPump(conn *PolicyConn) {
	defer func() {
		h.hub.Unregister(conn)
		if wsConn, ok := conn.conn.(*websocket.Conn); ok {
			_ = wsConn.Close()
		}
		log.Printf("Policy WebSocket disconnected: employee=%s", conn.EmployeeID)
	}()

	wsConn, ok := conn.conn.(*websocket.Conn)
	if !ok {
		return
	}

	wsConn.SetReadLimit(maxMessageSize)
	_ = wsConn.SetReadDeadline(time.Now().Add(pongWait))
	wsConn.SetPongHandler(func(string) error {
		return wsConn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, message, err := wsConn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Policy WebSocket error: %v", err)
			}
			break
		}

		// Handle pong messages from proxy
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Currently only handle pong messages
		if msgType, ok := msg["type"].(string); ok && msgType == "pong" {
			// Heartbeat acknowledged
		}
	}
}

// writePump handles outgoing messages to the proxy
func (h *PolicyHandler) writePump(conn *PolicyConn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if wsConn, ok := conn.conn.(*websocket.Conn); ok {
			_ = wsConn.Close()
		}
	}()

	wsConn, ok := conn.conn.(*websocket.Conn)
	if !ok {
		return
	}

	for {
		select {
		case message, channelOpen := <-conn.send:
			_ = wsConn.SetWriteDeadline(time.Now().Add(writeWait))
			if !channelOpen {
				_ = wsConn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := wsConn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = wsConn.SetWriteDeadline(time.Now().Add(writeWait))

			// Send ping message as JSON
			pingMsg := PolicyMessage{Type: PolicyMessageTypePing}
			pingBytes, _ := json.Marshal(pingMsg)
			if err := wsConn.WriteMessage(websocket.TextMessage, pingBytes); err != nil {
				return
			}
		}
	}
}

// dbPolicyToPolicyData converts a database policy to WebSocket PolicyData
func dbPolicyToPolicyData(p db.ToolPolicy) PolicyData {
	pd := PolicyData{
		ID:       p.ID,
		OrgID:    p.OrgID,
		ToolName: p.ToolName,
		Action:   p.Action,
	}

	// Handle nullable Reason
	if p.Reason != nil {
		pd.Reason = *p.Reason
	}

	if p.TeamID.Valid {
		tid := uuid.UUID(p.TeamID.Bytes)
		pd.TeamID = &tid
	}

	if p.EmployeeID.Valid {
		eid := uuid.UUID(p.EmployeeID.Bytes)
		pd.EmployeeID = &eid
	}

	if len(p.Conditions) > 0 {
		var conditions map[string]interface{}
		if err := json.Unmarshal(p.Conditions, &conditions); err == nil {
			pd.Conditions = conditions
		}
	}

	// Determine scope
	if p.EmployeeID.Valid {
		pd.Scope = "employee"
	} else if p.TeamID.Valid {
		pd.Scope = "team"
	} else {
		pd.Scope = "organization"
	}

	if p.CreatedAt.Valid {
		pd.CreatedAt = p.CreatedAt.Time
	}

	if p.UpdatedAt.Valid {
		pd.UpdatedAt = &p.UpdatedAt.Time
	}

	return pd
}
