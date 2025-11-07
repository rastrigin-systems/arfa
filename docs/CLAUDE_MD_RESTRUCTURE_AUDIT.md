# CLAUDE.md Restructure Audit

**Date:** 2025-11-05
**Issue:** #118
**Goal:** Reduce CLAUDE.md from ~14K tokens to ~8-10K tokens (~40% reduction)

---

## Current State Analysis

### File Size Metrics
- **Current Size:** 42KB (5,554 words, 1,340 lines)
- **Estimated Tokens:** ~14,000 tokens (based on 1 token ≈ 0.75 words with markdown)
- **Target Size:** ~8-10K tokens (~25-30KB)
- **Reduction Needed:** ~30-40%

### Content Breakdown by Section

| Section | Lines | Estimated Tokens | Status | Action |
|---------|-------|------------------|--------|--------|
| **Foundation** | ~250 | ~2,000 | Essential | Keep with minor edits |
| System Overview | 40 | 300 | Essential | Keep |
| Architecture | 45 | 400 | Essential | Keep |
| Database Schema | 40 | 350 | Summary only | Move details to DATABASE.md |
| Technology Stack | 10 | 80 | Essential | Keep |
| Project Structure | 70 | 600 | Essential | Keep (critical reference) |
| **Documentation Map** | ~130 | ~1,000 | Essential | Keep but streamline |
| **Quick Reference** | ~130 | ~1,200 | Redundant | Move to QUICK_REFERENCE.md |
| Common Commands | 30 | 250 | Move | → QUICK_REFERENCE.md |
| Database Access | 10 | 80 | Move | → DATABASE.md |
| MCP Servers | 60 | 600 | Move | → MCP_SERVERS.md |
| **Development** | ~470 | ~5,500 | **Excessive** | **Major reduction needed** |
| Standard PR Workflow | 90 | 800 | Move | → DEV_WORKFLOW.md (exists!) |
| First-Time Setup | 10 | 80 | Keep | Essential |
| Making Changes | 50 | 500 | Move | → DEVELOPMENT.md (exists!) |
| CI-Aware Workflow | 100 | 1,200 | Move | → DEV_WORKFLOW.md (exists!) |
| Milestone Planning | 130 | 1,500 | Move | → WORKFLOWS.md (new) |
| Code Generation | 15 | 150 | Keep | Essential |
| **Critical Notes** | ~220 | ~2,500 | **Excessive** | **Reduce to essentials** |
| Code Generation | 30 | 300 | Keep | Essential |
| Multi-Tenancy | 15 | 120 | Keep | Essential |
| Testing Strategy | 40 | 400 | Link only | → TESTING.md (exists!) |
| UI Workflow | 30 | 300 | Link only | → DEVELOPMENT.md |
| Release Management | 50 | 500 | Link only | → Release Manager Skill |
| Debugging | 90 | 1,000 | Move | → DEBUGGING.md (new) |
| **Status & Roadmap** | ~240 | ~2,800 | **Excessive** | **Major reduction** |
| Current Status | 110 | 1,200 | Streamline | Keep summary only |
| Success Metrics | 30 | 300 | Move | → MILESTONES_ARCHIVE.md |
| Roadmap | 30 | 200 | Keep | Essential |
| **Critical Usage Notes** | ~180 | ~2,000 | **Redundant** | **Move entirely** |
| Qdrant MCP | 35 | 400 | Move | → MCP_SERVERS.md |
| PostgreSQL MCP | 80 | 900 | Move | → MCP_SERVERS.md |

---

## Content to Move to Dedicated Files

### 1. MCP_SERVERS.md (NEW)
**Size:** ~2,000 tokens → Save ~2,000 tokens in CLAUDE.md

**Content to move:**
- ✅ Qdrant MCP setup and usage
- ✅ PostgreSQL MCP setup and usage
- ✅ GitHub MCP server details (from Quick Reference)
- ✅ Playwright MCP details

**Keep in CLAUDE.md:**
- List of configured MCP servers (1-2 lines)
- Link to MCP_SERVERS.md

---

### 2. QUICK_REFERENCE.md (NEW)
**Size:** ~1,200 tokens → Save ~1,000 tokens in CLAUDE.md

**Content to move:**
- ✅ Quick Start commands
- ✅ Common Commands (all make targets)
- ✅ Database access information
- ✅ Git commands cheat sheet

**Keep in CLAUDE.md:**
- Most essential 3-5 commands only
- Link to QUICK_REFERENCE.md

---

### 3. WORKFLOWS.md (NEW)
**Size:** ~1,500 tokens → Save ~1,500 tokens in CLAUDE.md

**Content to move:**
- ✅ Milestone planning workflow
- ✅ Milestone transition workflow
- ✅ Task splitting guidelines
- ✅ Milestone best practices

**Keep in CLAUDE.md:**
- Link to Release Manager Skill
- Link to WORKFLOWS.md

---

### 4. DEBUGGING.md (NEW)
**Size:** ~1,000 tokens → Save ~900 tokens in CLAUDE.md

**Content to move:**
- ✅ Debugging best practices
- ✅ Common pitfalls
- ✅ Debugging workflow
- ✅ Real-world examples

**Keep in CLAUDE.md:**
- "Check the data, not just the code" principle (1-2 lines)
- Link to DEBUGGING.md

---

### 5. DATABASE.md (NEW)
**Size:** ~500 tokens → Save ~300 tokens in CLAUDE.md

**Content to move:**
- ✅ Detailed table descriptions
- ✅ Database access methods
- ✅ Schema management
- ✅ Migration procedures

**Keep in CLAUDE.md:**
- High-level schema overview (5 table groups)
- Link to ERD.md
- Link to DATABASE.md

---

### 6. Update Existing Files

#### DEV_WORKFLOW.md (EXISTS)
**Add content from CLAUDE.md:**
- ✅ CI-Aware Development Workflow section
- ✅ Helper script examples
- ✅ Agent configuration references

#### DEVELOPMENT.md (EXISTS)
**Add content from CLAUDE.md:**
- ✅ Making Changes example
- ✅ UI Development Workflow
- ✅ Code generation pipeline details

#### TESTING.md (EXISTS)
**Verify contains:**
- ✅ TDD workflow (already there)
- ✅ Testing strategy (already there)

---

## New CLAUDE.md Structure

### Proposed Outline (~8-10K tokens)

```markdown
# Ubik Enterprise — AI Agent Management Platform

## 📑 Table of Contents
[Streamlined to 3 main sections]

---

# 1. FOUNDATION (~2,000 tokens)
- System Overview (brief)
- Architecture (high-level)
- Database Schema (summary only)
- Technology Stack
- Project Structure

---

# 2. DOCUMENTATION MAP (~1,500 tokens)
**Start Here:**
- Quick links to QUICKSTART, ERD, ROADMAP

**Core Documentation:**
- Development: DEVELOPMENT.md, DEV_WORKFLOW.md, TESTING.md
- Database: DATABASE.md, ERD.md
- Operations: MCP_SERVERS.md, QUICK_REFERENCE.md, DEBUGGING.md
- Workflows: WORKFLOWS.md, Release Manager Skill
- CLI: CLI_CLIENT.md, CLI_PHASE*.md

**Configuration:**
- Links only

**Archived:**
- Links only

---

# 3. QUICK REFERENCE (~800 tokens)
**Essential Commands Only:**
```bash
# Database
make db-up / db-down / db-reset

# Development
make generate
make test

# See QUICK_REFERENCE.md for complete list
```

**MCP Servers:**
- List of configured servers (2-3 lines)
- See MCP_SERVERS.md for setup

---

# 4. DEVELOPMENT (~2,000 tokens)
**Standard Workflow:**
- Link to DEV_WORKFLOW.md
- First-time setup commands

**Critical Rules:**
- Never edit generated/
- All queries must be org-scoped
- Always follow TDD
- See TESTING.md, DEVELOPMENT.md for details

**Code Generation:**
- High-level pipeline
- When to regenerate

---

# 5. CRITICAL NOTES (~1,000 tokens)
**Code Generation:** Essential warnings
**Multi-Tenancy:** Org-scoping requirement
**Testing:** Link to TESTING.md
**Release Management:** Link to Release Manager Skill
**Debugging:** "Check data, not code" + link to DEBUGGING.md

---

# 6. STATUS & ROADMAP (~1,500 tokens)
**Current Status:**
- Version, branch, status (brief)
- Latest milestone (link to MILESTONE_v0.X.md)

**Achievements:**
- Phase 1-4 summaries (very brief)

**Roadmap:**
- Link to IMPLEMENTATION_ROADMAP.md

---

# 7. KEY LINKS (~200 tokens)
- Quick links to essential docs
```

---

## Token Savings Breakdown

| Action | Tokens Saved |
|--------|-------------|
| Move MCP server details → MCP_SERVERS.md | ~2,000 |
| Move commands → QUICK_REFERENCE.md | ~1,000 |
| Move workflows → WORKFLOWS.md | ~1,500 |
| Move debugging → DEBUGGING.md | ~900 |
| Move DB details → DATABASE.md | ~300 |
| Streamline Status section | ~1,000 |
| Remove redundant content | ~300 |
| **Total Reduction** | **~7,000 tokens** |

**New Size:** ~14,000 - 7,000 = **~7,000 tokens** (better than target!)

---

## Existing Documentation Files

### Already Exist (Use These)
- ✅ QUICKSTART.md (7.3K)
- ✅ TESTING.md (11K)
- ✅ DEVELOPMENT.md (9.4K)
- ✅ DEV_WORKFLOW.md (11K)
- ✅ CLI_CLIENT.md (19K)
- ✅ ERD.md (9.3K)
- ✅ RELEASES.md (5.7K)
- ✅ MILESTONE_v0.1.md (9.5K)

### Need to Create
- ❌ MCP_SERVERS.md
- ❌ QUICK_REFERENCE.md
- ❌ WORKFLOWS.md
- ❌ DEBUGGING.md
- ❌ DATABASE.md

---

## Implementation Order

1. ✅ **Create MCP_SERVERS.md** - Largest savings (~2K tokens)
2. ✅ **Create WORKFLOWS.md** - Major workflow documentation (~1.5K tokens)
3. ✅ **Create DEBUGGING.md** - Best practices (~900 tokens)
4. ✅ **Create QUICK_REFERENCE.md** - Commands reference (~1K tokens)
5. ✅ **Create DATABASE.md** - Database details (~300 tokens)
6. ✅ **Update DEV_WORKFLOW.md** - Add CI workflow content
7. ✅ **Update DEVELOPMENT.md** - Add Making Changes example
8. ✅ **Restructure CLAUDE.md** - Final reorganization
9. ✅ **Verify all links** - Ensure discoverability
10. ✅ **Validate token count** - Confirm <10K tokens

---

## Success Criteria

- [x] CLAUDE.md reduced to ~7-10K tokens (~40-50% reduction)
- [ ] All critical information preserved
- [ ] Clear navigation between docs
- [ ] No broken links
- [ ] Documentation Map updated
- [ ] All new files created
- [ ] Cross-references added

---

## Notes

- Keep CLAUDE.md as the **high-level index** and **critical reference**
- Detailed procedures go in dedicated docs
- Every section should have a "See [link] for details" reference
- Maintain discoverability - docs should be easy to find
