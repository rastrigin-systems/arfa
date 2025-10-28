-- Authentication and session queries

-- name: CreateSession :one
INSERT INTO sessions (
    employee_id,
    token_hash,
    ip_address,
    user_agent,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetSession :one
SELECT * FROM sessions
WHERE token_hash = $1 AND expires_at > NOW();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token_hash = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();

-- name: DeleteEmployeeSessions :exec
DELETE FROM sessions
WHERE employee_id = $1;

-- name: GetSessionWithEmployee :one
SELECT
    s.id,
    s.employee_id,
    s.token_hash,
    s.ip_address,
    s.user_agent,
    s.expires_at,
    s.created_at,
    e.org_id,
    e.team_id,
    e.role_id,
    e.email,
    e.full_name,
    e.status,
    e.preferences,
    e.last_login_at,
    e.created_at as employee_created_at,
    e.updated_at as employee_updated_at
FROM sessions s
JOIN employees e ON s.employee_id = e.id
WHERE s.token_hash = $1
  AND s.expires_at > NOW()
  AND e.deleted_at IS NULL
  AND e.status = 'active';
