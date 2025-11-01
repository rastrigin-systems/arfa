# Development Setup Snapshot

**Date:** 2025-11-01
**Version:** Pre-Coordinator (v0.3.0-dev)
**Backup Location:** `~/claude-code-dev-setup-backup-20251101-195047.tar.gz`

## Purpose

This document captures the **current development configuration** used to build Ubik Enterprise before implementing major orchestration changes. This setup has proven highly effective and may inform future Ubik features.

---

## Current Architecture

### Claude Code Configuration

**Location:** `~/.claude/`

**Agents Defined:**
1. **go-backend-developer** (`~/.claude/agents/go-backend-developer.md`)
   - Implements Go backend features (API, CLI, DB)
   - Uses Git worktrees for parallel development
   - Follows strict TDD workflow
   - Waits for CI checks before marking PR ready
   - Updates GitHub Project status automatically

2. **frontend-developer** (`~/.claude/agents/frontend-developer.md`)
   - Implements Next.js UI components and pages
   - Follows TDD with Vitest + Playwright
   - Ensures WCAG AA accessibility
   - Coordinates with backend agent for API needs
   - Uses Git worktrees for parallel development

3. **product-strategist** (`~/.claude/agents/product-strategist.md`)
   - Prioritizes features by business value
   - Uses GitHub Projects as source of truth
   - Creates and updates issues with strategic context
   - Stores decisions in Qdrant MCP
   - Recommends next highest-value tasks

4. **tech-lead** (`~/.claude/agents/tech-lead.md`)
   - Makes architectural decisions
   - Coordinates between specialized agents
   - Breaks down epics into tasks
   - Reviews technical changes
   - Maintains system architecture integrity

5. **pr-reviewer** (`~/.claude/agents/pr-reviewer.md`)
   - Reviews pull requests
   - Resolves merge conflicts
   - Waits for CI/CD checks
   - Merges approved PRs
   - Cleans up branches and worktrees
   - Updates GitHub Project status to "Done"

### Development Workflow

**Multi-Terminal Setup:**
- 2-3 Claude Code CLI instances open simultaneously
- Each instance can run different agents
- Manual coordination via GitHub Projects board
- Manual agent invocation (user triggers each agent)

**Parallel Development via Git Worktrees:**
```bash
# Agent 1 works in:
/Users/sergeirastrigin/Projects/ubik-enterprise (main branch)

# Agent 2 works in:
/Users/sergeirastrigin/Projects/ubik-issue-123 (issue-123 branch)

# Agent 3 works in:
/Users/sergeirastrigin/Projects/ubik-issue-234 (issue-234 branch)
```

**GitHub Projects Integration:**
- Issues labeled with: `status/ready`, `status/in-progress`, `status/waiting-for-review`, `status/done`
- Agents update issue status at each phase
- Project board reflects real-time progress
- Script: `./scripts/update-project-status.sh` for status updates

**CI/CD Integration:**
- Agents wait for GitHub Actions checks before marking PR ready
- Command: `gh pr checks $PR_NUM --watch --interval 10`
- Auto-update to "In Review" only when CI passes
- Failed checks → agent investigates and fixes

**Agent Coordination (Manual):**
1. User reads GitHub Projects board
2. User decides which task to work on
3. User invokes appropriate agent (e.g., "Implement issue #123")
4. Agent creates branch, worktree, implements feature
5. Agent creates PR, waits for CI, updates status
6. User invokes pr-reviewer agent to merge
7. Repeat

---

## Key Patterns & Practices

### 1. Strict TDD Workflow
- Write failing tests FIRST (mandatory)
- Implement minimal code to pass tests
- Refactor with tests passing
- Target: 85% code coverage

### 2. Git Worktree Usage
- One worktree per issue
- Format: `../ubik-issue-<NUM>`
- Branch format: `issue-<NUM>-<description>`
- Enables true parallel development

### 3. GitHub as Source of Truth
- All tasks tracked in GitHub Issues
- Project board shows status
- Issue comments for agent communication
- Labels for priority, area, status

### 4. Qdrant MCP for Knowledge
- Store architectural decisions
- Store business strategy insights
- Store past solutions and patterns
- Query before making decisions

### 5. Agent Specialization
- Each agent has clear domain
- Agents consult others (manual user coordination)
- No agent conflicts (manual coordination prevents)

---

## What This Setup Does Well

✅ **Parallel Development:** Multiple features simultaneously without conflicts
✅ **Quality Enforcement:** TDD + CI checks mandatory
✅ **Visibility:** GitHub Projects shows real-time progress
✅ **Automation:** Agents handle full dev workflow (branch → PR → CI → review)
✅ **Documentation:** Agents update docs alongside code
✅ **Knowledge Retention:** Qdrant MCP stores institutional knowledge
✅ **Clean Process:** Standard PR workflow enforced by agents

---

## What This Setup Lacks (Opportunities)

❌ **Autonomous Coordination:** Requires manual agent triggering
❌ **Agent Communication:** Agents don't coordinate directly
❌ **Conflict Detection:** No automatic detection of overlapping work
❌ **Health Monitoring:** No tracking if agent gets stuck
❌ **Auto-Assignment:** User decides which agent works on which task
❌ **Milestone Planning:** Manual milestone tracking
❌ **Load Balancing:** No agent capacity management
❌ **Recovery Mechanism:** No automatic retry on agent failure

---

## Comparison: Dev Setup vs. Ubik Product

### Dev Setup (Layer 3) - What We Have
- 5 specialized agents with workflows
- Git worktree orchestration
- GitHub Projects integration
- Automated PR lifecycle
- CI/CD awareness
- Qdrant knowledge base
- Multi-agent parallelism

### Ubik Product (Layer 2) - What Users Get
- Agent config sync (`ubik sync`)
- Docker container management
- Interactive agent sessions
- Basic agent management commands

**Gap:** Dev setup is 10x more sophisticated than product!

---

## Strategic Decision Point

### Option A: Keep Separate
- **Dev setup** = Internal tool for building Ubik
- **Ubik product** = Simple config sync for end users
- **Rationale:** Enterprises may not need full orchestration

### Option B: Merge into Product (Ubik Pro)
- Extract dev setup into **Ubik Advanced Features**
- Offer as premium tier for enterprises
- **Features:**
  - Multi-agent orchestration
  - GitHub Projects integration
  - Automated workflows
  - Team collaboration
- **Rationale:** Dev setup IS what enterprises want!

### Option C: Hybrid Approach
- **Ubik Basic:** Config sync + Docker (current)
- **Ubik Workflows:** Agent orchestration (dev setup features)
- **Ubik Enterprise:** + Multi-tenant + Team management
- **Rationale:** Tiered product for different customer segments

---

## Recommended Next Steps

1. **Preserve Current Setup** ✅ (this document + backup)
2. **Experiment with Coordination** (add coordinator agent)
3. **Validate with Beta Users** (test dev setup with team)
4. **Decide Product Direction** (Option A/B/C above)
5. **Roadmap Advanced Features** (if merging into product)

---

## Backup & Restore

### Backup Created
```bash
~/claude-code-dev-setup-backup-20251101-195047.tar.gz
```

### Restore Command
```bash
cd ~/.claude
tar -xzf ~/claude-code-dev-setup-backup-20251101-195047.tar.gz
```

### Files Backed Up
- `~/.claude/agents/go-backend-developer.md`
- `~/.claude/agents/frontend-developer.md`
- `~/.claude/agents/product-strategist.md`
- `~/.claude/agents/tech-lead.md`
- `~/.claude/agents/pr-reviewer.md`
- `~/.claude/CLAUDE.md` (global instructions)

---

## Agent Configuration Locations

### Current Development Setup
- **Location:** `~/.claude/agents/`
- **Purpose:** Build Ubik platform
- **Scope:** Personal development environment
- **Version:** Custom, evolving

### Ubik Product Configuration (Future)
- **Location:** `~/.ubik/configs/` (synced from platform)
- **Purpose:** End-user agent configurations
- **Scope:** Organization-managed
- **Version:** Controlled by Ubik platform

**Note:** These are intentionally separate! Dev setup is meta-tooling.

---

## Lessons Learned

1. **Dog-fooding reveals product gaps:** Building Ubik with Claude Code revealed what advanced users need
2. **Meta-tooling often surpasses product:** Development environments naturally evolve faster
3. **Reference implementations are valuable:** This setup can guide Ubik's roadmap
4. **Separation is important:** Don't conflate dev tooling with product features
5. **Documentation prevents loss:** Capturing this setup preserves institutional knowledge

---

## References

- [CLAUDE.md](../CLAUDE.md) - Complete Ubik system documentation
- [IMPLEMENTATION_ROADMAP.md](../IMPLEMENTATION_ROADMAP.md) - Ubik feature roadmap
- [MARKETING.md](../MARKETING.md) - Ubik product strategy
- [docs/CLI_CLIENT.md](./CLI_CLIENT.md) - Ubik CLI architecture

---

**Preserved:** 2025-11-01
**Status:** Active Development Setup
**Decision Pending:** Whether to integrate into Ubik product
