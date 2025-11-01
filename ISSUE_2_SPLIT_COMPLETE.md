# Issue #2 Split Complete ✅

**Date:** 2025-11-01
**Parent Issue:** [#2 - Web UI Dashboard](https://github.com/sergei-rastrigin/ubik-enterprise/issues/2)
**Reason:** XL size (80 hours) → Split into 5 manageable tasks (54 hours)

---

## 📋 Created Sub-Issues

### 1. [#12 - Web UI Foundation & Authentication](https://github.com/sergei-rastrigin/ubik-enterprise/issues/12)
**Size:** M (12 hours) | **Priority:** P0 🔴 | **Start First** ⭐

**Scope:**
- Next.js 14 App Router setup
- TypeScript + Tailwind CSS + shadcn/ui
- API client generation (openapi-typescript)
- Login page with JWT authentication
- Protected route middleware
- Logout functionality

**Dependencies:** None (can start immediately)

**Deliverables:**
- `services/web/app/(auth)/login/page.tsx`
- `services/web/app/(dashboard)/layout.tsx`
- `services/web/lib/api-client.ts`
- `services/web/middleware.ts`

---

### 2. [#13 - Agent Catalog Page](https://github.com/sergei-rastrigin/ubik-enterprise/issues/13)
**Size:** S (6 hours) | **Priority:** P0 🔴

**Scope:**
- Agent catalog grid/cards view
- Display agent name, provider, description, logo
- "Enable for Org" / "Configure" buttons
- Loading states and error handling

**Dependencies:** #12 (needs Web UI foundation)

**API:** `GET /agents`

**Deliverables:**
- `app/(dashboard)/agents/page.tsx`
- `components/agents/AgentCatalog.tsx`
- `components/agents/AgentCard.tsx`

---

### 3. [#14 - Organization Agent Configuration](https://github.com/sergei-rastrigin/ubik-enterprise/issues/14)
**Size:** L (16 hours) | **Priority:** P0 🔴 | **Core Feature** 🔥

**Scope:**
- List org agent configurations
- Enable agent for org (create config)
- Config editor modal (JSON or form)
- Update agent configuration
- Delete/disable agent configuration
- Validation (required fields, invalid JSON)

**Dependencies:** #13 (needs catalog)

**APIs:**
- `GET /organizations/current/agent-configs`
- `POST /organizations/current/agent-configs`
- `PATCH /organizations/current/agent-configs/{id}`
- `DELETE /organizations/current/agent-configs/{id}`

**Deliverables:**
- `app/(dashboard)/settings/agents/page.tsx`
- `components/agents/OrgAgentConfigs.tsx`
- `components/agents/ConfigEditorModal.tsx`
- `components/agents/MonacoEditor.tsx` (if using Monaco)

---

### 4. [#15 - Team Assignment UI](https://github.com/sergei-rastrigin/ubik-enterprise/issues/15)
**Size:** M (8 hours) | **Priority:** P1 🟠

**Scope:**
- Team selection/assignment interface
- Assign agent config to teams
- View assigned teams per agent
- Remove team assignments

**Dependencies:** #14 (needs org configs)

**APIs:**
- `GET /teams`
- `POST /teams/{team_id}/agent-configs`
- `DELETE /teams/{team_id}/agent-configs/{id}`

**Deliverables:**
- `components/agents/TeamAssignment.tsx`
- `components/agents/TeamSelector.tsx`

---

### 5. [#16 - E2E Testing & Polish](https://github.com/sergei-rastrigin/ubik-enterprise/issues/16)
**Size:** M (12 hours) | **Priority:** P1 🟠

**Scope:**
- E2E test: Login → Catalog → Configure → Assign
- Loading states for all async operations
- Error handling for all API calls
- Responsive design verification (mobile/tablet/desktop)
- Accessibility audit (WCAG AA)
- Performance optimization (Lighthouse >85)

**Dependencies:** #12, #13, #14, #15 (all features complete)

**Deliverables:**
- `tests/e2e/agent-dashboard.spec.ts`
- `components/ErrorBoundary.tsx`
- `components/ui/loading.tsx`

---

## 📊 Effort Breakdown

| Issue | Title | Size | Hours | Priority | Status |
|-------|-------|------|-------|----------|--------|
| #12 | Web UI Foundation & Authentication | M | 12 | P0 🔴 | To Do |
| #13 | Agent Catalog Page | S | 6 | P0 🔴 | To Do |
| #14 | Organization Agent Configuration | L | 16 | P0 🔴 | To Do |
| #15 | Team Assignment UI | M | 8 | P1 🟠 | To Do |
| #16 | E2E Testing & Polish | M | 12 | P1 🟠 | To Do |
| **TOTAL** | | | **54** | | |

**Original Estimate (Issue #2):** 80 hours (XL)
**New Total Estimate:** 54 hours (5 smaller tasks)
**Difference:** -26 hours (more accurate breakdown)

---

## 🎯 Implementation Sequence

```
┌─────────────────────────────────────────────┐
│ Week 1: Foundation + Core Features         │
├─────────────────────────────────────────────┤
│ Day 1-2 → #12 Web UI Foundation (12h)      │
│ Day 3   → #13 Agent Catalog (6h)           │
│ Day 4-5 → #14 Org Config (16h) 🔥          │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│ Week 2: Team Features + Polish             │
├─────────────────────────────────────────────┤
│ Day 1   → #15 Team Assignment (8h)         │
│ Day 2-3 → #16 E2E Testing & Polish (12h)   │
└─────────────────────────────────────────────┘
```

**Critical Path:** #12 → #13 → #14 → #15 → #16 (sequential)

---

## ✅ Benefits of Splitting

### Before Split:
- ❌ Single XL issue (80 hours) - intimidating
- ❌ Hard to track progress (all-or-nothing)
- ❌ Difficult to parallelize work
- ❌ Unclear what to start with
- ❌ Risk of scope creep

### After Split:
- ✅ 5 focused issues (S, M, M, M, L) - manageable
- ✅ Clear progress tracking (5 checkpoints)
- ✅ Can parallelize #15 and #16 if needed
- ✅ Clear starting point (#12)
- ✅ Well-defined scope per issue
- ✅ More accurate time estimates (54h vs 80h)

---

## 🔗 Links

**Parent Issue:** https://github.com/sergei-rastrigin/ubik-enterprise/issues/2

**Sub-Issues:**
- #12: https://github.com/sergei-rastrigin/ubik-enterprise/issues/12
- #13: https://github.com/sergei-rastrigin/ubik-enterprise/issues/13
- #14: https://github.com/sergei-rastrigin/ubik-enterprise/issues/14
- #15: https://github.com/sergei-rastrigin/ubik-enterprise/issues/15
- #16: https://github.com/sergei-rastrigin/ubik-enterprise/issues/16

**Project Board:** https://github.com/users/sergei-rastrigin/projects/3

**Milestone:** v0.3.0 - Web UI MVP (Due: 2025-11-15)

---

## 📈 Progress Tracking

**Definition of Done (Epic #2):**
- [ ] Sub-issue #12 completed and closed
- [ ] Sub-issue #13 completed and closed
- [ ] Sub-issue #14 completed and closed
- [ ] Sub-issue #15 completed and closed
- [ ] Sub-issue #16 completed and closed
- [ ] All acceptance criteria from original #2 met
- [ ] Web UI dashboard functional and tested
- [ ] Close parent issue #2

**Progress:** 0/5 sub-issues complete (0%)

---

## 🚀 Getting Started

**Next Action:** Start with [Issue #12 - Web UI Foundation & Authentication](https://github.com/sergei-rastrigin/ubik-enterprise/issues/12)

**Command:**
```bash
# Check out issue #12
gh issue view 12 --repo sergei-rastrigin/ubik-enterprise

# Assign to yourself (if not already)
gh issue edit 12 --assignee @me

# Move to "In Progress" on project board
# (do this in GitHub UI or via gh project CLI)
```

---

## 📝 Notes

**Why 5 sub-issues?**
- Logical separation of concerns
- Each can be completed in 1-3 days
- Clear dependencies and sequence
- Easier to review and test incrementally

**Can tasks run in parallel?**
- No: #12 → #13 → #14 → #15 must be sequential (dependencies)
- Maybe: #15 and #16 could overlap if #14 is complete
- Recommendation: Follow sequence for simplicity

**What if scope changes?**
- Adjust individual sub-issues (easier than XL issue)
- Can add new sub-issues if needed
- Parent issue #2 tracks overall progress

---

## ✅ Verification

All sub-issues:
- ✅ Created in GitHub Issues
- ✅ Added to v0.3.0 milestone
- ✅ Added to Engineering project board (#3)
- ✅ Labeled with priority/area/size
- ✅ Assigned to @me
- ✅ Linked to parent issue #2
- ✅ Parent issue #2 updated with sub-issue links

---

**Split Complete:** 2025-11-01
**Ready to Start:** ✅ YES (begin with #12)

---

🚀 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
