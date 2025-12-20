package control

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// LogEntry represents a log entry to be queued and uploaded.
type LogEntry struct {
	// Ownership fields
	EmployeeID string `json:"employee_id"`
	OrgID      string `json:"org_id"`
	SessionID  string `json:"session_id"`
	AgentID    string `json:"agent_id"`

	// Event metadata
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`

	// Content
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// QueuedEntry wraps a LogEntry with queue metadata.
type QueuedEntry struct {
	ID       string   // Filename without extension
	FilePath string   // Full path to file
	Entry    LogEntry // The actual log entry
}

// QueueConfig configures the disk queue behavior.
type QueueConfig struct {
	// QueueDir is the directory to store queued log files.
	QueueDir string

	// FlushInterval is how often to check for pending entries and upload.
	FlushInterval time.Duration

	// MaxBatchSize is the maximum number of entries to upload in one batch.
	MaxBatchSize int
}

// Uploader defines the interface for uploading log entries.
type Uploader interface {
	Upload(entries []LogEntry) error
}

// DiskQueue implements a disk-based queue for log entries.
// Entries are written to individual JSON files and uploaded by a background worker.
type DiskQueue struct {
	config   QueueConfig
	uploader Uploader
	mu       sync.Mutex
}

// NewDiskQueue creates a new disk queue.
// If uploader is nil, entries are queued but not uploaded (useful for testing).
func NewDiskQueue(config QueueConfig, uploader Uploader) (*DiskQueue, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(config.QueueDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create queue directory: %w", err)
	}

	return &DiskQueue{
		config:   config,
		uploader: uploader,
	}, nil
}

// Enqueue writes a log entry to the disk queue.
// This is non-blocking and returns immediately after writing to disk.
func (q *DiskQueue) Enqueue(entry LogEntry) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Generate unique filename based on timestamp
	filename := fmt.Sprintf("%d.json", time.Now().UnixNano())
	filepath := filepath.Join(q.config.QueueDir, filename)

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal entry: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0600); err != nil {
		return fmt.Errorf("failed to write entry: %w", err)
	}

	return nil
}

// Pending returns all queued entries waiting to be uploaded.
func (q *DiskQueue) Pending() ([]QueuedEntry, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	files, err := filepath.Glob(filepath.Join(q.config.QueueDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to list queue files: %w", err)
	}

	// Sort by filename (timestamp order)
	sort.Strings(files)

	var entries []QueuedEntry
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			// Skip files we can't read
			continue
		}

		var entry LogEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			// Skip files we can't parse
			continue
		}

		id := filepath.Base(file)
		id = id[:len(id)-5] // Remove .json extension

		entries = append(entries, QueuedEntry{
			ID:       id,
			FilePath: file,
			Entry:    entry,
		})
	}

	return entries, nil
}

// Remove deletes a queued entry by ID.
func (q *DiskQueue) Remove(id string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	filepath := filepath.Join(q.config.QueueDir, id+".json")
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove entry: %w", err)
	}

	return nil
}

// StartWorker starts the background worker that uploads queued entries.
// The worker runs until the context is cancelled.
func (q *DiskQueue) StartWorker(ctx context.Context) {
	ticker := time.NewTicker(q.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			q.flush()
		}
	}
}

// flush uploads pending entries in batches.
func (q *DiskQueue) flush() {
	if q.uploader == nil {
		return
	}

	pending, err := q.Pending()
	if err != nil || len(pending) == 0 {
		return
	}

	// Process in batches
	for i := 0; i < len(pending); i += q.config.MaxBatchSize {
		end := i + q.config.MaxBatchSize
		if end > len(pending) {
			end = len(pending)
		}

		batch := pending[i:end]
		entries := make([]LogEntry, len(batch))
		for j, qe := range batch {
			entries[j] = qe.Entry
		}

		// Try to upload
		if err := q.uploader.Upload(entries); err != nil {
			// Upload failed, entries stay in queue for retry
			continue
		}

		// Upload succeeded, remove entries
		for _, qe := range batch {
			_ = q.Remove(qe.ID)
		}
	}
}
