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

-- name: CreateRole :one
INSERT INTO roles (
    name,
    permissions
) VALUES (
    $1, $2
)
RETURNING *;
