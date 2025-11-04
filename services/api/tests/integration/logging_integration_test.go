package integration

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/service"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
)

// TestLoggingService_CreateLog_Integration tests creating I/O logs in real database
func TestLoggingService_CreateLog_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "test@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	// Get existing agent or create test agent with unique name
	agents, err := queries.ListAgents(ctx)
	var agent db.Agent
	if err != nil || len(agents) == 0 {
		// Create test agent with unique name
		agent, err = queries.CreateAgent(ctx, db.CreateAgentParams{
			Name:          "Test Agent " + uuid.New().String()[:8],
			Type:          "coding",
			Description:   "Test AI agent",
			Provider:      "Test",
			DefaultConfig: json.RawMessage(`{}`),
			Capabilities:  json.RawMessage(`{"coding":true}`),
			LlmProvider:   "test",
			LlmModel:      "test-model",
			IsPublic:      true,
		})
		require.NoError(t, err)
	} else {
		agent = agents[0]
	}

	loggingSvc := service.NewLoggingService(queries)

	sessionID := uuid.New()

	tests := []struct {
		name      string
		entry     service.LogEntry
		wantErr   bool
		checkFunc func(*testing.T, db.ActivityLog)
	}{
		{
			name: "create session_start log",
			entry: service.LogEntry{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				SessionID:     sessionID,
				AgentID:       agent.ID,
				EventType:     "session_start",
				EventCategory: "io",
				Payload: map[string]interface{}{
					"agent":   "claude-code",
					"version": "1.0.0",
					"command": "ubik",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, log db.ActivityLog) {
				assert.Equal(t, org.ID, log.OrgID)
				assert.Equal(t, "session_start", log.EventType)
				assert.Equal(t, "io", log.EventCategory)
				assert.True(t, log.SessionID.Valid)
				assert.True(t, log.AgentID.Valid)
			},
		},
		{
			name: "create input log with content",
			entry: service.LogEntry{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				SessionID:     sessionID,
				AgentID:       agent.ID,
				EventType:     "input",
				EventCategory: "io",
				Content:       "User typed: implement feature X",
				Payload: map[string]interface{}{
					"command": "chat",
					"length":  28,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, log db.ActivityLog) {
				assert.Equal(t, "input", log.EventType)
				assert.NotNil(t, log.Content)
				assert.Equal(t, "User typed: implement feature X", *log.Content)
			},
		},
		{
			name: "create output log with content",
			entry: service.LogEntry{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				SessionID:     sessionID,
				AgentID:       agent.ID,
				EventType:     "output",
				EventCategory: "io",
				Content:       "Assistant: I'll help you implement feature X...",
				Payload: map[string]interface{}{
					"tokens": 150,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, log db.ActivityLog) {
				assert.Equal(t, "output", log.EventType)
				assert.NotNil(t, log.Content)
			},
		},
		{
			name: "create error log",
			entry: service.LogEntry{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				SessionID:     sessionID,
				AgentID:       agent.ID,
				EventType:     "error",
				EventCategory: "io",
				Content:       "Error: connection timeout",
				Payload: map[string]interface{}{
					"error_code": "ETIMEDOUT",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, log db.ActivityLog) {
				assert.Equal(t, "error", log.EventType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loggingSvc.CreateLog(ctx, tt.entry)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			// Verify log was created by fetching session logs
			logs, err := queries.GetLogsBySession(ctx, pgtype.UUID{Bytes: sessionID, Valid: true})
			require.NoError(t, err)
			assert.NotEmpty(t, logs)

			// Run custom check function if provided
			if tt.checkFunc != nil {
				lastLog := logs[len(logs)-1]
				tt.checkFunc(t, lastLog)
			}
		})
	}
}

// TestLoggingService_GetLogsBySession_Integration tests retrieving session logs
func TestLoggingService_GetLogsBySession_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "test@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	loggingSvc := service.NewLoggingService(queries)
	sessionID := uuid.New()

	// Create multiple logs for the session
	logEntries := []service.LogEntry{
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			SessionID:     sessionID,
			EventType:     "session_start",
			EventCategory: "io",
			Payload:       map[string]interface{}{"agent": "claude-code"},
		},
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			SessionID:     sessionID,
			EventType:     "input",
			EventCategory: "io",
			Content:       "test input",
			Payload:       map[string]interface{}{},
		},
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			SessionID:     sessionID,
			EventType:     "output",
			EventCategory: "io",
			Content:       "test output",
			Payload:       map[string]interface{}{},
		},
	}

	for _, entry := range logEntries {
		err := loggingSvc.CreateLog(ctx, entry)
		require.NoError(t, err)
	}

	// Retrieve all logs for the session
	logs, err := loggingSvc.GetLogsBySession(ctx, sessionID)
	require.NoError(t, err)
	assert.Len(t, logs, 3)

	// Verify logs are ordered by created_at ASC
	assert.Equal(t, "session_start", logs[0].EventType)
	assert.Equal(t, "input", logs[1].EventType)
	assert.Equal(t, "output", logs[2].EventType)
}

// TestLoggingService_GetLogsByEmployee_Integration tests filtering employee logs
func TestLoggingService_GetLogsByEmployee_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "test@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	loggingSvc := service.NewLoggingService(queries)

	// Create logs with different categories and timestamps
	now := time.Now()

	logEntries := []service.LogEntry{
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			EventType:     "session_start",
			EventCategory: "io",
			Payload:       map[string]interface{}{},
		},
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			EventType:     "agent.installed",
			EventCategory: "agent",
			Payload:       map[string]interface{}{},
		},
		{
			OrgID:         org.ID,
			EmployeeID:    employee.ID,
			EventType:     "mcp.configured",
			EventCategory: "mcp",
			Payload:       map[string]interface{}{},
		},
	}

	for _, entry := range logEntries {
		err := loggingSvc.CreateLog(ctx, entry)
		require.NoError(t, err)
	}

	tests := []struct {
		name      string
		filters   service.LogFilters
		wantCount int
		wantTypes []string
	}{
		{
			name: "filter by io category",
			filters: service.LogFilters{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				EventCategory: "io",
				Limit:         50,
			},
			wantCount: 1,
			wantTypes: []string{"session_start"},
		},
		{
			name: "filter by agent category",
			filters: service.LogFilters{
				OrgID:         org.ID,
				EmployeeID:    employee.ID,
				EventCategory: "agent",
				Limit:         50,
			},
			wantCount: 1,
			wantTypes: []string{"agent.installed"},
		},
		{
			name: "no category filter (all logs)",
			filters: service.LogFilters{
				OrgID:      org.ID,
				EmployeeID: employee.ID,
				Limit:      50,
			},
			wantCount: 3,
		},
		{
			name: "filter by time range (last 1 hour)",
			filters: service.LogFilters{
				OrgID:      org.ID,
				EmployeeID: employee.ID,
				Since:      now.Add(-1 * time.Hour),
				Limit:      50,
			},
			wantCount: 3,
		},
		{
			name: "filter by time range (future - should be empty)",
			filters: service.LogFilters{
				OrgID:      org.ID,
				EmployeeID: employee.ID,
				Since:      now.Add(1 * time.Hour),
				Limit:      50,
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs, err := loggingSvc.GetLogsByEmployee(ctx, tt.filters)
			require.NoError(t, err)
			assert.Len(t, logs, tt.wantCount)

			if len(tt.wantTypes) > 0 {
				for i, expectedType := range tt.wantTypes {
					assert.Equal(t, expectedType, logs[i].EventType)
				}
			}
		})
	}
}

// TestLoggingService_DeleteOldLogs_Integration tests retention policy cleanup
func TestLoggingService_DeleteOldLogs_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "test@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	loggingSvc := service.NewLoggingService(queries)

	// Create a log entry
	err := loggingSvc.CreateLog(ctx, service.LogEntry{
		OrgID:         org.ID,
		EmployeeID:    employee.ID,
		EventType:     "test",
		EventCategory: "io",
		Payload:       map[string]interface{}{},
	})
	require.NoError(t, err)

	// Verify log exists
	logs, err := queries.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  org.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, logs, "log should exist before deletion")

	// Delete logs older than now (should delete the log we just created since it's in the past)
	// In real use, we'd delete logs older than 30+ days
	err = loggingSvc.DeleteOldLogs(ctx, time.Now().Add(1*time.Hour))
	require.NoError(t, err)

	// Verify log was deleted
	logs, err = queries.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  org.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Empty(t, logs, "logs should be deleted")
}

// TestRetentionPolicy_CleanupOldLogs_Integration tests the retention policy service
func TestRetentionPolicy_CleanupOldLogs_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "test@example.com",
		FullName: "Test User",
		Status:   "active",
	})

	loggingSvc := service.NewLoggingService(queries)

	// Create a test log
	err := loggingSvc.CreateLog(ctx, service.LogEntry{
		OrgID:         org.ID,
		EmployeeID:    employee.ID,
		EventType:     "test",
		EventCategory: "io",
		Payload:       map[string]interface{}{},
	})
	require.NoError(t, err)

	// Create retention policy with 0 days (defaults to 30 days)
	rp := service.NewRetentionPolicy(queries, 0)
	assert.Equal(t, 30, rp.RetentionDays, "should default to 30 days")

	// Run cleanup
	err = rp.CleanupOldLogs(ctx)
	require.NoError(t, err)

	// Verify logs still exist (created just now, within 30 day window)
	logs, err := queries.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  org.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, logs, "logs should still exist within 30 day retention")
}

// TestLoggingService_MultiTenancy_Integration verifies org-level isolation
func TestLoggingService_MultiTenancy_Integration(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two test organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "admin")

	employee1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "user1@org1.com",
		FullName: "User 1",
		Status:   "active",
	})

	employee2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "user2@org2.com",
		FullName: "User 2",
		Status:   "active",
	})

	loggingSvc := service.NewLoggingService(queries)

	// Create logs for org1
	err := loggingSvc.CreateLog(ctx, service.LogEntry{
		OrgID:         org1.ID,
		EmployeeID:    employee1.ID,
		EventType:     "test_org1",
		EventCategory: "io",
		Payload:       map[string]interface{}{},
	})
	require.NoError(t, err)

	// Create logs for org2
	err = loggingSvc.CreateLog(ctx, service.LogEntry{
		OrgID:         org2.ID,
		EmployeeID:    employee2.ID,
		EventType:     "test_org2",
		EventCategory: "io",
		Payload:       map[string]interface{}{},
	})
	require.NoError(t, err)

	// Verify org1 only sees their logs
	logs1, err := queries.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  org1.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Len(t, logs1, 1)
	assert.Equal(t, "test_org1", logs1[0].EventType)

	// Verify org2 only sees their logs
	logs2, err := queries.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  org2.ID,
		Limit:  10,
		Offset: 0,
	})
	require.NoError(t, err)
	assert.Len(t, logs2, 1)
	assert.Equal(t, "test_org2", logs2[0].EventType)
}
