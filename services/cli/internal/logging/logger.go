package logging

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/types"
)

// loggerImpl implements the Logger interface
type loggerImpl struct {
	config        *Config
	api           APIClient
	sessionID     uuid.UUID
	clientName    string
	clientVersion string
	startTime     time.Time
	buffer        []LogEntry
	bufferMu      sync.Mutex
	done          chan struct{}
	wg            sync.WaitGroup
	queueDir      string
	ctx           context.Context
	cancel        context.CancelFunc

	// Classified logs (parsed API requests/responses)
	classifiedLogs   []types.ClassifiedLogEntry
	classifiedLogsMu sync.RWMutex
}

// NewLogger creates a new logger instance
func NewLogger(config *Config, api APIClient) (Logger, error) {
	// Check for opt-out via environment variable
	if os.Getenv("UBIK_NO_LOGGING") != "" {
		return nil, nil
	}

	// Check if logging is disabled in config
	if !config.Enabled {
		return nil, nil
	}

	// Ensure queue directory exists
	if config.QueueDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		config.QueueDir = filepath.Join(home, ".ubik", "log_queue")
	}

	if err := os.MkdirAll(config.QueueDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}

	// Set defaults
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.BatchInterval <= 0 {
		config.BatchInterval = 5 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 5
	}
	if config.RetryBackoff <= 0 {
		config.RetryBackoff = 1 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	logger := &loggerImpl{
		config:   config,
		api:      api,
		buffer:   make([]LogEntry, 0, config.BatchSize),
		done:     make(chan struct{}),
		queueDir: config.QueueDir,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Start background workers
	logger.wg.Add(2)
	go logger.batchSender()
	go logger.queueProcessor()

	return logger, nil
}

// StartSession begins a new logging session
func (l *loggerImpl) StartSession() uuid.UUID {
	l.sessionID = uuid.New()
	l.startTime = time.Now()

	// Log session_start event
	l.LogEvent("session_start", "session", "", map[string]interface{}{
		"start_time": l.startTime,
	})

	return l.sessionID
}

// SetClient sets the client name and version for all subsequent log entries
func (l *loggerImpl) SetClient(clientName, clientVersion string) {
	l.clientName = clientName
	l.clientVersion = clientVersion
}

// EndSession marks the end of the current session
func (l *loggerImpl) EndSession() {
	duration := time.Since(l.startTime)

	l.LogEvent("session_end", "session", "", map[string]interface{}{
		"end_time":         time.Now(),
		"duration_seconds": duration.Seconds(),
	})
}

// LogEvent logs a custom event
func (l *loggerImpl) LogEvent(eventType, category, content string, metadata map[string]interface{}) {
	// Use session_id from metadata if provided (e.g., from proxy requests)
	// This allows the daemon logger to log events with the correct session ID
	sessionID := l.sessionID.String()
	if metadata != nil {
		if sid, ok := metadata["session_id"].(string); ok && sid != "" {
			sessionID = sid
		}
	}

	// Use client info from metadata if provided
	clientName := l.clientName
	clientVersion := l.clientVersion
	if metadata != nil {
		if cn, ok := metadata["client_name"].(string); ok && cn != "" {
			clientName = cn
		}
		if cv, ok := metadata["client_version"].(string); ok && cv != "" {
			clientVersion = cv
		}
	}

	entry := LogEntry{
		SessionID:     sessionID,
		ClientName:    clientName,
		ClientVersion: clientVersion,
		EventType:     eventType,
		EventCategory: category,
		Content:       content,
		Payload:       metadata,
		Timestamp:     time.Now(),
	}

	l.bufferMu.Lock()
	l.buffer = append(l.buffer, entry)
	shouldFlush := len(l.buffer) >= l.config.BatchSize
	l.bufferMu.Unlock()

	// Trigger immediate flush if batch size reached
	if shouldFlush {
		go l.flushBuffer()
	}
}

// LogClassified logs a classified log entry (parsed from API requests/responses)
func (l *loggerImpl) LogClassified(entry types.ClassifiedLogEntry) {
	// Ensure entry has session ID and timestamp
	if entry.SessionID == "" {
		entry.SessionID = l.sessionID.String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	// Store in memory for current session display
	l.classifiedLogsMu.Lock()
	l.classifiedLogs = append(l.classifiedLogs, entry)
	l.classifiedLogsMu.Unlock()

	// Send to API for persistence (primary storage)
	payload := map[string]interface{}{
		"entry_type":     string(entry.EntryType),
		"provider":       string(entry.Provider),
		"model":          entry.Model,
		"tokens_input":   entry.TokensInput,
		"tokens_output":  entry.TokensOutput,
		"session_id":     entry.SessionID,     // Pass session_id to LogEvent for proper tracking
		"client_name":    entry.ClientName,    // Pass client info to LogEvent for proper tracking
		"client_version": entry.ClientVersion, // Pass client info to LogEvent for proper tracking
	}

	if entry.ToolName != "" {
		payload["tool_name"] = entry.ToolName
		payload["tool_id"] = entry.ToolID
		payload["tool_input"] = entry.ToolInput
	}

	content := entry.Content
	if content == "" && entry.ToolOutput != "" {
		content = entry.ToolOutput
	}
	if content == "" && entry.ErrorMessage != "" {
		content = entry.ErrorMessage
	}

	l.LogEvent(string(entry.EntryType), "classified", content, payload)
}

// GetClassifiedLogs returns classified logs for the current session
func (l *loggerImpl) GetClassifiedLogs() []types.ClassifiedLogEntry {
	l.classifiedLogsMu.RLock()
	defer l.classifiedLogsMu.RUnlock()

	// Return a copy to prevent concurrent modification
	logs := make([]types.ClassifiedLogEntry, len(l.classifiedLogs))
	copy(logs, l.classifiedLogs)
	return logs
}

// Flush forces immediate sending of buffered logs
func (l *loggerImpl) Flush() {
	l.flushBuffer()
}

// Close shuts down the logger and flushes remaining logs
func (l *loggerImpl) Close() error {
	// Flush any remaining logs
	l.flushBuffer()

	// Signal shutdown
	close(l.done)
	l.cancel()

	// Wait for background workers to finish
	l.wg.Wait()

	return nil
}

// batchSender periodically sends batched logs
func (l *loggerImpl) batchSender() {
	defer l.wg.Done()

	ticker := time.NewTicker(l.config.BatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.flushBuffer()
		case <-l.done:
			return
		}
	}
}

// flushBuffer sends buffered logs to the API
func (l *loggerImpl) flushBuffer() {
	l.bufferMu.Lock()
	if len(l.buffer) == 0 {
		l.bufferMu.Unlock()
		return
	}

	// Take current buffer and create new one
	toSend := l.buffer
	l.buffer = make([]LogEntry, 0, l.config.BatchSize)
	l.bufferMu.Unlock()

	// Try to send
	if err := l.sendWithRetry(toSend); err != nil {
		// If send fails, queue to disk
		if err := l.queueToDisk(toSend); err != nil {
			// Log error but don't crash
			fmt.Fprintf(os.Stderr, "Failed to queue logs: %v\n", err)
		}
	}
}

// sendWithRetry sends logs with exponential backoff retry
func (l *loggerImpl) sendWithRetry(entries []LogEntry) error {
	backoff := l.config.RetryBackoff

	for attempt := 0; attempt <= l.config.MaxRetries; attempt++ {
		err := l.api.CreateLogBatch(l.ctx, entries)
		if err == nil {
			return nil
		}

		// Don't retry on last attempt
		if attempt == l.config.MaxRetries {
			return fmt.Errorf("max retries exceeded: %w", err)
		}

		// Exponential backoff: 1s, 2s, 4s, 8s, 16s
		select {
		case <-time.After(backoff):
			backoff *= 2
			if backoff > 16*time.Second {
				backoff = 16 * time.Second
			}
		case <-l.ctx.Done():
			return l.ctx.Err()
		}
	}

	return fmt.Errorf("failed to send logs after retries")
}

// queueToDisk saves logs to disk for later processing
func (l *loggerImpl) queueToDisk(entries []LogEntry) error {
	filename := filepath.Join(l.queueDir, fmt.Sprintf("logs_%d.json", time.Now().UnixNano()))

	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write queue file: %w", err)
	}

	return nil
}

// queueProcessor processes queued logs from disk
func (l *loggerImpl) queueProcessor() {
	defer l.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.processQueue()
		case <-l.done:
			// Process remaining queue on shutdown
			l.processQueue()
			return
		}
	}
}

// processQueue processes all queued log files
func (l *loggerImpl) processQueue() {
	files, err := os.ReadDir(l.queueDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			l.processQueueFile(filepath.Join(l.queueDir, file.Name()))
		}
	}
}

// processQueueFile processes a single queued log file
func (l *loggerImpl) processQueueFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	var entries []LogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		// Invalid file, delete it
		os.Remove(filename)
		return
	}

	// Try to send
	if err := l.sendWithRetry(entries); err != nil {
		// Leave file for next attempt
		return
	}

	// Success, delete file
	os.Remove(filename)
}
