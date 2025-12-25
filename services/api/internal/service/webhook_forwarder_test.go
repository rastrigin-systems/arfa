package service

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
)

// ============================================================================
// matchesEventFilter Tests
// ============================================================================

func TestMatchesEventFilter_EmptyFilter_MatchesAll(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	// Empty filter should match all event types
	assert.True(t, wf.matchesEventFilter("tool_call", []string{}))
	assert.True(t, wf.matchesEventFilter("permission_denied", []string{}))
	assert.True(t, wf.matchesEventFilter("any_event", []string{}))
	assert.True(t, wf.matchesEventFilter("", []string{}))
}

func TestMatchesEventFilter_WildcardFilter_MatchesAll(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	// Wildcard filter should match all
	assert.True(t, wf.matchesEventFilter("tool_call", []string{"*"}))
	assert.True(t, wf.matchesEventFilter("permission_denied", []string{"*"}))
	assert.True(t, wf.matchesEventFilter("any_event", []string{"*", "other"}))
}

func TestMatchesEventFilter_SpecificFilter_MatchesExact(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	// Specific filter should match exactly
	assert.True(t, wf.matchesEventFilter("tool_call", []string{"tool_call"}))
	assert.True(t, wf.matchesEventFilter("tool_call", []string{"other", "tool_call"}))
	assert.False(t, wf.matchesEventFilter("tool_call", []string{"other"}))
	assert.False(t, wf.matchesEventFilter("tool_call", []string{"permission_denied"}))
}

func TestMatchesEventFilter_MultipleFilters(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	filters := []string{"tool_call", "permission_denied", "api_request"}

	assert.True(t, wf.matchesEventFilter("tool_call", filters))
	assert.True(t, wf.matchesEventFilter("permission_denied", filters))
	assert.True(t, wf.matchesEventFilter("api_request", filters))
	assert.False(t, wf.matchesEventFilter("other_event", filters))
}

// ============================================================================
// computeSignature Tests
// ============================================================================

func TestComputeSignature_ValidPayload(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	payload := []byte(`{"event":"test","data":{"key":"value"}}`)
	secret := "test-secret"

	signature := wf.computeSignature(payload, secret)

	// Signature should start with sha256=
	assert.True(t, len(signature) > 7)
	assert.Equal(t, "sha256=", signature[:7])

	// Should be deterministic
	signature2 := wf.computeSignature(payload, secret)
	assert.Equal(t, signature, signature2)
}

func TestComputeSignature_DifferentSecrets(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	payload := []byte(`{"event":"test"}`)

	sig1 := wf.computeSignature(payload, "secret1")
	sig2 := wf.computeSignature(payload, "secret2")

	// Different secrets should produce different signatures
	assert.NotEqual(t, sig1, sig2)
}

func TestComputeSignature_DifferentPayloads(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	secret := "test-secret"

	sig1 := wf.computeSignature([]byte(`{"a":1}`), secret)
	sig2 := wf.computeSignature([]byte(`{"a":2}`), secret)

	// Different payloads should produce different signatures
	assert.NotEqual(t, sig1, sig2)
}

// ============================================================================
// buildPayload Tests
// ============================================================================

func TestBuildPayload_MinimalLog(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	orgID := uuid.New()
	logID := uuid.New()

	log := db.ActivityLog{
		ID:            logID,
		OrgID:         orgID,
		EventType:     "tool_call",
		EventCategory: "agent_activity",
		Payload:       []byte(`{}`),
		CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	payload := wf.buildPayload(log)

	assert.Equal(t, logID, payload.ID)
	assert.Equal(t, orgID, payload.OrgID)
	assert.Equal(t, "tool_call", payload.EventType)
	assert.Equal(t, "agent_activity", payload.EventCategory)
	assert.Nil(t, payload.EmployeeID)
	assert.Nil(t, payload.SessionID)
	assert.Empty(t, payload.Content)
}

func TestBuildPayload_FullLog(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	orgID := uuid.New()
	logID := uuid.New()
	empID := uuid.New()
	sessionID := uuid.New()
	content := "Test content"

	log := db.ActivityLog{
		ID:            logID,
		OrgID:         orgID,
		EmployeeID:    pgtype.UUID{Bytes: empID, Valid: true},
		SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
		EventType:     "permission_denied",
		EventCategory: "security",
		Content:       &content,
		Payload:       []byte(`{"tool":"Bash","blocked":true}`),
		CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	payload := wf.buildPayload(log)

	assert.Equal(t, logID, payload.ID)
	assert.Equal(t, orgID, payload.OrgID)
	assert.Equal(t, "permission_denied", payload.EventType)
	assert.Equal(t, "security", payload.EventCategory)
	assert.NotNil(t, payload.EmployeeID)
	assert.Equal(t, empID, *payload.EmployeeID)
	assert.NotNil(t, payload.SessionID)
	assert.Equal(t, sessionID, *payload.SessionID)
	assert.Equal(t, "Test content", payload.Content)
	assert.NotNil(t, payload.Payload)
	assert.Equal(t, true, payload.Payload["blocked"])
}

func TestBuildPayload_InvalidCreatedAt(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	log := db.ActivityLog{
		ID:            uuid.New(),
		OrgID:         uuid.New(),
		EventType:     "test",
		EventCategory: "test",
		Payload:       []byte(`{}`),
		CreatedAt:     pgtype.Timestamp{Valid: false}, // Invalid timestamp
	}

	payload := wf.buildPayload(log)

	// Should use current time as fallback
	assert.False(t, payload.Timestamp.IsZero())
}

// ============================================================================
// addAuth Tests
// ============================================================================

func TestAddAuth_Bearer(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	authConfig := []byte(`{"token":"test-bearer-token"}`)

	wf.addAuth(req, "bearer", authConfig)

	assert.Equal(t, "Bearer test-bearer-token", req.Header.Get("Authorization"))
}

func TestAddAuth_Header(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	authConfig := []byte(`{"header_name":"X-API-Key","header_value":"my-secret-key"}`)

	wf.addAuth(req, "header", authConfig)

	assert.Equal(t, "my-secret-key", req.Header.Get("X-API-Key"))
}

func TestAddAuth_Basic(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)
	authConfig := []byte(`{"username":"user","password":"pass"}`)

	wf.addAuth(req, "basic", authConfig)

	// Basic auth should be set
	username, password, ok := req.BasicAuth()
	assert.True(t, ok)
	assert.Equal(t, "user", username)
	assert.Equal(t, "pass", password)
}

func TestAddAuth_None(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	wf.addAuth(req, "none", []byte(`{}`))

	assert.Empty(t, req.Header.Get("Authorization"))
}

func TestAddAuth_EmptyConfig(t *testing.T) {
	wf := NewWebhookForwarder(nil)

	req := httptest.NewRequest(http.MethodPost, "/webhook", nil)

	// Empty config should not crash
	wf.addAuth(req, "bearer", []byte{})
	wf.addAuth(req, "bearer", nil)

	assert.Empty(t, req.Header.Get("Authorization"))
}

// ============================================================================
// ProcessDeliveries Integration Tests
// ============================================================================

func TestProcessDeliveries_NoEnabledDestinations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// No enabled destinations
	mockDB.EXPECT().
		ListEnabledDestinations(gomock.Any()).
		Return([]db.ListEnabledDestinationsRow{}, nil)

	wf := NewWebhookForwarder(mockDB)

	err := wf.ProcessDeliveries(t.Context())
	require.NoError(t, err)
}

func TestProcessDeliveries_WithDestination_NoLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	destID := uuid.New()

	// Return one enabled destination
	mockDB.EXPECT().
		ListEnabledDestinations(gomock.Any()).
		Return([]db.ListEnabledDestinationsRow{
			{
				ID:             destID,
				OrgID:          orgID,
				Name:           "Test Webhook",
				Url:            "https://example.com/webhook",
				AuthType:       "none",
				EventTypes:     []string{},
				BatchSize:      100,
				TimeoutMs:      5000,
				RetryMax:       3,
				RetryBackoffMs: 1000,
			},
		}, nil)

	// No undelivered logs
	mockDB.EXPECT().
		GetUndeliveredLogs(gomock.Any(), gomock.Any()).
		Return([]uuid.UUID{}, nil)

	// No pending deliveries
	mockDB.EXPECT().
		GetPendingDeliveries(gomock.Any(), gomock.Any()).
		Return([]db.GetPendingDeliveriesRow{}, nil)

	wf := NewWebhookForwarder(mockDB)

	err := wf.ProcessDeliveries(t.Context())
	require.NoError(t, err)
}

func TestProcessDeliveries_CreatesDeliveryForUndeliveredLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	destID := uuid.New()
	logID := uuid.New()

	// Return one enabled destination
	mockDB.EXPECT().
		ListEnabledDestinations(gomock.Any()).
		Return([]db.ListEnabledDestinationsRow{
			{
				ID:             destID,
				OrgID:          orgID,
				Name:           "Test Webhook",
				Url:            "https://example.com/webhook",
				AuthType:       "none",
				EventTypes:     []string{}, // Match all
				BatchSize:      100,
				TimeoutMs:      5000,
				RetryMax:       3,
				RetryBackoffMs: 1000,
			},
		}, nil)

	// One undelivered log
	mockDB.EXPECT().
		GetUndeliveredLogs(gomock.Any(), gomock.Any()).
		Return([]uuid.UUID{logID}, nil)

	// Get the log entry
	mockDB.EXPECT().
		GetActivityLog(gomock.Any(), logID).
		Return(db.ActivityLog{
			ID:            logID,
			OrgID:         orgID,
			EventType:     "tool_call",
			EventCategory: "agent_activity",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	// Expect delivery to be created
	mockDB.EXPECT().
		CreateWebhookDelivery(gomock.Any(), db.CreateWebhookDeliveryParams{
			DestinationID: destID,
			LogID:         logID,
		}).
		Return(db.WebhookDelivery{
			ID:            uuid.New(),
			DestinationID: destID,
			LogID:         logID,
			Status:        "pending",
		}, nil)

	// No pending deliveries (we just created it, it won't be picked up until next cycle)
	mockDB.EXPECT().
		GetPendingDeliveries(gomock.Any(), gomock.Any()).
		Return([]db.GetPendingDeliveriesRow{}, nil)

	wf := NewWebhookForwarder(mockDB)

	err := wf.ProcessDeliveries(t.Context())
	require.NoError(t, err)
}

func TestProcessDeliveries_SkipsFilteredEventTypes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	destID := uuid.New()
	logID := uuid.New()

	// Destination only accepts "permission_denied" events
	mockDB.EXPECT().
		ListEnabledDestinations(gomock.Any()).
		Return([]db.ListEnabledDestinationsRow{
			{
				ID:             destID,
				OrgID:          orgID,
				Name:           "Security Webhook",
				Url:            "https://example.com/webhook",
				AuthType:       "none",
				EventTypes:     []string{"permission_denied"}, // Only security events
				BatchSize:      100,
				TimeoutMs:      5000,
				RetryMax:       3,
				RetryBackoffMs: 1000,
			},
		}, nil)

	// One undelivered log
	mockDB.EXPECT().
		GetUndeliveredLogs(gomock.Any(), gomock.Any()).
		Return([]uuid.UUID{logID}, nil)

	// Get the log entry - it's a tool_call, not permission_denied
	mockDB.EXPECT().
		GetActivityLog(gomock.Any(), logID).
		Return(db.ActivityLog{
			ID:            logID,
			OrgID:         orgID,
			EventType:     "tool_call", // Doesn't match filter
			EventCategory: "agent_activity",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	// CreateWebhookDelivery should NOT be called because event doesn't match filter

	// No pending deliveries
	mockDB.EXPECT().
		GetPendingDeliveries(gomock.Any(), gomock.Any()).
		Return([]db.GetPendingDeliveriesRow{}, nil)

	wf := NewWebhookForwarder(mockDB)

	err := wf.ProcessDeliveries(t.Context())
	require.NoError(t, err)
}
