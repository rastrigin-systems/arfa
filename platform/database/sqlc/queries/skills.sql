-- ============================================================================
-- Skills Catalog Queries
-- ============================================================================

-- name: ListSkills :many
SELECT * FROM skill_catalog
WHERE is_active = true
ORDER BY category, name;

-- name: ListAllSkills :many
SELECT * FROM skill_catalog
ORDER BY category, name;

-- name: GetSkill :one
SELECT * FROM skill_catalog
WHERE id = $1;

-- name: GetSkillByName :one
SELECT * FROM skill_catalog
WHERE name = $1;

-- name: CreateSkill :one
INSERT INTO skill_catalog (
    name,
    description,
    category,
    version,
    files,
    dependencies,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: UpdateSkill :one
UPDATE skill_catalog
SET
    description = COALESCE(sqlc.narg(description), description),
    category = COALESCE(sqlc.narg(category), category),
    version = COALESCE(sqlc.narg(version), version),
    files = COALESCE(sqlc.narg(files), files),
    dependencies = COALESCE(sqlc.narg(dependencies), dependencies),
    is_active = COALESCE(sqlc.narg(is_active), is_active),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteSkill :exec
DELETE FROM skill_catalog
WHERE id = $1;

-- name: DeactivateSkill :exec
UPDATE skill_catalog
SET is_active = false, updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- Employee Skills Queries
-- ============================================================================

-- name: ListEmployeeSkills :many
SELECT
    sc.id,
    sc.name,
    sc.description,
    sc.category,
    sc.version,
    sc.files,
    sc.dependencies,
    sc.is_active,
    es.is_enabled,
    es.config,
    es.created_at as installed_at
FROM skill_catalog sc
JOIN employee_skills es ON es.skill_id = sc.id
WHERE es.employee_id = $1
ORDER BY sc.category, sc.name;

-- name: GetEmployeeSkill :one
SELECT
    sc.*,
    es.is_enabled,
    es.config,
    es.created_at as installed_at
FROM skill_catalog sc
JOIN employee_skills es ON es.skill_id = sc.id
WHERE es.employee_id = $1 AND es.skill_id = $2;

-- name: AssignSkillToEmployee :one
INSERT INTO employee_skills (
    employee_id,
    skill_id,
    is_enabled,
    config
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateEmployeeSkill :one
UPDATE employee_skills
SET
    is_enabled = COALESCE(sqlc.narg(is_enabled), is_enabled),
    config = COALESCE(sqlc.narg(config), config),
    updated_at = NOW()
WHERE employee_id = sqlc.arg(employee_id) AND skill_id = sqlc.arg(skill_id)
RETURNING *;

-- name: RemoveSkillFromEmployee :exec
DELETE FROM employee_skills
WHERE employee_id = $1 AND skill_id = $2;

-- name: CountEmployeeSkills :one
SELECT COUNT(*)
FROM employee_skills
WHERE employee_id = $1;

-- name: GetSkillUsageCount :one
SELECT COUNT(*)
FROM employee_skills
WHERE skill_id = $1 AND is_enabled = true;
