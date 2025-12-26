package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rastrigin-systems/arfa/generated/db"
)

// WebhookForwarder processes activity logs and forwards them to webhook destinations
type WebhookForwarder struct {
	db         db.Querier
	httpClient *http.Client
	batchSize  int32
}

// WebhookPayload is the payload sent to webhook destinations
type WebhookPayload struct {
	ID             uuid.UUID              `json:"id"`
	EventType      string                 `json:"event_type"`
	EventCategory  string                 `json:"event_category"`
	Timestamp      time.Time              `json:"timestamp"`
	OrgID          uuid.UUID              `json:"org_id"`
	EmployeeID     *uuid.UUID             `json:"employee_id,omitempty"`
	ProxySessionID *uuid.UUID             `json:"proxy_session_id,omitempty"`
	ClientName     string                 `json:"client_name,omitempty"`
	ClientVersion  string                 `json:"client_version,omitempty"`
	Content        string                 `json:"content"`
	Payload        map[string]interface{} `json:"payload,omitempty"`
}

// NewWebhookForwarder creates a new webhook forwarder
func NewWebhookForwarder(database db.Querier) *WebhookForwarder {
	return &WebhookForwarder{
		db: database,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		batchSize: 100, // Process up to 100 logs per destination per cycle
	}
}

// StartForwarderWorker starts a background worker that processes webhook deliveries
func (wf *WebhookForwarder) StartForwarderWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Println("Webhook forwarder started")

	for {
		select {
		case <-ticker.C:
			if err := wf.ProcessDeliveries(ctx); err != nil {
				log.Printf("Error processing webhook deliveries: %v", err)
			}
		case <-ctx.Done():
			log.Println("Webhook forwarder stopped")
			return
		}
	}
}

// ProcessDeliveries processes all pending webhook deliveries
func (wf *WebhookForwarder) ProcessDeliveries(ctx context.Context) error {
	// Step 1: Get all enabled webhook destinations
	destinations, err := wf.db.ListEnabledDestinations(ctx)
	if err != nil {
		return fmt.Errorf("failed to list enabled destinations: %w", err)
	}

	if len(destinations) == 0 {
		return nil // No destinations to process
	}

	// Step 2: For each destination, find undelivered logs and create delivery records
	for _, dest := range destinations {
		if err := wf.processDestination(ctx, dest); err != nil {
			log.Printf("Error processing destination %s: %v", dest.Name, err)
			// Continue with other destinations
		}
	}

	// Step 3: Process pending deliveries
	return wf.processPendingDeliveries(ctx)
}

// processDestination finds undelivered logs for a destination and creates delivery records
func (wf *WebhookForwarder) processDestination(ctx context.Context, dest db.ListEnabledDestinationsRow) error {
	// Find logs that haven't been delivered to this destination yet
	// Look back 24 hours for undelivered logs
	since := time.Now().Add(-24 * time.Hour)

	undeliveredLogs, err := wf.db.GetUndeliveredLogs(ctx, db.GetUndeliveredLogsParams{
		DestinationID: dest.ID,
		OrgID:         dest.OrgID,
		CreatedAt:     pgtype.Timestamp{Time: since, Valid: true},
		Limit:         wf.batchSize,
	})
	if err != nil {
		return fmt.Errorf("failed to get undelivered logs: %w", err)
	}

	if len(undeliveredLogs) == 0 {
		return nil
	}

	// Create delivery records for each log
	for _, logID := range undeliveredLogs {
		// Check if log matches event type filter
		logEntry, err := wf.db.GetActivityLog(ctx, logID)
		if err != nil {
			log.Printf("Failed to get log %s: %v", logID, err)
			continue
		}

		if !wf.matchesEventFilter(logEntry.EventType, dest.EventTypes) {
			continue
		}

		// Create delivery record
		_, err = wf.db.CreateWebhookDelivery(ctx, db.CreateWebhookDeliveryParams{
			DestinationID: dest.ID,
			LogID:         logID,
		})
		if err != nil {
			log.Printf("Failed to create delivery for log %s: %v", logID, err)
		}
	}

	return nil
}

// matchesEventFilter checks if the log event type matches the destination's event filter
func (wf *WebhookForwarder) matchesEventFilter(eventType string, eventTypes []string) bool {
	// Empty filter means match all
	if len(eventTypes) == 0 {
		return true
	}

	for _, et := range eventTypes {
		if et == eventType || et == "*" {
			return true
		}
	}
	return false
}

// processPendingDeliveries processes all pending deliveries
func (wf *WebhookForwarder) processPendingDeliveries(ctx context.Context) error {
	// Get pending deliveries that are ready for processing
	deliveries, err := wf.db.GetPendingDeliveries(ctx, wf.batchSize)
	if err != nil {
		return fmt.Errorf("failed to get pending deliveries: %w", err)
	}

	for _, delivery := range deliveries {
		if err := wf.processDelivery(ctx, delivery); err != nil {
			log.Printf("Error processing delivery %s: %v", delivery.ID, err)
		}
	}

	return nil
}

// processDelivery sends a single webhook delivery
func (wf *WebhookForwarder) processDelivery(ctx context.Context, delivery db.GetPendingDeliveriesRow) error {
	// Get the log entry
	logEntry, err := wf.db.GetActivityLog(ctx, delivery.LogID)
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	// Get the destination (need full details including auth)
	dest, err := wf.db.GetWebhookDestination(ctx, db.GetWebhookDestinationParams{
		ID:    delivery.DestinationID,
		OrgID: logEntry.OrgID,
	})
	if err != nil {
		return fmt.Errorf("failed to get destination: %w", err)
	}

	// Build the payload
	payload := wf.buildPayload(logEntry)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", dest.Url, bytes.NewReader(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Arfa-Webhook/1.0")
	req.Header.Set("X-Arfa-Event-Type", logEntry.EventType)
	req.Header.Set("X-Arfa-Delivery-ID", delivery.ID.String())

	// Add HMAC signature if signing secret is configured
	if dest.SigningSecret != nil && *dest.SigningSecret != "" {
		signature := wf.computeSignature(payloadBytes, *dest.SigningSecret)
		req.Header.Set("X-Arfa-Signature", signature)
	}

	// Add authentication
	wf.addAuth(req, dest.AuthType, dest.AuthConfig)

	// Set timeout from destination config
	client := &http.Client{
		Timeout: time.Duration(dest.TimeoutMs) * time.Millisecond,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		// Mark as failed
		errMsg := err.Error()
		nextRetry := time.Now().Add(time.Duration(dest.RetryBackoffMs) * time.Millisecond * time.Duration(1<<delivery.Attempts))
		_ = wf.db.MarkDeliveryFailed(ctx, db.MarkDeliveryFailedParams{
			ID:             delivery.ID,
			ResponseStatus: nil,
			ResponseBody:   nil,
			ErrorMessage:   &errMsg,
			Attempts:       dest.RetryMax,
			NextRetryAt:    pgtype.Timestamp{Time: nextRetry, Valid: true},
		})
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body (limited to 1KB)
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	bodyStr := string(bodyBytes)
	statusCode := int32(resp.StatusCode)

	// Check if successful (2xx status code)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_ = wf.db.MarkDeliverySuccess(ctx, db.MarkDeliverySuccessParams{
			ID:             delivery.ID,
			ResponseStatus: &statusCode,
			ResponseBody:   &bodyStr,
		})
		return nil
	}

	// Mark as failed for non-2xx responses
	errMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, bodyStr)
	nextRetry := time.Now().Add(time.Duration(dest.RetryBackoffMs) * time.Millisecond * time.Duration(1<<delivery.Attempts))
	_ = wf.db.MarkDeliveryFailed(ctx, db.MarkDeliveryFailedParams{
		ID:             delivery.ID,
		ResponseStatus: &statusCode,
		ResponseBody:   &bodyStr,
		ErrorMessage:   &errMsg,
		Attempts:       dest.RetryMax,
		NextRetryAt:    pgtype.Timestamp{Time: nextRetry, Valid: true},
	})

	return fmt.Errorf("delivery failed: HTTP %d", resp.StatusCode)
}

// buildPayload creates the webhook payload from a log entry
func (wf *WebhookForwarder) buildPayload(logEntry db.ActivityLog) WebhookPayload {
	var content string
	if logEntry.Content != nil {
		content = *logEntry.Content
	}

	var timestamp time.Time
	if logEntry.CreatedAt.Valid {
		timestamp = logEntry.CreatedAt.Time
	} else {
		timestamp = time.Now()
	}

	payload := WebhookPayload{
		ID:            logEntry.ID,
		EventType:     logEntry.EventType,
		EventCategory: logEntry.EventCategory,
		Timestamp:     timestamp,
		OrgID:         logEntry.OrgID,
		Content:       content,
	}

	if logEntry.EmployeeID.Valid {
		id := logEntry.EmployeeID.Bytes
		payload.EmployeeID = (*uuid.UUID)(&id)
	}
	if logEntry.ProxySessionID.Valid {
		id := logEntry.ProxySessionID.Bytes
		payload.ProxySessionID = (*uuid.UUID)(&id)
	}
	if logEntry.ClientName != nil {
		payload.ClientName = *logEntry.ClientName
	}
	if logEntry.ClientVersion != nil {
		payload.ClientVersion = *logEntry.ClientVersion
	}

	// Parse JSON payload if present
	if len(logEntry.Payload) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(logEntry.Payload, &data); err == nil {
			payload.Payload = data
		}
	}

	return payload
}

// computeSignature computes HMAC-SHA256 signature for the payload
func (wf *WebhookForwarder) computeSignature(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// addAuth adds authentication headers based on auth type
func (wf *WebhookForwarder) addAuth(req *http.Request, authType string, authConfig []byte) {
	if len(authConfig) == 0 {
		return
	}

	var config map[string]interface{}
	if err := json.Unmarshal(authConfig, &config); err != nil {
		return
	}

	switch authType {
	case "bearer":
		if token, ok := config["token"].(string); ok {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "header":
		if headerName, ok := config["header_name"].(string); ok {
			if headerValue, ok := config["header_value"].(string); ok {
				req.Header.Set(headerName, headerValue)
			}
		}
	case "basic":
		if username, ok := config["username"].(string); ok {
			if password, ok := config["password"].(string); ok {
				req.SetBasicAuth(username, password)
			}
		}
	}
}
