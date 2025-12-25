package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	assert.NotNil(t, hub)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.broadcast)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
}

func TestHub_RegisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	client := &Client{
		orgID: uuid.New(),
		send:  make(chan []byte, 256),
	}

	// Register client
	hub.register <- client

	// Give hub time to process
	time.Sleep(10 * time.Millisecond)

	// Verify client is registered
	assert.True(t, hub.IsClientRegistered(client))
}

func TestHub_UnregisterClient(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	client := &Client{
		orgID: uuid.New(),
		send:  make(chan []byte, 256),
	}

	// Register then unregister client
	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Verify client is unregistered
	assert.False(t, hub.IsClientRegistered(client))
}

func TestHub_BroadcastToMatchingClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID := uuid.New()
	sessionID := uuid.New()

	// Create client with session filter
	client := &Client{
		orgID: orgID,
		send:  make(chan []byte, 256),
		filters: ClientFilters{
			SessionID: sessionID,
		},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast log matching the filter
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID,
		SessionID:     sessionID,
		EventType:     "cli.start",
		EventCategory: "cli",
		Content:       "Test log",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(10 * time.Millisecond)

	// Client should receive the message
	select {
	case msg := <-client.send:
		var received LogMessage
		err := json.Unmarshal(msg, &received)
		require.NoError(t, err)
		assert.Equal(t, logMsg.Content, received.Content)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Client did not receive message")
	}
}

func TestHub_BroadcastFiltersNonMatchingClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID := uuid.New()
	sessionID1 := uuid.New()
	sessionID2 := uuid.New()

	// Create client with session filter
	client := &Client{
		orgID: orgID,
		send:  make(chan []byte, 256),
		filters: ClientFilters{
			SessionID: sessionID1,
		},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast log with different session ID
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID,
		SessionID:     sessionID2,
		EventType:     "cli.start",
		EventCategory: "cli",
		Content:       "Test log",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(10 * time.Millisecond)

	// Client should NOT receive the message
	select {
	case <-client.send:
		t.Fatal("Client received message it should have filtered")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message
	}
}

func TestHub_BroadcastMultipleTenantIsolation(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID1 := uuid.New()
	orgID2 := uuid.New()

	// Create clients from different organizations
	client1 := &Client{
		orgID: orgID1,
		send:  make(chan []byte, 256),
	}

	client2 := &Client{
		orgID: orgID2,
		send:  make(chan []byte, 256),
	}

	hub.register <- client1
	hub.register <- client2
	time.Sleep(10 * time.Millisecond)

	// Broadcast log for orgID1 only
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID1,
		EventType:     "cli.start",
		EventCategory: "cli",
		Content:       "Org 1 log",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(10 * time.Millisecond)

	// Client1 should receive, Client2 should not
	select {
	case <-client1.send:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Client1 did not receive message")
	}

	select {
	case <-client2.send:
		t.Fatal("Client2 received message from wrong org")
	case <-time.After(50 * time.Millisecond):
		// Expected - no message for client2
	}
}

func TestHub_BroadcastFilterByEmployeeID(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID := uuid.New()
	employeeID := uuid.New()

	client := &Client{
		orgID: orgID,
		send:  make(chan []byte, 256),
		filters: ClientFilters{
			EmployeeID: employeeID,
		},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast log matching employee filter
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID,
		EmployeeID:    employeeID,
		EventType:     "cli.start",
		EventCategory: "cli",
		Content:       "Employee log",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(10 * time.Millisecond)

	// Client should receive
	select {
	case <-client.send:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Client did not receive message")
	}
}

func TestHub_BroadcastFilterByAgentID(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID := uuid.New()
	employeeID := uuid.New()

	client := &Client{
		orgID: orgID,
		send:  make(chan []byte, 256),
		filters: ClientFilters{
			EmployeeID: employeeID,
		},
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast log matching employee filter
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID,
		EmployeeID:    employeeID,
		EventType:     "agent.invoked",
		EventCategory: "agent",
		Content:       "Agent log",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(10 * time.Millisecond)

	// Client should receive
	select {
	case <-client.send:
		// Expected
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Client did not receive message")
	}
}

func TestHub_MultipleConcurrentClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	orgID := uuid.New()
	sessionID := uuid.New()

	// Register 10 clients with same filter
	clients := make([]*Client, 10)
	for i := 0; i < 10; i++ {
		clients[i] = &Client{
			orgID: orgID,
			send:  make(chan []byte, 256),
			filters: ClientFilters{
				SessionID: sessionID,
			},
		}
		hub.register <- clients[i]
	}

	time.Sleep(50 * time.Millisecond)

	// Broadcast one log
	logMsg := LogMessage{
		ID:            uuid.New(),
		OrgID:         orgID,
		SessionID:     sessionID,
		EventType:     "cli.output",
		EventCategory: "cli",
		Content:       "Broadcast test",
		Timestamp:     time.Now(),
	}

	hub.broadcast <- logMsg
	time.Sleep(50 * time.Millisecond)

	// All clients should receive
	for i, client := range clients {
		select {
		case <-client.send:
			// Expected
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("Client %d did not receive message", i)
		}
	}
}
