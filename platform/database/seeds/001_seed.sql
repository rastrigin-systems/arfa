-- Seed Data for Arfa Enterprise - Security Proxy Mode
-- Version: 3.0.0 - Matches actual schema

-- =============================================================================
-- Organizations
-- =============================================================================

INSERT INTO organizations (id, name, slug, settings, created_at, updated_at) VALUES
    ('e5d10009-0988-44b6-b313-67ffbbbb1ef8', 'Acme Corporation', 'acme-corp',
     '{"features": {"tool_blocking": true, "webhook_export": true}}',
     NOW(), NOW()),
    ('f6e21110-1a99-55c7-c424-78ffcccc2fa9', 'TechStart Inc', 'techstart',
     '{"features": {"tool_blocking": true, "webhook_export": false}}',
     NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Subscriptions
-- =============================================================================

INSERT INTO subscriptions (id, org_id, plan_type, monthly_budget_usd, current_spending_usd, billing_period_start, billing_period_end, status) VALUES
    ('a1b2c3d4-e5f6-4a5b-8c9d-0e1f2a3b4c5d', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'enterprise', 10000.00, 0.00, NOW(), NOW() + INTERVAL '1 month', 'active'),
    ('b2c3d4e5-f6a7-5b6c-9d0e-1f2a3b4c5d6e', 'f6e21110-1a99-55c7-c424-78ffcccc2fa9',
     'team', 1000.00, 0.00, NOW(), NOW() + INTERVAL '1 month', 'active')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Roles (Global - no org_id)
-- =============================================================================

INSERT INTO roles (id, name, description, permissions) VALUES
    ('11111111-1111-1111-1111-111111111111', 'Admin', 'Full administrative access', '["*"]'),
    ('22222222-2222-2222-2222-222222222222', 'Developer', 'Standard developer access', '["logs:read", "policies:read"]'),
    ('33333333-3333-3333-3333-333333333333', 'Viewer', 'Read-only access', '["logs:read"]')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Teams
-- =============================================================================

INSERT INTO teams (id, org_id, name, created_at, updated_at) VALUES
    -- Acme Corporation teams
    ('aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'Engineering', NOW(), NOW()),
    ('bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'Security', NOW(), NOW()),
    -- TechStart teams
    ('cccc3333-cccc-cccc-cccc-cccccccccccc', 'f6e21110-1a99-55c7-c424-78ffcccc2fa9',
     'Product', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Employees
-- =============================================================================

-- Password: admin123 (bcrypt hash with cost 10)
INSERT INTO employees (id, org_id, team_id, role_id, email, full_name, password_hash, status, preferences) VALUES
    -- Acme Corporation employees
    ('6a41beee-cf2d-4c59-affd-80e3f58466d6', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '11111111-1111-1111-1111-111111111111',
     'admin@acme.com', 'Admin User',
     '$2a$10$tvQL2A2wWAXissld8AyFUegJH5OYm5vRmhl1t/CPq0rbgmVoQKKV.', 'active', '{}'),
    ('7b52cfff-d03e-5d6a-bffe-91f4069577e7', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa', '22222222-2222-2222-2222-222222222222',
     'dev@acme.com', 'Developer User',
     '$2a$10$tvQL2A2wWAXissld8AyFUegJH5OYm5vRmhl1t/CPq0rbgmVoQKKV.', 'active', '{}'),
    -- TechStart employees
    ('8c63d000-e04f-6e7b-c00f-a2050706880f', 'f6e21110-1a99-55c7-c424-78ffcccc2fa9',
     'cccc3333-cccc-cccc-cccc-cccccccccccc', '11111111-1111-1111-1111-111111111111',
     'admin@techstart.com', 'TechStart Admin',
     '$2a$10$tvQL2A2wWAXissld8AyFUegJH5OYm5vRmhl1t/CPq0rbgmVoQKKV.', 'active', '{}')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Policies (Global security policies)
-- =============================================================================

INSERT INTO policies (id, name, type, rules, severity) VALUES
    ('c1111111-1111-1111-1111-111111111111', 'Production Safety', 'tool_blocking',
     '{"blocked_patterns": ["rm -rf /", "sudo", "chmod 777"]}', 'block'),
    ('c2222222-2222-2222-2222-222222222222', 'Audit All Tools', 'audit',
     '{"audit_all": true}', 'warn'),
    ('c3333333-3333-3333-3333-333333333333', 'File Write Restriction', 'tool_blocking',
     '{"blocked_paths": ["/etc/*", "/var/*", "~/.ssh/*"]}', 'block')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Tool Policies (Org-specific tool blocking rules)
-- =============================================================================

INSERT INTO tool_policies (id, org_id, team_id, employee_id, tool_name, conditions, action, reason) VALUES
    -- Acme Corporation - Block dangerous bash commands org-wide
    ('d1111111-1111-1111-1111-111111111111', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     NULL, NULL, 'Bash', '{"patterns": ["rm -rf", "sudo", "chmod 777"]}', 'deny', 'Dangerous command blocked by org policy'),
    -- Acme Corporation - Audit all file writes for security team
    ('d2222222-2222-2222-2222-222222222222', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb', NULL, 'Write', '{"paths": ["*"]}', 'audit', 'All file writes audited for security team'),
    -- TechStart - Block rm -rf org-wide
    ('d3333333-3333-3333-3333-333333333333', 'f6e21110-1a99-55c7-c424-78ffcccc2fa9',
     NULL, NULL, 'Bash', '{"patterns": ["rm -rf"]}', 'deny', 'Dangerous delete blocked')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Employee Policies (Assign policies to employees)
-- =============================================================================

INSERT INTO employee_policies (id, employee_id, policy_id, overrides) VALUES
    ('e1111111-1111-1111-1111-111111111111', '6a41beee-cf2d-4c59-affd-80e3f58466d6',
     'c1111111-1111-1111-1111-111111111111', '{}'),
    ('e2222222-2222-2222-2222-222222222222', '6a41beee-cf2d-4c59-affd-80e3f58466d6',
     'c2222222-2222-2222-2222-222222222222', '{}'),
    ('e3333333-3333-3333-3333-333333333333', '7b52cfff-d03e-5d6a-bffe-91f4069577e7',
     'c2222222-2222-2222-2222-222222222222', '{}'),
    ('e4444444-4444-4444-4444-444444444444', '8c63d000-e04f-6e7b-c00f-a2050706880f',
     'c1111111-1111-1111-1111-111111111111', '{}')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Webhook Destinations (SIEM integration)
-- =============================================================================

INSERT INTO webhook_destinations (id, org_id, name, url, auth_type, auth_config, event_types, enabled, signing_secret) VALUES
    ('f1111111-1111-1111-1111-111111111111', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'Security SIEM', 'https://siem.acme.com/webhooks/arfa', 'bearer',
     '{"token": "siem-bearer-token-123"}', ARRAY['tool_blocked', 'policy_violation'], true, 'acme-webhook-signing-secret'),
    ('f2222222-2222-2222-2222-222222222222', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     'Audit Log Export', 'https://audit.acme.com/logs', 'header',
     '{"header_name": "X-API-Key", "header_value": "audit-api-key-456"}', ARRAY['*'], true, 'acme-audit-signing-secret')
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- Sample Activity Logs (for testing)
-- =============================================================================

INSERT INTO activity_logs (id, org_id, employee_id, session_id, client_name, client_version, event_type, event_category, content, payload, created_at) VALUES
    ('a0a11111-1111-1111-1111-111111111111', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     '6a41beee-cf2d-4c59-affd-80e3f58466d6', 'b0b11111-1111-1111-1111-111111111111',
     'claude-code', '1.0.25', 'tool_call', 'development',
     'Read file package.json', '{"tool": "Read", "file": "package.json", "status": "allowed"}', NOW() - INTERVAL '1 hour'),
    ('a0a22222-2222-2222-2222-222222222222', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     '6a41beee-cf2d-4c59-affd-80e3f58466d6', 'b0b11111-1111-1111-1111-111111111111',
     'claude-code', '1.0.25', 'tool_blocked', 'security',
     'Blocked dangerous command: rm -rf /', '{"tool": "Bash", "command": "rm -rf /", "reason": "matches blocked pattern"}', NOW() - INTERVAL '30 minutes'),
    ('a0a33333-3333-3333-3333-333333333333', 'e5d10009-0988-44b6-b313-67ffbbbb1ef8',
     '7b52cfff-d03e-5d6a-bffe-91f4069577e7', 'b0b22222-2222-2222-2222-222222222222',
     'cursor', '0.43.0', 'tool_call', 'development',
     'Edit file src/index.ts', '{"tool": "Edit", "file": "src/index.ts", "status": "allowed"}', NOW() - INTERVAL '15 minutes')
ON CONFLICT (id) DO NOTHING;
