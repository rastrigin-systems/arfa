package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
	authmiddleware "github.com/rastrigin-systems/arfa/services/api/internal/middleware"
	ws "github.com/rastrigin-systems/arfa/services/api/internal/websocket"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

func TestWebSocketIntegration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer func() { _ = conn.Close(testutil.GetContext(t)) }()
	ctx := testutil.GetContext(t)

	// Create test organization, role, and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "wstest@example.com",
		FullName: "WebSocket Tester",
		Status:   "active",
	})

	// Generate JWT token for authentication
	token, err := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	require.NoError(t, err)

	// Create session in database for JWT middleware
	tokenHash := auth.HashToken(token)
	_, err = queries.CreateSession(ctx, db.CreateSessionParams{
		EmployeeID: employee.ID,
		TokenHash:  tokenHash,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	require.NoError(t, err)

	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create handlers
	logsHandler := handlers.NewLogsHandler(queries, hub)
	wsHandler := ws.NewHandler(hub)

	// Setup router
	router := chi.NewRouter()

	// Apply JWT middleware only to REST endpoints (not WebSocket)
	// WebSocket handler has its own authentication logic
	router.Group(func(r chi.Router) {
		r.Use(authmiddleware.JWTAuth(queries))
		r.Post("/api/v1/logs", logsHandler.CreateLog)
	})

	// WebSocket endpoint without middleware (handles auth internally)
	router.Get("/api/v1/logs/stream", wsHandler.ServeHTTP)

	// Start test server
	server := httptest.NewServer(router)
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/v1/logs/stream"

	t.Run("WebSocket connection with valid JWT header", func(t *testing.T) {
		// Connect to WebSocket with JWT token in header (standard method)
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+token)

		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, headers)
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		defer func() { _ = conn.Close() }()

		// Connection successful
		assert.NotNil(t, conn)
	})

	t.Run("WebSocket connection with valid JWT query param", func(t *testing.T) {
		// Connect to WebSocket with JWT token in query parameter
		// This simulates browser WebSocket which can't set custom headers
		wsURLWithToken := wsURL + "?token=" + token

		conn, resp, err := websocket.DefaultDialer.Dial(wsURLWithToken, nil)
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		defer func() { _ = conn.Close() }()

		// Connection successful
		assert.NotNil(t, conn)
	})

	t.Run("WebSocket header takes precedence over query param", func(t *testing.T) {
		// When both header and query param present, header should be used
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+token)
		wsURLWithToken := wsURL + "?token=invalid-token"

		// Should succeed because valid token in header
		conn, resp, err := websocket.DefaultDialer.Dial(wsURLWithToken, headers)
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		defer func() { _ = conn.Close() }()

		assert.NotNil(t, conn)
	})

	t.Run("WebSocket connection with invalid query param token fails", func(t *testing.T) {
		// Invalid token in query parameter should fail
		wsURLWithToken := wsURL + "?token=invalid-token"

		_, resp, err := websocket.DefaultDialer.Dial(wsURLWithToken, nil)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("WebSocket connection without JWT fails", func(t *testing.T) {
		_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
		assert.Error(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Real-time log streaming", func(t *testing.T) {
		sessionID := uuid.New()

		// Connect to WebSocket with session filter
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+token)
		wsURLWithFilter := wsURL + "?session_id=" + sessionID.String()

		conn, _, err := websocket.DefaultDialer.Dial(wsURLWithFilter, headers)
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		// Channel to receive messages
		received := make(chan ws.LogMessage, 1)
		go func() {
			_, message, err := conn.ReadMessage()
			if err == nil {
				var logMsg ws.LogMessage
				if json.Unmarshal(message, &logMsg) == nil {
					received <- logMsg
				}
			}
		}()

		// Create a log via POST /api/v1/logs
		logReq := api.CreateLogRequest{
			EventType:     "session_start",
			EventCategory: "io",
			Content:       stringPtr("Test log message"),
		}
		reqBody, _ := json.Marshal(logReq)

		req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)

		// Wait for WebSocket message
		select {
		case msg := <-received:
			assert.Equal(t, "session_start", msg.EventType)
			assert.Equal(t, "io", msg.EventCategory)
			assert.Equal(t, "Test log message", msg.Content)
		case <-time.After(2 * time.Second):
			t.Fatal("Did not receive log message via WebSocket")
		}
	})

	// NOTE: Session ID filtering test removed - session_id param no longer exists in API

	t.Run("Multi-tenancy isolation", func(t *testing.T) {
		// Create second organization
		org2 := testutil.CreateTestOrg(t, queries, ctx)
		role2 := testutil.CreateTestRole(t, queries, ctx, "admin")
		employee2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org2.ID,
			RoleID:   role2.ID,
			Email:    "wstest2@example.com",
			FullName: "WebSocket Tester 2",
			Status:   "active",
		})

		// Generate token for org2 employee
		token2, err := auth.GenerateJWT(employee2.ID, org2.ID, 24*time.Hour)
		require.NoError(t, err)

		// Connect as org2 employee
		headers := http.Header{}
		headers.Set("Authorization", "Bearer "+token2)

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, headers)
		require.NoError(t, err)
		defer func() { _ = conn.Close() }()

		received := make(chan bool, 1)
		go func() {
			_, _, err := conn.ReadMessage()
			if err == nil {
				received <- true
			}
		}()

		// Create log for org1 (should NOT be received by org2 connection)
		logReq := api.CreateLogRequest{
			EventType:     "session_start",
			EventCategory: "io",
			Content:       stringPtr("Org 1 log"),
		}
		reqBody, _ := json.Marshal(logReq)

		req := httptest.NewRequest("POST", "/api/v1/logs", bytes.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)

		// Should NOT receive message (different org)
		select {
		case <-received:
			t.Fatal("Received message from different organization")
		case <-time.After(500 * time.Millisecond):
			// Expected - no message received due to multi-tenancy isolation
		}
	})
}

func stringPtr(s string) *string {
	return &s
}
