# Ubik Enterprise — AI Agent Management Platform

**Multi-tenant SaaS platform for centralized AI agent and MCP configuration management**

---

## 📑 Table of Contents

### Foundation
- [System Overview](#system-overview)
- [Architecture](#architecture)
- [Database Schema](#database-schema)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)

### Documentation & Resources
- [Documentation Map](#-documentation-map)
- [Quick Start](#quick-start)

### Development
- [Development Essentials](#development-essentials)
- [Critical Rules](#critical-rules)

### Status & Roadmap
- [Current Status](#current-status)
- [Roadmap](#roadmap)

---

# FOUNDATION

*Stable foundation of the system - rarely changes*

---

## System Overview

### Purpose

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

**Monorepo Benefits:**
- **Cleaner Dependencies**: CLI doesn't carry 50+ server deps
- **Independent Versioning**: API v0.5 + CLI v1.0 possible
- **Smaller Binaries**: CLI binary ~60% smaller (no DB drivers, HTTP handlers)
- **Better Modularity**: Clear service boundaries
- **Future-Ready**: Easy to add web UI, workers, etc.

**Hybrid Schema/API**: Database schema and API spec maintained separately, both generate code automatically.

---

## Database Schema

### Overview

**20 Tables + 3 Views**

| Category | Tables | Count |
|----------|--------|-------|
| Core Organization | organizations, subscriptions, teams, roles, employees | 5 |
| Agent Management | agent_catalog, tools, policies, agent_tools, agent_policies, team_policies, employee_agent_configs | 7 |
| MCP Configuration | mcp_categories, mcp_catalog, employee_mcp_configs | 3 |
| Authentication | sessions | 1 |
| Approvals | agent_requests, approvals | 2 |
| Analytics | activity_logs, usage_records | 2 |
| **Views** | v_employee_agents, v_employee_mcps, v_pending_approvals | 3 |

**See [docs/ERD.md](./docs/ERD.md) for complete visual schema.**
**See [docs/DATABASE.md](./docs/DATABASE.md) for database operations and best practices.**

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
│
├── services/                     # 🎯 Microservices
│   ├── api/                      # API Server Module
│   └── cli/                      # CLI Client Module
│
├── pkg/types/                    # 📦 Shared Go Code
├── shared/                       # 🔧 Cross-language Shared
│   ├── openapi/spec.yaml         # OpenAPI 3.0.3 spec
│   └── schema/schema.sql         # PostgreSQL schema
│
├── generated/                    # ⚠️ AUTO-GENERATED (don't edit!)
├── sqlc/                         # SQL queries
├── docs/                         # Documentation
└── scripts/                      # Utility scripts
```

**See project structure details in specific service README files.**

---

# DOCUMENTATION

*Links to all project documentation organized by purpose*

---

## 📚 Documentation Map

### 🔥 START HERE

**New to the project?**
1. **[docs/QUICKSTART.md](./docs/QUICKSTART.md)** - 5-minute setup guide
2. **[docs/ERD.md](./docs/ERD.md)** - Visual database schema

### 📖 Core Documentation

**Development:**
- **[docs/QUICK_REFERENCE.md](./docs/QUICK_REFERENCE.md)** - Commands and operations cheat sheet
- **[docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)** - Development workflow and best practices
- **[docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md)** ⭐ - Standard PR & Git workflow (mandatory)
- **[docs/TESTING.md](./docs/TESTING.md)** ⭐ - Complete testing guide (TDD workflow, patterns, commands)
- **[docs/DEBUGGING.md](./docs/DEBUGGING.md)** - Debugging strategies and common pitfalls

**Database:**
- **[docs/DATABASE.md](./docs/DATABASE.md)** - Database operations, access, and best practices
- **[docs/ERD.md](./docs/ERD.md)** ⭐ - User-friendly ERD with categories (auto-generated)
- **[docs/README.md](./docs/README.md)** - Technical reference with table index (auto-generated by tbls)
- **[docs/public.*.md](./docs/)** - Per-table documentation (27 files, auto-generated by tbls)

**Operations:**
- **[docs/MCP_SERVERS.md](./docs/MCP_SERVERS.md)** - MCP server setup and configuration
- **[docs/WORKFLOWS.md](./docs/WORKFLOWS.md)** - Milestone planning, releases, task management

**CLI:**
- **[docs/CLI_CLIENT.md](./docs/CLI_CLIENT.md)** - CLI architecture and design
- **[docs/CLI_PHASE1_COMPLETE.md](./docs/CLI_PHASE1_COMPLETE.md)** - CLI Phase 1 details
- **[docs/CLI_PHASE2_COMPLETE.md](./docs/CLI_PHASE2_COMPLETE.md)** - CLI Phase 2 details
- **[docs/CLI_PHASE3_COMPLETE.md](./docs/CLI_PHASE3_COMPLETE.md)** - CLI Phase 3 details
- **[docs/CLI_PHASE4_COMPLETE.md](./docs/CLI_PHASE4_COMPLETE.md)** - CLI Phase 4 details

### 🚀 Release Management

- **[.claude/skills/release-manager/SKILL.md](./.claude/skills/release-manager/SKILL.md)** ⭐ - Release workflow and versioning
- **[docs/RELEASES.md](./docs/RELEASES.md)** - Complete release history and notes
- **[docs/WORKFLOWS.md](./docs/WORKFLOWS.md)** - Milestone planning and transitions

### 🤖 AI Agent Configurations

- **[.claude/agents/go-backend-developer.md](./.claude/agents/go-backend-developer.md)** - Backend API development
- **[.claude/agents/frontend-developer.md](./.claude/agents/frontend-developer.md)** - Frontend/Next.js development
- **[.claude/agents/product-designer.md](./.claude/agents/product-designer.md)** - Wireframes, UI/UX design & accessibility
- **[.claude/agents/coordinator.md](./.claude/agents/coordinator.md)** - Autonomous team orchestration
- **[.claude/agents/tech-lead.md](./.claude/agents/tech-lead.md)** - Architecture & technical leadership
- **[.claude/agents/product-strategist.md](./.claude/agents/product-strategist.md)** - Feature prioritization
- **[.claude/agents/pr-reviewer.md](./.claude/agents/pr-reviewer.md)** - Code review & QA

**See [.claude/agents/README.md](./.claude/agents/README.md) for complete agent documentation.**

---

## Quick Start

### First-Time Setup

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

# Run tests
make test

# View documentation
open docs/ERD.md
```

**See [docs/QUICKSTART.md](./docs/QUICKSTART.md) for detailed setup guide.**

### Essential Commands

```bash
# Database
make db-up              # Start PostgreSQL
make db-down            # Stop PostgreSQL
make db-reset           # Reset database (⚠️ deletes data)

# Code Generation
make generate           # Generate everything (ERD + API + DB + Mocks)

# Testing
make test               # Run all tests with coverage
make test-unit          # Run unit tests only (fast)
make test-integration   # Run integration tests (requires Docker)

# Development
make build              # Build binaries
make clean              # Clean generated files
```

**See [docs/QUICK_REFERENCE.md](./docs/QUICK_REFERENCE.md) for complete command reference.**

### MCP Servers

**Currently Configured:**
- ✅ **github** - GitHub operations (issues, PRs, repos, code search)
- ✅ **playwright** - Browser automation and web interaction
- ✅ **qdrant** - Vector search and knowledge management (ACTIVE - use for all knowledge operations!)
- ✅ **gcloud** - Google Cloud Platform operations (projects, services, compute, storage)
- ✅ **observability** - Google Cloud monitoring and logging
- ⚠️ **postgres** - Database operations (manual setup)

**See [docs/MCP_SERVERS.md](./docs/MCP_SERVERS.md) for complete setup and usage guide.**

---

# DEVELOPMENT

*Essential information for working with the codebase*

---

## Development Essentials

### Standard Workflow

**⚠️ MANDATORY: ALL code changes MUST follow the standard PR workflow.**

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

**See [docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md) for the complete mandatory workflow.**

### First-Time Setup

```bash
# Install code generation tools (one-time, if you plan to modify schema/API)
make install-tools

# Generate code before building/testing
make generate
```

### Code Generation Pipeline

```
shared/schema/schema.sql → PostgreSQL → tbls → ERD docs (auto-generated)
                        ↓
                       sqlc → generated/db/*.go

openapi/spec.yaml → oapi-codegen → generated/api/server.gen.go
```

**When to regenerate:**
- **CI/CD (automatic):** On every build/test in GitHub Actions
- **Local (manual):** After changing shared/schema/schema.sql, openapi/spec.yaml, or SQL queries
- **After pull:** When pulling changes that modify source files

```bash
# Regenerate everything
make generate

# Or specific parts
make generate-api  # After changing openapi/spec.yaml
make generate-db   # After changing SQL queries
make generate-erd  # After changing shared/schema/schema.sql
```

**Note:** The `generated/` directory is NOT committed to git. CI/CD handles generation automatically.

**See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for detailed development guide.**

---

## Critical Rules

### 1. Code Generation

**Never edit generated files** - they are completely regenerated!

**After changing database schema:**
```bash
# 1. Edit the schema
vim shared/schema/schema.sql

# 2. Regenerate EVERYTHING (code + docs)
make generate

# 3. Commit both schema and docs
git add shared/schema/schema.sql docs/
git commit -m "feat: Add new table"
```

**What gets committed:**
- ✅ Source files (`shared/schema/schema.sql`, `openapi/spec.yaml`, SQL queries)
- ✅ Documentation (`docs/` - ERD, README, per-table docs)
- ❌ Generated code (`generated/` - NOT committed)

**CI/CD enforces this:**
- Regenerates Go code automatically (not committed)
- Regenerates ERD docs and FAILS if they're stale
- This catches when developers forget to run `make generate-erd`

---

### 2. Multi-Tenancy

**All queries must be org-scoped:**

```go
// ✅ GOOD - Scoped to organization
employees, err := db.ListEmployees(ctx, orgID, status)

// ❌ BAD - Exposes all orgs!
employees, err := db.ListAllEmployees(ctx)
```

Use Row-Level Security (RLS) policies as safety net.

**See [docs/DATABASE.md](./docs/DATABASE.md#multi-tenancy) for RLS details.**

---

### 3. Testing Strategy

**⚠️ CRITICAL: ALWAYS FOLLOW STRICT TDD (Test-Driven Development)**

**Mandatory TDD Workflow:**
```
✅ 1. Write failing tests FIRST
✅ 2. Implement minimal code to pass tests
✅ 3. Refactor with tests passing
❌ NEVER write implementation before tests
```

**Target Coverage:** 85% overall (excluding generated code)

**See [docs/TESTING.md](./docs/TESTING.md) for complete testing guide with TDD workflow.**

---

### 4. UI Development

**⚠️ CRITICAL: Wireframes Required for All UI Changes**

**Mandatory UI Workflow:**
- ✅ Request wireframes from **product-designer agent** FIRST (for new pages)
- ✅ Request updated wireframes from **product-designer agent** (for page changes)
- ✅ Wait for wireframes before starting implementation
- ✅ Implement UI matching wireframes exactly
- ❌ NEVER implement new UI without wireframes from product-designer

**Wireframe Location:** `docs/wireframes/` directory

**Product Designer Agent:** Senior UX/UI expert responsible for all wireframes, user flows, and accessibility compliance.

**See:**
- [.claude/agents/product-designer.md](./.claude/agents/product-designer.md) - Product designer agent configuration
- [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md#ui-development) - Complete UI workflow

---

### 5. Release Management

**Use the Release Manager Skill for all releases** - ensures consistency across all agents.

**Quick Release Checklist:**
- ✅ All CI/CD checks passing
- ✅ All tests passing
- ✅ Milestone issues closed
- ✅ On main branch, clean working tree
- ✅ Documentation updated
- ✅ Create annotated git tag
- ✅ Push tag to remote
- ✅ Create GitHub Release

**See [.claude/skills/release-manager/SKILL.md](./.claude/skills/release-manager/SKILL.md) for complete workflow.**
**See [docs/WORKFLOWS.md](./docs/WORKFLOWS.md) for milestone planning.**

---

### 6. Debugging Best Practices

**The Golden Rule: Check the Data, Not Just the Code**

When integration tests or CLI operations fail unexpectedly:

1. ✅ **Add Request/Response Logging First**
2. ✅ **Verify Database State** - Foreign keys, seed data
3. ✅ **Check for Stale Cache** - `~/.ubik/`, binaries
4. ✅ **Rebuild Binaries** - Ensure latest code

**Common Pitfalls:**
- ❌ Assuming code is wrong when data is wrong
- ❌ Not checking cache invalidation
- ❌ Testing with stale binaries
- ❌ Missing org-level configurations

**See [docs/DEBUGGING.md](./docs/DEBUGGING.md) for complete debugging guide with real-world examples.**

---

### 7. PR-Based Development Workflow

**⚠️ CRITICAL: ALL code changes MUST go through Pull Requests**

**Branch Protection Enforcement:**
- ✅ Direct commits to `main` are **BLOCKED** by branch protection
- ✅ All changes require Pull Request approval
- ✅ Branches auto-delete after PR merge

**Mandatory PR Workflow:**

1. **Create Feature Branch:**
   ```bash
   # Format: feature/{issue-number}-{description}
   git checkout -b feature/138-update-prompts
   ```

2. **PR Title Format (REQUIRED):**
   ```
   feat: Description (#138)
   fix: Bug description (#139)
   chore: Maintenance task (#140)
   ```
   **The issue number in title is critical for automatic linking!**

3. **Automatic Status Transitions (GitHub Actions):**
   - ✅ PR opened → Issue gets `status/in-review` label + comment
   - ✅ PR merged → Issue gets `status/done` label, closes automatically
   - ✅ Branch deleted automatically after merge

4. **CI Checks (MANDATORY):**
   - ✅ All tests must pass
   - ✅ Lint checks must pass
   - ✅ Build must succeed
   - ❌ **NEVER merge with failing CI**

**What's Automated:**
- Issue status updates (`status/in-review`, `status/done`)
- Issue closure (via `Closes #123` or PR title)
- Branch deletion after merge
- Status comments on issues

**What You Still Do:**
- Create feature branch
- Write code following TDD
- Create PR with proper title format
- Wait for CI checks to pass
- Merge PR when approved

**See [docs/DEV_WORKFLOW.md](./docs/DEV_WORKFLOW.md) for complete workflow guide.**

---

### 8. Docker Testing

**⚠️ CRITICAL: ALWAYS Test Docker Builds Locally Before Deploying**

**Mandatory Docker Testing Workflow:**

```bash
# 1. Build Docker image locally
docker build -f services/api/Dockerfile.gcp -t ubik-api-test .

# 2. Verify files are in the image
docker run --rm ubik-api-test ls -la /app/
docker run --rm ubik-api-test ls -la /app/shared/openapi/

# 3. Test container locally
docker run --rm -p 8080:8080 \
  -e DATABASE_URL="postgres://ubik:ubik_dev_password@host.docker.internal:5432/ubik?sslmode=disable" \
  ubik-api-test

# 4. Verify endpoints work
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/docs/
```

**Why This Matters:**
- ❌ Local build ≠ Docker build
- ❌ Files in local filesystem may not be in Docker image
- ❌ Routes may behave differently in containers
- ✅ Testing Docker locally catches environment-specific bugs

**When Docker Testing is Required:**
- ✅ ANY change to Dockerfile
- ✅ ANY change to cloudbuild.yaml
- ✅ New API endpoints added
- ✅ New file resources needed (configs, specs, images)
- ✅ Environment variable changes

**See [docs/DOCKER_TESTING_CHECKLIST.md](./docs/DOCKER_TESTING_CHECKLIST.md) for complete Docker testing guide.**

---

### 8. Context Efficiency & Tool Selection

**⚠️ CRITICAL: Choose the Right Tool for the Task**

**Context Usage Awareness:**

Different tools have vastly different context costs:

```
Playwright browser snapshot:  ~12,000 tokens (Swagger UI page)
Screenshot:                   ~1,000 tokens
Curl API test:               ~100 tokens
```

**Golden Rule: Use the simplest tool that accomplishes the task**

**For API Testing:**
```bash
# ✅ GOOD - Direct and efficient
curl http://localhost:8080/api/v1/health

# ❌ BAD - Wastes ~50k tokens for same result
# playwright navigate → click endpoint → try it out → execute
```

**For UI Verification:**
```bash
# ✅ GOOD - Visual confirmation
browser_navigate + browser_take_screenshot  # ~1k tokens

# ❌ BAD - Full accessibility tree
browser_navigate + analyze full snapshot    # ~12k tokens
```

**When to Use Each Tool:**

**Playwright (browser automation):**
- ✅ Testing UI interactions (forms, buttons, navigation)
- ✅ Visual regression testing (screenshots)
- ✅ E2E workflows requiring browser state
- ❌ API endpoint testing (use curl)
- ❌ Browsing complex pages (use screenshots)

**Curl (HTTP requests):**
- ✅ API endpoint testing
- ✅ Health checks
- ✅ Quick response verification
- ✅ Testing authentication flows

**Read/Grep (file operations):**
- ✅ Searching code
- ✅ Verifying file contents
- ✅ Configuration inspection

**Best Practices:**
1. **Test APIs with curl, not Playwright**
2. **Use screenshots for visual verification**
3. **Minimize page snapshots** - only when you need to interact with specific elements
4. **Chain operations efficiently** - navigate → act → verify, don't browse
5. **Monitor token usage** - if a single operation uses >5k tokens, consider alternatives

**Example - Testing Swagger UI:**

```bash
# ❌ BAD - Uses ~48k tokens
playwright navigate to /api/docs
playwright click health endpoint      # 12k tokens
playwright click "Try it out"         # 12k tokens
playwright click "Execute"            # 12k tokens
verify response                       # 12k tokens

# ✅ GOOD - Uses ~2k tokens
playwright navigate to /api/docs      # 12k tokens (verify UI loads)
playwright take_screenshot            # 1k tokens
curl http://localhost:8080/api/v1/health  # 100 tokens (test endpoint)
```

**Token Budget Awareness:**
- 200k token context limit
- Large Playwright snapshots can consume 5-10% per interaction
- 4-5 page loads = 50k tokens = 25% of context
- Be mindful and efficient

---

# STATUS & ROADMAP

*Current progress and next steps*

---

## Current Status

**Last Updated:** 2025-11-05
**Version:** 0.2.0 🎉
**Status:** 🟢 **CLI Phase 4 Complete - Ready for v0.2.0 Release**
**Git Tag:** `v0.1.0` (v0.2.0 tag pending)

### 🎉 Milestone v0.1.0 Released!

**39 API endpoints implemented** | **144+ tests passing** | **73-88% coverage**

**See [docs/MILESTONE_v0.1.md](./docs/MILESTONE_v0.1.md) for complete release notes.**

### Key Achievements

**Phase 1 - Foundation ✅**
- Complete database schema (20 tables + 3 views)
- Code generation pipeline
- OpenAPI spec for all endpoints
- Local development environment

**Phase 2 - API ✅ (v0.1.0)**
- Complete authentication system (JWT + sessions)
- Employee, Organization, Team, Role management
- Agent catalog and configurations
- 144+ tests passing, 73-88% coverage

**Phase 3 - CLI ✅ (v0.2.0)**
- Authentication (`ubik login`, `ubik logout`)
- Config sync (`ubik sync`)
- Docker integration (container management)
- Interactive mode (`ubik` command)
- Agent management (`ubik agents`)
- **79 tests passing, 100% pass rate**

**See [docs/CLI_PHASE4_COMPLETE.md](./docs/CLI_PHASE4_COMPLETE.md) for CLI v0.2.0 details.**

---

## Roadmap

### Completed ✅
- **Phase 1:** Database schema, code generation, documentation
- **Phase 2:** Authentication, Employee/Org/Team/Role APIs (v0.1.0)
- **Phase 3:** CLI client (v0.2.0)

### In Progress 🎯
- **v0.3.0:** Web UI Foundation
  - Next.js 14 + shadcn/ui
  - Authentication & session management
  - Agent catalog page
  - Configuration management UI

### Planned
- **v0.4.0:** Analytics & Approvals
- **v0.5.0:** System Prompts & MCP Management
- **v1.0.0:** Production Release

**See [docs/WORKFLOWS.md](./docs/WORKFLOWS.md) for milestone planning.**

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

**For implementation details, see linked documents above.**

**Key Links:**
- 🚀 [Get Started](./docs/QUICKSTART.md)
- 🧪 [Testing Guide](./docs/TESTING.md)
- 🔧 [Development Guide](./docs/DEVELOPMENT.md)
- 📊 [Database ERD](./docs/ERD.md)
- 📖 [Quick Reference](./docs/QUICK_REFERENCE.md)
- 🔍 [Debugging Guide](./docs/DEBUGGING.md)
