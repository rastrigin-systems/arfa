# Ubik Enterprise - Release History

Complete history of all releases with links to GitHub Releases and detailed changelogs.

---

## Current Release

**v0.4.0** - Beta Ready (2025-11-03)

---

## All Releases

### v0.4.0 - Beta Ready (2025-11-03)

**Milestone:** Docker Development + Skills/MCP Management

**Major Features:**
- Complete Docker Compose stack (API, Web, Database, Adminer)
- Skills and MCP server management (API + CLI)
- Enhanced Claude Code sync workflow
- GitHub Project Manager agent
- Streamlined developer experience

**Components:**
- Docker containerized API server with health checks
- Docker containerized Next.js web UI with hot reload
- Skills API endpoints (CRUD operations)
- MCP server catalog and configuration
- Enhanced `ubik sync` with Claude Code support
- `ubik skills` command group (list/show/my)

**Statistics:**
- 11 issues closed in milestone
- 17 commits since v0.3.0
- Complete Docker development stack
- Skills + MCP management fully implemented

**Links:**
- [Git Tag](https://github.com/rastrigin-org/ubik-enterprise/releases/tag/v0.4.0)
- [Compare with v0.3.0](https://github.com/rastrigin-org/ubik-enterprise/compare/v0.3.0...v0.4.0)

---

### v0.3.0 - Complete Web UI (2025-11-02)

**Milestone:** Web Interface

**Major Features:**
- Complete Next.js 14 web UI with 11 pages
- Dark/Light mode theming
- Comprehensive E2E test suite with Playwright
- MSW API mocking for reliable tests
- GitHub workflow automation skills

**Statistics:**
- 69 commits since v0.2.0
- 11 web UI pages
- 24+ E2E tests
- 3 new GitHub workflow skills

**Links:**
- [Git Tag](https://github.com/rastrigin-org/ubik-enterprise/releases/tag/v0.3.0)
- [Compare with v0.2.0](https://github.com/rastrigin-org/ubik-enterprise/compare/v0.2.0...v0.3.0)

**Note:** v0.3.0 tag originally pointed to monorepo migration, but was intended for Web UI milestone completion.

---

### v0.2.0 - CLI Client (2025-10-29)

**Milestone:** Employee CLI

**Major Features:**
- CLI client with Docker integration
- Agent container orchestration
- Interactive agent sessions
- Agent management commands
- Session tracking and cleanup

**Components:**
- `ubik login` / `ubik logout` - Authentication
- `ubik sync` - Configuration synchronization
- `ubik agents` - Agent management
- `ubik start` / `ubik stop` - Container control
- Interactive `ubik` mode - Agent sessions

**Statistics:**
- 79 tests passing (unit + integration)
- 100% pass rate maintained
- Docker SDK integration complete

**Links:**
- [Git Tag](https://github.com/rastrigin-org/ubik-enterprise/releases/tag/v0.2.0)
- [Compare with v0.1.0](https://github.com/rastrigin-org/ubik-enterprise/compare/v0.1.0...v0.2.0)

---

### v0.1.0 - API Foundation (2025-10-29)

**Milestone:** Backend API + Authentication

**Major Features:**
- Complete authentication system with JWT + sessions
- Employee CRUD endpoints (5 endpoints)
- Organization management (2 endpoints)
- Team management (5 endpoints)
- Role management (5 endpoints)
- Agent catalog (2 endpoints)
- Agent configurations (16 endpoints)

**Statistics:**
- 39 API endpoints implemented
- 144+ tests passing (119 unit + 25 integration)
- 73-88% code coverage across all modules
- Multi-tenancy verified

**Architecture:**
- PostgreSQL 15 database (20 tables + 3 views)
- Go Chi HTTP router
- OpenAPI 3.0.3 specification
- sqlc for type-safe queries
- JWT authentication with sessions

**Links:**
- [Git Tag](https://github.com/rastrigin-org/ubik-enterprise/releases/tag/v0.1.0)
- [Release Notes](https://github.com/rastrigin-org/ubik-enterprise/blob/main/docs/MILESTONE_v0.1.md)

---

## Pre-Releases

### v0.0.1 - Initial Commit (2025-10-28)

Project initialization with:
- Database schema design
- Code generation pipeline
- Documentation structure
- Development environment

---

## Versioning Strategy

Ubik Enterprise follows [Semantic Versioning 2.0.0](https://semver.org/):

- **v0.x.y** (Pre-1.0): Rapid development, breaking changes allowed
  - **0.x.0**: New milestone features
  - **0.x.y**: Bug fixes and polish

- **v1.0.0+** (Post-launch): Production releases
  - **Major (1.0.0 → 2.0.0)**: Breaking API changes
  - **Minor (1.0.0 → 1.1.0)**: New features (backward compatible)
  - **Patch (1.0.0 → 1.0.1)**: Bug fixes only

---

## Release Process

See [Release Manager Skill](../.claude/skills/release-manager/SKILL.md) for complete release workflow.

### Quick Release Checklist

1. ✅ All CI/CD checks passing
2. ✅ All tests passing (make test)
3. ✅ Milestone issues closed
4. ✅ On main branch, clean working tree
5. ✅ Documentation updated
6. ✅ Changelog generated
7. ✅ Git tag created and pushed
8. ✅ GitHub Release published
9. ✅ Announcement posted

---

## Upcoming Releases

### v0.5.0 - Public Launch (Planned)

**Target:** Q1 2026

**Planned Features:**
- Public launch campaign and beta onboarding
- Platform stabilization
- Performance optimizations
- Production deployment setup
- User documentation and tutorials

### v1.0.0 - Production Launch (Planned)

**Target:** Q2 2026

**Planned Features:**
- Complete feature set
- Production-grade security
- Comprehensive documentation
- Migration tools
- Performance optimizations

---

## GitHub Releases

All releases are also published as [GitHub Releases](https://github.com/rastrigin-org/ubik-enterprise/releases) with:
- Detailed changelogs
- Downloadable binaries (future)
- Release notes
- Migration guides

---

## Related Documentation

- [CLAUDE.md](../CLAUDE.md) - Project overview and current status
- [IMPLEMENTATION_ROADMAP.md](../IMPLEMENTATION_ROADMAP.md) - Feature roadmap
- [Release Manager Skill](../.claude/skills/release-manager/SKILL.md) - Release workflow
- [GitHub Milestones](https://github.com/rastrigin-org/ubik-enterprise/milestones) - Track progress

---

**Last Updated:** 2025-11-02
