package handlers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/generated/mocks"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/handlers"
)

// ============================================================================
// GetCurrentSubscription Tests
// ============================================================================

// TDD Lesson: Testing subscription retrieval with budget calculations
func TestGetCurrentSubscription_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	subID := uuid.New()

	// Create mock subscription
	subscription := db.Subscription{
		ID:                 subID,
		OrgID:              orgID,
		PlanType:           "professional",
		MonthlyBudgetUsd:   mustParseInt(t, 100000), // $1000.00
		CurrentSpendingUsd: mustParseInt(t, 45000),  // $450.00
		BillingPeriodStart: pgtype.Timestamp{Time: time.Now().AddDate(0, 0, -15), Valid: true},
		BillingPeriodEnd:   pgtype.Timestamp{Time: time.Now().AddDate(0, 0, 15), Valid: true},
		Status:             "active",
		CreatedAt:          pgtype.Timestamp{Time: time.Now().AddDate(0, -1, 0), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	// Mock subscription query
	mockDB.EXPECT().
		GetSubscriptionByOrgID(gomock.Any(), orgID).
		Return(subscription, nil)

	handler := handlers.NewSubscriptionsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, subID.String(), response["id"])
	assert.Equal(t, orgID.String(), response["org_id"])
	assert.Equal(t, "professional", response["plan_type"])
	assert.Equal(t, "active", response["status"])

	// Verify budget values
	monthlyBudget, ok := response["monthly_budget_usd"].(float64)
	require.True(t, ok)
	assert.Greater(t, monthlyBudget, 0.0)

	currentSpending, ok := response["current_spending_usd"].(float64)
	require.True(t, ok)
	assert.Greater(t, currentSpending, 0.0)

	// Verify percentage calculation (450/1000 = 45%)
	spendingPercentage, ok := response["spending_percentage"].(float64)
	require.True(t, ok)
	assert.Greater(t, spendingPercentage, 0.0)
	assert.Less(t, spendingPercentage, 100.0)

	// Verify timestamps
	assert.NotNil(t, response["billing_period_start"])
	assert.NotNil(t, response["billing_period_end"])
	assert.NotNil(t, response["created_at"])
	assert.NotNil(t, response["updated_at"])
}

// TDD Lesson: Testing subscription with zero budget (division by zero safety)
func TestGetCurrentSubscription_ZeroBudget(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	subID := uuid.New()

	// Create subscription with zero budget
	subscription := db.Subscription{
		ID:                 subID,
		OrgID:              orgID,
		PlanType:           "free",
		MonthlyBudgetUsd:   mustParseInt(t, 0),    // $0.00
		CurrentSpendingUsd: mustParseInt(t, 1000), // $10.00
		BillingPeriodStart: pgtype.Timestamp{Time: time.Now().AddDate(0, 0, -15), Valid: true},
		BillingPeriodEnd:   pgtype.Timestamp{Time: time.Now().AddDate(0, 0, 15), Valid: true},
		Status:             "active",
		CreatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	mockDB.EXPECT().
		GetSubscriptionByOrgID(gomock.Any(), orgID).
		Return(subscription, nil)

	handler := handlers.NewSubscriptionsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 200 OK (graceful handling of division by zero)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify percentage is 0 (not NaN or infinity)
	spendingPercentage := response["spending_percentage"].(float64)
	assert.Equal(t, float64(0), spendingPercentage)
}

// TDD Lesson: Testing subscription with 100% budget spent
func TestGetCurrentSubscription_FullBudgetSpent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	subID := uuid.New()

	// Create subscription with budget fully spent
	subscription := db.Subscription{
		ID:                 subID,
		OrgID:              orgID,
		PlanType:           "starter",
		MonthlyBudgetUsd:   mustParseInt(t, 50000), // $500.00
		CurrentSpendingUsd: mustParseInt(t, 50000), // $500.00 (100%)
		BillingPeriodStart: pgtype.Timestamp{Time: time.Now().AddDate(0, 0, -15), Valid: true},
		BillingPeriodEnd:   pgtype.Timestamp{Time: time.Now().AddDate(0, 0, 15), Valid: true},
		Status:             "active",
		CreatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	mockDB.EXPECT().
		GetSubscriptionByOrgID(gomock.Any(), orgID).
		Return(subscription, nil)

	handler := handlers.NewSubscriptionsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify 100% spending
	spendingPercentage := response["spending_percentage"].(float64)
	assert.Equal(t, float64(100), spendingPercentage)
}

// TDD Lesson: Testing subscription over budget (>100%)
func TestGetCurrentSubscription_OverBudget(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	subID := uuid.New()

	// Create subscription with spending over budget
	subscription := db.Subscription{
		ID:                 subID,
		OrgID:              orgID,
		PlanType:           "starter",
		MonthlyBudgetUsd:   mustParseInt(t, 50000), // $500.00
		CurrentSpendingUsd: mustParseInt(t, 65000), // $650.00 (130%)
		BillingPeriodStart: pgtype.Timestamp{Time: time.Now().AddDate(0, 0, -15), Valid: true},
		BillingPeriodEnd:   pgtype.Timestamp{Time: time.Now().AddDate(0, 0, 15), Valid: true},
		Status:             "active",
		CreatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	mockDB.EXPECT().
		GetSubscriptionByOrgID(gomock.Any(), orgID).
		Return(subscription, nil)

	handler := handlers.NewSubscriptionsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify over 100% spending
	spendingPercentage := response["spending_percentage"].(float64)
	assert.Greater(t, spendingPercentage, float64(100))
}

// TDD Lesson: Testing unauthorized access (no org_id)
func TestGetCurrentSubscription_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewSubscriptionsHandler(mockDB)

	// Create request WITHOUT org_id in context
	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", response["error"])
}

// TDD Lesson: Testing subscription not found
func TestGetCurrentSubscription_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// No subscription found
	mockDB.EXPECT().
		GetSubscriptionByOrgID(gomock.Any(), orgID).
		Return(db.Subscription{}, assert.AnError)

	handler := handlers.NewSubscriptionsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetCurrentSubscription(rec, req)

	// Assert HTTP 404 Not Found
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Subscription not found", response["error"])
}

// TDD Lesson: Testing various subscription plan types
func TestGetCurrentSubscription_DifferentPlanTypes(t *testing.T) {
	testCases := []struct {
		planType string
		status   string
	}{
		{"free", "active"},
		{"starter", "active"},
		{"professional", "active"},
		{"enterprise", "active"},
		{"professional", "suspended"},
		{"professional", "canceled"},
	}

	for _, tc := range testCases {
		t.Run(tc.planType+"_"+tc.status, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			orgID := uuid.New()
			subID := uuid.New()

			subscription := db.Subscription{
				ID:                 subID,
				OrgID:              orgID,
				PlanType:           tc.planType,
				MonthlyBudgetUsd:   mustParseInt(t, 100000),
				CurrentSpendingUsd: mustParseInt(t, 25000),
				BillingPeriodStart: pgtype.Timestamp{Time: time.Now(), Valid: true},
				BillingPeriodEnd:   pgtype.Timestamp{Time: time.Now(), Valid: true},
				Status:             tc.status,
				CreatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:          pgtype.Timestamp{Time: time.Now(), Valid: true},
			}

			mockDB.EXPECT().
				GetSubscriptionByOrgID(gomock.Any(), orgID).
				Return(subscription, nil)

			handler := handlers.NewSubscriptionsHandler(mockDB)

			req := httptest.NewRequest(http.MethodGet, "/subscription", nil)
			req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
			rec := httptest.NewRecorder()

			handler.GetCurrentSubscription(rec, req)

			// Assert HTTP 200 OK
			assert.Equal(t, http.StatusOK, rec.Code)

			var response map[string]interface{}
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)

			// Verify plan type and status
			assert.Equal(t, tc.planType, response["plan_type"])
			assert.Equal(t, tc.status, response["status"])
		})
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

// Helper to create pgtype.Numeric from int64 (cents)
func mustParseInt(t *testing.T, val int64) pgtype.Numeric {
	// For testing, create a Numeric value
	// The value represents cents (e.g., 100000 = $1000.00)
	// Convert to decimal string format
	dollars := float64(val) / 100.0
	dollarStr := fmt.Sprintf("%.2f", dollars)

	var result pgtype.Numeric
	err := result.Scan(dollarStr)
	if err != nil {
		t.Fatalf("Failed to create numeric from %d: %v", val, err)
	}
	return result
}
