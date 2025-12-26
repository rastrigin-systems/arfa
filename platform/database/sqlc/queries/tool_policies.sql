-- name: GetToolPoliciesForEmployee :many
-- Get all tool policies that apply to an employee (org-level, team-level, and employee-level)
-- Policy resolution order: employee > team > org (but deny always wins regardless of level)
SELECT
    id,
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by,
    created_at,
    updated_at
FROM tool_policies
WHERE org_id = sqlc.arg(org_id)
    AND (
        -- Org-wide policies (team_id IS NULL AND employee_id IS NULL)
        (team_id IS NULL AND employee_id IS NULL)
        -- Team-level policies
        OR (team_id = sqlc.narg(team_id) AND employee_id IS NULL)
        -- Employee-specific policies
        OR employee_id = sqlc.arg(employee_id)
    )
ORDER BY
    -- Order by specificity: employee first, then team, then org
    CASE
        WHEN employee_id IS NOT NULL THEN 1
        WHEN team_id IS NOT NULL THEN 2
        ELSE 3
    END,
    tool_name;

-- name: CreateToolPolicy :one
-- Create a new tool policy
INSERT INTO tool_policies (
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetToolPolicy :one
-- Get a specific tool policy by ID
SELECT
    id,
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by,
    created_at,
    updated_at
FROM tool_policies
WHERE id = $1;

-- name: GetToolPolicyByIdAndOrg :one
-- Get a specific tool policy by ID with org_id check (for authorization)
SELECT
    id,
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by,
    created_at,
    updated_at
FROM tool_policies
WHERE id = $1 AND org_id = $2;

-- name: ListToolPoliciesByOrg :many
-- List all tool policies for an organization (admin view)
SELECT
    id,
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by,
    created_at,
    updated_at
FROM tool_policies
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: ListToolPoliciesFiltered :many
-- List tool policies with optional filters
SELECT
    id,
    org_id,
    team_id,
    employee_id,
    tool_name,
    conditions,
    action,
    reason,
    created_by,
    created_at,
    updated_at
FROM tool_policies
WHERE org_id = sqlc.arg(org_id)
    AND (sqlc.narg(team_id)::uuid IS NULL OR team_id = sqlc.narg(team_id))
    AND (sqlc.narg(employee_id)::uuid IS NULL OR employee_id = sqlc.narg(employee_id))
    AND (sqlc.narg(scope)::text IS NULL OR (
        CASE
            WHEN sqlc.narg(scope) = 'organization' THEN team_id IS NULL AND employee_id IS NULL
            WHEN sqlc.narg(scope) = 'team' THEN team_id IS NOT NULL AND employee_id IS NULL
            WHEN sqlc.narg(scope) = 'employee' THEN employee_id IS NOT NULL
            ELSE TRUE
        END
    ))
ORDER BY created_at DESC;

-- name: UpdateToolPolicy :one
-- Update an existing tool policy
UPDATE tool_policies
SET
    tool_name = $2,
    conditions = $3,
    action = $4,
    reason = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateToolPolicyByOrg :one
-- Update an existing tool policy with org_id check (for authorization)
UPDATE tool_policies
SET
    tool_name = COALESCE(sqlc.narg(tool_name), tool_name),
    conditions = COALESCE(sqlc.narg(conditions), conditions),
    action = COALESCE(sqlc.narg(action), action),
    reason = COALESCE(sqlc.narg(reason), reason),
    updated_at = NOW()
WHERE id = sqlc.arg(id) AND org_id = sqlc.arg(org_id)
RETURNING *;

-- name: DeleteToolPolicy :exec
-- Delete a tool policy
DELETE FROM tool_policies
WHERE id = $1;

-- name: DeleteToolPolicyByOrg :exec
-- Delete a tool policy with org_id check (for authorization)
DELETE FROM tool_policies
WHERE id = $1 AND org_id = $2;

-- name: CountToolPoliciesByOrg :one
-- Count tool policies for an organization
SELECT COUNT(*) FROM tool_policies
WHERE org_id = $1;
