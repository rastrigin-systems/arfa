-- Migration: Add Skills and MCP Server tables + agent_configs.content column
-- Version: 001
-- Description: Support Claude Code configuration management with skills and MCP servers
-- Date: 2025-11-02

-- ============================================================================
-- UP MIGRATION
-- ============================================================================

-- 1. Add content column to agent_configs table (if it exists, else create placeholder)
-- Note: agent_configs doesn't exist yet in schema.sql, but this prepares for it
-- For now, we'll add to employee_agent_configs which is the closest equivalent

ALTER TABLE employee_agent_configs
ADD COLUMN IF NOT EXISTS content TEXT;

COMMENT ON COLUMN employee_agent_configs.content IS 'Full .md file content for agent configuration';

-- 2. Skill Catalog - Available skills (release-manager, github-task-manager, etc.)
CREATE TABLE IF NOT EXISTS skill_catalog (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50), -- 'workflow', 'development', 'devops'
    version VARCHAR(20) NOT NULL,
    files JSONB NOT NULL, -- [{path: "SKILL.md", content: "..."}]
    dependencies JSONB, -- {mcp_servers: ["github"], skills: []}
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skill_catalog_name ON skill_catalog(name);
CREATE INDEX IF NOT EXISTS idx_skill_catalog_category ON skill_catalog(category);
CREATE INDEX IF NOT EXISTS idx_skill_catalog_is_active ON skill_catalog(is_active);

COMMENT ON TABLE skill_catalog IS 'Available skills for Claude Code agents';
COMMENT ON COLUMN skill_catalog.files IS 'Array of file objects with path and content';
COMMENT ON COLUMN skill_catalog.dependencies IS 'Required MCP servers and other skills';

-- 3. Employee Skills - Which skills each employee has
CREATE TABLE IF NOT EXISTS employee_skills (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skill_catalog(id) ON DELETE CASCADE,
    is_enabled BOOLEAN DEFAULT true,
    config JSONB, -- Skill-specific config overrides
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_employee_skill UNIQUE(employee_id, skill_id)
);

CREATE INDEX IF NOT EXISTS idx_employee_skills_employee_id ON employee_skills(employee_id);
CREATE INDEX IF NOT EXISTS idx_employee_skills_skill_id ON employee_skills(skill_id);
CREATE INDEX IF NOT EXISTS idx_employee_skills_is_enabled ON employee_skills(is_enabled);

COMMENT ON TABLE employee_skills IS 'Skills assigned to each employee';
COMMENT ON COLUMN employee_skills.config IS 'Employee-specific configuration overrides';

-- 4. MCP Catalog Update - This table already exists, but we'll ensure it has the fields we need
-- The existing mcp_catalog has most fields we need, but let's add docker_image if missing
ALTER TABLE mcp_catalog
ADD COLUMN IF NOT EXISTS docker_image VARCHAR(255);

ALTER TABLE mcp_catalog
ADD COLUMN IF NOT EXISTS config_template JSONB;

ALTER TABLE mcp_catalog
ADD COLUMN IF NOT EXISTS required_env_vars JSONB;

COMMENT ON COLUMN mcp_catalog.docker_image IS 'Docker image for MCP server (if containerized)';
COMMENT ON COLUMN mcp_catalog.config_template IS 'Template configuration for MCP server';
COMMENT ON COLUMN mcp_catalog.required_env_vars IS 'Required environment variables (e.g., ["GITHUB_TOKEN"])';

-- 5. Update employee_mcp_configs - Already exists, verify it has the fields we need
-- The existing table already has connection_config which maps to our config field
-- Add is_enabled if not present
ALTER TABLE employee_mcp_configs
ADD COLUMN IF NOT EXISTS is_enabled BOOLEAN DEFAULT true;

COMMENT ON COLUMN employee_mcp_configs.connection_config IS 'Actual configuration with secrets';

-- Create index for is_enabled if it doesn't exist
CREATE INDEX IF NOT EXISTS idx_employee_mcp_configs_is_enabled ON employee_mcp_configs(is_enabled);

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Add updated_at triggers for new tables
CREATE TRIGGER update_skill_catalog_updated_at BEFORE UPDATE ON skill_catalog
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_employee_skills_updated_at BEFORE UPDATE ON employee_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- SEED DATA
-- ============================================================================

-- Sample Skills
INSERT INTO skill_catalog (name, description, category, version, files, dependencies, is_active) VALUES
    (
        'release-manager',
        'Automated release management and versioning',
        'devops',
        '1.0.0',
        '[{"path": "SKILL.md", "content": "# Release Manager Skill\\n\\nAutomates release workflows..."}]'::JSONB,
        '{"mcp_servers": ["github"], "skills": []}'::JSONB,
        true
    ),
    (
        'github-task-manager',
        'GitHub issue and project management',
        'workflow',
        '1.0.0',
        '[{"path": "SKILL.md", "content": "# GitHub Task Manager\\n\\nManages GitHub issues and projects..."}]'::JSONB,
        '{"mcp_servers": ["github"], "skills": []}'::JSONB,
        true
    ),
    (
        'code-reviewer',
        'Automated code review and quality checks',
        'development',
        '1.0.0',
        '[{"path": "SKILL.md", "content": "# Code Reviewer\\n\\nProvides automated code reviews..."}]'::JSONB,
        '{"mcp_servers": ["github"], "skills": []}'::JSONB,
        true
    )
ON CONFLICT (name) DO NOTHING;

-- Update MCP catalog with docker images where applicable
UPDATE mcp_catalog
SET
    docker_image = CASE name
        WHEN 'Filesystem' THEN 'modelcontextprotocol/server-filesystem:latest'
        WHEN 'GitHub' THEN 'modelcontextprotocol/server-github:latest'
        WHEN 'PostgreSQL' THEN 'modelcontextprotocol/server-postgres:latest'
        WHEN 'Slack' THEN 'modelcontextprotocol/server-slack:latest'
        ELSE NULL
    END,
    config_template = CASE name
        WHEN 'GitHub' THEN '{"token": "${GITHUB_TOKEN}"}'::JSONB
        WHEN 'PostgreSQL' THEN '{"connectionString": "${DATABASE_URL}"}'::JSONB
        WHEN 'Slack' THEN '{"token": "${SLACK_TOKEN}"}'::JSONB
        ELSE '{}'::JSONB
    END,
    required_env_vars = CASE name
        WHEN 'GitHub' THEN '["GITHUB_TOKEN"]'::JSONB
        WHEN 'PostgreSQL' THEN '["DATABASE_URL"]'::JSONB
        WHEN 'Slack' THEN '["SLACK_TOKEN"]'::JSONB
        ELSE '[]'::JSONB
    END
WHERE name IN ('Filesystem', 'GitHub', 'PostgreSQL', 'Slack');

-- ============================================================================
-- DOWN MIGRATION (for rollback)
-- ============================================================================

-- To rollback this migration, run:
/*
DROP TABLE IF EXISTS employee_skills CASCADE;
DROP TABLE IF EXISTS skill_catalog CASCADE;
ALTER TABLE employee_agent_configs DROP COLUMN IF EXISTS content;
ALTER TABLE mcp_catalog DROP COLUMN IF EXISTS docker_image;
ALTER TABLE mcp_catalog DROP COLUMN IF EXISTS config_template;
ALTER TABLE mcp_catalog DROP COLUMN IF EXISTS required_env_vars;
ALTER TABLE employee_mcp_configs DROP COLUMN IF EXISTS is_enabled;
*/
