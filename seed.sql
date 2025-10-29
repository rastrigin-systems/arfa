-- Seed Data for Pivot - AI Agent Management Platform
-- Purpose: Reusable test data for development and testing
-- Password for all test users: "password123"
-- Password hash: $2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC

-- Temporarily disable Row-Level Security for seeding
ALTER TABLE teams DISABLE ROW LEVEL SECURITY;
ALTER TABLE employees DISABLE ROW LEVEL SECURITY;
ALTER TABLE employee_agent_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE employee_mcp_configs DISABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs DISABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records DISABLE ROW LEVEL SECURITY;

-- Clean existing data (in reverse dependency order)
TRUNCATE TABLE employee_agent_configs CASCADE;
TRUNCATE TABLE team_agent_configs CASCADE;
TRUNCATE TABLE org_agent_configs CASCADE;
TRUNCATE TABLE sessions CASCADE;
TRUNCATE TABLE employees CASCADE;
TRUNCATE TABLE teams CASCADE;
TRUNCATE TABLE roles CASCADE;
TRUNCATE TABLE organizations CASCADE;
TRUNCATE TABLE agents CASCADE;

-- ============================================================================
-- ORGANIZATIONS
-- ============================================================================

INSERT INTO organizations (id, name, slug, plan, settings, max_employees, max_agents_per_employee) VALUES
('11111111-1111-1111-1111-111111111111', 'Acme Corporation', 'acme-corp', 'enterprise', '{"features": ["sso", "audit_logs"]}'::jsonb, 500, 10),
('22222222-2222-2222-2222-222222222222', 'Tech Startup Inc', 'tech-startup', 'professional', '{"features": ["audit_logs"]}'::jsonb, 50, 5),
('33333333-3333-3333-3333-333333333333', 'Small Business LLC', 'small-biz', 'starter', '{}'::jsonb, 10, 3);

-- ============================================================================
-- ROLES
-- ============================================================================

INSERT INTO roles (id, name, permissions) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'Super Admin', '["*"]'::jsonb),
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'Admin', '["read", "write", "delete", "manage_employees"]'::jsonb),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'Manager', '["read", "write", "manage_team"]'::jsonb),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'Developer', '["read", "write"]'::jsonb),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'Viewer', '["read"]'::jsonb);

-- ============================================================================
-- TEAMS
-- ============================================================================

INSERT INTO teams (id, org_id, name, description) VALUES
-- Acme Corporation teams
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Engineering', 'Software development team'),
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Product', 'Product management team'),
('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', 'Design', 'UX/UI design team'),
('77777777-7777-7777-7777-777777777777', '11111111-1111-1111-1111-111111111111', 'Sales', 'Sales and business development'),

-- Tech Startup teams
('88888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'Full Stack Team', 'Cross-functional development team'),
('99999999-9999-9999-9999-999999999999', '22222222-2222-2222-2222-222222222222', 'Marketing', 'Marketing and growth team');

-- ============================================================================
-- EMPLOYEES
-- ============================================================================
-- Password for all users: "password123"

INSERT INTO employees (id, org_id, role_id, team_id, email, full_name, password_hash, status, preferences) VALUES
-- Acme Corporation employees
('e1111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '44444444-4444-4444-4444-444444444444', 'alice@acme.com', 'Alice Anderson', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{"theme": "dark", "notifications": true}'::jsonb),
('e2222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '44444444-4444-4444-4444-444444444444', 'bob@acme.com', 'Bob Builder', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),
('e3333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '44444444-4444-4444-4444-444444444444', 'charlie@acme.com', 'Charlie Chen', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),
('e4444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'cccccccc-cccc-cccc-cccc-cccccccccccc', '55555555-5555-5555-5555-555555555555', 'diana@acme.com', 'Diana Davis', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),
('e5555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '66666666-6666-6666-6666-666666666666', 'eve@acme.com', 'Eve Edwards', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),
('e6666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', '77777777-7777-7777-7777-777777777777', 'frank@acme.com', 'Frank Foster', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'suspended', '{}'::jsonb),

-- Tech Startup employees
('e7777777-7777-7777-7777-777777777777', '22222222-2222-2222-2222-222222222222', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', '88888888-8888-8888-8888-888888888888', 'grace@techstartup.com', 'Grace Green', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),
('e8888888-8888-8888-8888-888888888888', '22222222-2222-2222-2222-222222222222', 'dddddddd-dddd-dddd-dddd-dddddddddddd', '88888888-8888-8888-8888-888888888888', 'henry@techstartup.com', 'Henry Harris', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb),

-- Small Business employees
('e9999999-9999-9999-9999-999999999999', '33333333-3333-3333-3333-333333333333', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', NULL, 'iris@smallbiz.com', 'Iris Irving', '$2a$10$LoAcqAqR2r6aCRRtmtcJROs0d1JUXmg3tkxQwVxTf3uJvUc5ttRiC', 'active', '{}'::jsonb);

-- ============================================================================
-- AGENTS (AI Agents Catalog)
-- ============================================================================

INSERT INTO agents (id, name, type, description, provider, default_config, capabilities, llm_provider, llm_model, is_public) VALUES
('a1111111-1111-1111-1111-111111111111', 'Claude Code', 'ide_assistant', 'Claude-powered IDE assistant for VS Code', 'Anthropic',
'{"max_tokens": 8000, "temperature": 0.7}'::jsonb, '["code_completion", "chat", "refactoring"]'::jsonb, 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a2222222-2222-2222-2222-222222222222', 'Cursor', 'ide_assistant', 'AI-first code editor with pair programming', 'Anysphere',
'{"max_tokens": 4000, "temperature": 0.5}'::jsonb, '["code_completion", "chat"]'::jsonb, 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a3333333-3333-3333-3333-333333333333', 'Windsurf', 'ide_assistant', 'AI-powered coding assistant', 'Codeium',
'{"max_tokens": 4000}'::jsonb, '["code_completion", "suggestions"]'::jsonb, 'anthropic', 'claude-3-5-sonnet-20241022', true),

('a4444444-4444-4444-4444-444444444444', 'GitHub Copilot', 'code_completion', 'AI pair programmer for code completion', 'GitHub',
'{"enable_completions": true, "enable_chat": true}'::jsonb, '["code_completion"]'::jsonb, 'openai', 'gpt-4', true),

('a5555555-5555-5555-5555-555555555555', 'Continue', 'ide_assistant', 'Open source AI code assistant', 'Continue.dev',
'{"max_tokens": 4000}'::jsonb, '["code_completion", "chat", "custom_commands"]'::jsonb, 'anthropic', 'claude-3-5-sonnet-20241022', true);

-- ============================================================================
-- ORG AGENT CONFIGS (Base configurations)
-- ============================================================================

INSERT INTO org_agent_configs (id, org_id, agent_id, config, is_enabled) VALUES
-- Acme Corp configs
('c1111111-1111-1111-1111-111111111111', '11111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
'{"model": "claude-3-5-sonnet-20241022", "max_tokens": 8000, "temperature": 0.7}'::jsonb, true),

('c2222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 'a2222222-2222-2222-2222-222222222222',
'{"model": "claude-3-5-sonnet-20241022", "max_tokens": 4000}'::jsonb, true),

-- Tech Startup configs
('c3333333-3333-3333-3333-333333333333', '22222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111',
'{"model": "claude-3-5-sonnet-20241022", "max_tokens": 4000, "temperature": 0.5}'::jsonb, true);

-- ============================================================================
-- TEAM AGENT CONFIGS (Team overrides)
-- ============================================================================

INSERT INTO team_agent_configs (id, team_id, agent_id, config_override, is_enabled) VALUES
-- Engineering team at Acme gets higher token limits
('f1111111-1111-1111-1111-111111111111', '44444444-4444-4444-4444-444444444444', 'a1111111-1111-1111-1111-111111111111',
'{"max_tokens": 10000, "system_prompt": "You are helping a senior software engineer."}'::jsonb, true),

-- Product team has different settings
('f2222222-2222-2222-2222-222222222222', '55555555-5555-5555-5555-555555555555', 'a1111111-1111-1111-1111-111111111111',
'{"temperature": 0.9, "system_prompt": "You are helping with product documentation."}'::jsonb, true);

-- ============================================================================
-- EMPLOYEE AGENT CONFIGS (Individual overrides)
-- ============================================================================

INSERT INTO employee_agent_configs (id, employee_id, agent_id, config_override, is_enabled) VALUES
-- Alice (super admin) has custom settings
('d1111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111',
'{"max_tokens": 12000, "temperature": 0.8, "custom_prompt": "Expert mode enabled"}'::jsonb, true),

-- Bob has Cursor configured
('d2222222-2222-2222-2222-222222222222', 'e2222222-2222-2222-2222-222222222222', 'a2222222-2222-2222-2222-222222222222',
'{"model": "claude-3-opus-20240229", "max_tokens": 5000}'::jsonb, true);

-- ============================================================================
-- SUMMARY
-- ============================================================================

SELECT '=== SEED DATA SUMMARY ===' AS status;
SELECT 'Organizations' AS entity, COUNT(*) AS count FROM organizations
UNION ALL
SELECT 'Roles', COUNT(*) FROM roles
UNION ALL
SELECT 'Teams', COUNT(*) FROM teams
UNION ALL
SELECT 'Employees', COUNT(*) FROM employees
UNION ALL
SELECT 'Agents', COUNT(*) FROM agents
UNION ALL
SELECT 'Org Agent Configs', COUNT(*) FROM org_agent_configs
UNION ALL
SELECT 'Team Agent Configs', COUNT(*) FROM team_agent_configs
UNION ALL
SELECT 'Employee Agent Configs', COUNT(*) FROM employee_agent_configs;

SELECT '=== TEST CREDENTIALS ===' AS status;
SELECT email, full_name, status, 'password123' AS password
FROM employees
ORDER BY org_id, email;

-- ============================================================================
-- ACTIVITY LOGS (Sample Data)
-- ============================================================================

INSERT INTO activity_logs (org_id, employee_id, event_type, event_category, payload, created_at) VALUES
-- Acme Corporation activities
('11111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', 'employee.created', 'admin', '{"employee_name": "Bob Developer"}'::jsonb, NOW() - INTERVAL '2 hours'),
('11111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', 'agent.installed', 'agent', '{"agent_name": "Claude Code"}'::jsonb, NOW() - INTERVAL '3 hours'),
('11111111-1111-1111-1111-111111111111', 'e2222222-2222-2222-2222-222222222222', 'mcp.configured', 'mcp', '{"mcp_name": "Filesystem"}'::jsonb, NOW() - INTERVAL '5 hours'),
('11111111-1111-1111-1111-111111111111', 'e3333333-3333-3333-3333-333333333333', 'team.created', 'admin', '{"team_name": "Engineering"}'::jsonb, NOW() - INTERVAL '1 day'),
('11111111-1111-1111-1111-111111111111', 'e1111111-1111-1111-1111-111111111111', 'agent.updated', 'agent', '{"agent_name": "Claude Code", "setting": "max_tokens"}'::jsonb, NOW() - INTERVAL '1 day'),
('11111111-1111-1111-1111-111111111111', 'e2222222-2222-2222-2222-222222222222', 'employee.updated', 'admin', '{"employee_name": "Bob Developer", "field": "permissions"}'::jsonb, NOW() - INTERVAL '2 days'),

-- Tech Startup activities
('22222222-2222-2222-2222-222222222222', 'e7777777-7777-7777-7777-777777777777', 'auth.login', 'auth', '{}'::jsonb, NOW() - INTERVAL '30 minutes'),
('22222222-2222-2222-2222-222222222222', 'e8888888-8888-8888-8888-888888888888', 'agent.installed', 'agent', '{"agent_name": "Cursor"}'::jsonb, NOW() - INTERVAL '4 hours'),
('22222222-2222-2222-2222-222222222222', 'e9999999-9999-9999-9999-999999999999', 'mcp.configured', 'mcp', '{"mcp_name": "Git"}'::jsonb, NOW() - INTERVAL '6 hours');

-- Re-enable Row-Level Security
ALTER TABLE teams ENABLE ROW LEVEL SECURITY;
ALTER TABLE employees ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_agent_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE employee_mcp_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE activity_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_records ENABLE ROW LEVEL SECURITY;

SELECT '=== RLS RE-ENABLED ===' AS status;
