package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// PolicyMessage types for server-to-client communication
const (
	PolicyMessageTypeInit   = "init"
	PolicyMessageTypeUpsert = "upsert"
	PolicyMessageTypeDelete = "delete"
	PolicyMessageTypeRevoke = "revoke"
	PolicyMessageTypePing   = "ping"
)

// PolicyMessage represents a message sent from server to proxy
type PolicyMessage struct {
	Type     string        `json:"type"`
	Policies []PolicyData  `json:"policies,omitempty"` // For init
	Policy   *PolicyData   `json:"policy,omitempty"`   // For upsert
	PolicyID *uuid.UUID    `json:"policy_id,omitempty"` // For delete
	Reason   string        `json:"reason,omitempty"`    // For revoke
	Version  int64         `json:"version,omitempty"`   // For init
}

// PolicyData represents a policy in WebSocket messages
type PolicyData struct {
	ID         uuid.UUID              `json:"id"`
	OrgID      uuid.UUID              `json:"org_id"`
	TeamID     *uuid.UUID             `json:"team_id,omitempty"`
	EmployeeID *uuid.UUID             `json:"employee_id,omitempty"`
	ToolName   string                 `json:"tool_name"`
	Action     string                 `json:"action"`
	Reason     string                 `json:"reason,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Scope      string                 `json:"scope"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  *time.Time             `json:"updated_at,omitempty"`
}

// PolicyChangeNotification represents a notification from PostgreSQL NOTIFY
type PolicyChangeNotification struct {
	Action     string      `json:"action"` // create, update, delete, revoke
	Policy     *PolicyData `json:"policy,omitempty"`
	PolicyID   *uuid.UUID  `json:"policy_id,omitempty"`
	OrgID      uuid.UUID   `json:"org_id"`
	TeamID     *uuid.UUID  `json:"team_id,omitempty"`
	EmployeeID *uuid.UUID  `json:"employee_id,omitempty"`
}

// PolicyConn represents a proxy's WebSocket connection
type PolicyConn struct {
	ID          string
	OrgID       uuid.UUID
	EmployeeID  uuid.UUID
	TeamID      *uuid.UUID
	ConnectedAt time.Time
	send        chan []byte
	conn        interface{} // *websocket.Conn in real implementation
}

// PolicyHub manages proxy connections for policy updates
type PolicyHub struct {
	// All connections by ID
	connections map[string]*PolicyConn

	// Indexed for efficient broadcast
	byOrg      map[uuid.UUID]map[string]*PolicyConn // org_id -> conn_id -> conn
	byTeam     map[uuid.UUID]map[string]*PolicyConn // team_id -> conn_id -> conn
	byEmployee map[uuid.UUID]map[string]*PolicyConn // employee_id -> conn_id -> conn

	// Channels for registration/unregistration
	register   chan *PolicyConn
	unregister chan *PolicyConn

	// Channel for policy change notifications
	policyChange chan PolicyChangeNotification

	// Stop signal
	stop chan struct{}

	mu sync.RWMutex
}

// NewPolicyHub creates a new policy hub
func NewPolicyHub() *PolicyHub {
	return &PolicyHub{
		connections:  make(map[string]*PolicyConn),
		byOrg:        make(map[uuid.UUID]map[string]*PolicyConn),
		byTeam:       make(map[uuid.UUID]map[string]*PolicyConn),
		byEmployee:   make(map[uuid.UUID]map[string]*PolicyConn),
		register:     make(chan *PolicyConn),
		unregister:   make(chan *PolicyConn),
		policyChange: make(chan PolicyChangeNotification, 256),
		stop:         make(chan struct{}),
	}
}

// Run starts the hub's main event loop
func (h *PolicyHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.registerConnection(conn)

		case conn := <-h.unregister:
			h.unregisterConnection(conn)

		case notification := <-h.policyChange:
			h.handlePolicyChange(notification)

		case <-h.stop:
			return
		}
	}
}

// Stop signals the hub to stop
func (h *PolicyHub) Stop() {
	close(h.stop)
}

// Register adds a connection to the hub
func (h *PolicyHub) Register(conn *PolicyConn) {
	h.register <- conn
}

// Unregister removes a connection from the hub
func (h *PolicyHub) Unregister(conn *PolicyConn) {
	h.unregister <- conn
}

// NotifyPolicyChange sends a policy change notification to the hub
func (h *PolicyHub) NotifyPolicyChange(notification PolicyChangeNotification) {
	select {
	case h.policyChange <- notification:
	default:
		// Channel full, drop notification
	}
}

// registerConnection adds a connection to all indexes
func (h *PolicyHub) registerConnection(conn *PolicyConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Add to main map
	h.connections[conn.ID] = conn

	// Index by org
	if h.byOrg[conn.OrgID] == nil {
		h.byOrg[conn.OrgID] = make(map[string]*PolicyConn)
	}
	h.byOrg[conn.OrgID][conn.ID] = conn

	// Index by team if present
	if conn.TeamID != nil {
		if h.byTeam[*conn.TeamID] == nil {
			h.byTeam[*conn.TeamID] = make(map[string]*PolicyConn)
		}
		h.byTeam[*conn.TeamID][conn.ID] = conn
	}

	// Index by employee
	if h.byEmployee[conn.EmployeeID] == nil {
		h.byEmployee[conn.EmployeeID] = make(map[string]*PolicyConn)
	}
	h.byEmployee[conn.EmployeeID][conn.ID] = conn
}

// unregisterConnection removes a connection from all indexes
func (h *PolicyHub) unregisterConnection(conn *PolicyConn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.connections[conn.ID]; !ok {
		return
	}

	// Remove from main map
	delete(h.connections, conn.ID)

	// Remove from org index
	if orgConns := h.byOrg[conn.OrgID]; orgConns != nil {
		delete(orgConns, conn.ID)
		if len(orgConns) == 0 {
			delete(h.byOrg, conn.OrgID)
		}
	}

	// Remove from team index
	if conn.TeamID != nil {
		if teamConns := h.byTeam[*conn.TeamID]; teamConns != nil {
			delete(teamConns, conn.ID)
			if len(teamConns) == 0 {
				delete(h.byTeam, *conn.TeamID)
			}
		}
	}

	// Remove from employee index
	if empConns := h.byEmployee[conn.EmployeeID]; empConns != nil {
		delete(empConns, conn.ID)
		if len(empConns) == 0 {
			delete(h.byEmployee, conn.EmployeeID)
		}
	}

	// Close send channel
	close(conn.send)
}

// handlePolicyChange broadcasts policy changes to affected connections
func (h *PolicyHub) handlePolicyChange(notification PolicyChangeNotification) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var affectedConns map[string]*PolicyConn

	switch notification.Action {
	case "revoke":
		// Revocation targets a specific employee
		if notification.EmployeeID != nil {
			affectedConns = h.byEmployee[*notification.EmployeeID]
		}

	case "create", "update", "delete":
		// Determine affected connections based on policy scope
		if notification.EmployeeID != nil {
			// Employee-scoped policy: only that employee
			affectedConns = h.byEmployee[*notification.EmployeeID]
		} else if notification.TeamID != nil {
			// Team-scoped policy: all team members
			affectedConns = h.byTeam[*notification.TeamID]
		} else {
			// Org-scoped policy: all org members
			affectedConns = h.byOrg[notification.OrgID]
		}
	}

	if len(affectedConns) == 0 {
		return
	}

	// Build message based on action
	var msg PolicyMessage
	switch notification.Action {
	case "create", "update":
		msg = PolicyMessage{
			Type:   PolicyMessageTypeUpsert,
			Policy: notification.Policy,
		}
	case "delete":
		msg = PolicyMessage{
			Type:     PolicyMessageTypeDelete,
			PolicyID: notification.PolicyID,
		}
	case "revoke":
		msg = PolicyMessage{
			Type:   PolicyMessageTypeRevoke,
			Reason: "Employee account deactivated",
		}
	}

	// Serialize message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return
	}

	// Send to all affected connections
	for _, conn := range affectedConns {
		select {
		case conn.send <- msgBytes:
		default:
			// Send channel full, skip
		}
	}
}

// SendInitMessage sends the initial policy sync message to a connection
func (h *PolicyHub) SendInitMessage(conn *PolicyConn, policies []PolicyData) error {
	msg := PolicyMessage{
		Type:     PolicyMessageTypeInit,
		Policies: policies,
		Version:  time.Now().Unix(),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case conn.send <- msgBytes:
		return nil
	default:
		return nil // Channel full
	}
}

// GetConnectionCount returns the total number of connections (for monitoring)
func (h *PolicyHub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// GetConnectionCountByOrg returns connections for an organization (for monitoring)
func (h *PolicyHub) GetConnectionCountByOrg(orgID uuid.UUID) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.byOrg[orgID])
}

// IsConnected checks if a connection ID is currently registered (for testing)
func (h *PolicyHub) IsConnected(connID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.connections[connID]
	return ok
}
