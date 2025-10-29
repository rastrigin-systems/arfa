-- name: GetSubscriptionByOrgID :one
SELECT * FROM subscriptions
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: CreateSubscription :one
INSERT INTO subscriptions (
    org_id,
    plan_type,
    monthly_budget_usd,
    billing_period_start,
    billing_period_end,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: UpdateSubscriptionSpending :one
UPDATE subscriptions
SET
    current_spending_usd = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: IncrementSubscriptionSpending :one
UPDATE subscriptions
SET
    current_spending_usd = current_spending_usd + $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;
