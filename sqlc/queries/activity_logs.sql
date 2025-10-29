-- name: ListActivityLogs :many
-- List recent activity logs for an organization with pagination
SELECT
    id,
    org_id,
    employee_id,
    event_type,
    event_category,
    payload,
    created_at
FROM activity_logs
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateActivityLog :one
-- Create a new activity log entry
INSERT INTO activity_logs (
    org_id,
    employee_id,
    event_type,
    event_category,
    payload
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: CountActivityLogs :one
-- Count total activity logs for an organization
SELECT COUNT(*) FROM activity_logs
WHERE org_id = $1;
