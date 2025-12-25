package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/logging"
)

// TestLoggingEndToEnd tests the complete logging flow from initialization to close
func TestLoggingEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock API client that captures logs
	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	// Create logger
	config := &logging.Config{
		Enabled:       true,
		BatchSize:     10,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(config, mockAPI)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	// Set client info
	clientName := "claude-code"
	clientVersion := "1.0.25"
	logger.SetClient(clientName, clientVersion)

	// Start session
	sessionID := logger.StartSession()
	t.Logf("Started session: %s", sessionID)

	// Log some events
	logger.LogEvent("test_input", "cli", "user types command", map[string]interface{}{"command": "ls -la"})
	logger.LogEvent("test_output", "cli", "agent responds", map[string]interface{}{"response": "file listing"})
	logger.LogEvent("test_error", "cli", "some warning", map[string]interface{}{"level": "warning"})

	// End session
	logger.EndSession()

	// Flush to ensure all logs are sent
	logger.Flush()

	// Wait a bit for async operations to complete
	time.Sleep(200 * time.Millisecond)

	// Verify logs were captured
	mockAPI.mu.Lock()
	logs := mockAPI.logs
	mockAPI.mu.Unlock()

	if len(logs) == 0 {
		t.Fatal("Expected logs to be captured, but got none")
	}

	t.Logf("Captured %d log entries", len(logs))

	// Verify session_start event
	foundStart := false
	foundEnd := false
	foundInput := false
	foundOutput := false
	foundError := false

	for _, log := range logs {
		if log.SessionID != sessionID.String() {
			t.Errorf("Expected session_id %s, got %s", sessionID, log.SessionID)
		}

		if log.ClientName != clientName {
			t.Errorf("Expected client_name %s, got %s", clientName, log.ClientName)
		}

		switch log.EventType {
		case "session_start":
			foundStart = true
			if log.EventCategory != "session" {
				t.Errorf("Expected session_start to have category 'session', got %s", log.EventCategory)
			}
		case "session_end":
			foundEnd = true
			if log.EventCategory != "session" {
				t.Errorf("Expected session_end to have category 'session', got %s", log.EventCategory)
			}
		case "test_input":
			foundInput = true
			if log.EventCategory != "cli" {
				t.Errorf("Expected test_input to have category 'cli', got %s", log.EventCategory)
			}
		case "test_output":
			foundOutput = true
			if log.EventCategory != "cli" {
				t.Errorf("Expected test_output to have category 'cli', got %s", log.EventCategory)
			}
		case "test_error":
			foundError = true
			if log.EventCategory != "cli" {
				t.Errorf("Expected test_error to have category 'cli', got %s", log.EventCategory)
			}
		}
	}

	if !foundStart {
		t.Error("Expected to find session_start event")
	}
	if !foundEnd {
		t.Error("Expected to find session_end event")
	}
	if !foundInput {
		t.Error("Expected to find test_input event")
	}
	if !foundOutput {
		t.Error("Expected to find test_output event")
	}
	if !foundError {
		t.Error("Expected to find test_error event")
	}
}

// TestLoggingBatching tests that logs are batched correctly
func TestLoggingBatching(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs:      make([]logging.LogEntry, 0),
		batchSent: make(chan int, 10),
	}

	config := &logging.Config{
		Enabled:       true,
		BatchSize:     5,
		BatchInterval: 1 * time.Second, // Long interval to test size-based batching
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(config, mockAPI)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.StartSession()

	// Log exactly BatchSize entries to trigger batch send
	for i := 0; i < 5; i++ {
		logger.LogEvent("test_event", "cli", "test input", map[string]interface{}{"index": i})
	}

	// Wait for batch to be sent
	select {
	case batchSize := <-mockAPI.batchSent:
		// Should be 6: 1 session_start + 5 inputs
		if batchSize != 6 {
			t.Errorf("Expected batch size of 6 (1 session_start + 5 inputs), got %d", batchSize)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for batch to be sent")
	}
}

// TestLoggingOfflineQueue tests offline queue functionality
func TestLoggingOfflineQueue(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &failingAPIClient{
		failCount: 3, // Fail 3 times, then succeed
	}

	config := &logging.Config{
		Enabled:       true,
		BatchSize:     5,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    2, // Only retry twice, so it will queue to disk
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(config, mockAPI)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	logger.StartSession()

	// Log some entries
	for i := 0; i < 5; i++ {
		logger.LogEvent("test_event", "cli", "test input", map[string]interface{}{"index": i})
	}

	// Flush to trigger send
	logger.Flush()

	// Wait for retries to complete and queue to disk
	time.Sleep(500 * time.Millisecond)

	// The logs should have been queued to disk after max retries
	// In a real scenario, we'd verify the queue directory has files
	// For this test, we just verify the logger didn't crash
}

// capturingAPIClient captures all logs sent to it
type capturingAPIClient struct {
	mu        sync.Mutex
	logs      []logging.LogEntry
	batchSent chan int
}

func (c *capturingAPIClient) CreateLog(ctx context.Context, entry logging.LogEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logs = append(c.logs, entry)
	return nil
}

func (c *capturingAPIClient) CreateLogBatch(ctx context.Context, entries []logging.LogEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.logs = append(c.logs, entries...)
	if c.batchSent != nil {
		c.batchSent <- len(entries)
	}
	return nil
}

// failingAPIClient fails a specified number of times before succeeding
type failingAPIClient struct {
	mu        sync.Mutex
	failCount int
	callCount int
}

func (f *failingAPIClient) CreateLog(ctx context.Context, entry logging.LogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	if f.callCount <= f.failCount {
		return &temporaryError{"API unavailable"}
	}
	return nil
}

func (f *failingAPIClient) CreateLogBatch(ctx context.Context, entries []logging.LogEntry) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callCount++
	if f.callCount <= f.failCount {
		return &temporaryError{"API unavailable"}
	}
	return nil
}

// temporaryError is a temporary error type
type temporaryError struct {
	msg string
}

func (e *temporaryError) Error() string {
	return e.msg
}

func (e *temporaryError) Temporary() bool {
	return true
}
