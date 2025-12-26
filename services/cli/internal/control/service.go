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

	// Queue configuration
	QueueDir      string
	FlushInterval time.Duration
	MaxBatchSize  int

	// Uploader for sending logs to API (optional, can be set later)
	Uploader Uploader

	// PolicyClient configuration (optional, for real-time policy updates)
	APIURL string // API base URL for WebSocket connection
	Token  string // JWT token for authentication
}

// Service is the main Control Service that orchestrates the pipeline.
type Service struct {
	config        ServiceConfig
	sessionID     string
	ctx           *HandlerContext
	pipeline      *Pipeline
	queue         *DiskQueue
	policyClient  *PolicyClient
	policyHandler *PolicyHandler
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
	ctx := NewHandlerContext(config.EmployeeID, config.OrgID, sessionID)

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

	// Register client detector handler FIRST (highest priority)
	// This detects the AI client from User-Agent headers before other handlers run
	clientDetector := NewClientDetectorHandler()
	pipeline.Register(clientDetector)

	// Register default handlers
	loggerHandler := NewLoggerHandler(queue)
	pipeline.Register(loggerHandler)

	// Register policy handler (loads policies from ~/.arfa/policies.json by default)
	policyHandler := NewPolicyHandler()
	policyHandler.SetQueue(queue) // Enable logging of blocked tools
	pipeline.Register(policyHandler)

	// Register tool call logger (extracts and logs tool_use events)
	toolCallLogger := NewToolCallLoggerHandler(queue)
	pipeline.Register(toolCallLogger)

	return &Service{
		config:        config,
		sessionID:     sessionID,
		ctx:           ctx,
		pipeline:      pipeline,
		queue:         queue,
		policyHandler: policyHandler,
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

// EnablePolicyBlocking registers a PolicyHandler with the given deny list.
// This enables blocking of specific tools based on organization policies.
// Example: denyList := map[string]string{"Bash": "Shell commands blocked"}
func (s *Service) EnablePolicyBlocking(denyList map[string]string) {
	handler := NewPolicyHandlerWithDenyList(denyList)
	s.pipeline.Register(handler)
}

// EnableRealtimePolicies connects to the API WebSocket for real-time policy updates.
// This replaces file-based policy loading with live updates from the server.
// Call this before Start() to enable real-time policy enforcement.
func (s *Service) EnableRealtimePolicies(ctx context.Context, apiURL, token string) error {
	clientConfig := PolicyClientConfig{
		APIURL:           apiURL,
		Token:            token,
		GracePeriod:      5 * time.Minute,
		ReconnectBackoff: 1 * time.Second,
		MaxReconnectWait: 30 * time.Second,
	}

	s.policyClient = NewPolicyClient(clientConfig)
	s.policyHandler.SetPolicyClient(s.policyClient)

	// Start connection with retry in background
	go s.policyClient.ConnectWithRetry(ctx)

	return nil
}

// WaitForPolicies waits until initial policies are received or timeout.
// Returns error if timeout expires before policies are loaded.
func (s *Service) WaitForPolicies(ctx context.Context, timeout time.Duration) error {
	if s.policyClient == nil {
		return nil // Not using real-time policies
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return s.policyClient.WaitReady(ctx)
}

// PolicyClient returns the policy client (for status checks).
func (s *Service) PolicyClient() *PolicyClient {
	return s.policyClient
}

// ShouldBlockAllRequests returns true if all requests should be blocked.
// This happens when policy client is disconnected past grace period or revoked.
func (s *Service) ShouldBlockAllRequests() (string, bool) {
	return s.policyHandler.ShouldBlockAll()
}
