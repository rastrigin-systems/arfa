package logging

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rastrigin-systems/arfa/services/cli/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogger tests logger initialization
func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		expectNil   bool
		description string
	}{
		{
			name:        "enabled logger",
			enabled:     true,
			expectNil:   false,
			description: "should create logger when enabled",
		},
		{
			name:        "disabled logger",
			enabled:     false,
			expectNil:   true,
			description: "should return nil when disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Enabled:       tt.enabled,
				BatchSize:     100,
				BatchInterval: 5 * time.Second,
				QueueDir:      t.TempDir(),
			}

			logger, err := NewLogger(config, nil)

			if tt.expectNil {
				assert.Nil(t, logger, "expected nil logger when disabled")
				assert.NoError(t, err, "no error expected when disabled")
			} else {
				require.NotNil(t, logger, "expected non-nil logger when enabled")
				require.NoError(t, err, "no error expected")

				// Clean up
				logger.Close()
			}
		})
	}
}

// TestLoggerDisabledViaEnv tests ARFA_NO_LOGGING environment variable
func TestLoggerDisabledViaEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("ARFA_NO_LOGGING", "1")
	defer os.Unsetenv("ARFA_NO_LOGGING")

	config := &Config{
		Enabled:       true, // explicitly enabled in config
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	logger, err := NewLogger(config, nil)
	assert.Nil(t, logger, "logger should be nil when ARFA_NO_LOGGING=1")
	assert.NoError(t, err)
}

// TestSessionTracking tests session start/end events
func TestSessionTracking(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	// Start session
	sessionID := logger.StartSession()
	assert.NotEqual(t, uuid.Nil, sessionID, "session ID should not be nil")

	// Give time for log to be recorded
	time.Sleep(100 * time.Millisecond)

	// End session
	logger.EndSession()

	// Force flush
	logger.Flush()

	// Wait for async processing
	time.Sleep(200 * time.Millisecond)

	// Verify session_start and session_end events
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	assert.GreaterOrEqual(t, len(mockAPI.logs), 2, "should have at least session_start and session_end")

	// Find session_start
	var foundStart, foundEnd bool
	for _, log := range mockAPI.logs {
		if log.EventType == "session_start" {
			foundStart = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
		if log.EventType == "session_end" {
			foundEnd = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}

	assert.True(t, foundStart, "should have session_start event")
	assert.True(t, foundEnd, "should have session_end event")
}

// TestLogEvent tests the LogEvent method
func TestLogEvent(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	sessionID := logger.StartSession()

	// Log a custom event
	logger.LogEvent("api_request", "proxy", "POST https://api.anthropic.com", map[string]interface{}{
		"method": "POST",
		"url":    "https://api.anthropic.com/v1/messages",
	})

	// Force flush
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify log captured
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundEvent bool
	for _, log := range mockAPI.logs {
		if log.EventType == "api_request" && log.EventCategory == "proxy" {
			foundEvent = true
			assert.Equal(t, sessionID.String(), log.SessionID)
			assert.Equal(t, "POST https://api.anthropic.com", log.Content)
		}
	}

	assert.True(t, foundEvent, "should capture custom event")
}

// TestBatchSending tests batching behavior
func TestBatchSending(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     5,                // small batch for testing
		BatchInterval: 10 * time.Second, // long interval, we'll hit size limit first
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	sessionID := logger.StartSession()

	// Log exactly batch size entries
	for i := 0; i < 5; i++ {
		logger.LogEvent("test_event", "test", "test content", map[string]interface{}{"index": i})
	}

	// Wait for batch to be sent
	time.Sleep(300 * time.Millisecond)

	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	// Should have session_start + 5 events = 6 total
	assert.GreaterOrEqual(t, logCount, 5, "should have sent batch")

	// Verify session ID in all logs
	mockAPI.mu.Lock()
	for _, log := range mockAPI.logs {
		if log.EventType == "test_event" {
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}
	mockAPI.mu.Unlock()
}

// TestOfflineQueue tests queuing when API is unavailable
func TestOfflineQueue(t *testing.T) {
	queueDir := t.TempDir()

	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      queueDir,
		MaxRetries:    2, // Fast retries for testing
		RetryBackoff:  50 * time.Millisecond,
	}

	// Create API client that always fails
	mockAPI := &mockAPIClient{
		logs:       make([]LogEntry, 0),
		shouldFail: true,
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)

	_ = logger.StartSession()
	logger.LogEvent("test_event", "test", "test content", nil)
	logger.EndSession()

	// Flush to trigger send (which will fail after retries)
	logger.Flush()
	// Wait for retries: 50ms + 100ms + queue = ~300ms
	time.Sleep(400 * time.Millisecond)

	// Close logger
	logger.Close()

	// Verify queue files exist
	files, err := os.ReadDir(queueDir)
	require.NoError(t, err)

	queueFiles := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			queueFiles++
		}
	}

	assert.Greater(t, queueFiles, 0, "should have queued logs to disk")

	// Now create logger with working API
	mockAPI.shouldFail = false

	logger2, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger2)

	// Give time for queue processing
	time.Sleep(500 * time.Millisecond)

	logger2.Close()

	// Verify logs were sent
	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	assert.Greater(t, logCount, 0, "queued logs should be sent when API available")
}

// TestRetryWithBackoff tests exponential backoff retry logic
func TestRetryWithBackoff(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
		MaxRetries:    3,
	}

	mockAPI := &mockAPIClient{
		logs:       make([]LogEntry, 0),
		failCount:  2, // fail first 2 attempts, succeed on 3rd
		shouldFail: true,
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()
	logger.LogEvent("test_event", "test", "test retry", nil)

	// Flush to trigger send
	logger.Flush()

	// Wait for retries
	time.Sleep(2 * time.Second)

	mockAPI.mu.Lock()
	attempts := mockAPI.attempts
	mockAPI.mu.Unlock()

	assert.GreaterOrEqual(t, attempts, 2, "should have retried at least twice")
}

// TestConcurrentLogging tests thread safety
func TestConcurrentLogging(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()

	// Spawn multiple goroutines logging concurrently
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 5; j++ {
				logger.LogEvent("test_event", "test", "concurrent test", map[string]interface{}{
					"goroutine": id,
					"iteration": j,
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	logger.Flush()
	time.Sleep(300 * time.Millisecond)

	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	// Should have session_start + 50 events
	assert.GreaterOrEqual(t, logCount, 50, "should handle concurrent logging")
}

// TestLoggerClose tests cleanup on close
func TestLoggerClose(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.StartSession()
	logger.LogEvent("test_event", "test", "test before close", nil)

	// Close should flush pending logs
	logger.Close()

	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	assert.Greater(t, logCount, 0, "should flush logs on close")
}

// TestQueuePersistence tests that queue survives restarts
func TestQueuePersistence(t *testing.T) {
	queueDir := t.TempDir()

	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      queueDir,
		MaxRetries:    2, // Fast retries for testing
		RetryBackoff:  50 * time.Millisecond,
	}

	// First logger with failing API
	mockAPI1 := &mockAPIClient{
		logs:       make([]LogEntry, 0),
		shouldFail: true,
	}

	logger1, err := NewLogger(config, mockAPI1)
	require.NoError(t, err)
	require.NotNil(t, logger1)

	logger1.StartSession()
	logger1.LogEvent("test_event", "test", "persisted message", nil)
	logger1.Flush()
	// Wait for retries: 50ms + 100ms + queue = ~300ms
	time.Sleep(400 * time.Millisecond)
	logger1.Close()

	// Verify queue file exists
	files, err := os.ReadDir(queueDir)
	require.NoError(t, err)
	require.Greater(t, len(files), 0, "queue file should exist")

	// Second logger with working API
	mockAPI2 := &mockAPIClient{
		logs:       make([]LogEntry, 0),
		shouldFail: false,
	}

	logger2, err := NewLogger(config, mockAPI2)
	require.NoError(t, err)
	require.NotNil(t, logger2)

	// Give time to process queue
	time.Sleep(500 * time.Millisecond)
	logger2.Close()

	// Verify persisted log was sent
	mockAPI2.mu.Lock()
	defer mockAPI2.mu.Unlock()

	var foundPersisted bool
	for _, log := range mockAPI2.logs {
		if log.EventType == "test_event" && strings.Contains(log.Content, "persisted message") {
			foundPersisted = true
			break
		}
	}

	assert.True(t, foundPersisted, "persisted log should be sent after restart")
}

// TestSetClient tests setting client name and version
func TestSetClient(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()
	logger.SetClient("claude-code", "1.0.25")

	logger.LogEvent("test_event", "test", "test content", nil)

	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundWithClient bool
	for _, log := range mockAPI.logs {
		if log.EventType == "test_event" && log.ClientName == "claude-code" && log.ClientVersion == "1.0.25" {
			foundWithClient = true
			break
		}
	}

	assert.True(t, foundWithClient, "log should have client name and version")
}

// Mock API client for testing
type mockAPIClient struct {
	mu         sync.Mutex
	logs       []LogEntry
	shouldFail bool
	failCount  int
	attempts   int
}

func (m *mockAPIClient) CreateLog(ctx context.Context, entry LogEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.attempts++

	if m.shouldFail {
		// If failCount > 0, decrement and fail
		if m.failCount > 0 {
			m.failCount--
			return fmt.Errorf("API temporarily unavailable")
		}
		// If failCount == 0 and shouldFail is still true, keep failing forever
		return fmt.Errorf("API unavailable")
	}

	m.logs = append(m.logs, entry)
	return nil
}

func (m *mockAPIClient) CreateLogBatch(ctx context.Context, entries []LogEntry) error {
	for _, entry := range entries {
		if err := m.CreateLog(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

// TestLogClassified tests the LogClassified method
func TestLogClassified(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	sessionID := logger.StartSession()
	logger.SetClient("claude-code", "1.0.25")

	// Log a classified entry
	entry := types.ClassifiedLogEntry{
		EntryType:    types.LogTypeUserPrompt,
		Provider:     types.LogProviderAnthropic,
		Model:        "claude-3-opus",
		Content:      "Hello, Claude!",
		TokensInput:  10,
		TokensOutput: 0,
	}

	logger.LogClassified(entry)

	// Force flush
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify classified log was captured
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundClassified bool
	for _, log := range mockAPI.logs {
		if log.EventType == string(types.LogTypeUserPrompt) && log.EventCategory == "classified" {
			foundClassified = true
			assert.Equal(t, sessionID.String(), log.SessionID)
			assert.Equal(t, "claude-code", log.ClientName)
			assert.Equal(t, "Hello, Claude!", log.Content)
			// Check payload
			if log.Payload != nil {
				assert.Equal(t, "anthropic", log.Payload["provider"])
				assert.Equal(t, "claude-3-opus", log.Payload["model"])
			}
			break
		}
	}

	assert.True(t, foundClassified, "should capture classified log entry")
}

// TestLogClassifiedWithToolUse tests classified logging with tool use
func TestLogClassifiedWithToolUse(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()
	logger.SetClient("claude-code", "1.0.25")

	// Log a tool use entry
	entry := types.ClassifiedLogEntry{
		EntryType:  types.LogTypeToolCall,
		Provider:   types.LogProviderAnthropic,
		Model:      "claude-3-opus",
		ToolName:   "read_file",
		ToolID:     "tool_123",
		ToolInput:  map[string]interface{}{"path": "/tmp/test.txt"},
		ToolOutput: "file contents here",
	}

	logger.LogClassified(entry)

	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundTool bool
	for _, log := range mockAPI.logs {
		if log.EventType == string(types.LogTypeToolCall) {
			foundTool = true
			if log.Payload != nil {
				assert.Equal(t, "read_file", log.Payload["tool_name"])
				assert.Equal(t, "tool_123", log.Payload["tool_id"])
			}
			break
		}
	}

	assert.True(t, foundTool, "should capture tool use log entry")
}

// TestGetClassifiedLogs tests the GetClassifiedLogs method
func TestGetClassifiedLogs(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()
	logger.SetClient("claude-code", "1.0.25")

	// Log multiple classified entries
	entries := []types.ClassifiedLogEntry{
		{
			EntryType: types.LogTypeUserPrompt,
			Provider:  types.LogProviderAnthropic,
			Content:   "Message 1",
		},
		{
			EntryType: types.LogTypeAIText,
			Provider:  types.LogProviderAnthropic,
			Content:   "Message 2",
		},
		{
			EntryType: types.LogTypeToolCall,
			Provider:  types.LogProviderAnthropic,
			ToolName:  "bash",
		},
	}

	for _, entry := range entries {
		logger.LogClassified(entry)
	}

	// Get classified logs
	logs := logger.GetClassifiedLogs()

	assert.Equal(t, 3, len(logs), "should have 3 classified logs")

	// Verify they are copies (modifying shouldn't affect original)
	if len(logs) > 0 {
		logs[0].Content = "modified"
		originalLogs := logger.GetClassifiedLogs()
		assert.NotEqual(t, "modified", originalLogs[0].Content, "should return copies")
	}
}

// TestLogClassifiedWithError tests classified logging with error message
func TestLogClassifiedWithError(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession()

	// Log an error entry
	entry := types.ClassifiedLogEntry{
		EntryType:    types.LogTypeToolResult,
		Provider:     types.LogProviderAnthropic,
		ErrorMessage: "Tool execution failed: permission denied",
	}

	logger.LogClassified(entry)

	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundError bool
	for _, log := range mockAPI.logs {
		if log.EventType == string(types.LogTypeToolResult) {
			foundError = true
			assert.Equal(t, "Tool execution failed: permission denied", log.Content)
			break
		}
	}

	assert.True(t, foundError, "should capture error log entry")
}

// TestLogEventWithSessionOverride tests session ID override in metadata
func TestLogEventWithSessionOverride(t *testing.T) {
	config := &Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
		QueueDir:      t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	logger.StartSession() // This creates a session ID
	logger.SetClient("original-client", "1.0.0")

	// Log event with overridden session and client info
	logger.LogEvent("test_event", "test", "content", map[string]interface{}{
		"session_id":     "override-session-123",
		"client_name":    "override-client",
		"client_version": "2.0.0",
	})

	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var found bool
	for _, log := range mockAPI.logs {
		if log.EventType == "test_event" {
			found = true
			assert.Equal(t, "override-session-123", log.SessionID)
			assert.Equal(t, "override-client", log.ClientName)
			assert.Equal(t, "2.0.0", log.ClientVersion)
			break
		}
	}

	assert.True(t, found, "should find event with overridden session")
}
