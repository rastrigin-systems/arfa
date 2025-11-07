-- Organization and team queries

-- name: GetOrganization :one
SELECT * FROM organizations
WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1;

-- name: CreateOrganization :one
INSERT INTO organizations (
    name,
    slug
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateOrganization :one
UPDATE organizations
SET
    name = COALESCE(NULLIF(sqlc.arg(name), ''), name),
    settings = COALESCE(sqlc.narg(settings), settings),
    max_employees = COALESCE(NULLIF(sqlc.arg(max_employees), 0), max_employees),
    max_agents_per_employee = COALESCE(NULLIF(sqlc.arg(max_agents_per_employee), 0), max_agents_per_employee),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListTeams :many
SELECT * FROM teams
WHERE org_id = $1
ORDER BY name;

-- name: GetTeam :one
SELECT * FROM teams
WHERE id = $1 AND org_id = $2;

-- name: CreateTeam :one
INSERT INTO teams (
    org_id,
    name,
    description
) VALUES (
    $1, $2, $3
)
RETURNING *;

-- name: UpdateTeam :one
UPDATE teams
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTeam :exec
DELETE FROM teams
WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM roles
ORDER BY name;

-- name: GetRole :one
SELECT * FROM roles
WHERE id = $1;

-- name: GetRoleByName :one
SELECT * FROM roles
WHERE name = $1;

-- name: CreateRole :one
INSERT INTO roles (
    name,
    permissions
) VALUES (
    $1, $2
)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET
    name = COALESCE($2, name),
    permissions = COALESCE($3, permissions),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;
