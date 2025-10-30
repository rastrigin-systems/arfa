-- ============================================================================
-- SEED DATA FOR UBIK ENTERPRISE - AI AGENT MANAGEMENT PLATFORM
-- ============================================================================
--
-- Purpose: Realistic test data demonstrating complete user journeys
-- Password for all users: "password123"
-- Password hash: $2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC
--
-- USER JOURNEYS INCLUDED:
-- 1. Acme Corp (Mature Enterprise) - Fully configured with teams and agents
-- 2. TechCo (Growing Startup) - Mid-setup, some pending approvals
-- 3. NewCorp (Brand New) - Just registered, minimal setup
-- 4. Solo Startup (Testing Platform) - Single user testing features
--
-- ============================================================================

-- Temporarily disable Row-Level Security for seeding
ALTER TABLE teams DISABLE ROW LEVEL SECURITY;
ALTER TABLE employees DISABLE ROW LEVEL SECURITY;
ALTER TABLE employee_agent_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE employee_mcp_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records DISABLE ROW LEVEL SECURITY;
ALTER TABLE agent_requests DISABLE ROW LEVEL SECURITY;
ALTER TABLE approvals DISABLE ROW LEVEL SECURITY;

-- Clean existing data (in reverse dependency order)
TRUNCATE TABLE approvals CASCADE;
TRUNCATE TABLE agent_requests CASCADE;
TRUNCATE TABLE usage_records CASCADE;
TRUNCATE TABLE activity_logs CASCADE;
TRUNCATE TABLE employee_mcp_configs CASCADE;
TRUNCATE TABLE employee_agent_configs CASCADE;
TRUNCATE TABLE team_agent_configs CASCADE;
TRUNCATE TABLE org_agent_configs CASCADE;
TRUNCATE TABLE sessions CASCADE;
TRUNCATE TABLE employees CASCADE;
TRUNCATE TABLE teams CASCADE;
TRUNCATE TABLE roles CASCADE;
TRUNCATE TABLE subscriptions CASCADE;
TRUNCATE TABLE organizations CASCADE;
TRUNCATE TABLE mcp_catalog CASCADE;
TRUNCATE TABLE mcp_categories CASCADE;
TRUNCATE TABLE agents CASCADE;

-- ============================================================================
-- GLOBAL RESOURCES (Available to all organizations)
-- ============================================================================

-- Roles
INSERT INTO roles (id, name, permissions) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Owner', '["*"]'::jsonb),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Admin', '["read", "write", "delete", "manage_employees", "manage_teams", "approve_requests"]'::jsonb),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Manager', '["read", "write", "manage_team", "approve_requests"]'::jsonb),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Developer', '["read", "write", "request_agents"]'::jsonb),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Viewer', '["read"]'::jsonb);

-- Agents Catalog
INSERT INTO agents (id, name, type, description, provider, default_config, capabilities, llm_provider, llm_model, is_public) VALUES
('a1111111-1111-1111-1111-111111111111', 'Claude Code', 'ide_assistant',
 'AI-powered coding assistant with deep codebase understanding', 'Anthropic',
 '{"max_tokens": 8000, "temperature": 0.7}'::jsonb,
 '["code_completion", "chat", "refactoring", "documentation"]'::jsonb,
 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a2222222-2222-2222-2222-222222222222', 'Cursor', 'ide_assistant',
 'AI-first code editor with pair programming', 'Anysphere',
 '{"max_tokens": 4000, "temperature": 0.5}'::jsonb,
 '["code_completion", "chat", "multi_file_edit"]'::jsonb,
 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a3333333-3333-3333-3333-333333333333', 'Windsurf', 'ide_assistant',
 'AI-powered coding with flow state', 'Codeium',
 '{"max_tokens": 4000}'::jsonb,
 '["code_completion", "suggestions", "cascade_edit"]'::jsonb,
 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a4444444-4444-4444-4444-444444444444', 'GitHub Copilot', 'code_completion',
 'AI pair programmer from GitHub', 'GitHub',
 '{"enable_completions": true, "enable_chat": true}'::jsonb,
 '["code_completion", "chat"]'::jsonb,
 'openai', 'gpt-4', true),

('a5555555-5555-5555-5555-555555555555', 'Continue', 'ide_assistant',
 'Open source AI code assistant', 'Continue.dev',
 '{"max_tokens": 4000}'::jsonb,
 '["code_completion", "chat", "custom_commands"]'::jsonb,
 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a6666666-6666-6666-6666-666666666666', 'Aider', 'terminal_assistant',
 'AI pair programming in your terminal', 'Aider',
 '{"auto_commits": false}'::jsonb,
 '["code_generation", "refactoring", "git_integration"]'::jsonb,
 'anthropic', 'claude-3-5-sonnet-20241022', true);

-- MCP Categories
INSERT INTO mcp_categories (id, name, description) VALUES
('c1111111-1111-1111-1111-111111111111', 'Development Tools', 'Tools for software development'),
('c2222222-2222-2222-2222-222222222222', 'Data Sources', 'Database and API integrations'),
('c3333333-3333-3333-3333-333333333333', 'Cloud Services', 'Cloud platform integrations');

-- MCP Catalog
INSERT INTO mcp_catalog (id, category_id, name, description, provider, version, connection_schema, capabilities, requires_credentials, is_approved) VALUES
('0a111111-1111-1111-1111-111111111111', 'c1111111-1111-1111-1111-111111111111', 'Filesystem',
 'Access local filesystem with read/write capabilities', 'Anthropic', '1.0.0',
 '{"type": "object", "properties": {"allowed_directories": {"type": "array"}}}'::jsonb,
 '["read", "write", "list"]'::jsonb, false, true),

('0a222222-2222-2222-2222-222222222222', 'c1111111-1111-1111-1111-111111111111', 'Git',
 'Git repository operations and history', 'Anthropic', '1.0.0',
 '{"type": "object", "properties": {"repository_path": {"type": "string"}}}'::jsonb,
 '["read", "commit", "branch"]'::jsonb, false, true),

('0a333333-3333-3333-3333-333333333333', 'c1111111-1111-1111-1111-111111111111', 'GitHub',
 'GitHub API integration for issues, PRs, repos', 'Anthropic', '1.0.0',
 '{"type": "object", "properties": {"github_token": {"type": "string"}}}'::jsonb,
 '["issues", "pull_requests", "repositories"]'::jsonb, true, true),

('0a444444-4444-4444-4444-444444444444', 'c2222222-2222-2222-2222-222222222222', 'PostgreSQL',
 'PostgreSQL database queries and operations', 'Anthropic', '1.0.0',
 '{"type": "object", "properties": {"connection_string": {"type": "string"}}}'::jsonb,
 '["query", "schema"]'::jsonb, true, true);

-- ============================================================================
-- JOURNEY 1: ACME CORP - Mature Enterprise (Fully Configured)
-- ============================================================================

INSERT INTO organizations (id, name, slug, plan, settings, max_employees, max_agents_per_employee, claude_api_token) VALUES
('11111111-1111-1111-1111-111111111111', 'Acme Corporation', 'acme-corp', 'enterprise',
 '{"features": ["sso", "audit_logs", "custom_policies"], "sso_provider": "okta"}'::jsonb,
 500, 10, 'sk-ant-api03-acme-company-token');

INSERT INTO subscriptions (org_id, plan_type, monthly_budget_usd, current_spending_usd, billing_period_start, billing_period_end, status) VALUES
('11111111-1111-1111-1111-111111111111', 'enterprise', 10000.00, 3456.78,
 '2025-10-01 00:00:00', '2025-10-31 23:59:59', 'active');

-- Teams (using valid UUIDs with numeric/hex prefixes)
INSERT INTO teams (id, org_id, name, description) VALUES
('01111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'Platform Engineering', 'Core platform and infrastructure'),
('02222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'Frontend Team', 'Web and mobile UI development'),
('03333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'Data Team', 'Analytics and ML engineering');

-- Employees
INSERT INTO employees (id, org_id, role_id, team_id, email, full_name, password_hash, status, preferences) VALUES
('e1111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL,
 'sarah.cto@acme.com', 'Sarah Chen (CTO)', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active',
 '{"theme": "dark", "notifications": true}'::jsonb),

('e2222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '01111111-1111-1111-1111-111111111111',
 'alex.manager@acme.com', 'Alex Kumar (Eng Manager)', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),

('e3333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '01111111-1111-1111-1111-111111111111',
 'maria.senior@acme.com', 'Maria Rodriguez (Senior SWE)', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active',
 '{"theme": "dark", "editor": "vscode"}'::jsonb),

('e4444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '02222222-2222-2222-2222-222222222222',
 'emma.frontend@acme.com', 'Emma Thompson (Frontend Dev)', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb);

-- Org-level Agent Configurations
INSERT INTO org_agent_configs (org_id, agent_id, config, is_enabled) VALUES
('11111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
 '{"max_tokens": 8000, "temperature": 0.7}'::jsonb, true),
('11111111-1111-1111-1111-111111111111', 'a2222222-2222-2222-2222-222222222222',
 '{"max_tokens": 4000}'::jsonb, true);

-- Team-level Agent Configurations
INSERT INTO team_agent_configs (team_id, agent_id, config_override, is_enabled) VALUES
('01111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
 '{"temperature": 0.5}'::jsonb, true);

-- Employee-level Agent Configurations
INSERT INTO employee_agent_configs (employee_id, agent_id, config_override, is_enabled) VALUES
('e3333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111',
 '{"max_tokens": 12000}'::jsonb, true),
('e3333333-3333-3333-3333-333333333333', 'a2222222-2222-2222-2222-222222222222',
 '{}'::jsonb, true);

-- Employee MCP Configurations
INSERT INTO employee_mcp_configs (employee_id, mcp_catalog_id, status, connection_config) VALUES
('e3333333-3333-3333-3333-333333333333', '0a111111-1111-1111-1111-111111111111', 'active',
 '{"allowed_directories": ["/home/maria/projects"]}'::jsonb),
('e3333333-3333-3333-3333-333333333333', '0a222222-2222-2222-2222-222222222222', 'active',
 '{"repository_path": "/home/maria/projects/main-repo"}'::jsonb);

-- Activity Logs
INSERT INTO activity_logs (org_id, employee_id, event_type, event_category, payload) VALUES
('11111111-1111-1111-1111-111111111111', 'e3333333-3333-3333-3333-333333333333', 'agent.configured', 'agent',
 '{"agent_name": "Claude Code", "action": "enabled"}'::jsonb),
('11111111-1111-1111-1111-111111111111', 'e4444444-4444-4444-4444-444444444444', 'mcp.configured', 'mcp',
 '{"mcp_name": "Filesystem", "action": "connected"}'::jsonb);

-- Usage Records
INSERT INTO usage_records (org_id, employee_id, agent_config_id, resource_type, quantity, cost_usd, period_start, period_end, metadata, token_source) VALUES
('11111111-1111-1111-1111-111111111111', 'e3333333-3333-3333-3333-333333333333',
 (SELECT id FROM employee_agent_configs WHERE employee_id = 'e3333333-3333-3333-3333-333333333333' AND agent_id = 'a1111111-1111-1111-1111-111111111111' LIMIT 1),
 'llm_tokens', 125000, 12.50, '2025-10-01 00:00:00', '2025-10-07 23:59:59',
 '{"model": "claude-3-5-sonnet-20241022", "input_tokens": 100000, "output_tokens": 25000}'::jsonb, 'company');

-- ============================================================================
-- JOURNEY 2: TECHCO - Growing Startup (Mid-Setup, Pending Approvals)
-- ============================================================================

INSERT INTO organizations (id, name, slug, plan, max_employees, max_agents_per_employee) VALUES
('22222222-2222-2222-2222-222222222222', 'TechCo Inc', 'techco', 'professional', 50, 5);

INSERT INTO subscriptions (org_id, plan_type, monthly_budget_usd, current_spending_usd, billing_period_start, billing_period_end, status) VALUES
('22222222-2222-2222-2222-222222222222', 'professional', 1000.00, 234.56,
 '2025-10-01 00:00:00', '2025-10-31 23:59:59', 'active');

INSERT INTO teams (id, org_id, name, description) VALUES
('04444444-4444-4444-4444-444444444444', '22222222-2222-2222-2222-222222222222', 'Engineering', 'Product development team');

INSERT INTO employees (id, org_id, role_id, team_id, email, full_name, password_hash, status) VALUES
('e5555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL,
 'jane.founder@techco.com', 'Jane Founder', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active'),

('e6666666-6666-6666-6666-666666666666', '22222222-2222-2222-2222-222222222222', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '04444444-4444-4444-4444-444444444444',
 'tom.dev@techco.com', 'Tom Developer', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active');

-- Basic org config
INSERT INTO org_agent_configs (org_id, agent_id, config, is_enabled) VALUES
('22222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111',
 '{"max_tokens": 4000}'::jsonb, true);

-- Employee config
INSERT INTO employee_agent_configs (employee_id, agent_id, config_override, is_enabled) VALUES
('e5555555-5555-5555-5555-555555555555', 'a1111111-1111-1111-1111-111111111111', '{}'::jsonb, true);

-- Pending agent request (Tom wants Cursor)
INSERT INTO agent_requests (id, employee_id, request_type, request_data, status, reason) VALUES
('0b111111-1111-1111-1111-111111111111', 'e6666666-6666-6666-6666-666666666666', 'new_agent',
 '{"agent_id": "a2222222-2222-2222-2222-222222222222", "agent_name": "Cursor", "justification": "Need for pair programming"}'::jsonb,
 'pending', 'Want to try Cursor for better code completion');

-- ============================================================================
-- JOURNEY 3: NEWCORP - Brand New Company (Just Registered)
-- ============================================================================

INSERT INTO organizations (id, name, slug, plan, max_employees, max_agents_per_employee) VALUES
('33333333-3333-3333-3333-333333333333', 'NewCorp LLC', 'newcorp', 'starter', 10, 2);

INSERT INTO subscriptions (org_id, plan_type, monthly_budget_usd, billing_period_start, billing_period_end, status) VALUES
('33333333-3333-3333-3333-333333333333', 'trial', 0.00,
 '2025-10-25 00:00:00', '2025-11-25 23:59:59', 'trial');

INSERT INTO employees (id, org_id, role_id, team_id, email, full_name, password_hash, status) VALUES
('e7777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL,
 'owner@newcorp.com', 'New Owner', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active');

-- No agent configs yet (brand new)

-- ============================================================================
-- JOURNEY 4: SOLO STARTUP - Testing Platform (Solo Developer)
-- ============================================================================

INSERT INTO organizations (id, name, slug, plan, max_employees, max_agents_per_employee, claude_api_token) VALUES
('44444444-4444-4444-4444-444444444444', 'Solo Startup', 'solo-startup', 'professional', 10, 5,
 'sk-ant-api03-solo-personal-token');

INSERT INTO subscriptions (org_id, plan_type, monthly_budget_usd, current_spending_usd, billing_period_start, billing_period_end, status) VALUES
('44444444-4444-4444-4444-444444444444', 'professional', 500.00, 45.00,
 '2025-10-01 00:00:00', '2025-10-31 23:59:59', 'active');

INSERT INTO employees (id, org_id, role_id, team_id, email, full_name, password_hash, status, personal_claude_token) VALUES
('e8888888-8888-8888-8888-888888888888', '44444444-4444-4444-4444-444444444444', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL,
 'john@solostartup.com', 'John Solo', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active',
 'sk-ant-api03-john-personal-token');

-- Testing multiple agents
INSERT INTO org_agent_configs (org_id, agent_id, config, is_enabled) VALUES
('44444444-4444-4444-4444-444444444444', 'a1111111-1111-1111-1111-111111111111', '{}'::jsonb, true),
('44444444-4444-4444-4444-444444444444', 'a2222222-2222-2222-2222-222222222222', '{}'::jsonb, true),
('44444444-4444-4444-4444-444444444444', 'a6666666-6666-6666-6666-666666666666', '{}'::jsonb, true);

INSERT INTO employee_agent_configs (employee_id, agent_id, config_override, is_enabled) VALUES
('e8888888-8888-8888-8888-888888888888', 'a1111111-1111-1111-1111-111111111111', '{}'::jsonb, true),
('e8888888-8888-8888-8888-888888888888', 'a2222222-2222-2222-2222-222222222222', '{}'::jsonb, true);

-- ============================================================================
-- RE-ENABLE ROW-LEVEL SECURITY
-- ============================================================================

ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_agent_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_mcp_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE approvals ENABLE ROW LEVEL SECURITY;

-- ============================================================================
-- SEED DATA SUMMARY
-- ============================================================================

SELECT '========================= SEED DATA SUMMARY =========================' AS info;

SELECT 'Organizations' AS entity, COUNT(*) AS count FROM organizations
UNION ALL
SELECT 'Teams', COUNT(*) FROM teams
UNION ALL
SELECT 'Employees', COUNT(*) FROM employees
UNION ALL
SELECT 'Agents', COUNT(*) FROM agents
UNION ALL
SELECT 'MCP Servers', COUNT(*) FROM mcp_catalog
UNION ALL
SELECT 'Org Configs', COUNT(*) FROM org_agent_configs
UNION ALL
SELECT 'Team Configs', COUNT(*) FROM team_agent_configs
UNION ALL
SELECT 'Employee Configs', COUNT(*) FROM employee_agent_configs
UNION ALL
SELECT 'Pending Requests', COUNT(*) FROM agent_requests WHERE status = 'pending';

SELECT '========================= TEST CREDENTIALS =========================' AS info;

-- Journey 1: Acme Corp
SELECT '=== JOURNEY 1: ACME CORP (Mature Enterprise) ===' AS company;
SELECT email, full_name, 'password123' AS password, status
FROM employees
WHERE org_id = '11111111-1111-1111-1111-111111111111'
ORDER BY full_name;

-- Journey 2: TechCo
SELECT '=== JOURNEY 2: TECHCO (Growing Startup) ===' AS company;
SELECT email, full_name, 'password123' AS password, status
FROM employees
WHERE org_id = '22222222-2222-2222-2222-222222222222'
ORDER BY full_name;

-- Journey 3: NewCorp
SELECT '=== JOURNEY 3: NEWCORP (Brand New) ===' AS company;
SELECT email, full_name, 'password123' AS password, status
FROM employees
WHERE org_id = '33333333-3333-3333-3333-333333333333'
ORDER BY full_name;

-- Journey 4: Solo Startup
SELECT '=== JOURNEY 4: SOLO STARTUP (Testing Platform) ===' AS company;
SELECT email, full_name, 'password123' AS password, status
FROM employees
WHERE org_id = '44444444-4444-4444-4444-444444444444'
ORDER BY full_name;

SELECT '========================= PENDING APPROVALS =========================' AS info;
SELECT ar.reason, e.full_name AS requester, ar.request_data->>'agent_name' AS agent_requested
FROM agent_requests ar
JOIN employees e ON ar.employee_id = e.id
WHERE ar.status = 'pending';

SELECT '=== RLS RE-ENABLED === Seed complete!' AS status;
