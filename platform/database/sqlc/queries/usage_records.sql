-- name: CreateUsageRecord :one
INSERT INTO usage_records (
    org_id,
    employee_id,
    agent_config_id,
    resource_type,
    quantity,
    cost_usd,
    period_start,
    period_end,
    metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetUsageRecordsByEmployee :many
SELECT * FROM usage_records
WHERE employee_id = $1
  AND period_start >= $2
  AND period_end <= $3
ORDER BY period_start DESC;

-- name: GetUsageRecordsByOrg :many
SELECT * FROM usage_records
WHERE org_id = $1
  AND period_start >= $2
  AND period_end <= $3
ORDER BY period_start DESC;

-- name: GetEmployeeUsageStats :one
SELECT
    COUNT(*) as total_records,
    COALESCE(SUM(CASE WHEN resource_type = 'api_calls' THEN quantity ELSE 0 END), 0) as total_api_calls,
    COALESCE(SUM(CASE WHEN resource_type = 'llm_tokens' THEN quantity ELSE 0 END), 0) as total_tokens,
    COALESCE(SUM(cost_usd), 0) as total_cost_usd
FROM usage_records
WHERE employee_id = sqlc.arg(employee_id)::uuid
  AND period_start >= sqlc.arg(period_start)
  AND period_end <= sqlc.arg(period_end);

-- name: GetOrgUsageStats :one
SELECT
    COUNT(*) as total_records,
    COALESCE(SUM(CASE WHEN resource_type = 'api_calls' THEN quantity ELSE 0 END), 0) as total_api_calls,
    COALESCE(SUM(CASE WHEN resource_type = 'llm_tokens' THEN quantity ELSE 0 END), 0) as total_tokens,
    COALESCE(SUM(cost_usd), 0) as total_cost_usd
FROM usage_records
WHERE org_id = sqlc.arg(org_id)
  AND period_start >= sqlc.arg(period_start)
  AND period_end <= sqlc.arg(period_end);
