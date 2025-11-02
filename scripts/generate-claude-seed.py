#!/usr/bin/env python3
"""
Generate SQL seed data from .claude/ directory structure.

This script reads:
- Agent .md files from .claude/agents/
- Skill directories from .claude/skills/
- Generates INSERT statements for agents and skills
"""

import json
import os
import re
from pathlib import Path
from typing import Dict, List

# Base directory
BASE_DIR = Path(__file__).parent.parent
CLAUDE_DIR = BASE_DIR / ".claude"
AGENTS_DIR = CLAUDE_DIR / "agents"
SKILLS_DIR = CLAUDE_DIR / "skills"


def escape_sql_string(s: str) -> str:
    """Escape single quotes for SQL."""
    return s.replace("'", "''")


def read_agent_file(agent_file: Path) -> Dict:
    """Read agent .md file and extract metadata."""
    with open(agent_file, 'r', encoding='utf-8') as f:
        content = f.read()

    # Extract agent name from filename
    name = agent_file.stem

    # Try to extract description from first line or heading
    description = "AI development agent"
    lines = content.split('\n')
    for line in lines:
        if line.startswith('# '):
            description = line[2:].strip()
            break

    # Determine agent type based on name
    type_map = {
        'go-backend-developer': 'claude-code',
        'frontend-developer': 'claude-code',
        'coordinator': 'claude-code',
        'tech-lead': 'claude-code',
        'product-strategist': 'claude-code',
        'pr-reviewer': 'claude-code',
    }

    agent_type = type_map.get(name, 'claude-code')

    return {
        'name': name,
        'type': agent_type,
        'description': description,
        'content': content,
        'provider': 'anthropic',
        'llm_provider': 'anthropic',
        'llm_model': 'claude-sonnet-4-5-20250929'
    }


def read_skill_directory(skill_dir: Path) -> Dict:
    """Read all files in a skill directory and build JSON structure."""
    skill_name = skill_dir.name
    files = []

    # Walk through skill directory
    for file_path in sorted(skill_dir.rglob('*')):
        if file_path.is_file():
            # Get relative path from skill root
            rel_path = file_path.relative_to(skill_dir)

            # Read file content
            try:
                with open(file_path, 'r', encoding='utf-8') as f:
                    content = f.read()
            except UnicodeDecodeError:
                # Skip binary files
                continue

            files.append({
                'path': str(rel_path),
                'content': content
            })

    # Try to extract description from SKILL.md
    description = f"Skill: {skill_name}"
    skill_md = skill_dir / "SKILL.md"
    if skill_md.exists():
        with open(skill_md, 'r', encoding='utf-8') as f:
            lines = f.read().split('\n')
            for line in lines:
                if line.startswith('# '):
                    description = line[2:].strip()
                    break

    # Determine category
    category_map = {
        'release-manager': 'devops',
        'github-task-manager': 'workflow',
        'github-dev-workflow': 'workflow',
        'github-pr-workflow': 'workflow',
    }

    category = category_map.get(skill_name, 'development')

    # Determine dependencies (all github-related skills need github MCP)
    dependencies = {"mcp_servers": [], "skills": []}
    if 'github' in skill_name.lower():
        dependencies["mcp_servers"].append("github")

    return {
        'name': skill_name,
        'description': description,
        'category': category,
        'version': '1.0.0',
        'files': files,
        'dependencies': dependencies
    }


def generate_agent_sql(agents: List[Dict]) -> str:
    """Generate SQL INSERT statements for agents."""
    if not agents:
        return ""

    sql_parts = []
    sql_parts.append("-- ============================================================================")
    sql_parts.append("-- AGENT CATALOG - Real agents from .claude/agents/")
    sql_parts.append("-- ============================================================================")
    sql_parts.append("")

    for agent in agents:
        # Update existing agent or insert new
        sql_parts.append(f"-- Agent: {agent['name']}")
        sql_parts.append(f"""INSERT INTO agents (name, type, description, provider, llm_provider, llm_model, default_config, capabilities)
VALUES (
    '{agent['name']}',
    '{agent['type']}',
    '{escape_sql_string(agent['description'])}',
    '{agent['provider']}',
    '{agent['llm_provider']}',
    '{agent['llm_model']}',
    '{{}}'::JSONB,
    '["code_generation", "debugging", "refactoring"]'::JSONB
)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    llm_model = EXCLUDED.llm_model,
    updated_at = NOW();
""")
        sql_parts.append("")

    # Now update employee_agent_configs with content
    sql_parts.append("-- Update employee_agent_configs.content for existing agent assignments")
    sql_parts.append("-- (Only if employee_agent_configs exist)")
    for agent in agents:
        content_escaped = escape_sql_string(agent['content'])
        sql_parts.append(f"""UPDATE employee_agent_configs
SET content = '{content_escaped}'
WHERE agent_id = (SELECT id FROM agents WHERE name = '{agent['name']}');
""")
        sql_parts.append("")

    return "\n".join(sql_parts)


def generate_skill_sql(skills: List[Dict]) -> str:
    """Generate SQL INSERT statements for skills."""
    if not skills:
        return ""

    sql_parts = []
    sql_parts.append("-- ============================================================================")
    sql_parts.append("-- SKILL CATALOG - Real skills from .claude/skills/")
    sql_parts.append("-- ============================================================================")
    sql_parts.append("")

    for skill in skills:
        # Convert files to JSON
        files_json = json.dumps(skill['files'])
        files_json_escaped = escape_sql_string(files_json)

        # Convert dependencies to JSON
        deps_json = json.dumps(skill['dependencies'])
        deps_json_escaped = escape_sql_string(deps_json)

        sql_parts.append(f"-- Skill: {skill['name']} ({len(skill['files'])} files)")
        sql_parts.append(f"""INSERT INTO skill_catalog (name, description, category, version, files, dependencies, is_active)
VALUES (
    '{skill['name']}',
    '{escape_sql_string(skill['description'])}',
    '{skill['category']}',
    '{skill['version']}',
    '{files_json_escaped}'::JSONB,
    '{deps_json_escaped}'::JSONB,
    true
)
ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    files = EXCLUDED.files,
    dependencies = EXCLUDED.dependencies,
    version = EXCLUDED.version,
    updated_at = NOW();
""")
        sql_parts.append("")

    return "\n".join(sql_parts)


def generate_mcp_sql() -> str:
    """Generate SQL INSERT statements for MCP servers."""
    sql_parts = []
    sql_parts.append("-- ============================================================================")
    sql_parts.append("-- MCP SERVERS - Docker-based MCP server configurations")
    sql_parts.append("-- ============================================================================")
    sql_parts.append("")

    # Add playwright MCP
    sql_parts.append("-- Playwright MCP Server")
    sql_parts.append("""INSERT INTO mcp_catalog (
    name,
    provider,
    version,
    description,
    connection_schema,
    capabilities,
    requires_credentials,
    is_approved,
    category_id,
    docker_image,
    config_template,
    required_env_vars
)
VALUES (
    'Playwright',
    '@executeautomation/mcp-playwright',
    '1.0.0',
    'Browser automation and testing with Playwright',
    '{"type": "object", "properties": {"headless": {"type": "boolean"}}}'::JSONB,
    '["browser_automation", "web_testing", "screenshots"]'::JSONB,
    false,
    true,
    (SELECT id FROM mcp_categories WHERE name = 'Development'),
    'ghcr.io/executeautomation/mcp-playwright:latest',
    '{"headless": true}'::JSONB,
    '[]'::JSONB
)
ON CONFLICT (name) DO UPDATE SET
    docker_image = EXCLUDED.docker_image,
    config_template = EXCLUDED.config_template,
    updated_at = NOW();
""")
    sql_parts.append("")

    # GitHub MCP already exists, just update docker_image
    sql_parts.append("-- Update GitHub MCP Server with Docker image")
    sql_parts.append("""UPDATE mcp_catalog
SET
    docker_image = 'ghcr.io/github/github-mcp-server:latest',
    config_template = '{"GITHUB_PERSONAL_ACCESS_TOKEN": "${GITHUB_TOKEN}"}'::JSONB,
    required_env_vars = '["GITHUB_PERSONAL_ACCESS_TOKEN"]'::JSONB,
    updated_at = NOW()
WHERE name = 'GitHub';
""")
    sql_parts.append("")

    return "\n".join(sql_parts)


def main():
    """Main function to generate seed SQL."""
    # Read all agents
    agents = []
    for agent_file in sorted(AGENTS_DIR.glob("*.md")):
        # Skip README
        if agent_file.stem == "README":
            continue
        agent = read_agent_file(agent_file)
        agents.append(agent)

    print(f"Found {len(agents)} agents", file=os.sys.stderr)

    # Read all skills
    skills = []
    for skill_dir in sorted(SKILLS_DIR.iterdir()):
        if skill_dir.is_dir():
            skill = read_skill_directory(skill_dir)
            skills.append(skill)

    print(f"Found {len(skills)} skills", file=os.sys.stderr)

    # Generate SQL file
    output_file = BASE_DIR / "shared" / "schema" / "seeds" / "002_claude_config.sql"
    output_file.parent.mkdir(parents=True, exist_ok=True)

    with open(output_file, 'w', encoding='utf-8') as f:
        f.write("-- Seed Data: Claude Code Configuration\n")
        f.write("-- Auto-generated from .claude/ directory\n")
        f.write(f"-- Agents: {len(agents)}, Skills: {len(skills)}\n")
        f.write("-- Generated by: scripts/generate-claude-seed.py\n")
        f.write("--\n")
        f.write("-- This file imports real agent and skill configurations from the project.\n")
        f.write("-- Run with: psql $DATABASE_URL -f shared/schema/seeds/002_claude_config.sql\n")
        f.write("\n")

        # Generate sections
        f.write(generate_agent_sql(agents))
        f.write("\n")
        f.write(generate_skill_sql(skills))
        f.write("\n")
        f.write(generate_mcp_sql())

    print(f"Generated: {output_file}", file=os.sys.stderr)
    print(f"  - {len(agents)} agents", file=os.sys.stderr)
    print(f"  - {len(skills)} skills", file=os.sys.stderr)
    print(f"  - 2 MCP servers", file=os.sys.stderr)


if __name__ == "__main__":
    main()
