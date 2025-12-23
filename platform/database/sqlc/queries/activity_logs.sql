-- name: GetActivityLog :one
-- Get a single activity log by ID
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
WHERE id = $1;

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

-- name: ListActivityLogsFiltered :many
-- List activity logs with comprehensive filtering
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
    AND (sqlc.narg(employee_id)::UUID IS NULL OR employee_id = sqlc.narg(employee_id))
    AND (sqlc.narg(session_id)::UUID IS NULL OR session_id = sqlc.narg(session_id))
    AND (sqlc.narg(agent_id)::UUID IS NULL OR agent_id = sqlc.narg(agent_id))
    AND (sqlc.narg(event_type)::VARCHAR IS NULL OR event_type = sqlc.narg(event_type))
    AND (sqlc.narg(event_category)::VARCHAR IS NULL OR event_category = sqlc.narg(event_category))
    AND (sqlc.narg(start_date)::TIMESTAMP IS NULL OR created_at >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date)::TIMESTAMP IS NULL OR created_at <= sqlc.narg(end_date))
ORDER BY created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountActivityLogsFiltered :one
-- Count activity logs with comprehensive filtering
SELECT COUNT(*) FROM activity_logs
WHERE org_id = sqlc.arg(org_id)
    AND (sqlc.narg(employee_id)::UUID IS NULL OR employee_id = sqlc.narg(employee_id))
    AND (sqlc.narg(session_id)::UUID IS NULL OR session_id = sqlc.narg(session_id))
    AND (sqlc.narg(agent_id)::UUID IS NULL OR agent_id = sqlc.narg(agent_id))
    AND (sqlc.narg(event_type)::VARCHAR IS NULL OR event_type = sqlc.narg(event_type))
    AND (sqlc.narg(event_category)::VARCHAR IS NULL OR event_category = sqlc.narg(event_category))
    AND (sqlc.narg(start_date)::TIMESTAMP IS NULL OR created_at >= sqlc.narg(start_date))
    AND (sqlc.narg(end_date)::TIMESTAMP IS NULL OR created_at <= sqlc.narg(end_date));
