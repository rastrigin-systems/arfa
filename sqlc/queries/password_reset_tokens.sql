-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (
    employee_id,
    token,
    expires_at
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token = $1
  AND expires_at > NOW()
  AND used_at IS NULL
LIMIT 1;

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens
SET used_at = NOW()
WHERE token = $1;

-- name: CountRecentPasswordResetRequests :one
SELECT COUNT(*) FROM password_reset_tokens
WHERE employee_id = $1
  AND created_at > NOW() - INTERVAL '1 hour';

-- name: UpdateEmployeePassword :exec
UPDATE employees
SET password_hash = $1,
    updated_at = NOW()
WHERE id = $2;
