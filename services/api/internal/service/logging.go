package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/rastrigin-systems/arfa/generated/db"
)

// LogEntry represents a log entry to be created
type LogEntry struct {
	OrgID         uuid.UUID
	EmployeeID    uuid.UUID
	SessionID     uuid.UUID
	ClientName    string // e.g., "claude-code", "cursor"
	ClientVersion string // e.g., "1.0.25"
	EventType     string
	EventCategory string
	Content       string
	Payload       map[string]interface{}
}

// Validate checks if the log entry has required fields
func (e *LogEntry) Validate() error {
	if e.OrgID == uuid.Nil {
		return fmt.Errorf("org_id is required")
	}
	if e.EventType == "" {
		return fmt.Errorf("event_type is required")
	}
	if e.EventCategory == "" {
		return fmt.Errorf("event_category is required")
	}
	return nil
}

// LogFilters represents filters for querying logs
type LogFilters struct {
	OrgID         uuid.UUID
	EmployeeID    uuid.UUID
	EventCategory string
	Since         time.Time
	Limit         int32
	Offset        int32
}

// LoggingService handles activity log operations
type LoggingService struct {
	db db.Querier
}

// NewLoggingService creates a new logging service
func NewLoggingService(db db.Querier) *LoggingService {
	return &LoggingService{
		db: db,
	}
}

// CreateLog creates a new activity log entry
func (s *LoggingService) CreateLog(ctx context.Context, entry LogEntry) error {
	// Validate entry
	if err := entry.Validate(); err != nil {
		return fmt.Errorf("invalid log entry: %w", err)
	}

	// Marshal payload to JSON
	payloadJSON, err := json.Marshal(entry.Payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Convert UUIDs to pgtype.UUID
	sessionID := pgtype.UUID{Valid: false}
	if entry.SessionID != uuid.Nil {
		sessionID = pgtype.UUID{Bytes: entry.SessionID, Valid: true}
	}

	employeeID := pgtype.UUID{Valid: false}
	if entry.EmployeeID != uuid.Nil {
		employeeID = pgtype.UUID{Bytes: entry.EmployeeID, Valid: true}
	}

	// Create nullable string pointers for client info
	var clientName *string
	if entry.ClientName != "" {
		clientName = &entry.ClientName
	}
	var clientVersion *string
	if entry.ClientVersion != "" {
		clientVersion = &entry.ClientVersion
	}

	// Create content pointer (nullable field)
	var content *string
	if entry.Content != "" {
		content = &entry.Content
	}

	// Create the log entry
	_, err = s.db.CreateActivityLog(ctx, db.CreateActivityLogParams{
		OrgID:         entry.OrgID,
		EmployeeID:    employeeID,
		SessionID:     sessionID,
		ClientName:    clientName,
		ClientVersion: clientVersion,
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
		Content:       content,
		Payload:       payloadJSON,
	})

	if err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

// GetLogsBySession retrieves all logs for a specific CLI session
func (s *LoggingService) GetLogsBySession(ctx context.Context, sessionID uuid.UUID) ([]db.ActivityLog, error) {
	if sessionID == uuid.Nil {
		return nil, fmt.Errorf("session_id is required")
	}

	logs, err := s.db.GetLogsBySession(ctx, pgtype.UUID{Bytes: sessionID, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("failed to get logs by session: %w", err)
	}

	return logs, nil
}

// GetLogsByEmployee retrieves logs for an employee with filters
func (s *LoggingService) GetLogsByEmployee(ctx context.Context, filters LogFilters) ([]db.ActivityLog, error) {
	// Validate required filters
	if filters.OrgID == uuid.Nil {
		return nil, fmt.Errorf("org_id is required")
	}
	if filters.EmployeeID == uuid.Nil {
		return nil, fmt.Errorf("employee_id is required")
	}

	// Set defaults
	if filters.Limit == 0 {
		filters.Limit = 50
	}

	// Convert filters to query parameters
	var category *string
	if filters.EventCategory != "" {
		category = &filters.EventCategory
	}

	var since pgtype.Timestamp
	if !filters.Since.IsZero() {
		since = pgtype.Timestamp{Time: filters.Since, Valid: true}
	}

	logs, err := s.db.GetLogsByEmployee(ctx, db.GetLogsByEmployeeParams{
		OrgID:         filters.OrgID,
		EmployeeID:    pgtype.UUID{Bytes: filters.EmployeeID, Valid: true},
		EventCategory: category,
		Since:         since,
		QueryLimit:    filters.Limit,
		QueryOffset:   filters.Offset,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get logs by employee: %w", err)
	}

	return logs, nil
}

// DeleteOldLogs deletes activity logs older than the specified timestamp
func (s *LoggingService) DeleteOldLogs(ctx context.Context, olderThan time.Time) error {
	if olderThan.IsZero() {
		return fmt.Errorf("olderThan timestamp is required")
	}

	err := s.db.DeleteOldLogs(ctx, pgtype.Timestamp{Time: olderThan, Valid: true})
	if err != nil {
		return fmt.Errorf("failed to delete old logs: %w", err)
	}

	return nil
}
