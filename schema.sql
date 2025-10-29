-- Enterprise AI Agent Management System - PostgreSQL Schema (Simplified)
-- Version: 2.0.0 (Ubik Enterprise - Configuration Management Focus)
-- Description: Multi-tenant system for managing employee AI agent and MCP configurations

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- CORE: Organizations, Teams, Employees
-- ============================================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    plan VARCHAR(50) NOT NULL DEFAULT 'starter', -- starter, professional, enterprise
    settings JSONB NOT NULL DEFAULT '{}',
    max_employees INT NOT NULL DEFAULT 10,
    max_agents_per_employee INT NOT NULL DEFAULT 3,
    claude_api_token TEXT, -- Company-wide Claude token (from 'claude setup-token')
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    plan_type VARCHAR(50) NOT NULL,
    monthly_budget_usd DECIMAL(10, 2) NOT NULL DEFAULT 100.00,
    current_spending_usd DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    billing_period_start TIMESTAMP NOT NULL,
    billing_period_end TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, cancelled
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_team_name_per_org UNIQUE (org_id, name)
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL, -- bcrypt hash
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, inactive
    preferences JSONB NOT NULL DEFAULT '{}',
    personal_claude_token TEXT, -- Employee personal Claude token (takes precedence over org token)
    last_login_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP -- Soft delete timestamp
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- AGENT CONFIGURATION
-- ============================================================================

CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    type VARCHAR(100) NOT NULL, -- claude-code, cursor, windsurf, continue, etc.
    description TEXT NOT NULL,
    provider VARCHAR(100) NOT NULL, -- anthropic, openai, custom
    default_config JSONB NOT NULL DEFAULT '{}',
    capabilities JSONB NOT NULL DEFAULT '[]',
    llm_provider VARCHAR(50) NOT NULL DEFAULT 'anthropic',
    llm_model VARCHAR(100) NOT NULL DEFAULT 'claude-3-5-sonnet-20241022',
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- fs, git, http, shell, docker, aws
    description TEXT NOT NULL,
    schema JSONB NOT NULL DEFAULT '{}',
    requires_approval BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- path_restriction, rate_limit, cost_limit, approval_required
    rules JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warn', -- block, warn, log
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Agent-Tool mappings
CREATE TABLE agent_tools (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    tool_id UUID NOT NULL REFERENCES tools(id) ON DELETE CASCADE,
    config JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (agent_id, tool_id)
);

-- Agent-Policy mappings
CREATE TABLE agent_policies (
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (agent_id, policy_id)
);

-- Team-level policy overrides
CREATE TABLE team_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    overrides JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_team_policy UNIQUE (team_id, policy_id)
);

-- ============================================================================
-- HIERARCHICAL AGENT CONFIGURATION (Org → Team → Employee)
-- ============================================================================

-- Org-level: Default agent configurations for entire organization
CREATE TABLE org_agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,

    -- Agent settings (model, temperature, max_tokens, etc.)
    config JSONB NOT NULL DEFAULT '{}',

    -- Status at org level
    is_enabled BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_org_agent UNIQUE (org_id, agent_id)
);

-- Team-level: Overrides org-level config for specific teams
-- Team members automatically inherit team's agents (access model a)
CREATE TABLE team_agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,

    -- Overrides org config (merged at resolution time)
    -- Only include fields you want to override
    config_override JSONB NOT NULL DEFAULT '{}',

    -- Status at team level (can disable for team even if org has it)
    is_enabled BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_team_agent UNIQUE (team_id, agent_id)
);

-- Employee-level: Final overrides for individual employees
-- Only needed when employee needs different config than team
CREATE TABLE employee_agent_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    agent_id UUID NOT NULL REFERENCES agents(id) ON DELETE RESTRICT,

    -- Overrides team/org config
    -- Only include fields you want to override
    config_override JSONB NOT NULL DEFAULT '{}',

    -- Status at employee level
    is_enabled BOOLEAN NOT NULL DEFAULT true,

    -- CLI sync metadata
    sync_token VARCHAR(255) UNIQUE,
    last_synced_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_employee_agent UNIQUE (employee_id, agent_id)
);

-- System Prompts: Additive across hierarchy (org + team + employee)
-- Priority determines concatenation order
CREATE TABLE system_prompts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Polymorphic scope: which level this prompt applies to
    scope_type VARCHAR(20) NOT NULL CHECK (scope_type IN ('org', 'team', 'employee')),
    scope_id UUID NOT NULL,  -- References org_id, team_id, or employee_id

    -- Which agent (NULL = all agents at this scope)
    agent_id UUID REFERENCES agents(id) ON DELETE CASCADE,

    -- The actual prompt text
    prompt TEXT NOT NULL,

    -- Priority for concatenation order (lower = higher priority)
    priority INT NOT NULL DEFAULT 0,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_system_prompt UNIQUE (scope_type, scope_id, agent_id, priority)
);

-- Index for efficient lookups during config resolution
CREATE INDEX idx_system_prompts_scope ON system_prompts(scope_type, scope_id, agent_id);

-- Employee-level policy overrides (completes the hierarchy: agent → team → employee)
CREATE TABLE employee_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    overrides JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_employee_policy UNIQUE (employee_id, policy_id)
);

-- ============================================================================
-- MCP CONFIGURATION
-- ============================================================================

CREATE TABLE mcp_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE mcp_catalog (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    provider VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    connection_schema JSONB NOT NULL,
    capabilities JSONB NOT NULL DEFAULT '[]',
    requires_credentials BOOLEAN NOT NULL DEFAULT false,
    is_approved BOOLEAN NOT NULL DEFAULT false,
    category_id UUID REFERENCES mcp_categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE employee_mcp_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    mcp_catalog_id UUID NOT NULL REFERENCES mcp_catalog(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, active, disabled, error
    connection_config JSONB NOT NULL DEFAULT '{}',
    credentials_encrypted TEXT,
    sync_token VARCHAR(255) UNIQUE,
    last_sync_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_employee_mcp UNIQUE (employee_id, mcp_catalog_id)
);

-- ============================================================================
-- APPROVAL WORKFLOWS
-- ============================================================================

CREATE TABLE agent_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    request_type VARCHAR(50) NOT NULL, -- new_agent, new_mcp, increase_budget
    request_data JSONB NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, approved, rejected, cancelled
    reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

CREATE TABLE approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id UUID NOT NULL REFERENCES agent_requests(id) ON DELETE CASCADE,
    approver_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- approved, rejected
    comment TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMP
);

-- ============================================================================
-- AUDIT & ANALYTICS
-- ============================================================================

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL, -- agent.installed, mcp.configured, config.synced
    event_category VARCHAR(50) NOT NULL, -- agent, mcp, auth, admin
    payload JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE usage_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    agent_config_id UUID REFERENCES employee_agent_configs(id) ON DELETE SET NULL,
    resource_type VARCHAR(50) NOT NULL, -- llm_tokens, api_calls, storage_mb
    quantity BIGINT NOT NULL,
    cost_usd DECIMAL(10, 4) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    token_source VARCHAR(20) DEFAULT 'company' CHECK (token_source IN ('company', 'personal')), -- Track which token was used
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- INDEXES
-- ============================================================================

-- Organizations & Teams
CREATE INDEX idx_organizations_slug ON organizations(slug);
CREATE INDEX idx_teams_org_id ON teams(org_id);

-- Employees
CREATE INDEX idx_employees_org_id ON employees(org_id);
CREATE INDEX idx_employees_team_id ON employees(team_id);
CREATE INDEX idx_employees_email ON employees(email);
CREATE INDEX idx_employees_status ON employees(status);

-- Sessions
CREATE INDEX idx_sessions_employee_id ON sessions(employee_id);
CREATE INDEX idx_sessions_token_hash ON sessions(token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Agent Configs
CREATE INDEX idx_employee_agent_configs_employee_id ON employee_agent_configs(employee_id);
CREATE INDEX idx_employee_agent_configs_agent_id ON employee_agent_configs(agent_id);
CREATE INDEX idx_employee_agent_configs_is_enabled ON employee_agent_configs(is_enabled);
CREATE INDEX idx_employee_agent_configs_sync_token ON employee_agent_configs(sync_token) WHERE sync_token IS NOT NULL;

-- MCP Configs
CREATE INDEX idx_employee_mcp_configs_employee_id ON employee_mcp_configs(employee_id);
CREATE INDEX idx_employee_mcp_configs_mcp_catalog_id ON employee_mcp_configs(mcp_catalog_id);
CREATE INDEX idx_employee_mcp_configs_status ON employee_mcp_configs(status);
CREATE INDEX idx_employee_mcp_configs_sync_token ON employee_mcp_configs(sync_token) WHERE sync_token IS NOT NULL;

-- Activity Logs
CREATE INDEX idx_activity_logs_org_id ON activity_logs(org_id);
CREATE INDEX idx_activity_logs_employee_id ON activity_logs(employee_id);
CREATE INDEX idx_activity_logs_event_type ON activity_logs(event_type);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);

-- Usage Records
CREATE INDEX idx_usage_records_org_id ON usage_records(org_id);
CREATE INDEX idx_usage_records_employee_id ON usage_records(employee_id);
CREATE INDEX idx_usage_records_agent_config_id ON usage_records(agent_config_id);
CREATE INDEX idx_usage_records_period ON usage_records(period_start, period_end);

-- Agent Requests
CREATE INDEX idx_agent_requests_employee_id ON agent_requests(employee_id);
CREATE INDEX idx_agent_requests_status ON agent_requests(status);
CREATE INDEX idx_agent_requests_created_at ON agent_requests(created_at DESC);

-- Approvals
CREATE INDEX idx_approvals_request_id ON approvals(request_id);
CREATE INDEX idx_approvals_approver_id ON approvals(approver_id);

-- ============================================================================
-- ROW-LEVEL SECURITY
-- ============================================================================

ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_agent_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_mcp_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;

-- Example RLS policy (requires app to set current_setting('app.current_org_id'))
-- CREATE POLICY org_isolation ON employees
--     FOR ALL
--     USING (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_organizations_updated_at BEFORE UPDATE ON organizations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_agents_updated_at BEFORE UPDATE ON agents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tools_updated_at BEFORE UPDATE ON tools
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employee_agent_configs_updated_at BEFORE UPDATE ON employee_agent_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_mcp_catalog_updated_at BEFORE UPDATE ON mcp_catalog
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employee_mcp_configs_updated_at BEFORE UPDATE ON employee_mcp_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Generate sync tokens automatically
CREATE OR REPLACE FUNCTION generate_sync_token()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.sync_token IS NULL THEN
        NEW.sync_token := encode(gen_random_bytes(32), 'hex');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER generate_agent_config_sync_token BEFORE INSERT ON employee_agent_configs
    FOR EACH ROW EXECUTE FUNCTION generate_sync_token();

CREATE TRIGGER generate_mcp_config_sync_token BEFORE INSERT ON employee_mcp_configs
    FOR EACH ROW EXECUTE FUNCTION generate_sync_token();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Roles
INSERT INTO roles (name, description, permissions) VALUES
    ('admin', 'Full system access', '["*"]'),
    ('manager', 'Team management and approvals', '["teams:manage", "approvals:review", "agents:assign"]'),
    ('developer', 'Standard access', '["agents:use", "mcps:request"]'),
    ('viewer', 'Read-only access', '["analytics:view"]');

-- MCP Categories
INSERT INTO mcp_categories (name, description) VALUES
    ('Development', 'Code-related tools and IDEs'),
    ('Data', 'Database and data processing'),
    ('Cloud', 'Cloud provider integrations'),
    ('Communication', 'Messaging and notifications'),
    ('Productivity', 'General productivity tools');

-- Tools
INSERT INTO tools (name, type, description, schema, requires_approval) VALUES
    ('filesystem', 'fs', 'Read and write files', '{"type":"object","properties":{"path":{"type":"string"}}}', false),
    ('git', 'git', 'Git operations', '{"type":"object","properties":{"command":{"type":"string"}}}', false),
    ('http', 'http', 'HTTP requests', '{"type":"object","properties":{"url":{"type":"string"}}}', false),
    ('shell', 'shell', 'Execute shell commands', '{"type":"object","properties":{"command":{"type":"string"}}}', true),
    ('docker', 'docker', 'Docker operations', '{"type":"object","properties":{"command":{"type":"string"}}}', true);

-- Policies
INSERT INTO policies (name, type, rules, severity) VALUES
    ('restricted_paths', 'path_restriction', '{"denied_paths":["/etc","/root","/sys"]}', 'block'),
    ('rate_limit_basic', 'rate_limit', '{"max_requests_per_hour":100}', 'warn'),
    ('cost_limit_daily', 'cost_limit', '{"max_usd_per_day":10.0}', 'warn'),
    ('approval_required_prod', 'approval_required', '{"patterns":[".*prod.*",".*production.*"]}', 'block');

-- Agents (Popular AI coding assistants)
INSERT INTO agents (name, type, description, provider, default_config, capabilities, llm_provider, llm_model) VALUES
    ('Claude Code', 'claude-code', 'Anthropic Claude Code CLI', 'anthropic', '{"temperature":0.2}', '["code_generation","debugging","refactoring","research"]', 'anthropic', 'claude-3-5-sonnet-20241022'),
    ('Cursor', 'cursor', 'Cursor AI IDE', 'cursor', '{"temperature":0.3}', '["code_generation","autocomplete","chat"]', 'openai', 'gpt-4o'),
    ('Windsurf', 'windsurf', 'Windsurf AI IDE', 'codeium', '{"temperature":0.2}', '["code_generation","chat","cascade"]', 'anthropic', 'claude-3-5-sonnet-20241022'),
    ('Continue', 'continue', 'Continue VS Code Extension', 'continue', '{"temperature":0.2}', '["autocomplete","chat","edit"]', 'anthropic', 'claude-3-5-sonnet-20241022'),
    ('GitHub Copilot', 'copilot', 'GitHub Copilot', 'github', '{"temperature":0.3}', '["autocomplete","chat"]', 'openai', 'gpt-4o');

-- Link agents with tools
INSERT INTO agent_tools (agent_id, tool_id)
SELECT a.id, t.id
FROM agents a, tools t
WHERE (a.name = 'Claude Code' AND t.name IN ('filesystem', 'git', 'http'))
   OR (a.name = 'Cursor' AND t.name IN ('filesystem', 'git'))
   OR (a.name = 'Windsurf' AND t.name IN ('filesystem', 'git'))
   OR (a.name = 'Continue' AND t.name IN ('filesystem', 'git'))
   OR (a.name = 'GitHub Copilot' AND t.name IN ('filesystem'));

-- Link agents with policies
INSERT INTO agent_policies (agent_id, policy_id)
SELECT a.id, p.id
FROM agents a, policies p
WHERE p.name IN ('restricted_paths', 'rate_limit_basic', 'cost_limit_daily');

-- Sample MCP Servers
INSERT INTO mcp_catalog (name, provider, version, description, connection_schema, capabilities, requires_credentials, is_approved, category_id) VALUES
    ('Filesystem', '@modelcontextprotocol/server-filesystem', '1.0.0', 'Local filesystem access', '{"type":"object","properties":{"rootPath":{"type":"string"}}}', '["read_file","write_file","list_dir"]', false, true, (SELECT id FROM mcp_categories WHERE name = 'Development')),
    ('GitHub', '@modelcontextprotocol/server-github', '1.0.0', 'GitHub API integration', '{"type":"object","properties":{"token":{"type":"string"}}}', '["repo_access","pr_management","issues"]', true, true, (SELECT id FROM mcp_categories WHERE name = 'Development')),
    ('PostgreSQL', '@modelcontextprotocol/server-postgres', '1.0.0', 'PostgreSQL database access', '{"type":"object","properties":{"connectionString":{"type":"string"}}}', '["query","schema_inspection"]', true, true, (SELECT id FROM mcp_categories WHERE name = 'Data')),
    ('Slack', '@modelcontextprotocol/server-slack', '1.0.0', 'Slack messaging', '{"type":"object","properties":{"token":{"type":"string"}}}', '["send_message","read_channels"]', true, false, (SELECT id FROM mcp_categories WHERE name = 'Communication'));

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- Employee agent configurations with catalog details
-- Employee agents view (simplified - shows only employee-level configs)
-- For full resolution (org + team + employee), use application code
CREATE VIEW v_employee_agents AS
SELECT
    eac.id,
    eac.employee_id,
    e.full_name as employee_name,
    e.email as employee_email,
    a.name as agent_name,
    a.type as agent_type,
    a.provider,
    eac.is_enabled,
    eac.config_override,
    eac.sync_token,
    eac.last_synced_at,
    eac.created_at
FROM employee_agent_configs eac
JOIN employees e ON eac.employee_id = e.id
JOIN agents a ON eac.agent_id = a.id;

-- Employee MCP configurations with catalog details
CREATE VIEW v_employee_mcps AS
SELECT 
    emc.id,
    emc.employee_id,
    e.full_name as employee_name,
    e.email as employee_email,
    mc.name as mcp_name,
    mc.provider,
    mc.version,
    emc.status,
    emc.sync_token,
    emc.last_sync_at,
    emc.created_at
FROM employee_mcp_configs emc
JOIN employees e ON emc.employee_id = e.id
JOIN mcp_catalog mc ON emc.mcp_catalog_id = mc.id;

-- Pending approval requests with requester details
CREATE VIEW v_pending_approvals AS
SELECT 
    ar.id as request_id,
    ar.request_type,
    ar.request_data,
    ar.reason,
    ar.created_at as requested_at,
    e.id as employee_id,
    e.full_name as requester_name,
    e.email as requester_email,
    t.name as team_name,
    o.name as org_name
FROM agent_requests ar
JOIN employees e ON ar.employee_id = e.id
JOIN organizations o ON e.org_id = o.id
LEFT JOIN teams t ON e.team_id = t.id
WHERE ar.status = 'pending';

COMMENT ON VIEW v_employee_agents IS 'Complete view of employee agent configurations with catalog details';
COMMENT ON VIEW v_employee_mcps IS 'Complete view of employee MCP configurations with catalog details';
COMMENT ON VIEW v_pending_approvals IS 'Pending approval requests with full requester context';

-- ============================================================================
-- CLAUDE TOKEN MANAGEMENT (Hybrid Auth Model)
-- ============================================================================

-- Index for quick lookup of employees with personal tokens
CREATE INDEX idx_employees_personal_token ON employees(org_id, personal_claude_token)
WHERE personal_claude_token IS NOT NULL;

-- Function to get effective Claude token for an employee
CREATE OR REPLACE FUNCTION get_effective_claude_token(emp_id UUID)
RETURNS TABLE (
    token TEXT,
    source VARCHAR(20),
    org_id UUID
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(e.personal_claude_token, o.claude_api_token) as token,
        CASE
            WHEN e.personal_claude_token IS NOT NULL THEN 'personal'::VARCHAR(20)
            ELSE 'company'::VARCHAR(20)
        END as source,
        e.org_id
    FROM employees e
    JOIN organizations o ON e.org_id = o.id
    WHERE e.id = emp_id
    AND e.deleted_at IS NULL;
END;
$$ LANGUAGE plpgsql STABLE;

COMMENT ON FUNCTION get_effective_claude_token(UUID) IS
'Returns the effective Claude token for an employee (personal if available, otherwise company token)';
-- Test comment
