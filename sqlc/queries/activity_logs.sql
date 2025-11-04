-- name: ListActivityLogs :many
-- List recent activity logs for an organization with pagination
SELECT
    id,
    org_id,
    employee_id,
    session_id,
    agent_id,
    event_type,
    event_category,
    content,
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
    session_id,
    agent_id,
    event_type,
    event_category,
    content,
    payload
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: CountActivityLogs :one
-- Count total activity logs for an organization
SELECT COUNT(*) FROM activity_logs
WHERE org_id = $1;

-- name: GetLogsBySession :many
-- Get all logs for a specific CLI session
SELECT
    id,
    org_id,
    employee_id,
    session_id,
    agent_id,
    event_type,
    event_category,
    content,
    payload,
    created_at
FROM activity_logs
WHERE session_id = $1
ORDER BY created_at ASC;

-- name: GetLogsByEmployee :many
-- Get logs for a specific employee with filters
SELECT
    id,
    org_id,
    employee_id,
    session_id,
    agent_id,
    event_type,
    event_category,
    content,
    payload,
    created_at
FROM activity_logs
WHERE org_id = sqlc.arg(org_id)
    AND employee_id = sqlc.arg(employee_id)
    AND (sqlc.narg(event_category)::VARCHAR IS NULL OR event_category = sqlc.narg(event_category))
    AND (sqlc.narg(since)::TIMESTAMP IS NULL OR created_at >= sqlc.narg(since))
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: DeleteOldLogs :exec
-- Delete activity logs older than specified timestamp
DELETE FROM activity_logs
WHERE created_at < $1;
