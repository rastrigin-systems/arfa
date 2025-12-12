-- Employee queries for sqlc code generation

-- name: GetEmployee :one
SELECT
  e.*,
  t.name as team_name
FROM employees e
LEFT JOIN teams t ON e.team_id = t.id
WHERE e.id = $1;

-- name: GetEmployeeByEmail :one
SELECT * FROM employees
WHERE email = $1;

-- name: ListEmployees :many
SELECT
  e.*,
  t.name as team_name
FROM employees e
LEFT JOIN teams t ON e.team_id = t.id
WHERE e.org_id = sqlc.arg(org_id)
  AND (sqlc.narg(status)::text IS NULL OR e.status = sqlc.narg(status)::text)
  AND (sqlc.narg(team_id)::uuid IS NULL OR e.team_id = sqlc.narg(team_id)::uuid)
  AND (sqlc.narg(search)::text IS NULL OR e.full_name ILIKE '%' || sqlc.narg(search)::text || '%' OR e.email ILIKE '%' || sqlc.narg(search)::text || '%')
ORDER BY e.created_at DESC
LIMIT sqlc.arg(query_limit) OFFSET sqlc.arg(query_offset);

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE org_id = sqlc.arg(org_id)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
  AND (sqlc.narg(team_id)::uuid IS NULL OR team_id = sqlc.narg(team_id)::uuid)
  AND (sqlc.narg(search)::text IS NULL OR full_name ILIKE '%' || sqlc.narg(search)::text || '%' OR email ILIKE '%' || sqlc.narg(search)::text || '%');

-- name: CreateEmployee :one
INSERT INTO employees (
    org_id,
    team_id,
    role_id,
    email,
    full_name,
    password_hash,
    status,
    preferences
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateEmployee :one
UPDATE employees
SET
    full_name = COALESCE($2, full_name),
    team_id = COALESCE($3, team_id),
    role_id = COALESCE($4, role_id),
    status = COALESCE($5, status),
    preferences = COALESCE($6, preferences),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateEmployeeLastLogin :exec
UPDATE employees
SET last_login_at = NOW()
WHERE id = $1;

-- name: DeleteEmployee :exec
DELETE FROM employees
WHERE id = $1;

-- name: GetEmployeesByTeam :many
SELECT * FROM employees
WHERE team_id = $1
ORDER BY full_name;

-- name: GetEmployeeWithRole :one
SELECT
    e.*,
    r.name as role_name,
    r.permissions as role_permissions
FROM employees e
JOIN roles r ON e.role_id = r.id
WHERE e.id = $1;

-- name: CountEmployeesByTeam :one
SELECT COUNT(*) as count
FROM employees
WHERE team_id = $1;

-- name: CountEmployeesByRole :one
SELECT COUNT(*) as count
FROM employees
WHERE role_id = $1;
