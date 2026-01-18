package control

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
)

// ProxyState represents the current state of the proxy connection
type ProxyState string

const (
	StateConnecting   ProxyState = "connecting"   // Waiting for initial sync
	StateReady        ProxyState = "ready"        // Normal operation
	StateDisconnected ProxyState = "disconnected" // Lost connection (grace period)
	StateRevoked      ProxyState = "revoked"      // Access revoked (block all)
)

// PolicyClientConfig holds configuration for PolicyClient
type PolicyClientConfig struct {
	APIURL           string        // Base API URL (e.g., http://localhost:3001)
	Token            string        // JWT token for authentication
	GracePeriod      time.Duration // Time to allow cached policies after disconnect (default: 5m)
	ReconnectBackoff time.Duration // Initial backoff for reconnection (default: 1s)
	MaxReconnectWait time.Duration // Max backoff for reconnection (default: 30s)
}

// PolicyMessage represents a message from the server
type PolicyMessage struct {
	Type     string       `json:"type"`
	Policies []PolicyData `json:"policies,omitempty"`
	Policy   *PolicyData  `json:"policy,omitempty"`
	PolicyID *string      `json:"policy_id,omitempty"`
	Reason   string       `json:"reason,omitempty"`
	Version  int64        `json:"version,omitempty"`
}

// PolicyData represents a policy in WebSocket messages
type PolicyData struct {
	ID         string                 `json:"id"`
	OrgID      string                 `json:"org_id"`
	TeamID     *string                `json:"team_id,omitempty"`
	EmployeeID *string                `json:"employee_id,omitempty"`
	ToolName   string                 `json:"tool_name"`
	Action     string                 `json:"action"`
	Reason     string                 `json:"reason,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	Scope      string                 `json:"scope"`
}

// PolicyClient manages WebSocket connection for real-time policy updates
type PolicyClient struct {
	config PolicyClientConfig
	conn   *websocket.Conn

	// Policy storage
	policies map[string]PolicyData // id -> policy
	mu       sync.RWMutex

	// State management
	state          ProxyState
	stateMu        sync.RWMutex
	lastContact    time.Time
	disconnectedAt time.Time

	// Callbacks
	onStateChange     func(ProxyState)
	onPoliciesChanged func()

	// Control channels
	done   chan struct{}
	initCh chan struct{} // Closed when initial policies received
}

// NewPolicyClient creates a new PolicyClient
func NewPolicyClient(config PolicyClientConfig) *PolicyClient {
	if config.GracePeriod == 0 {
		config.GracePeriod = 5 * time.Minute
	}
	if config.ReconnectBackoff == 0 {
		config.ReconnectBackoff = 1 * time.Second
	}
	if config.MaxReconnectWait == 0 {
		config.MaxReconnectWait = 30 * time.Second
	}

	return &PolicyClient{
		config:   config,
		policies: make(map[string]PolicyData),
		state:    StateConnecting,
		done:     make(chan struct{}),
		initCh:   make(chan struct{}),
	}
}

// SetOnStateChange sets callback for state changes
func (c *PolicyClient) SetOnStateChange(fn func(ProxyState)) {
	c.onStateChange = fn
}

// SetOnPoliciesChanged sets callback for policy changes
func (c *PolicyClient) SetOnPoliciesChanged(fn func()) {
	c.onPoliciesChanged = fn
}

// Connect establishes WebSocket connection and starts receiving policies
func (c *PolicyClient) Connect(ctx context.Context) error {
	// Build WebSocket URL
	wsURL, err := c.buildWebSocketURL()
	if err != nil {
		return fmt.Errorf("invalid API URL: %w", err)
	}

	// Set up headers
	header := http.Header{}
	header.Set("Authorization", "Bearer "+c.config.Token)

	// Connect
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return fmt.Errorf("websocket connection failed: %w", err)
	}

	c.conn = conn
	c.lastContact = time.Now()

	// Start read loop in goroutine
	go c.readLoop(ctx)

	return nil
}

// ConnectWithRetry connects with automatic reconnection
func (c *PolicyClient) ConnectWithRetry(ctx context.Context) {
	backoff := c.config.ReconnectBackoff

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.done:
			return
		default:
		}

		err := c.Connect(ctx)
		if err == nil {
			// Connected successfully, wait for init message
			select {
			case <-c.initCh:
				// Initial policies received
				c.setState(StateReady)
				log.Println("Policy client connected and ready")
				backoff = c.config.ReconnectBackoff // Reset backoff

				// Wait for disconnect or done
				<-c.done
			case <-ctx.Done():
				return
			case <-time.After(30 * time.Second):
				// Timeout waiting for init
				log.Println("Timeout waiting for initial policies")
				c.Close()
			}
		} else {
			log.Printf("Policy client connection failed: %v", err)
		}

		// Connection lost or failed - enter disconnected state
		c.stateMu.Lock()
		if c.state != StateRevoked {
			c.state = StateDisconnected
			c.disconnectedAt = time.Now()
		}
		c.stateMu.Unlock()

		if c.onStateChange != nil {
			c.onStateChange(StateDisconnected)
		}

		// Wait before retry
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}

		// Increase backoff (exponential)
		backoff = backoff * 2
		if backoff > c.config.MaxReconnectWait {
			backoff = c.config.MaxReconnectWait
		}

		// Reset channels for new connection
		c.done = make(chan struct{})
		c.initCh = make(chan struct{})
	}
}

// buildWebSocketURL constructs the WebSocket URL from API URL
func (c *PolicyClient) buildWebSocketURL() (string, error) {
	u, err := url.Parse(c.config.APIURL)
	if err != nil {
		return "", err
	}

	// Convert http(s) to ws(s)
	switch u.Scheme {
	case "http":
		u.Scheme = "ws"
	case "https":
		u.Scheme = "wss"
	}

	u.Path = "/api/v1/ws/policies"
	return u.String(), nil
}

// readLoop reads messages from WebSocket
func (c *PolicyClient) readLoop(ctx context.Context) {
	defer func() {
		if c.conn != nil {
			c.conn.Close()
		}
		close(c.done)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Policy WebSocket error: %v", err)
			}
			return
		}

		c.lastContact = time.Now()
		c.handleMessage(message)
	}
}

// handleMessage processes incoming WebSocket messages
func (c *PolicyClient) handleMessage(data []byte) {
	var msg PolicyMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("Failed to parse policy message: %v", err)
		return
	}

	switch msg.Type {
	case "init":
		c.handleInit(msg)
	case "upsert":
		c.handleUpsert(msg)
	case "delete":
		c.handleDelete(msg)
	case "revoke":
		c.handleRevoke(msg)
	case "ping":
		c.handlePing()
	}
}

// handleInit processes initial policy sync
func (c *PolicyClient) handleInit(msg PolicyMessage) {
	c.mu.Lock()
	c.policies = make(map[string]PolicyData)
	for _, p := range msg.Policies {
		c.policies[p.ID] = p
	}
	c.mu.Unlock()

	log.Printf("Received %d policies (version %d)", len(msg.Policies), msg.Version)

	// Signal that init is complete
	select {
	case <-c.initCh:
		// Already closed
	default:
		close(c.initCh)
	}

	if c.onPoliciesChanged != nil {
		c.onPoliciesChanged()
	}
}

// handleUpsert processes policy create/update
func (c *PolicyClient) handleUpsert(msg PolicyMessage) {
	if msg.Policy == nil {
		return
	}

	c.mu.Lock()
	c.policies[msg.Policy.ID] = *msg.Policy
	c.mu.Unlock()

	log.Printf("Policy upserted: %s (%s)", msg.Policy.ToolName, msg.Policy.Action)

	if c.onPoliciesChanged != nil {
		c.onPoliciesChanged()
	}
}

// handleDelete processes policy deletion
func (c *PolicyClient) handleDelete(msg PolicyMessage) {
	if msg.PolicyID == nil {
		return
	}

	c.mu.Lock()
	delete(c.policies, *msg.PolicyID)
	c.mu.Unlock()

	log.Printf("Policy deleted: %s", *msg.PolicyID)

	if c.onPoliciesChanged != nil {
		c.onPoliciesChanged()
	}
}

// handleRevoke processes access revocation
func (c *PolicyClient) handleRevoke(msg PolicyMessage) {
	c.setState(StateRevoked)
	log.Printf("Access revoked: %s", msg.Reason)

	// Close connection
	if c.conn != nil {
		c.conn.Close()
	}
}

// handlePing responds to server ping
func (c *PolicyClient) handlePing() {
	if c.conn == nil {
		return
	}

	pong := map[string]string{"type": "pong"}
	data, _ := json.Marshal(pong)
	_ = c.conn.WriteMessage(websocket.TextMessage, data)
}

// setState updates the state and triggers callback
func (c *PolicyClient) setState(state ProxyState) {
	c.stateMu.Lock()
	c.state = state
	c.stateMu.Unlock()

	if c.onStateChange != nil {
		c.onStateChange(state)
	}
}

// GetState returns current proxy state
func (c *PolicyClient) GetState() ProxyState {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state
}

// ShouldBlockAll returns true if all requests should be blocked
func (c *PolicyClient) ShouldBlockAll() bool {
	c.stateMu.RLock()
	state := c.state
	disconnectedAt := c.disconnectedAt
	c.stateMu.RUnlock()

	switch state {
	case StateConnecting:
		return true // Block until ready
	case StateRevoked:
		return true // Block all after revocation
	case StateDisconnected:
		// Check if grace period expired
		if time.Since(disconnectedAt) > c.config.GracePeriod {
			return true
		}
		return false // Still within grace period
	default:
		return false
	}
}

// GetPolicies returns a copy of all current policies as api.ToolPolicy slice
func (c *PolicyClient) GetPolicies() []api.ToolPolicy {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]api.ToolPolicy, 0, len(c.policies))
	for _, p := range c.policies {
		result = append(result, c.toPolicyAPI(p))
	}
	return result
}

// toPolicyAPI converts PolicyData to api.ToolPolicy
func (c *PolicyClient) toPolicyAPI(p PolicyData) api.ToolPolicy {
	policy := api.ToolPolicy{
		ToolName:   p.ToolName,
		Action:     api.ToolPolicyAction(p.Action),
		Conditions: p.Conditions,
	}

	if p.Reason != "" {
		policy.Reason = &p.Reason
	}

	return policy
}

// Close closes the WebSocket connection
func (c *PolicyClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// WaitReady blocks until the client is ready or context is cancelled
func (c *PolicyClient) WaitReady(ctx context.Context) error {
	select {
	case <-c.initCh:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// PolicyCount returns the number of cached policies
func (c *PolicyClient) PolicyCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.policies)
}

// IsConnected returns true if WebSocket is connected
func (c *PolicyClient) IsConnected() bool {
	c.stateMu.RLock()
	defer c.stateMu.RUnlock()
	return c.state == StateReady
}
