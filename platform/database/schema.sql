-- Arfa - PostgreSQL Schema (OSS v1.0)
-- Version: 1.0.0 - Open Source Release
-- Description: Multi-tenant system for AI agent security monitoring and policy enforcement

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- CORE: Organizations, Teams, Employees
-- ============================================================================

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    settings JSONB NOT NULL DEFAULT '{}',
    max_employees INT NOT NULL DEFAULT 10,
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
    password_hash VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- active, suspended, inactive
    preferences JSONB NOT NULL DEFAULT '{}',
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

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- POLICIES (Tool Blocking)
-- ============================================================================

CREATE TABLE policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    type VARCHAR(50) NOT NULL, -- path_restriction, rate_limit, cost_limit, approval_required
    rules JSONB NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'warn', -- block, warn, log
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Tool policies for blocking/auditing LLM tool usage
CREATE TABLE tool_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE CASCADE,

    -- What to match
    tool_name VARCHAR(255) NOT NULL,  -- "Bash", "Read", "mcp__playwright__%", "*"

    -- Conditions (optional, for param-based blocking)
    conditions JSONB,  -- {"any": [{"param_path": "command", "operator": "contains", "value": "rm -rf"}]}

    -- Action
    action VARCHAR(20) NOT NULL DEFAULT 'deny' CHECK (action IN ('deny', 'audit')),
    reason TEXT,  -- Human-readable explanation shown to agent

    -- Metadata
    created_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT tool_policies_has_org CHECK (org_id IS NOT NULL)
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

-- Employee-level policy overrides
CREATE TABLE employee_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    policy_id UUID NOT NULL REFERENCES policies(id) ON DELETE CASCADE,
    overrides JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_employee_policy UNIQUE (employee_id, policy_id)
);

-- ============================================================================
-- ACTIVITY LOGS
-- ============================================================================

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    session_id UUID, -- CLI session tracking

    -- Client detection
    client_name VARCHAR(100),      -- e.g., "claude-code", "cursor", "continue"
    client_version VARCHAR(50),    -- e.g., "1.0.25"

    -- Event details
    event_type VARCHAR(100) NOT NULL, -- tool_call, policy_violation, session_start, session_end
    event_category VARCHAR(50) NOT NULL, -- classified, raw, auth, admin
    content TEXT, -- Actual I/O text for input/output/error events
    payload JSONB NOT NULL DEFAULT '{}', -- Metadata: tool_name, tool_input, blocked, etc.
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ============================================================================
-- WEBHOOKS (SIEM Integration)
-- ============================================================================

CREATE TABLE webhook_destinations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,

    -- Destination config
    name VARCHAR(100) NOT NULL,
    url TEXT NOT NULL,

    -- Authentication
    auth_type VARCHAR(50) NOT NULL DEFAULT 'none' CHECK (auth_type IN ('none', 'bearer', 'header', 'basic')),
    auth_config JSONB DEFAULT '{}',

    -- Event filtering
    event_types TEXT[] DEFAULT '{}',  -- Empty = all, or specific: ['tool_call', 'policy_violation']
    event_filter JSONB DEFAULT '{}',  -- {"blocked": true} = only blocked events

    -- Delivery settings
    enabled BOOLEAN NOT NULL DEFAULT true,
    batch_size INT NOT NULL DEFAULT 1 CHECK (batch_size >= 1 AND batch_size <= 100),
    timeout_ms INT NOT NULL DEFAULT 5000 CHECK (timeout_ms >= 1000 AND timeout_ms <= 30000),
    retry_max INT NOT NULL DEFAULT 3 CHECK (retry_max >= 0 AND retry_max <= 10),
    retry_backoff_ms INT NOT NULL DEFAULT 1000 CHECK (retry_backoff_ms >= 100),

    -- Security
    signing_secret VARCHAR(255),  -- For X-Arfa-Signature header (HMAC)

    -- Metadata
    created_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    UNIQUE(org_id, name)
);

CREATE TABLE webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    destination_id UUID NOT NULL REFERENCES webhook_destinations(id) ON DELETE CASCADE,
    log_id UUID NOT NULL REFERENCES activity_logs(id) ON DELETE CASCADE,

    -- Delivery status
    status VARCHAR(50) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'delivered', 'failed', 'dead')),
    attempts INT NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    next_retry_at TIMESTAMP,

    -- Response info
    response_status INT,
    response_body TEXT,
    error_message TEXT,

    -- Timing
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP,

    UNIQUE(destination_id, log_id)
);

-- ============================================================================
-- INVITATIONS & ONBOARDING
-- ============================================================================

CREATE TABLE invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id),
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    token VARCHAR(64) NOT NULL UNIQUE,
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, accepted, expired, cancelled
    expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '7 days'),
    accepted_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    accepted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_pending_invitation UNIQUE (org_id, email, status)
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

-- Password Reset Tokens
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_employee_id ON password_reset_tokens(employee_id);

-- Tool Policies
CREATE INDEX idx_tool_policies_org_id ON tool_policies(org_id);
CREATE INDEX idx_tool_policies_team_id ON tool_policies(team_id) WHERE team_id IS NOT NULL;
CREATE INDEX idx_tool_policies_employee_id ON tool_policies(employee_id) WHERE employee_id IS NOT NULL;
CREATE INDEX idx_tool_policies_lookup ON tool_policies(org_id, team_id, employee_id, tool_name);

-- Activity Logs
CREATE INDEX idx_activity_logs_org_id ON activity_logs(org_id);
CREATE INDEX idx_activity_logs_employee_id ON activity_logs(employee_id);
CREATE INDEX idx_activity_logs_session_id ON activity_logs(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_activity_logs_client ON activity_logs(client_name);
CREATE INDEX idx_activity_logs_client_version ON activity_logs(client_name, client_version);
CREATE INDEX idx_activity_logs_event_type ON activity_logs(event_type);
CREATE INDEX idx_activity_logs_created_at ON activity_logs(created_at DESC);
CREATE INDEX idx_activity_logs_session_created ON activity_logs(session_id, created_at) WHERE session_id IS NOT NULL;

-- Invitations
CREATE INDEX idx_invitations_org_id ON invitations(org_id);
CREATE INDEX idx_invitations_email ON invitations(email);
CREATE INDEX idx_invitations_token ON invitations(token);
CREATE INDEX idx_invitations_status ON invitations(status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);
CREATE INDEX idx_invitations_inviter_id ON invitations(inviter_id);

-- Webhooks
CREATE INDEX idx_webhook_destinations_org_id ON webhook_destinations(org_id);
CREATE INDEX idx_webhook_destinations_enabled ON webhook_destinations(org_id, enabled) WHERE enabled = true;
CREATE INDEX idx_webhook_deliveries_pending ON webhook_deliveries(status, next_retry_at)
    WHERE status IN ('pending', 'failed');
CREATE INDEX idx_webhook_deliveries_destination ON webhook_deliveries(destination_id);
CREATE INDEX idx_webhook_deliveries_log ON webhook_deliveries(log_id);

-- ============================================================================
-- ROW-LEVEL SECURITY
-- ============================================================================

ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE invitations ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_destinations ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_deliveries ENABLE ROW LEVEL SECURITY;

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

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employees_updated_at BEFORE UPDATE ON employees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_policies_updated_at BEFORE UPDATE ON policies
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_invitations_updated_at BEFORE UPDATE ON invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Generate invitation tokens automatically
CREATE OR REPLACE FUNCTION generate_invitation_token()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.token IS NULL OR NEW.token = '' THEN
        NEW.token := encode(gen_random_bytes(32), 'hex');
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER generate_invitation_token_trigger BEFORE INSERT ON invitations
    FOR EACH ROW EXECUTE FUNCTION generate_invitation_token();

-- Expire old invitations automatically
CREATE OR REPLACE FUNCTION expire_old_invitations()
RETURNS void AS $$
BEGIN
    UPDATE invitations
    SET status = 'expired'
    WHERE status = 'pending'
    AND expires_at < NOW();
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Default Roles (required for signup)
INSERT INTO roles (name, description, permissions) VALUES
    ('admin', 'Full administrative access', '["*"]'),
    ('member', 'Standard employee access', '["logs:read", "policies:read"]'),
    ('viewer', 'Read-only access', '["logs:read"]');

-- Default Policies
INSERT INTO policies (name, type, rules, severity) VALUES
    ('restricted_paths', 'path_restriction', '{"denied_paths":["/etc","/root","/sys"]}', 'block'),
    ('rate_limit_basic', 'rate_limit', '{"max_requests_per_hour":100}', 'warn');
