package control

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// ServiceConfig configures the Control Service.
type ServiceConfig struct {
	// Ownership fields
	EmployeeID string
	OrgID      string
	AgentID    string

	// Queue configuration
	QueueDir      string
	FlushInterval time.Duration
	MaxBatchSize  int

	// Uploader for sending logs to API (optional, can be set later)
	Uploader Uploader
}

// Service is the main Control Service that orchestrates the pipeline.
type Service struct {
	config    ServiceConfig
	sessionID string
	ctx       *HandlerContext
	pipeline  *Pipeline
	queue     *DiskQueue
}

// NewService creates a new Control Service.
func NewService(config ServiceConfig) (*Service, error) {
	// Apply defaults
	if config.FlushInterval == 0 {
		config.FlushInterval = 5 * time.Second
	}
	if config.MaxBatchSize == 0 {
		config.MaxBatchSize = 10
	}

	// Generate session ID
	sessionID := uuid.New().String()

	// Create handler context
	ctx := NewHandlerContext(config.EmployeeID, config.OrgID, sessionID, config.AgentID)

	// Create disk queue
	queueConfig := QueueConfig{
		QueueDir:      config.QueueDir,
		FlushInterval: config.FlushInterval,
		MaxBatchSize:  config.MaxBatchSize,
	}
	queue, err := NewDiskQueue(queueConfig, config.Uploader)
	if err != nil {
		return nil, err
	}

	// Create pipeline
	pipeline := NewPipeline()

	// Register default handlers
	loggerHandler := NewLoggerHandler(queue)
	pipeline.Register(loggerHandler)

	return &Service{
		config:    config,
		sessionID: sessionID,
		ctx:       ctx,
		pipeline:  pipeline,
		queue:     queue,
	}, nil
}

// SessionID returns the unique session ID for this service instance.
func (s *Service) SessionID() string {
	return s.sessionID
}

// Context returns the handler context with ownership fields.
func (s *Service) Context() *HandlerContext {
	return s.ctx
}

// Pipeline returns the handler pipeline for custom handler registration.
func (s *Service) Pipeline() *Pipeline {
	return s.pipeline
}

// RegisterHandler adds a custom handler to the pipeline.
func (s *Service) RegisterHandler(h Handler) {
	s.pipeline.Register(h)
}

// HandleRequest processes an outgoing request through the pipeline.
// Returns the result indicating whether to continue or block.
func (s *Service) HandleRequest(req *http.Request) Result {
	return s.pipeline.ExecuteRequest(s.ctx, req)
}

// HandleResponse processes an incoming response through the pipeline.
// Returns the result indicating whether to continue or block.
func (s *Service) HandleResponse(res *http.Response) Result {
	return s.pipeline.ExecuteResponse(s.ctx, res)
}

// Start starts the background workers (queue uploader).
// Blocks until context is cancelled.
func (s *Service) Start(ctx context.Context) {
	s.queue.StartWorker(ctx)
}

// Stop performs a synchronous flush of all pending log entries.
// Call this before exiting to ensure all logs are uploaded.
func (s *Service) Stop() {
	s.queue.flush()
}

// SetUploader sets the uploader for sending logs to the API.
// Can be called after creation if uploader wasn't available at init time.
func (s *Service) SetUploader(uploader Uploader) {
	// Note: This creates a race condition if called while worker is running.
	// In practice, this should be called before Start().
	s.queue = &DiskQueue{
		config: QueueConfig{
			QueueDir:      s.config.QueueDir,
			FlushInterval: s.config.FlushInterval,
			MaxBatchSize:  s.config.MaxBatchSize,
		},
		uploader: uploader,
	}
}
