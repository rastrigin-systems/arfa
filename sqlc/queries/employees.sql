-- Employee queries for sqlc code generation

-- name: GetEmployee :one
SELECT * FROM employees
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetEmployeeByEmail :one
SELECT * FROM employees
WHERE email = $1 AND deleted_at IS NULL;

-- name: ListEmployees :many
SELECT * FROM employees
WHERE org_id = $1 
  AND deleted_at IS NULL
  AND ($2::text IS NULL OR status = $2)
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountEmployees :one
SELECT COUNT(*) FROM employees
WHERE org_id = $1 
  AND deleted_at IS NULL
  AND ($2::text IS NULL OR status = $2);

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
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateEmployeeLastLogin :exec
UPDATE employees
SET last_login_at = NOW()
WHERE id = $1;

-- name: SoftDeleteEmployee :exec
UPDATE employees
SET deleted_at = NOW()
WHERE id = $1;

-- name: GetEmployeesByTeam :many
SELECT * FROM employees
WHERE team_id = $1 
  AND deleted_at IS NULL
ORDER BY full_name;

-- name: GetEmployeeWithRole :one
SELECT 
    e.*,
    r.name as role_name,
    r.permissions as role_permissions
FROM employees e
JOIN roles r ON e.role_id = r.id
WHERE e.id = $1 AND e.deleted_at IS NULL;
