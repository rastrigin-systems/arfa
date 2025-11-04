package logging

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
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
				Enabled:    tt.enabled,
				BatchSize:  100,
				BatchInterval: 5 * time.Second,
				QueueDir:   t.TempDir(),
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

// TestLoggerDisabledViaEnv tests UBIK_NO_LOGGING environment variable
func TestLoggerDisabledViaEnv(t *testing.T) {
	// Set environment variable
	os.Setenv("UBIK_NO_LOGGING", "1")
	defer os.Unsetenv("UBIK_NO_LOGGING")

	config := &Config{
		Enabled:    true, // explicitly enabled in config
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
	}

	logger, err := NewLogger(config, nil)
	assert.Nil(t, logger, "logger should be nil when UBIK_NO_LOGGING=1")
	assert.NoError(t, err)
}

// TestSessionTracking tests session start/end events
func TestSessionTracking(t *testing.T) {
	config := &Config{
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
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

// TestStdinCapture tests capturing stdin
func TestStdinCapture(t *testing.T) {
	config := &Config{
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	sessionID := logger.StartSession()

	// Create stdin interceptor
	originalStdin := strings.NewReader("test input\nanother line\n")
	interceptedStdin := logger.InterceptStdin(originalStdin)

	// Read from intercepted stream
	buf := make([]byte, 1024)
	n, err := interceptedStdin.Read(buf)
	require.NoError(t, err)
	assert.Greater(t, n, 0)

	// Force flush
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify logs captured
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundInput bool
	for _, log := range mockAPI.logs {
		if log.EventType == "input" {
			foundInput = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}

	assert.True(t, foundInput, "should capture stdin")
}

// TestIOCapture tests capturing stdin/stdout/stderr
func TestIOCapture(t *testing.T) {
	config := &Config{
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer logger.Close()

	sessionID := logger.StartSession()

	// Create intercepted writers
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	interceptedStdout := logger.InterceptStdout(stdout)
	interceptedStderr := logger.InterceptStderr(stderr)

	// Write to intercepted streams (with newlines to trigger logging)
	testOutput := "test output message"
	testError := "test error message"

	interceptedStdout.Write([]byte(testOutput + "\n"))
	interceptedStderr.Write([]byte(testError + "\n"))

	// Verify original streams received data
	assert.Contains(t, stdout.String(), testOutput)
	assert.Contains(t, stderr.String(), testError)

	// Force flush
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify logs captured
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	var foundOutput, foundError bool
	for _, log := range mockAPI.logs {
		if log.EventType == "output" && strings.Contains(log.Content, testOutput) {
			foundOutput = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
		if log.EventType == "error" && strings.Contains(log.Content, testError) {
			foundError = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}

	assert.True(t, foundOutput, "should capture stdout")
	assert.True(t, foundError, "should capture stderr")
}

// TestBatchSending tests batching behavior
func TestBatchSending(t *testing.T) {
	config := &Config{
		Enabled:    true,
		BatchSize:  5, // small batch for testing
		BatchInterval: 10 * time.Second, // long interval, we'll hit size limit first
		QueueDir:   t.TempDir(),
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
		logger.LogInput("test input", map[string]interface{}{"index": i})
	}

	// Wait for batch to be sent
	time.Sleep(300 * time.Millisecond)

	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	// Should have session_start + 5 inputs = 6 total
	assert.GreaterOrEqual(t, logCount, 5, "should have sent batch")

	// Verify session ID in all logs
	mockAPI.mu.Lock()
	for _, log := range mockAPI.logs {
		if log.EventType == "input" {
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}
	mockAPI.mu.Unlock()
}

// TestOfflineQueue tests queuing when API is unavailable
func TestOfflineQueue(t *testing.T) {
	queueDir := t.TempDir()

	config := &Config{
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   queueDir,
		MaxRetries: 2, // Fast retries for testing
		RetryBackoff: 50 * time.Millisecond,
	}

	// Create API client that always fails
	mockAPI := &mockAPIClient{
		logs:      make([]LogEntry, 0),
		shouldFail: true,
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)

	_ = logger.StartSession() // session ID
	logger.LogInput("test input", nil)
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
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
		MaxRetries: 3,
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
	logger.LogInput("test retry", nil)

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
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
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
				logger.LogInput("concurrent test", map[string]interface{}{
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

	// Should have session_start + 50 inputs
	assert.GreaterOrEqual(t, logCount, 50, "should handle concurrent logging")
}

// TestLoggerClose tests cleanup on close
func TestLoggerClose(t *testing.T) {
	config := &Config{
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   t.TempDir(),
	}

	mockAPI := &mockAPIClient{
		logs: make([]LogEntry, 0),
	}

	logger, err := NewLogger(config, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.StartSession()
	logger.LogInput("test before close", nil)

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
		Enabled:    true,
		BatchSize:  100,
		BatchInterval: 5 * time.Second,
		QueueDir:   queueDir,
		MaxRetries: 2, // Fast retries for testing
		RetryBackoff: 50 * time.Millisecond,
	}

	// First logger with failing API
	mockAPI1 := &mockAPIClient{
		logs:      make([]LogEntry, 0),
		shouldFail: true,
	}

	logger1, err := NewLogger(config, mockAPI1)
	require.NoError(t, err)
	require.NotNil(t, logger1)

	logger1.StartSession()
	logger1.LogInput("persisted message", nil)
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
		logs:      make([]LogEntry, 0),
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
		if log.EventType == "input" && strings.Contains(log.Content, "persisted message") {
			foundPersisted = true
			break
		}
	}

	assert.True(t, foundPersisted, "persisted log should be sent after restart")
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
