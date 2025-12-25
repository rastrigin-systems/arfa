package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
)

func TestLoggingService_CreateLog(t *testing.T) {
	tests := []struct {
		name      string
		entry     LogEntry
		mockSetup func(*mocks.MockQuerier)
		wantErr   bool
	}{
		{
			name: "successful I/O log creation",
			entry: LogEntry{
				OrgID:         uuid.New(),
				EmployeeID:    uuid.New(),
				SessionID:     uuid.New(),
				AgentID:       uuid.New(),
				EventType:     "input",
				EventCategory: "io",
				Content:       "User typed: hello world",
				Payload:       map[string]interface{}{"command": "chat"},
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					CreateActivityLog(gomock.Any(), gomock.Any()).
					Return(db.ActivityLog{
						ID:            uuid.New(),
						OrgID:         uuid.New(),
						EventType:     "input",
						EventCategory: "io",
					}, nil)
			},
			wantErr: false,
		},
		{
			name: "session start log",
			entry: LogEntry{
				OrgID:         uuid.New(),
				EmployeeID:    uuid.New(),
				SessionID:     uuid.New(),
				AgentID:       uuid.New(),
				EventType:     "session_start",
				EventCategory: "io",
				Payload:       map[string]interface{}{"agent": "claude-code", "version": "1.0"},
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					CreateActivityLog(gomock.Any(), gomock.Any()).
					Return(db.ActivityLog{}, nil)
			},
			wantErr: false,
		},
		{
			name: "error log with content",
			entry: LogEntry{
				OrgID:         uuid.New(),
				EmployeeID:    uuid.New(),
				SessionID:     uuid.New(),
				AgentID:       uuid.New(),
				EventType:     "error",
				EventCategory: "io",
				Content:       "Error: connection refused",
				Payload:       map[string]interface{}{"error_code": "ECONNREFUSED"},
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					CreateActivityLog(gomock.Any(), gomock.Any()).
					Return(db.ActivityLog{}, nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.mockSetup(mockDB)

			svc := NewLoggingService(mockDB)
			err := svc.CreateLog(context.Background(), tt.entry)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoggingService_GetLogsBySession(t *testing.T) {
	sessionID := uuid.New()
	orgID := uuid.New()
	employeeID := uuid.New()

	tests := []struct {
		name      string
		sessionID uuid.UUID
		mockSetup func(*mocks.MockQuerier)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "retrieve session logs",
			sessionID: sessionID,
			mockSetup: func(m *mocks.MockQuerier) {
				logs := []db.ActivityLog{
					{
						ID:            uuid.New(),
						OrgID:         orgID,
						EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
						SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
						EventType:     "session_start",
						EventCategory: "io",
					},
					{
						ID:            uuid.New(),
						OrgID:         orgID,
						EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
						SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
						EventType:     "input",
						EventCategory: "io",
					},
					{
						ID:            uuid.New(),
						OrgID:         orgID,
						EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
						SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
						EventType:     "output",
						EventCategory: "io",
					},
				}
				m.EXPECT().
					GetLogsBySession(gomock.Any(), pgtype.UUID{Bytes: sessionID, Valid: true}).
					Return(logs, nil)
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name:      "empty session logs",
			sessionID: uuid.New(),
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetLogsBySession(gomock.Any(), gomock.Any()).
					Return([]db.ActivityLog{}, nil)
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.mockSetup(mockDB)

			svc := NewLoggingService(mockDB)
			logs, err := svc.GetLogsBySession(context.Background(), tt.sessionID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, logs, tt.wantCount)
			}
		})
	}
}

func TestLoggingService_GetLogsByEmployee(t *testing.T) {
	orgID := uuid.New()
	employeeID := uuid.New()

	tests := []struct {
		name      string
		filters   LogFilters
		mockSetup func(*mocks.MockQuerier)
		wantCount int
		wantErr   bool
	}{
		{
			name: "filter by category",
			filters: LogFilters{
				OrgID:         orgID,
				EmployeeID:    employeeID,
				EventCategory: "io",
				Limit:         50,
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetLogsByEmployee(gomock.Any(), gomock.Any()).
					Return([]db.ActivityLog{
						{EventCategory: "io"},
						{EventCategory: "io"},
					}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "filter by time range",
			filters: LogFilters{
				OrgID:      orgID,
				EmployeeID: employeeID,
				Since:      time.Now().Add(-24 * time.Hour),
				Limit:      50,
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetLogsByEmployee(gomock.Any(), gomock.Any()).
					Return([]db.ActivityLog{{}, {}}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "pagination",
			filters: LogFilters{
				OrgID:      orgID,
				EmployeeID: employeeID,
				Limit:      10,
				Offset:     20,
			},
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					GetLogsByEmployee(gomock.Any(), gomock.Any()).
					Return([]db.ActivityLog{{}, {}}, nil)
			},
			wantCount: 2,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.mockSetup(mockDB)

			svc := NewLoggingService(mockDB)
			logs, err := svc.GetLogsByEmployee(context.Background(), tt.filters)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, logs, tt.wantCount)
			}
		})
	}
}

func TestLoggingService_DeleteOldLogs(t *testing.T) {
	tests := []struct {
		name      string
		olderThan time.Time
		mockSetup func(*mocks.MockQuerier)
		wantErr   bool
	}{
		{
			name:      "delete logs older than 30 days",
			olderThan: time.Now().Add(-30 * 24 * time.Hour),
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					DeleteOldLogs(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "delete logs older than 90 days",
			olderThan: time.Now().Add(-90 * 24 * time.Hour),
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					DeleteOldLogs(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.mockSetup(mockDB)

			svc := NewLoggingService(mockDB)
			err := svc.DeleteOldLogs(context.Background(), tt.olderThan)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogEntry_Validation(t *testing.T) {
	tests := []struct {
		name    string
		entry   LogEntry
		wantErr bool
	}{
		{
			name: "valid I/O log",
			entry: LogEntry{
				OrgID:         uuid.New(),
				EmployeeID:    uuid.New(),
				SessionID:     uuid.New(),
				EventType:     "input",
				EventCategory: "io",
				Content:       "test content",
			},
			wantErr: false,
		},
		{
			name: "missing org_id",
			entry: LogEntry{
				EmployeeID:    uuid.New(),
				EventType:     "input",
				EventCategory: "io",
			},
			wantErr: true,
		},
		{
			name: "missing event_type",
			entry: LogEntry{
				OrgID:         uuid.New(),
				EmployeeID:    uuid.New(),
				EventCategory: "io",
			},
			wantErr: true,
		},
		{
			name: "missing event_category",
			entry: LogEntry{
				OrgID:      uuid.New(),
				EmployeeID: uuid.New(),
				EventType:  "input",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLogEntry_MarshalPayload(t *testing.T) {
	entry := LogEntry{
		OrgID:         uuid.New(),
		EmployeeID:    uuid.New(),
		EventType:     "input",
		EventCategory: "io",
		Payload: map[string]interface{}{
			"command":  "chat",
			"duration": 1.5,
			"tool":     "filesystem",
		},
	}

	payload, err := json.Marshal(entry.Payload)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(payload, &result)
	require.NoError(t, err)

	assert.Equal(t, "chat", result["command"])
	assert.Equal(t, 1.5, result["duration"])
	assert.Equal(t, "filesystem", result["tool"])
}
