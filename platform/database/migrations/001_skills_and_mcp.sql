-- Migration: Add Skills tables
-- Version: 001
-- Description: Support Claude Code configuration management with skills
-- Date: 2025-11-02

-- ============================================================================
-- UP MIGRATION
-- ============================================================================

-- 1. Skill Catalog - Available skills (release-manager, github-task-manager, etc.)
CREATE TABLE IF NOT EXISTS skill_catalog (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    category VARCHAR(50), -- 'workflow', 'development', 'devops'
    version VARCHAR(20) NOT NULL,
    files JSONB NOT NULL, -- [{path: "SKILL.md", content: "..."}]
    dependencies JSONB, -- {skills: []}
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skill_catalog_name ON skill_catalog(name);
CREATE INDEX IF NOT EXISTS idx_skill_catalog_category ON skill_catalog(category);
CREATE INDEX IF NOT EXISTS idx_skill_catalog_is_active ON skill_catalog(is_active);

COMMENT ON TABLE skill_catalog IS 'Available skills for AI agents';
COMMENT ON COLUMN skill_catalog.files IS 'Array of file objects with path and content';
COMMENT ON COLUMN skill_catalog.dependencies IS 'Required other skills';

-- 2. Employee Skills - Which skills each employee has
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
        '{"skills": []}'::JSONB,
        true
    ),
    (
        'github-task-manager',
        'GitHub issue and project management',
        'workflow',
        '1.0.0',
        '[{"path": "SKILL.md", "content": "# GitHub Task Manager\\n\\nManages GitHub issues and projects..."}]'::JSONB,
        '{"skills": []}'::JSONB,
        true
    ),
    (
        'code-reviewer',
        'Automated code review and quality checks',
        'development',
        '1.0.0',
        '[{"path": "SKILL.md", "content": "# Code Reviewer\\n\\nProvides automated code reviews..."}]'::JSONB,
        '{"skills": []}'::JSONB,
        true
    )
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- DOWN MIGRATION (for rollback)
-- ============================================================================

-- To rollback this migration, run:
/*
DROP TABLE IF EXISTS employee_skills CASCADE;
DROP TABLE IF EXISTS skill_catalog CASCADE;
*/
