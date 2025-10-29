-- Agent Queries
-- These queries manage the available AI agents (Claude Code, Cursor, Windsurf, etc.)

-- name: ListAgents :many
-- List all public/active agents
SELECT
    id,
    name,
    type,
    description,
    provider,
    default_config,
    capabilities,
    llm_provider,
    llm_model,
    is_public,
    created_at,
    updated_at
FROM agents
WHERE is_public = true
ORDER BY name ASC;

-- name: GetAgentByID :one
-- Get a specific agent by ID
SELECT
    id,
    name,
    type,
    description,
    provider,
    default_config,
    capabilities,
    llm_provider,
    llm_model,
    is_public,
    created_at,
    updated_at
FROM agents
WHERE id = $1;

-- name: GetAgentByName :one
-- Get a specific agent by name
SELECT
    id,
    name,
    type,
    description,
    provider,
    default_config,
    capabilities,
    llm_provider,
    llm_model,
    is_public,
    created_at,
    updated_at
FROM agents
WHERE name = $1;

-- name: CreateAgent :one
-- Create a new agent (for testing/admin)
INSERT INTO agents (
    name,
    type,
    description,
    provider,
    default_config,
    capabilities,
    llm_provider,
    llm_model,
    is_public
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- Employee Agent Configuration Queries
-- These queries manage agent assignments to employees

-- name: ListEmployeeAgentConfigs :many
-- List all agent configurations for a specific employee
SELECT
    eac.id,
    eac.employee_id,
    eac.agent_id,
    eac.config_override,
    eac.is_enabled,
    eac.sync_token,
    eac.last_synced_at,
    eac.created_at,
    eac.updated_at,
    a.name as agent_name,
    a.type as agent_type,
    a.provider as agent_provider,
    a.default_config as agent_default_config
FROM employee_agent_configs eac
JOIN agents a ON eac.agent_id = a.id
WHERE eac.employee_id = $1
ORDER BY eac.created_at DESC;

-- name: GetEmployeeAgentConfig :one
-- Get a specific agent configuration by ID
SELECT
    eac.id,
    eac.employee_id,
    eac.agent_id,
    eac.config_override,
    eac.is_enabled,
    eac.sync_token,
    eac.last_synced_at,
    eac.created_at,
    eac.updated_at,
    a.name as agent_name,
    a.type as agent_type,
    a.provider as agent_provider
FROM employee_agent_configs eac
JOIN agents a ON eac.agent_id = a.id
WHERE eac.id = $1;

-- name: CreateEmployeeAgentConfig :one
-- Create a new agent configuration for an employee
INSERT INTO employee_agent_configs (
    employee_id,
    agent_id,
    config_override,
    is_enabled
) VALUES ($1, $2, $3, $4)
RETURNING id, employee_id, agent_id, config_override, is_enabled, sync_token, last_synced_at, created_at, updated_at;

-- name: UpdateEmployeeAgentConfig :one
-- Update an existing agent configuration
UPDATE employee_agent_configs
SET
    config_override = COALESCE(sqlc.narg('config_override')::jsonb, config_override),
    is_enabled = COALESCE(sqlc.narg('is_enabled')::boolean, is_enabled),
    updated_at = NOW()
WHERE id = $1
RETURNING id, employee_id, agent_id, config_override, is_enabled, sync_token, last_synced_at, created_at, updated_at;

-- name: DeleteEmployeeAgentConfig :exec
-- Delete an agent configuration (hard delete)
DELETE FROM employee_agent_configs
WHERE id = $1;

-- name: CheckEmployeeAgentExists :one
-- Check if employee already has this agent assigned
SELECT EXISTS(
    SELECT 1 FROM employee_agent_configs
    WHERE employee_id = $1
      AND agent_id = $2
) AS exists;

-- name: CountEmployeeAgentConfigs :one
-- Count total agent configs for an employee
SELECT COUNT(*)
FROM employee_agent_configs
WHERE employee_id = $1;
