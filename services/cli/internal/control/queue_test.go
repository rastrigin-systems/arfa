package control

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDiskQueue(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Second,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)

	require.NoError(t, err)
	require.NotNil(t, q)
}

func TestDiskQueue_Enqueue_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour, // Don't auto-flush
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)

	entry := LogEntry{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
		EventType:  "api_request",
		Timestamp:  time.Now(),
		Payload:    map[string]interface{}{"key": "value"},
	}

	err = q.Enqueue(entry)
	require.NoError(t, err)

	// Check file was created
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 1)

	// Verify content
	data, err := os.ReadFile(files[0])
	require.NoError(t, err)

	var saved LogEntry
	err = json.Unmarshal(data, &saved)
	require.NoError(t, err)
	assert.Equal(t, "emp-123", saved.EmployeeID)
	assert.Equal(t, "org-456", saved.OrgID)
	assert.Equal(t, "api_request", saved.EventType)
}

func TestDiskQueue_Enqueue_MultipleEntries(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		entry := LogEntry{
			EmployeeID: "emp-123",
			OrgID:      "org-456",
			SessionID:  "sess-789",
			AgentID:    "agent-abc",
			EventType:  "api_request",
			Timestamp:  time.Now(),
		}
		err = q.Enqueue(entry)
		require.NoError(t, err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 5)
}

func TestDiskQueue_Pending_ReturnsQueuedEntries(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)

	// Enqueue entries
	for i := 0; i < 3; i++ {
		entry := LogEntry{
			EmployeeID: "emp-123",
			OrgID:      "org-456",
			SessionID:  "sess-789",
			AgentID:    "agent-abc",
			EventType:  "api_request",
			Timestamp:  time.Now(),
		}
		err = q.Enqueue(entry)
		require.NoError(t, err)
	}

	pending, err := q.Pending()
	require.NoError(t, err)
	assert.Len(t, pending, 3)
}

func TestDiskQueue_Remove_DeletesFile(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)

	entry := LogEntry{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
		EventType:  "api_request",
		Timestamp:  time.Now(),
	}
	err = q.Enqueue(entry)
	require.NoError(t, err)

	pending, err := q.Pending()
	require.NoError(t, err)
	require.Len(t, pending, 1)

	// Remove entry
	err = q.Remove(pending[0].ID)
	require.NoError(t, err)

	// Verify file is gone
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestDiskQueue_BackgroundWorker_UploadsEntries(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: 50 * time.Millisecond,
		MaxBatchSize:  10,
	}

	var uploadedCount int32
	uploader := &mockUploader{
		uploadFunc: func(entries []LogEntry) error {
			atomic.AddInt32(&uploadedCount, int32(len(entries)))
			return nil
		},
	}

	q, err := NewDiskQueue(config, uploader)
	require.NoError(t, err)

	// Enqueue entries
	for i := 0; i < 3; i++ {
		entry := LogEntry{
			EmployeeID: "emp-123",
			OrgID:      "org-456",
			SessionID:  "sess-789",
			AgentID:    "agent-abc",
			EventType:  "api_request",
			Timestamp:  time.Now(),
		}
		err = q.Enqueue(entry)
		require.NoError(t, err)
	}

	// Start worker
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go q.StartWorker(ctx)

	// Wait for upload
	time.Sleep(200 * time.Millisecond)

	assert.Equal(t, int32(3), atomic.LoadInt32(&uploadedCount))

	// Files should be removed after successful upload
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestDiskQueue_BackgroundWorker_RetriesOnError(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: 50 * time.Millisecond,
		MaxBatchSize:  10,
	}

	var attempts int32
	uploader := &mockUploader{
		uploadFunc: func(entries []LogEntry) error {
			atomic.AddInt32(&attempts, 1)
			if atomic.LoadInt32(&attempts) < 2 {
				return assert.AnError
			}
			return nil
		},
	}

	q, err := NewDiskQueue(config, uploader)
	require.NoError(t, err)

	entry := LogEntry{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
		EventType:  "api_request",
		Timestamp:  time.Now(),
	}
	err = q.Enqueue(entry)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go q.StartWorker(ctx)

	time.Sleep(300 * time.Millisecond)

	// Should have retried at least once
	assert.GreaterOrEqual(t, atomic.LoadInt32(&attempts), int32(2))

	// File should be removed after successful retry
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestDiskQueue_BatchSize_LimitedToConfig(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: 50 * time.Millisecond,
		MaxBatchSize:  2, // Only 2 at a time
	}

	var mu sync.Mutex
	var batchSizes []int
	uploader := &mockUploader{
		uploadFunc: func(entries []LogEntry) error {
			mu.Lock()
			batchSizes = append(batchSizes, len(entries))
			mu.Unlock()
			return nil
		},
	}

	q, err := NewDiskQueue(config, uploader)
	require.NoError(t, err)

	// Enqueue 5 entries
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			EmployeeID: "emp-123",
			OrgID:      "org-456",
			SessionID:  "sess-789",
			AgentID:    "agent-abc",
			EventType:  "api_request",
			Timestamp:  time.Now(),
		}
		err = q.Enqueue(entry)
		require.NoError(t, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go q.StartWorker(ctx)

	// Wait for worker to finish
	<-ctx.Done()
	time.Sleep(50 * time.Millisecond) // Allow final flush to complete

	// Should have uploaded in batches of 2 or less
	mu.Lock()
	sizes := make([]int, len(batchSizes))
	copy(sizes, batchSizes)
	mu.Unlock()

	for _, size := range sizes {
		assert.LessOrEqual(t, size, 2)
	}
}

func TestDiskQueue_CreatesDirectoryIfMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "queue")
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)
	require.NotNil(t, q)

	// Directory should exist
	_, err = os.Stat(dir)
	require.NoError(t, err)
}

func TestDiskQueue_Stop_GracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	config := QueueConfig{
		QueueDir:      dir,
		FlushInterval: time.Hour,
		MaxBatchSize:  10,
	}

	q, err := NewDiskQueue(config, nil)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		q.StartWorker(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Worker stopped
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
}

// mockUploader implements Uploader for testing
type mockUploader struct {
	uploadFunc func(entries []LogEntry) error
}

func (m *mockUploader) Upload(entries []LogEntry) error {
	if m.uploadFunc != nil {
		return m.uploadFunc(entries)
	}
	return nil
}
