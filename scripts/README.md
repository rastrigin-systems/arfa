# Scripts Directory

This directory contains **cross-cutting scripts** that operate across multiple services or at the repository level.

## Code Generation & Documentation

- **`generate-erd-overview.py`** - Generates user-friendly ERD documentation from database schema
- **`generate-claude-seed.py`** - Generates seed SQL from `.claude/` directory for Claude Code configuration

## Database & Testing

- **`create-test-users.sh`** - Creates test organizations and employees for development/testing
- **`seed-claude-config.sh`** - Seeds Claude Code configuration into database from `.claude/` directory

## Project Management

- **`archive-milestone.sh`** - Closes and archives all issues from a completed milestone
- **`start-milestone.sh`** - Creates a new milestone and populates it with prioritized issues
- **`split-large-tasks.sh`** - Helper for breaking down large tasks into subtasks
- **`update-project-status.sh`** - Updates GitHub Project status for an issue

## Git & Development Workflow

- **`install-hooks.sh`** - Git hooks installation script (deprecated as of v0.3.0)

---

## Service-Specific Scripts

**Service-specific scripts belong in their respective service directories:**

- API service scripts → `/services/api/scripts/`
- CLI service scripts → `/services/cli/scripts/`
- Web service scripts → `/services/web/scripts/`

---

## Usage

Most scripts include usage instructions in their headers. Run with `--help` or `-h` for more information:

```bash
./scripts/archive-milestone.sh --help
./scripts/start-milestone.sh --help
```

## Prerequisites

- **Python 3.x** - For Python scripts
- **Bash 4.x+** - For shell scripts
- **PostgreSQL client** - For database scripts (`psql`)
- **GitHub CLI (`gh`)** - For project management scripts
- **Go 1.24+** - For some database seeding scripts

## Environment Variables

Database scripts respect standard environment variables:

- `DATABASE_URL` - PostgreSQL connection string (default: `postgres://ubik:ubik_dev_password@localhost:5432/ubik`)

GitHub scripts use the `gh` CLI which reads from:

- `GITHUB_TOKEN` - GitHub personal access token
- Or use `gh auth login` for interactive authentication
