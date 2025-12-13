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
    docker_image,
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
    docker_image,
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
    docker_image,
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
    docker_image,
    default_config,
    capabilities,
    llm_provider,
    llm_model,
    is_public
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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

-- Hierarchical Config Resolution Queries
-- These queries fetch configs at different levels for merging

-- name: GetOrgAgentConfig :one
-- Get org-level config for an agent
SELECT
    id,
    org_id,
    agent_id,
    config,
    is_enabled,
    created_at,
    updated_at
FROM org_agent_configs
WHERE org_id = $1 AND agent_id = $2;

-- name: GetTeamAgentConfig :one
-- Get team-level config for an agent (requires team_id lookup)
SELECT
    id,
    team_id,
    agent_id,
    config_override,
    is_enabled,
    created_at,
    updated_at
FROM team_agent_configs
WHERE team_id = $1 AND agent_id = $2;

-- Team Agent Configuration CRUD Queries

-- name: ListTeamAgentConfigs :many
-- List all agent configurations for a specific team
SELECT
    tac.id,
    tac.team_id,
    tac.agent_id,
    tac.config_override,
    tac.is_enabled,
    tac.created_at,
    tac.updated_at,
    a.name as agent_name,
    a.type as agent_type,
    a.provider as agent_provider
FROM team_agent_configs tac
JOIN agents a ON tac.agent_id = a.id
WHERE tac.team_id = $1
ORDER BY tac.created_at DESC;

-- name: GetTeamAgentConfigByID :one
-- Get a specific team agent configuration by ID
SELECT
    tac.id,
    tac.team_id,
    tac.agent_id,
    tac.config_override,
    tac.is_enabled,
    tac.created_at,
    tac.updated_at,
    a.name as agent_name,
    a.type as agent_type,
    a.provider as agent_provider
FROM team_agent_configs tac
JOIN agents a ON tac.agent_id = a.id
WHERE tac.id = $1;

-- name: CreateTeamAgentConfig :one
-- Create a new team-level agent configuration override
INSERT INTO team_agent_configs (
    team_id,
    agent_id,
    config_override,
    is_enabled
) VALUES ($1, $2, $3, $4)
RETURNING id, team_id, agent_id, config_override, is_enabled, created_at, updated_at;

-- name: UpdateTeamAgentConfig :one
-- Update an existing team agent configuration
UPDATE team_agent_configs
SET
    config_override = COALESCE(sqlc.narg('config_override')::jsonb, config_override),
    is_enabled = COALESCE(sqlc.narg('is_enabled')::boolean, is_enabled),
    updated_at = NOW()
WHERE id = $1
RETURNING id, team_id, agent_id, config_override, is_enabled, created_at, updated_at;

-- name: DeleteTeamAgentConfig :exec
-- Delete a team agent configuration (hard delete)
DELETE FROM team_agent_configs
WHERE id = $1;

-- name: CheckTeamAgentConfigExists :one
-- Check if team already has this agent configured
SELECT EXISTS(
    SELECT 1 FROM team_agent_configs
    WHERE team_id = $1
      AND agent_id = $2
) AS exists;

-- name: GetEmployeeAgentConfigByAgent :one
-- Get employee-level config for a specific agent
SELECT
    id,
    employee_id,
    agent_id,
    config_override,
    is_enabled,
    sync_token,
    last_synced_at,
    created_at,
    updated_at
FROM employee_agent_configs
WHERE employee_id = $1 AND agent_id = $2;

-- name: GetSystemPrompts :many
-- Get all system prompts for org/team/employee + agent
-- Returns prompts ordered by scope hierarchy (org -> team -> employee) then priority
SELECT
    id,
    scope_type,
    scope_id,
    agent_id,
    prompt,
    priority,
    created_at,
    updated_at
FROM system_prompts
WHERE (scope_type = 'org' AND scope_id = $1 AND (agent_id = $2 OR agent_id IS NULL))
   OR (scope_type = 'team' AND scope_id = $3 AND (agent_id = $2 OR agent_id IS NULL))
   OR (scope_type = 'employee' AND scope_id = $4 AND (agent_id = $2 OR agent_id IS NULL))
ORDER BY
    CASE scope_type
        WHEN 'org' THEN 1
        WHEN 'team' THEN 2
        WHEN 'employee' THEN 3
    END,
    priority ASC;

-- name: ListOrgAgentConfigs :many
-- List all org-level agent configs for an organization
SELECT
    oac.id,
    oac.org_id,
    oac.agent_id,
    oac.config,
    oac.is_enabled,
    oac.created_at,
    oac.updated_at,
    a.name as agent_name,
    a.type as agent_type,
    a.provider as agent_provider,
    a.default_config as agent_default_config
FROM org_agent_configs oac
JOIN agents a ON oac.agent_id = a.id
WHERE oac.org_id = $1
ORDER BY a.name ASC;

-- name: CreateOrgAgentConfig :one
-- Create org-level agent config
INSERT INTO org_agent_configs (
    org_id,
    agent_id,
    config,
    is_enabled
) VALUES ($1, $2, $3, $4)
RETURNING id, org_id, agent_id, config, is_enabled, created_at, updated_at;

-- name: UpdateOrgAgentConfig :one
-- Update org-level agent config
UPDATE org_agent_configs
SET
    config = COALESCE(sqlc.narg('config')::jsonb, config),
    is_enabled = COALESCE(sqlc.narg('is_enabled')::boolean, is_enabled),
    updated_at = NOW()
WHERE id = $1
RETURNING id, org_id, agent_id, config, is_enabled, created_at, updated_at;

-- name: DeleteOrgAgentConfig :exec
-- Delete org-level agent config
DELETE FROM org_agent_configs
WHERE id = $1;

-- name: CheckOrgAgentConfigExists :one
-- Check if org already has this agent configured
SELECT EXISTS(
    SELECT 1 FROM org_agent_configs
    WHERE org_id = $1 AND agent_id = $2
) AS exists;


-- name: CountTeamAgentConfigs :one
-- Count agent configurations for a specific team
SELECT COUNT(*) as count
FROM team_agent_configs
WHERE team_id = $1;

