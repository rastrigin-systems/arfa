-- Claude Token Management Queries
-- These queries manage the hybrid authentication model for Claude Code

-- ============================================================================
-- ORGANIZATION TOKEN MANAGEMENT
-- ============================================================================

-- name: SetOrganizationClaudeToken :exec
UPDATE organizations
SET claude_api_token = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetOrganizationClaudeToken :one
SELECT id, name, claude_api_token, updated_at
FROM organizations
WHERE id = $1;

-- name: DeleteOrganizationClaudeToken :exec
UPDATE organizations
SET claude_api_token = NULL,
    updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- EMPLOYEE TOKEN MANAGEMENT
-- ============================================================================

-- name: SetEmployeePersonalToken :exec
UPDATE employees
SET personal_claude_token = $2,
    updated_at = NOW()
WHERE id = $1
AND deleted_at IS NULL;

-- name: GetEmployeePersonalToken :one
SELECT id, full_name, email, personal_claude_token, updated_at
FROM employees
WHERE id = $1
AND deleted_at IS NULL;

-- name: DeleteEmployeePersonalToken :exec
UPDATE employees
SET personal_claude_token = NULL,
    updated_at = NOW()
WHERE id = $1
AND deleted_at IS NULL;

-- ============================================================================
-- TOKEN RESOLUTION
-- ============================================================================

-- name: GetEffectiveClaudeToken :one
SELECT
    COALESCE(e.personal_claude_token, o.claude_api_token) as token,
    CASE
        WHEN e.personal_claude_token IS NOT NULL THEN 'personal'
        ELSE 'company'
    END as source,
    e.org_id,
    e.id as employee_id,
    o.name as org_name
FROM employees e
JOIN organizations o ON e.org_id = o.id
WHERE e.id = $1
AND e.deleted_at IS NULL;

-- name: GetEmployeeTokenStatus :one
SELECT
    e.id as employee_id,
    e.full_name,
    e.personal_claude_token IS NOT NULL as has_personal_token,
    o.claude_api_token IS NOT NULL as has_company_token,
    CASE
        WHEN e.personal_claude_token IS NOT NULL THEN 'personal'
        WHEN o.claude_api_token IS NOT NULL THEN 'company'
        ELSE 'none'
    END as active_token_source
FROM employees e
JOIN organizations o ON e.org_id = o.id
WHERE e.id = $1
AND e.deleted_at IS NULL;

-- ============================================================================
-- TOKEN STATISTICS
-- ============================================================================

-- name: CountEmployeesWithPersonalTokens :one
SELECT COUNT(*) as count
FROM employees
WHERE org_id = $1
AND personal_claude_token IS NOT NULL
AND deleted_at IS NULL;

-- name: GetTokenUsageBySource :many
SELECT
    token_source,
    COUNT(*) as usage_count,
    SUM(cost_usd) as total_cost
FROM usage_records
WHERE org_id = $1
AND period_start >= $2
AND period_end <= $3
GROUP BY token_source;
