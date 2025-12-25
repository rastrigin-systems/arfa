package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rastrigin-systems/arfa/generated/db"
)

type SubscriptionsHandler struct {
	db db.Querier
}

func NewSubscriptionsHandler(database db.Querier) *SubscriptionsHandler {
	return &SubscriptionsHandler{db: database}
}

// GetCurrentSubscription gets the subscription for the authenticated user's organization
func (h *SubscriptionsHandler) GetCurrentSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	subscription, err := h.db.GetSubscriptionByOrgID(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	// Convert pgtype.Numeric to float64 for percentage calculation
	monthlyBudgetVal, err := subscription.MonthlyBudgetUsd.Float64Value()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to convert budget value")
		return
	}
	currentSpendingVal, err := subscription.CurrentSpendingUsd.Float64Value()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to convert spending value")
		return
	}

	monthlyBudget := monthlyBudgetVal.Float64
	currentSpending := currentSpendingVal.Float64

	// Calculate percentage spent
	percentage := float64(0)
	if monthlyBudget > 0 {
		percentage = (currentSpending / monthlyBudget) * 100
	}

	response := map[string]interface{}{
		"id":                   subscription.ID,
		"org_id":               subscription.OrgID,
		"plan_type":            subscription.PlanType,
		"monthly_budget_usd":   monthlyBudget,
		"current_spending_usd": currentSpending,
		"spending_percentage":  percentage,
		"billing_period_start": subscription.BillingPeriodStart,
		"billing_period_end":   subscription.BillingPeriodEnd,
		"status":               subscription.Status,
		"created_at":           subscription.CreatedAt,
		"updated_at":           subscription.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
