package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

type SubscriptionsHandler struct {
	queries *db.Queries
}

func NewSubscriptionsHandler(queries *db.Queries) *SubscriptionsHandler {
	return &SubscriptionsHandler{queries: queries}
}

// GetCurrentSubscription gets the subscription for the authenticated user's organization
func (h *SubscriptionsHandler) GetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	subscription, err := h.queries.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		http.Error(w, "Subscription not found", http.StatusNotFound)
		return
	}

	// Calculate percentage spent
	percentage := float64(0)
	if subscription.MonthlyBudgetUsd.Float64 > 0 {
		percentage = (subscription.CurrentSpendingUsd.Float64 / subscription.MonthlyBudgetUsd.Float64) * 100
	}

	response := map[string]interface{}{
		"id":                  subscription.ID,
		"org_id":             subscription.OrgID,
		"plan_type":          subscription.PlanType,
		"monthly_budget_usd": subscription.MonthlyBudgetUsd.Float64,
		"current_spending_usd": subscription.CurrentSpendingUsd.Float64,
		"spending_percentage": percentage,
		"billing_period_start": subscription.BillingPeriodStart,
		"billing_period_end": subscription.BillingPeriodEnd,
		"status":             subscription.Status,
		"created_at":         subscription.CreatedAt,
		"updated_at":         subscription.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
