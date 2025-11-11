-- ============================================================================
-- MCP Catalog Queries
-- ============================================================================

-- name: ListMCPServers :many
SELECT * FROM mcp_catalog
WHERE is_approved = true
ORDER BY provider, name;

-- name: ListAllMCPServers :many
SELECT * FROM mcp_catalog
ORDER BY provider, name;

-- name: GetMCPServer :one
SELECT * FROM mcp_catalog
WHERE id = $1;

-- name: GetMCPServerByName :one
SELECT * FROM mcp_catalog
WHERE name = $1;

-- name: CreateMCPServer :one
INSERT INTO mcp_catalog (
    name,
    provider,
    version,
    description,
    connection_schema,
    capabilities,
    requires_credentials,
    is_approved,
    category_id,
    docker_image,
    config_template,
    required_env_vars
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
) RETURNING *;

-- name: UpdateMCPServer :one
UPDATE mcp_catalog
SET
    provider = COALESCE(sqlc.narg(provider), provider),
    version = COALESCE(sqlc.narg(version), version),
    description = COALESCE(sqlc.narg(description), description),
    connection_schema = COALESCE(sqlc.narg(connection_schema), connection_schema),
    capabilities = COALESCE(sqlc.narg(capabilities), capabilities),
    requires_credentials = COALESCE(sqlc.narg(requires_credentials), requires_credentials),
    is_approved = COALESCE(sqlc.narg(is_approved), is_approved),
    category_id = COALESCE(sqlc.narg(category_id), category_id),
    docker_image = COALESCE(sqlc.narg(docker_image), docker_image),
    config_template = COALESCE(sqlc.narg(config_template), config_template),
    required_env_vars = COALESCE(sqlc.narg(required_env_vars), required_env_vars),
    updated_at = NOW()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteMCPServer :exec
DELETE FROM mcp_catalog
WHERE id = $1;

-- name: ApproveMCPServer :exec
UPDATE mcp_catalog
SET is_approved = true, updated_at = NOW()
WHERE id = $1;

-- name: DisapproveMCPServer :exec
UPDATE mcp_catalog
SET is_approved = false, updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- Employee MCP Configuration Queries
-- ============================================================================

-- name: ListEmployeeMCPConfigs :many
SELECT
    mc.id,
    mc.name,
    mc.provider,
    mc.version,
    mc.description,
    mc.docker_image,
    mc.config_template,
    mc.required_env_vars,
    emc.connection_config,
    emc.is_enabled,
    emc.created_at
FROM mcp_catalog mc
JOIN employee_mcp_configs emc ON emc.mcp_catalog_id = mc.id
WHERE emc.employee_id = $1
ORDER BY mc.provider, mc.name;

-- name: GetEmployeeMCPConfig :one
SELECT
    mc.*,
    emc.connection_config,
    emc.is_enabled,
    emc.created_at as configured_at
FROM mcp_catalog mc
JOIN employee_mcp_configs emc ON emc.mcp_catalog_id = mc.id
WHERE emc.employee_id = $1 AND emc.mcp_catalog_id = $2;

-- name: CreateEmployeeMCPConfig :one
INSERT INTO employee_mcp_configs (
    employee_id,
    mcp_catalog_id,
    connection_config,
    is_enabled
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: UpdateEmployeeMCPConfig :one
UPDATE employee_mcp_configs
SET
    connection_config = COALESCE(sqlc.narg(connection_config), connection_config),
    is_enabled = COALESCE(sqlc.narg(is_enabled), is_enabled),
    updated_at = NOW()
WHERE employee_id = sqlc.arg(employee_id) AND mcp_catalog_id = sqlc.arg(mcp_catalog_id)
RETURNING *;

-- name: DeleteEmployeeMCPConfig :exec
DELETE FROM employee_mcp_configs
WHERE employee_id = $1 AND mcp_catalog_id = $2;

-- name: CountEmployeeMCPConfigs :one
SELECT COUNT(*)
FROM employee_mcp_configs
WHERE employee_id = $1;

-- name: GetMCPUsageCount :one
SELECT COUNT(*)
FROM employee_mcp_configs
WHERE mcp_catalog_id = $1 AND is_enabled = true;
