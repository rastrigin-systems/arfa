# Ubik Enterprise — AI Agent Management Platform

**Multi-tenant SaaS platform for centralized AI agent and MCP configuration management**

---

## 📑 Table of Contents

### Foundation (Rarely Changes)
- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)

### Documentation
- [Documentation Map](#-documentation-map)
- [Quick Reference](#quick-reference)

### Development
- [Common Commands](#common-commands)
- [Development Workflow](#development-workflow)
- [Critical Notes](#critical-notes)

### Status & Next Steps
- [Current Status](#current-status)
- [Success Metrics](#success-metrics)
- [Roadmap](#roadmap)

---

# FOUNDATION

*This section describes the stable foundation of the system that rarely changes.*

---

## System Overview

### System Purpose

Multi-tenant SaaS platform for companies to centrally manage AI agent (Claude Code, Cursor, Windsurf, etc.) and MCP server configurations for their employees.

**Core Value**: Centralized control, policy enforcement, and visibility into AI agent usage across an organization.

### What This Platform Does

**For Companies:**
- Manage employees, teams, and roles
- Control which AI agents employees can use
- Configure MCP servers and access per employee
- Set usage policies (path restrictions, rate limits, cost limits)
- Approve/reject employee requests for new agents or MCPs
- Track usage, costs, and activity across the organization
- Enforce compliance and security policies

**For Employees:**
- Sync agent configurations to local machines via CLI
- Request access to new agents or MCP servers
- View their assigned agents and policies
- Use AI agents with centrally-managed configurations

### Design Decisions

**Authentication Model**: Self-service employees with passwords (no separate admin tier)
- Each employee manages their own configs
- Roles define permissions (member, approver)
- No admin/user distinction - everyone is an employee
- Multi-tenant via org_id scoping

---

## Architecture

### System Design

```
🌟 Go Workspace Monorepo
    ├─→ services/api/       (API Server Module)
    ├─→ services/cli/       (CLI Client Module)
    ├─→ pkg/types/          (Shared Types Module)
    └─→ generated/          (Generated Code Module)

PostgreSQL Schema (DB source of truth)
    ↓
    ├─→ tbls → schema.json, README.md, public.*.md, schema.svg (auto-generated)
    ├─→ Python script → ERD.md (from schema.json, auto-generated)
    └─→ sqlc → Type-safe Go database code → generated/db/

OpenAPI Spec (API source of truth)
    ↓
    └─→ oapi-codegen → Go API types + Chi server → generated/api/

Services consume generated code
    ↓
    ├─→ services/api/ imports generated/api, generated/db, pkg/types
    └─→ services/cli/ imports pkg/types (no DB/API deps!)
```

**Monorepo Approach**: Go workspace with independent modules, shared dependencies via local replace directives.

**Why Monorepo?**
- **Cleaner Dependencies**: CLI doesn't carry 50+ server deps
- **Independent Versioning**: API v0.5 + CLI v1.0 possible
- **Smaller Binaries**: CLI binary ~60% smaller (no DB drivers, HTTP handlers)
- **Better Modularity**: Clear service boundaries
- **Future-Ready**: Easy to add web UI, workers, etc.

**Hybrid Schema/API**: Database schema and API spec maintained separately, both generate code automatically.

**Why Hybrid?**
- DB tables ≠ API DTOs (different concerns)
- DB can have more tables than API exposes
- API can aggregate/transform DB data

---

## Database Schema

### Overview

**20 Tables + 3 Views**

### Core Organization (5 tables)
- `organizations` - Top-level tenant
- `subscriptions` - Billing and budget tracking
- `teams` - Group employees
- `roles` - Define permissions
- `employees` - User accounts

### Agent Management (7 tables)
- `agent_catalog` - Available AI agents (Claude Code, Cursor, etc.)
- `tools` - Available tools (fs, git, http, etc.)
- `policies` - Usage policies and restrictions
- `agent_tools` - Many-to-many: agents ↔ tools
- `agent_policies` - Many-to-many: agents ↔ policies
- `team_policies` - Team-specific policy overrides
- `employee_agent_configs` - Per-employee agent instances

### MCP Configuration (3 tables)
- `mcp_categories` - Organize MCP servers
- `mcp_catalog` - Available MCP servers
- `employee_mcp_configs` - Per-employee MCP instances

### Authentication (1 table)
- `sessions` - JWT session tracking

### Approvals (2 tables)
- `agent_requests` - Employee requests for agents/MCPs
- `approvals` - Manager approval workflow

### Analytics (2 tables)
- `activity_logs` - Audit trail
- `usage_records` - Cost and resource tracking

### Views (3)
- `v_employee_agents` - Employee agents with catalog details
- `v_employee_mcps` - Employee MCPs with catalog details
- `v_pending_approvals` - Pending approval requests with context

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**

---

## Technology Stack

- **Language:** Go 1.24+
- **Database:** PostgreSQL 15+ (multi-tenant with RLS) - 20 tables + 3 views
- **API Specification:** OpenAPI 3.0.3
- **Code Generation:** oapi-codegen, sqlc, tbls
- **HTTP Router:** Chi
- **Testing:** testcontainers-go, gomock
- **Web UI:** Next.js 14 (future)
- **Deployment:** Docker, Docker Compose

---

## Project Structure

```
ubik-enterprise/                  # 🌟 Monorepo Root
├── go.work                       # Go workspace configuration
├── Makefile                      # Automation commands
├── docker-compose.yml            # Local environment
├── CLAUDE.md                     # This file - documentation root
├── README.md                     # Quick overview
├── IMPLEMENTATION_ROADMAP.md     # Next endpoints to build
│
├── services/                     # 🎯 Microservices
│   ├── api/                      # API Server Module
│   │   ├── go.mod                # API dependencies
│   │   ├── cmd/server/main.go    # Server entry point
│   │   ├── internal/
│   │   │   ├── handlers/         # HTTP handlers
│   │   │   ├── auth/             # JWT utilities
│   │   │   ├── middleware/       # Auth, RLS, logging
│   │   │   └── service/          # Business logic
│   │   └── tests/integration/    # Integration tests
│   │
│   └── cli/                      # CLI Client Module
│       ├── go.mod                # CLI dependencies
│       ├── cmd/ubik/main.go      # CLI entry point
│       ├── internal/
│       │   ├── client/           # API client
│       │   ├── config/           # Config management
│       │   ├── docker/           # Docker SDK wrapper
│       │   └── commands/         # Cobra commands
│       └── tests/                # CLI tests
│
├── pkg/                          # 📦 Shared Go Code
│   └── types/                    # Shared domain types
│       ├── go.mod
│       └── types.go              # Common models
│
├── shared/                       # 🔧 Cross-language Shared
│   ├── openapi/
│   │   └── spec.yaml             # OpenAPI 3.0.3 spec
│   ├── schema/
│   │   └── schema.sql            # PostgreSQL schema
│   └── docker/                   # Docker images
│
├── generated/                    # ⚠️ AUTO-GENERATED (don't edit!)
│   ├── go.mod                    # Generated code module
│   ├── api/                      # From OpenAPI spec
│   ├── db/                       # From SQL queries
│   └── mocks/                    # From interfaces
│
├── sqlc/
│   ├── sqlc.yaml                 # Generator config
│   └── queries/                  # SQL queries
│
├── docs/
│   ├── QUICKSTART.md             # 5-minute setup
│   ├── TESTING.md                # Testing guide
│   ├── DEVELOPMENT.md            # Development workflow
│   ├── MONOREPO_MIGRATION.md     # Migration plan
│   ├── ERD.md                    # ⭐ Auto-generated ERD
│   ├── README.md                 # Auto-generated table index
│   └── archive/                  # Historical docs
│
└── scripts/                      # Utility scripts
```

**Monorepo Benefits:**
- 🎯 **Clean Dependencies** - Each service has minimal, focused deps
- 📦 **Independent Versioning** - API and CLI can evolve separately
- 🚀 **Smaller Binaries** - No unused code in builds
- 🔧 **Better Modularity** - Clear service boundaries
- 🌐 **Web UI Ready** - Easy to add Next.js as services/web/

---

# DOCUMENTATION

*Links to all project documentation organized by purpose.*

---

## 📚 Documentation Map

### 🔥 START HERE

**New to the project?**
1. **[docs/QUICKSTART.md](./docs/QUICKSTART.md)** - 5-minute setup guide
2. **[docs/ERD.md](./docs/ERD.md)** - Visual database schema
3. **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** ⭐ - Next endpoints to build

### 📊 Database Documentation

- **[docs/ERD.md](./docs/ERD.md)** ⭐ - **Start here!** User-friendly ERD with categories (auto-generated)
- **[docs/README.md](./docs/README.md)** - Technical reference with table index and functions (auto-generated by tbls)
- **[docs/public.*.md](./docs/)** - Per-table documentation (27 files, auto-generated by tbls)
- **[docs/schema.json](./docs/schema.json)** - Machine-readable schema (auto-generated by tbls)
- **[docs/schema.svg](./docs/schema.svg)** - Visual ERD diagram (auto-generated by tbls)

**Note:** Both ERD.md and README.md contain full Mermaid ERD diagrams but serve different purposes:
- **ERD.md** = Human-friendly overview with grouping (🏢 Core, 🤖 Agent, 🔌 MCP, etc.)
- **README.md** = Technical reference with table index and function list

### 🧪 Development & Testing

- **[docs/TESTING.md](./docs/TESTING.md)** ⭐ - Complete testing guide (TDD workflow, patterns, commands)
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** ⭐ - Development workflow and best practices
- **[IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)** ⭐⭐⭐ - **PRIORITY ORDER** for next endpoints

### 🚀 Release Management

- **[.claude/skills/release-manager/SKILL.md](./.claude/skills/release-manager/SKILL.md)** ⭐ - Release workflow and versioning
- **[docs/RELEASES.md](./docs/RELEASES.md)** - Complete release history and notes
- **[Release Manager Examples](./.claude/skills/release-manager/examples/release-examples.md)** - Real-world release workflows
- **[Release Templates](./.claude/skills/release-manager/templates/)** - Tag and release note templates

### 🤖 AI Agent Configurations

**Development Agents:**
- **[.claude/agents/go-backend-developer.md](./.claude/agents/go-backend-developer.md)** - Backend API development
- **[.claude/agents/frontend-developer.md](./.claude/agents/frontend-developer.md)** - Frontend/Next.js development

**Management & Coordination:**
- **[.claude/agents/coordinator.md](./.claude/agents/coordinator.md)** - Autonomous team orchestration
- **[.claude/agents/tech-lead.md](./.claude/agents/tech-lead.md)** - Architecture & technical leadership

**Product & Strategy:**
- **[.claude/agents/product-strategist.md](./.claude/agents/product-strategist.md)** - Feature prioritization & business value

**Quality & Review:**
- **[.claude/agents/pr-reviewer.md](./.claude/agents/pr-reviewer.md)** - Code review & quality assurance

See **[.claude/agents/README.md](./.claude/agents/README.md)** for complete agent documentation.

### 🔧 Configuration Files

- **[openapi/spec.yaml](./openapi/spec.yaml)** - OpenAPI 3.0 specification
- **[sqlc/sqlc.yaml](./sqlc/sqlc.yaml)** - sqlc configuration
- **[sqlc/queries/*.sql](./sqlc/queries/)** - Type-safe SQL queries
- **[docker-compose.yml](./docker-compose.yml)** - Local development environment
- **[Makefile](./Makefile)** - Automation commands

### 📦 Archived Documentation

- **[docs/archive/MIGRATION_PLAN.md](./docs/archive/MIGRATION_PLAN.md)** - Original 10-week plan (historical reference)
- **[docs/archive/INIT_COMPLETE.md](./docs/archive/INIT_COMPLETE.md)** - Phase 1 completion summary
- **[docs/archive/SETUP_COMPLETE.md](./docs/archive/SETUP_COMPLETE.md)** - Initial setup notes
- **[docs/archive/DOCUMENTATION_COMPLETE.md](./docs/archive/DOCUMENTATION_COMPLETE.md)** - Docs overview

---

## Quick Reference

### Quick Start

```bash
# Start database
cd ubik-enterprise
make db-up

# Install tools (one-time)
make install-tools

# Install Git hooks (one-time, auto-generates code on commit)
make install-hooks

# Generate all code
make generate

# View documentation
open docs/ERD.md
```

### Common Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)

# Setup
make install-tools      # Install code generation tools
make install-hooks      # Install Git hooks (auto-generates code on commit)

# Code Generation
make generate           # Generate everything (ERD + API + DB + Mocks)
make generate-erd       # Generate ERD docs (ERD.md + README.md + schema.json + SVG)
make generate-api       # Generate API code only
make generate-db        # Generate DB code only
make generate-mocks     # Generate mocks only

# Testing
make test               # Run all tests with coverage
make test-unit          # Run unit tests only (fast)
make test-integration   # Run integration tests (requires Docker)
make test-coverage      # Generate HTML coverage report

# Development
make dev                # Start dev server (once implemented)
make build              # Build binaries
make clean              # Clean generated files

# Help
make help               # Show all commands
```

### Database Access

```bash
# PostgreSQL connection
postgres://ubik:ubik_dev_password@localhost:5432/ubik

# Adminer web UI
open http://localhost:8080

# psql CLI
docker exec ubik-postgres psql -U ubik -d ubik
```

### MCP Servers

**Claude Code** uses Model Context Protocol (MCP) servers for enhanced capabilities.

**Currently Configured:**
- ✅ **playwright** - Browser automation and web interaction
- ✅ **github** - GitHub operations (issues, PRs, repos, code search)

**Manage MCP Servers:**
```bash
# List configured servers
claude mcp list

# Get details about a server
claude mcp get github

# Add a new MCP server
claude mcp add <name> -- <command>

# Remove an MCP server
claude mcp remove <name> -s local
```

**GitHub MCP Server:**
```bash
# Verify GitHub MCP is connected
claude mcp list | grep github

# If disconnected, restart Claude Code
# The Docker container auto-starts with Claude Code

# If container issue, check Docker
docker ps | grep github-mcp-server
docker images | grep github-mcp-server

# Re-add if needed
claude mcp add github \
  -e GITHUB_PERSONAL_ACCESS_TOKEN=$(gh auth token) \
  -- docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server
```

**Troubleshooting:**
- **Container not running**: MCP containers auto-start when Claude Code needs them (not persistent)
- **Connection failed**: Check Docker is running: `docker ps`
- **Image missing**: Re-pull image: `docker pull ghcr.io/github/github-mcp-server`
- **Token expired**: Update token: `gh auth refresh` then `claude mcp remove github -s local` and re-add
- **Config location**: `~/.claude.json` (project-specific) or global config

**Available GitHub MCP Operations:**
- Create/update/list issues and PRs
- Search code across repositories
- Manage branches and files
- View repository details
- Monitor CI/CD workflows
- Code security scanning

**Note:** MCP servers are configured per-project in `~/.claude.json` (local scope) or globally. The GitHub MCP server uses the official Docker image `ghcr.io/github/github-mcp-server` maintained by GitHub.

---

# DEVELOPMENT

*Essential information for working with the codebase.*

---

## Development Workflow

### ⚠️ MANDATORY: Standard PR Workflow

**ALL code changes MUST follow the standard Git workflow.**

**See [docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md) for the complete workflow guide.**

**Quick summary - Required steps for EVERY change:**
1. ✅ Create feature branch from `main`
2. ✅ Implement changes (following TDD)
3. ✅ Run tests locally
4. ✅ Commit with descriptive message
5. ✅ Push to remote
6. ✅ Create Pull Request
7. ✅ Wait for CI/CD checks to pass
8. ✅ Review and merge
9. ✅ Delete feature branch

**No exceptions.** This applies to:
- Backend API changes (go-backend-developer agent)
- Frontend web changes (frontend-developer agent)
- CLI changes
- Documentation changes
- Database migrations

### First-Time Setup

```bash
# Install code generation tools (one-time)
make install-tools

# Install Git hooks for auto code generation (one-time)
make install-hooks
```

### Making Changes (Example)

```bash
# 1. Create feature branch
git checkout main && git pull
git checkout -b feature/my-feature

# 2. Update database schema (if needed)
vim schema.sql

# 3. Apply to database
make db-reset

# 4. Update OpenAPI spec (if API changes)
vim openapi/spec.yaml

# 5. Update SQL queries (if needed)
vim sqlc/queries/employees.sql

# 6. Implement handlers
vim internal/handlers/employees.go

# 7. Run tests
make test

# 8. Commit changes (Git hook auto-generates code!)
git add .
git commit -m "feat: Add new feature (#<issue>)

Details here...

Closes #<issue>"
# 🪝 Pre-commit hook will:
#   - Detect changes to source models (schema.sql, openapi/spec.yaml, sqlc queries)
#   - Run `make generate` automatically
#   - Add generated files to your commit

# 9. Push and create PR (see DEV_WORKFLOW.md for full PR workflow)
git push -u origin feature/my-feature
gh pr create --title "feat: My Feature (#<issue>)" --body "..."

# 10. Wait for CI/CD checks, then merge
```

**Note:** If you need to skip code generation (not recommended):
```bash
git commit --no-verify -m "your message"
```

### CI-Aware Development Workflow

**IMPORTANT:** Both `go-backend-developer` and `frontend-developer` agents now automatically wait for CI checks before completing tasks.

**Complete Workflow (Automated by Agents):**

```bash
# 1. Pick task from GitHub Projects
gh issue list --label="backend,status/ready"

# 2. Create branch + workspace
git checkout -b issue-123-feature
git worktree add ../ubik-issue-123 issue-123-feature
cd ../ubik-issue-123

# 3. Implement feature (TDD)
# - Write failing tests
# - Implement code
# - All tests pass locally

# 4. Create PR
gh pr create --title "..." --body "..." --label "backend"
PR_NUM=$(gh pr view --json number -q .number)

# 5. Wait for CI checks (CRITICAL!)
gh pr checks $PR_NUM --watch --interval 10

# 6. Verify CI passed
CI_STATUS=$(gh pr checks $PR_NUM --json state -q 'map(select(.state == "FAILURE" or .state == "CANCELLED")) | length')

if [ "$CI_STATUS" -eq 0 ]; then
  # All checks passed!
  ./scripts/update-project-status.sh --issue 123 --status "In Review"
  gh issue edit 123 --add-label "status/waiting-for-review"
  gh issue comment 123 --body "✅ All CI checks passed. PR #${PR_NUM} ready for review."
else
  # Checks failed - investigate and fix
  gh pr checks $PR_NUM
  gh issue comment 123 --body "❌ CI checks failed for PR #${PR_NUM}. Investigating..."
  # Fix failures and push again
fi
```

**Why This Matters:**

- ✅ **Quality Gate**: All tests must pass in CI before moving to review
- ✅ **Automated**: Agents handle the entire workflow without manual intervention
- ✅ **Visibility**: GitHub Project status auto-updates to "In Review" when ready
- ✅ **Fast Feedback**: Failures caught immediately, not during code review
- ✅ **Clean Pipeline**: No broken PRs waiting for review

**Helper Script:**

```bash
# Update GitHub Project status for an issue
./scripts/update-project-status.sh --issue 123 --status "In Review"
./scripts/update-project-status.sh --issue 123 --status "Done"
./scripts/update-project-status.sh --issue 123 --status "In Progress"

# For marketing board
./scripts/update-project-status.sh --issue 124 --status "Launched" --project marketing
```

**See agent configurations:**
- `.claude/agents/go-backend-developer.md` - Backend workflow (versioned in project)
- `.claude/agents/frontend-developer.md` - Frontend workflow (versioned in project)

### Milestone Planning and Transitions

**Complete workflow for planning milestones and transitioning between releases.**

#### After Releasing a Milestone

**1. Archive Completed Milestone**

```bash
# Archive all issues from completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0
```

This will:
- Label all milestone issues as "archived"
- Close any remaining open issues
- Close the milestone
- Update `docs/MILESTONES_ARCHIVE.md` with completion record

**2. Start New Milestone**

```bash
# Create new milestone and populate from backlog
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics Dashboard & Approval Workflows" \
  --due-date "2026-01-31" \
  --auto-split
```

This will:
1. Create GitHub milestone with description and due date
2. Query backlog for `priority/p0` and `priority/p1` issues
3. Display issues for review and confirmation
4. Add confirmed issues to milestone
5. Move issues to "Todo" status on project board
6. Flag large tasks (size/l, size/xl) for splitting
7. Create milestone kickoff issue

**3. Split Large Tasks**

```bash
# Find tasks flagged for splitting
gh issue list --label "needs-splitting" --milestone "v0.4.0"

# Split a large task
./scripts/split-large-tasks.sh --issue 51

# Or use auto-split with github-task-manager skill
./scripts/split-large-tasks.sh --issue 51 --auto
```

This helps break down size/xl and size/l tasks into manageable subtasks (size/s or size/m).

#### Milestone Planning Best Practices

**Before Starting New Milestone:**
- ✅ Review backlog and update priorities
- ✅ Ensure issue descriptions are clear
- ✅ Verify all issues have size labels
- ✅ Check for dependencies between issues
- ✅ Set realistic due date (4-6 weeks typical)

**When Populating Milestone:**
- ✅ Focus on p0/p1 priority issues
- ✅ Aim for mix of sizes (not all large tasks)
- ✅ Balance features vs bug fixes vs tech debt
- ✅ Include testing and documentation tasks
- ✅ Leave buffer for unexpected work (70-80% capacity)

**Task Splitting Guidelines:**
- ✅ Each subtask should be independently testable
- ✅ Subtasks should be size/s or size/m (1-3 days each)
- ✅ Use parent-child relationship (blockedBy in GitHub)
- ✅ Update parent with checklist of subtasks
- ✅ Close parent only when all subtasks complete

**Complete Workflow Example:**

```bash
# After releasing v0.3.0, transition to v0.4.0

# 1. Archive completed milestone
./scripts/archive-milestone.sh --milestone v0.3.0

# 2. Start new milestone
./scripts/start-milestone.sh \
  --version v0.4.0 \
  --description "Analytics & Approvals" \
  --due-date "2026-01-31" \
  --auto-split

# 3. Split flagged large tasks
for issue in $(gh issue list --label "needs-splitting" --milestone "v0.4.0" --json number -q '.[].number'); do
  ./scripts/split-large-tasks.sh --issue $issue
done

# 4. Verify milestone ready
gh issue list --milestone "v0.4.0" --json number,title,labels

# 5. Start working on first task
FIRST_TASK=$(gh issue list --milestone "v0.4.0" --label "priority/p0" --assignee "" --limit 1 --json number -q '.[0].number')
git checkout -b issue-$FIRST_TASK-feature
./scripts/update-project-status.sh --issue $FIRST_TASK --status "In Progress"
```

**See [Release Manager Skill](./.claude/skills/release-manager/SKILL.md) for complete release and milestone transition workflows.**

### Code Generation Pipeline

```
schema.sql → PostgreSQL → tbls → schema.json, README.md, public.*.md, schema.svg
                        ↓         ↓
                       sqlc      Python script → ERD.md (user-friendly)
                        ↓
                  generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go

Your code (internal/) → Uses generated types
```

**ERD Documentation:**
- `make generate-erd` creates **both** README.md (tbls) and ERD.md (custom script)
- README.md = Technical reference with table index
- ERD.md = User-friendly overview with categories

**See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for complete development guide.**

---

## Critical Notes

### Code Generation

**Never edit files in `generated/`** - they are completely regenerated!

**Automatic Code Generation (Recommended):**
```bash
# Install Git hooks (one-time setup)
make install-hooks

# Now code is auto-generated on commit!
git commit -m "feat: Update schema"
# 🪝 Pre-commit hook automatically:
#   - Detects changes to source models
#   - Runs `make generate`
#   - Adds generated files to commit
```

**Manual Code Generation (if needed):**
```bash
# After changing schema.sql
make db-reset && make generate-db && make generate-mocks

# After changing openapi/spec.yaml
make generate-api

# After changing SQL queries
make generate-db && make generate-mocks

# Or regenerate everything
make generate
```

### Multi-Tenancy

**All queries must be org-scoped:**

```go
// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ BAD - Exposes all orgs!
employees, err := db.ListAllEmployees(ctx)
```

Use Row-Level Security (RLS) policies as safety net.

### Testing Strategy

**⚠️ CRITICAL: ALWAYS FOLLOW STRICT TDD (Test-Driven Development)**

**Mandatory TDD Workflow:**
```
✅ 1. Write failing tests FIRST
✅ 2. Implement minimal code to pass tests
✅ 3. Refactor with tests passing
❌ NEVER write implementation before tests
```

**Example:**
```
✅ CORRECT: Write router tests → Implement router wiring → Tests pass
❌ WRONG:   Implement router wiring → Write tests later
❌ WRONG:   Implement handler → Write tests after
```

**Target Coverage:** 85% overall (excluding generated code)

**Why TDD is Mandatory:**
- Prevents regression bugs
- Forces good API design
- Ensures all code is testable
- Provides immediate feedback
- Builds confidence in changes

**See [docs/TESTING.md](./docs/TESTING.md) for complete testing guide with TDD workflow.**

### UI Development Workflow

**⚠️ CRITICAL: Wireframes Required for All UI Changes**

**Mandatory UI Workflow:**
```
✅ 1. Create wireframe FIRST (for new pages)
✅ 2. Update wireframes (for page changes)
✅ 3. Implement UI matching wireframe
✅ 4. Keep wireframes in sync with implementation
❌ NEVER implement new pages without wireframes
❌ NEVER modify pages without updating wireframes
```

**Why Wireframes are Mandatory:**
- Ensures design consistency across the platform
- Provides visual documentation of the UI
- Facilitates design review before implementation
- Prevents rework from design changes
- Serves as reference for future modifications

**Wireframe Location:**
- Store wireframes in `docs/wireframes/` directory
- Use descriptive names (e.g., `settings-page.png`, `employee-detail.png`)
- Include wireframes in pull requests with UI changes

### Release Management

**⚠️ CRITICAL: Follow Standardized Release Process**

**Use the Release Manager Skill for all releases** - ensures consistency across all agents.

**Quick Release Checklist:**
```
✅ 1. All CI/CD checks passing
✅ 2. All tests passing (make test)
✅ 3. Milestone issues closed
✅ 4. On main branch, clean working tree
✅ 5. Documentation updated
✅ 6. Create annotated git tag
✅ 7. Push tag to remote
✅ 8. Create GitHub Release
✅ 9. Update CLAUDE.md and docs/RELEASES.md
```

**Versioning Strategy:**
- **v0.x.0** - New milestone features (Web UI, Analytics, etc.)
- **v0.x.y** - Bug fixes and polish within a milestone
- **v1.0.0+** - Production releases (post-launch)

**Key Commands:**
```bash
# Check release readiness
gh run list --limit 1  # Verify CI green
make test              # Run all tests
gh issue list --milestone "v0.X.0" --state open  # Check milestone

# Create release
git tag -a v0.X.0 -m "Release v0.X.0 - [Description]"
git push origin v0.X.0
gh release create v0.X.0 --title "..." --notes "..."
```

**See [Release Manager Skill](./.claude/skills/release-manager/SKILL.md) for complete workflow.**

**Release History:** See [docs/RELEASES.md](./docs/RELEASES.md)

---

# STATUS & ROADMAP

*Current progress and next steps.*

---

## Current Status

**Last Updated:** 2025-10-29
**Version:** 0.2.0 🎉 + **Monorepo Migration Complete** 🌟
**Status:** 🟢 **CLI Phase 4 Complete - Ready for v0.2.0 Release**
**Git Tag:** `v0.1.0` (v0.2.0 tag pending)
**Branch:** `feature/monorepo-migration` (ready to merge)

### 🎉 Milestone v0.1.0 Released!

**39 API endpoints implemented** | **144+ tests passing** | **73-88% coverage**

See **[docs/MILESTONE_v0.1.md](./docs/MILESTONE_v0.1.md)** for complete release notes.

### Phase 1 Achievements ✅
- Complete database schema (20 tables + 3 views)
- Code generation pipeline working
- 60+ documentation files
- OpenAPI spec for all endpoints
- Type-safe SQL queries
- Local development environment
- Comprehensive Mermaid ERD

### Phase 2 Achievements ✅ (v0.1.0)
- **Complete authentication system** with JWT + sessions
- **Employee CRUD** - Full lifecycle management (5 endpoints)
- **Organization management** - Get/Update current org (2 endpoints)
- **Team management** - Full CRUD (5 endpoints)
- **Role management** - Full CRUD (5 endpoints)
- **Agent catalog** - List and get agents (2 endpoints)
- **Agent configurations** - Org/Team/Employee configs (16 endpoints)
- **144+ tests passing** (119 unit + 25+ integration)
- **73-88% code coverage** across all modules
- Multi-tenancy verified with integration tests
- TDD workflow throughout

**Test Coverage by Module:**
- `internal/handlers`: 73.3%
- `internal/auth`: 88.2%
- `internal/middleware`: 82.2%
- `internal/service`: 77.8%

### Phase 3 - CLI Development 🎯 (In Progress)

**Current Focus:** Employee CLI Client (v0.2.0)

**✅ Phase 0 - Docker Images (Complete)**
- Docker images for Claude Code and MCP servers built
- See [docker/README.md](./docker/README.md)

**✅ CLI Phase 1 - Foundation (Complete)**
- CLI project structure with cobra
- Authentication (`ubik login`, `ubik logout`)
- Platform API client
- Config management (`~/.ubik/config.json`)
- Sync service (`ubik sync` - fetch configs)
- 13 unit tests passing (100%)
- See **[docs/CLI_PHASE1_COMPLETE.md](./docs/CLI_PHASE1_COMPLETE.md)** for details

**✅ CLI Phase 2 - Docker Integration (Complete)**
- Docker SDK integration & client wrapper
- Container lifecycle management (start/stop/status)
- Network management (`ubik-network`)
- MCP server orchestration
- Enhanced `ubik sync --start-containers`
- New commands: `ubik start`, `ubik stop`
- **42 tests passing (24 unit + 18 integration)** ✅
- Coverage: ~22% (unit only), ~60-70% (with Docker)
- Comprehensive error handling and edge case coverage
- See **[docs/CLI_PHASE2_COMPLETE.md](./docs/CLI_PHASE2_COMPLETE.md)** for details

**✅ CLI Phase 3 - Interactive Mode (Complete)** ⭐
- Interactive workspace selection with prompt & validation
- I/O proxying to agent container (bidirectional streaming)
- Agent switching with --agent flag
- Session management (tracking, duration, metadata)
- Interactive `ubik` command (seamless agent interaction)
- **73 tests passing (~38 unit + ~35 integration)** ✅
- Coverage: ~25% (unit only), ~65-75% (with Docker)
- See **[docs/CLI_PHASE3_COMPLETE.md](./docs/CLI_PHASE3_COMPLETE.md)** for details

**✅ CLI Phase 4 - Agent Management (Complete)** ⭐⭐⭐
- `ubik agents list` - List available/local agents
- `ubik agents info <id>` - Get agent details
- `ubik agents request <id>` - Request agent access
- `ubik update` - Check for config updates (with --sync)
- `ubik cleanup` - Clean containers/config
- **TTY raw mode fix** - Interactive input working perfectly
- **Automatic container cleanup** - No more name conflicts
- **Docker images built** - Claude Code (1.8GB), MCP filesystem (252MB), MCP git (422MB)
- **79 tests passing (6 new + 3 skipped)** ✅
- **100% pass rate maintained**
- See **[docs/CLI_PHASE4_COMPLETE.md](./docs/CLI_PHASE4_COMPLETE.md)** for details
- See **[docs/CLI_TTY_FIX.md](./docs/CLI_TTY_FIX.md)** for TTY troubleshooting
- See **[CHANGELOG.md](./CHANGELOG.md)** for v0.2.0 release notes
- See **[INSTALL.md](./INSTALL.md)** for installation guide
- See **[MARKETING.md](./MARKETING.md)** for product strategy

**See [docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md) for complete architecture.**

**Future Phases:**
- Phase 4: Agent management & approvals
- Phase 5: Polish & telemetry
- v0.3+: System prompts, MCP management, Web UI

**See [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md) for detailed plan.**

---

## Success Metrics

### Phase 1 (Complete) ✅
- ✅ 20 tables + 3 views in PostgreSQL
- ✅ 60+ documentation files (streamlined)
- ✅ OpenAPI spec with 10+ endpoints
- ✅ 15+ type-safe SQL queries
- ✅ Code generation working end-to-end
- ✅ Complete automation via Makefile

### Phase 2 (Complete) ✅
- ✅ Authentication working (JWT + sessions)
- ✅ JWT Middleware implemented
- ✅ 43/43 tests passing
- ✅ ~88% code coverage
- ✅ Integration tests with real PostgreSQL

### Phase 3 (In Progress) 🎯
- [ ] Employee CRUD endpoints (0/5 complete)
- [ ] Organization management endpoints
- [ ] API response time <100ms (p95)
- [ ] 80%+ test coverage overall

---

## Roadmap

### Phase 1: ✅ COMPLETED (Foundation)
- Database schema, code generation, documentation

### Phase 2: ✅ COMPLETED (Authentication)
- JWT authentication, sessions, middleware

### Phase 3: In Progress (Employee Management)
- Employee CRUD, organization management

### Phase 4-8: Planned
- Agent/MCP configuration APIs
- Approval workflows
- Analytics endpoints
- Employee CLI client
- Admin web UI
- Production deployment

**See [docs/archive/MIGRATION_PLAN.md](./docs/archive/MIGRATION_PLAN.md) for original 10-week plan (historical reference).**

---

## Documentation Standards

### When to Update Docs

**Always update when:**
- Adding new tables → Regenerate ERD: `make generate-erd`
- Adding API endpoints → Update `openapi/spec.yaml`
- Adding SQL queries → Add to `sqlc/queries/*.sql`
- Changing architecture → Update this file (CLAUDE.md)

**How to update:**
```bash
# Database docs (auto-generated)
make generate-erd

# Manual docs (update manually)
vim CLAUDE.md
vim docs/TESTING.md
vim docs/DEVELOPMENT.md
```

---

**Foundation documentation is above this line. For implementation details, see linked documents.**

**Key Links:**
- 🚀 [Get Started](./docs/QUICKSTART.md)
- 🎯 [Next Tasks](./IMPLEMENTATION_ROADMAP.md)
- 🧪 [Testing Guide](./docs/TESTING.md)
- 🔧 [Development Guide](./docs/DEVELOPMENT.md)
- 📊 [Database ERD](./docs/ERD.md)

## Critical Usage Notes

### ✅ Qdrant MCP for Knowledge Management

**MANDATORY: Use Qdrant MCP for all knowledge operations**

- **Before any task**: Search using `mcp__code-search__qdrant-find`
- **During work**: Store findings using `mcp__code-search__qdrant-store`
- **After completion**: Update stored knowledge

1. Index important findings in Qdrant as you work
2. Use Qdrant for search/discovery →
  then read full .md file

  What to store in Qdrant:
  - Solutions to specific problems you
  solved
  - "Why we chose X over Y" decisions
  - Performance lessons ("approach X was
  10x faster")
  - Failed attempts and why they didn't
  work
  - Code patterns that work well in this
  codebase

  What to keep in .md:
  - Architecture overviews
  - Getting started guides
  - API references
  - Comprehensive feature docs

**Setup:**
```bash
docker run -d --name claude-qdrant -p 6333:6333 qdrant/qdrant:latest
```