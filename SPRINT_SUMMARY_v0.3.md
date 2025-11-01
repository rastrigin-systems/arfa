# Sprint v0.3.0 - Executive Summary

**Date:** 2025-11-01
**Sprint Duration:** 2 weeks (November 4-15, 2025)
**Primary Goal:** Enable beta launch with functional Web UI
**Version Target:** v0.3.0

---

## Quick Links

- **Full Sprint Plan:** [SPRINT_PLAN_v0.3.md](./SPRINT_PLAN_v0.3.md)
- **GitHub Tasks:** [GITHUB_TASKS_v0.3.md](./GITHUB_TASKS_v0.3.md)
- **Implementation Roadmap:** [IMPLEMENTATION_ROADMAP.md](./IMPLEMENTATION_ROADMAP.md)

---

## Strategic Context

### Business Priority
**Beta launch by February 1, 2025** with functional demo for:
- Landing page with signup form
- Product demo video (3-5 minutes)
- Beta customer outreach (target: 5-10 signups)

### Critical Path
```
v0.3.0 Web UI (Nov 4-15)
  → Demo Recording (Jan 15-20)
    → Landing Page Launch (Feb 1)
      → Beta Customer Outreach (Feb 1-15)
        → First Beta Customers (Feb 15)
```

---

## What We're Building

### Minimum Viable Web UI
1. **Authentication** - Login/logout with JWT
2. **Employee Management** - List, view, create, edit, deactivate
3. **Agent Configuration** - Org-level setup, employee overrides
4. **Testing** - E2E test suite with Playwright
5. **Polish** - Accessibility, loading states, error handling

### Out of Scope (Defer to v0.4.0)
- MCP Configuration UI
- System Prompts UI
- Approval Workflows UI
- Analytics Dashboard
- Usage Tracking

---

## Sprint Breakdown

### PHASE 1: Foundation (Day 1-2, 10 hours)
**Owner:** backend-api + frontend-web agents

1. **Fix Employee Integration Test** ⚠️ BLOCKER
   - **Why Critical:** Blocking all employee endpoint work
   - **Problem:** Response fields missing (Status, RoleId)
   - **Effort:** 3 SP (~4 hours)
   - **Success:** All 144+ tests passing

2. **Initialize Web UI Module**
   - **Deliverable:** Next.js 14 + TypeScript + Tailwind + ShadcnUI
   - **Key Feature:** API client auto-generated from OpenAPI spec
   - **Effort:** 5 SP (~6 hours)
   - **Success:** Login page working, dashboard protected

---

### PHASE 2: Employee Management (Day 3-5, 22 hours)
**Owner:** frontend-web agent

3. **Employee List Page**
   - **Features:** Table, filtering, pagination, search, actions
   - **API:** `GET /employees`, `GET /teams`
   - **Effort:** 8 SP (~8 hours)

4. **Employee Detail Page**
   - **Features:** Profile, agents, MCPs, activity log
   - **API:** `GET /employees/{id}`, `GET /employees/{id}/agent-configs/resolved`
   - **Effort:** 6 SP (~6 hours)

5. **Employee Form (Create/Edit)**
   - **Features:** Validation, team/role dropdowns, password strength
   - **API:** `POST /employees`, `PATCH /employees/{id}`
   - **Effort:** 8 SP (~8 hours)

---

### PHASE 3: Agent Configuration (Day 6-8, 20 hours)
**Owner:** frontend-web agent

6. **Org Agent Configs Page** ⭐ CORE VALUE PROP
   - **Features:** Catalog, config editor (Monaco), policy assignment
   - **API:** `GET /agents`, `POST /organizations/current/agent-configs`
   - **Effort:** 13 SP (~12 hours)
   - **⚠️ Complexity:** JSON editing, hierarchical config

7. **Employee Agent Overrides**
   - **Features:** Inherited agents, employee overrides, config merging
   - **API:** `GET /employees/{id}/agent-configs/resolved`
   - **Effort:** 8 SP (~8 hours)

---

### PHASE 4: Testing & Polish (Day 9-10, 16 hours)
**Owner:** frontend-web + backend-api agents

8. **E2E Test Suite (Playwright)**
   - **Coverage:** Auth, employees, agents, error handling
   - **Infrastructure:** Seed data, CI/CD integration
   - **Effort:** 8 SP (~10 hours)

9. **UI/UX Polish & Accessibility**
   - **Deliverables:** Design system, a11y audit, loading states, error boundaries
   - **Tools:** axe DevTools, Lighthouse
   - **Effort:** 5 SP (~6 hours)

---

## Resource Allocation

### Team Capacity (2 weeks)
- **Tech Lead (you):** 20 hours (coordination, reviews)
- **backend-api agent:** 40 hours (API fixes, support)
- **frontend-web agent:** 60 hours (Web UI development)
- **Total:** 120 hours

### Story Point Allocation (64 total)
| Phase | Tasks | Story Points | Hours |
|-------|-------|--------------|-------|
| Phase 1 | 2 | 8 | 10 |
| Phase 2 | 3 | 22 | 22 |
| Phase 3 | 2 | 21 | 20 |
| Phase 4 | 2 | 13 | 16 |
| **TOTAL** | **9** | **64** | **68** |

**Buffer:** 52 hours remaining (43% buffer for unknowns)

---

## Technical Decisions

### 1. Web UI Technology Stack
**Decision:** Next.js 14 App Router + TypeScript + Tailwind + ShadcnUI

**Rationale:**
- Next.js 14: Latest stable, server components for performance
- TypeScript: Type safety, matches Go backend philosophy
- Tailwind: Utility-first CSS, fast development
- ShadcnUI: Accessible components, customizable

**Alternatives Considered:**
- Remix (too early in maturity)
- Vite + React (no SSR out of the box)
- Vue/Svelte (team unfamiliar)

---

### 2. API Client Generation
**Decision:** `openapi-typescript` + `openapi-fetch`

**Rationale:**
- Auto-generates TypeScript types from OpenAPI spec
- Type-safe API calls with autocomplete
- Reduces type duplication (single source of truth)
- Easy to regenerate when API changes

**Implementation:**
```bash
npx openapi-typescript ../../shared/openapi/spec.yaml -o lib/api.ts
```

**Alternative:** Manual API client (too much duplication)

---

### 3. Authentication Strategy
**Decision:** Custom JWT in httpOnly cookie (no NextAuth.js)

**Rationale:**
- Simpler: Matches existing API design
- No extra dependencies
- Full control over session management
- NextAuth.js is overkill for our use case

**Implementation:**
```typescript
// lib/auth.ts
export async function login(email: string, password: string) {
  const { data } = await apiClient.POST('/auth/login', {
    body: { email, password }
  });
  // Store token in httpOnly cookie via middleware
  return data.token;
}
```

---

### 4. Config Editor
**Decision:** Monaco Editor (`@monaco-editor/react`)

**Rationale:**
- Industry standard (VSCode engine)
- JSON validation and autocomplete
- Syntax highlighting
- Familiar to developers

**Alternative:** Plain textarea (poor UX)

---

### 5. Testing Strategy
**Decision:** Playwright for E2E, React Testing Library for components

**Rationale:**
- Playwright: Fast, reliable, cross-browser
- RTL: Best practice for React testing
- Both integrate well with Next.js

**Coverage Targets:**
- E2E: >80% of critical user flows
- Component: >70% of UI components
- Integration: Existing 144+ API tests maintained

---

## Risk Assessment

### 🔴 HIGH RISK

**Risk:** Hierarchical config resolution UI too complex
**Impact:** Users confused, bugs in config merging
**Mitigation:**
- Start with simple UI (read-only inherited config + editable overrides)
- Show clear visual hierarchy (org → team → employee)
- Add tooltips explaining config inheritance
- Defer full JSON merge preview to v0.4.0 if too complex

**Contingency Plan:**
- If >80% of time spent on config UI, simplify to "override entire config" (no merging)
- Ship v0.3.1 with full merge preview later

---

### 🟡 MEDIUM RISK

**Risk:** E2E tests flaky or slow
**Impact:** CI/CD unreliable, slows development
**Mitigation:**
- Use testcontainers for isolated test DB
- Implement retry logic for flaky tests (Playwright built-in)
- Run tests in parallel (Playwright workers)
- Mock external APIs (if any)

**Contingency Plan:**
- Mark flaky tests as `.skip()` temporarily
- Ship v0.3.0 with manual testing, fix tests in v0.3.1

---

**Risk:** Integration test blocker delays sprint start
**Impact:** Web UI waits for API fix
**Mitigation:**
- Fix test on Day 1 (highest priority)
- Parallel work: UI foundation can start while fixing
- Use API mock server if needed (`openapi-mock`)

**Contingency Plan:**
- If test not fixed by end of Day 1, use mock API for UI development
- Fix test in parallel, integrate later

---

### 🟢 LOW RISK

**Risk:** Scope creep (too many features)
**Impact:** Sprint overcommitted, features incomplete
**Mitigation:**
- Strict MVP definition (no feature additions without removing others)
- Weekly sprint review (cut scope if behind)
- Defer MCP UI, system prompts, approvals to v0.4.0

**Enforcement:**
- Tech lead approval required for all scope changes
- "No new features" rule after Day 5

---

## Success Metrics

### Technical Metrics (Measured at Sprint End)
- [ ] **Test Coverage:** >80% overall (frontend + backend)
- [ ] **API Response Time:** <100ms (p95) - already achieved
- [ ] **Web UI Load Time:** <2s (p95)
- [ ] **E2E Tests Passing:** 100% (no flaky tests)
- [ ] **Accessibility Score:** >90 (Lighthouse)
- [ ] **Performance Score:** >85 (Lighthouse)
- [ ] **All 9 tasks complete:** P0 and P1 tasks done

### Business Metrics (Post-Sprint)
- [ ] **Internal Dogfooding:** 5+ team members using daily by Nov 20
- [ ] **Demo Video:** Can be recorded by Jan 15, 2025
- [ ] **Beta Signups:** Landing page ready to capture leads by Feb 1
- [ ] **Beta Launch Readiness:** Pass pre-launch checklist (security, performance, UX)

---

## Definition of Done

### For Each Task
- ✅ Code reviewed (self-review + tech lead approval)
- ✅ All tests passing (unit + integration + E2E)
- ✅ Test coverage >80% for new code
- ✅ No console errors or warnings
- ✅ Wireframe matches implementation (or wireframe updated)
- ✅ API types match OpenAPI spec
- ✅ Documentation updated (if architectural change)
- ✅ Accessibility checklist passed
- ✅ Responsive design verified (mobile, tablet, desktop)

### For Sprint (v0.3.0 Release)
- ✅ All P0 and P1 tasks complete (Tasks 1-7)
- ✅ E2E test suite passing (>80% coverage)
- ✅ No critical bugs (P0/P1)
- ✅ Performance: Page load <2s, API response <200ms
- ✅ Security: No exposed secrets, HTTPS enforced
- ✅ Demo video can be recorded (all features working)
- ✅ Beta landing page can link to working app
- ✅ Internal team can dogfood (use the platform daily)

---

## Next Steps (Immediate Actions)

### Day 1 (November 4, 2025)
**Owner:** backend-api agent
1. **Fix integration test** (Task 1) - Start immediately
   - Debug `CreateEmployee` handler response mapping
   - Verify all fields returned correctly
   - Run full test suite

**Owner:** frontend-web agent
2. **Initialize Web UI** (Task 2) - Can start in parallel
   - Set up Next.js 14 project
   - Install dependencies (TypeScript, Tailwind, ShadcnUI)
   - Generate API client from OpenAPI spec

### Day 2-3
**Owner:** frontend-web agent
3. **Build Employee List Page** (Task 3)
   - Create wireframe first
   - Implement table with filtering
   - Add E2E tests

### Day 4-5
**Owner:** frontend-web agent
4. **Build Employee Detail Page** (Task 4)
5. **Build Employee Form** (Task 5)

### Day 6-8
**Owner:** frontend-web agent
6. **Build Org Agent Configs** (Task 6) - Core feature
7. **Build Employee Overrides** (Task 7)

### Day 9-10
**Owner:** frontend-web + backend-api agents
8. **E2E Test Suite** (Task 8)
9. **UI/UX Polish** (Task 9)

---

## Post-Sprint Actions

### Retrospective (November 15, 2025)
**Agenda:**
1. What went well?
2. What could be improved?
3. Action items for v0.4.0 sprint

### Deployment Plan
**Target:** November 16-17, 2025
1. Merge `feature/web-ui` branch to `main`
2. Tag release `v0.3.0`
3. Deploy to staging environment
4. Internal testing (dogfooding)
5. Deploy to production (if stable)

### Demo Recording Plan
**Target:** January 15-20, 2025
1. Script demo video (3 scenes, 3-5 minutes)
2. Record with screen capture tool (Loom or ScreenFlow)
3. Add voiceover
4. Edit and upload to YouTube
5. Create 30s teaser for social media

---

## Questions & Clarifications

### Open Questions
1. **Authentication:** NextAuth.js or custom JWT?
   - **Decision:** Custom JWT (simpler, matches API)

2. **Config Editor:** Monaco or plain textarea?
   - **Decision:** Monaco (better UX, validation)

3. **Deployment:** Where to host Web UI?
   - **Pending:** Vercel (easy) or self-hosted (control)?

4. **Database:** Shared with API or separate?
   - **Decision:** Shared (same PostgreSQL, API handles all DB access)

### Decisions Needed From Product Team
- [ ] Exact demo video script (3 scenes)
- [ ] Landing page copy/messaging
- [ ] Pricing tier names and limits
- [ ] Beta customer target list (who to reach out to?)

---

## Appendix: File Structure

### New Files Created (Sprint v0.3.0)
```
services/web/                           # NEW MODULE
├── app/
│   ├── (auth)/login/page.tsx           # Task 2
│   ├── (dashboard)/
│   │   ├── employees/page.tsx          # Task 3
│   │   ├── employees/[id]/page.tsx     # Task 4
│   │   ├── employees/new/page.tsx      # Task 5
│   │   ├── employees/[id]/edit/page.tsx # Task 5
│   │   ├── employees/[id]/agents/page.tsx # Task 7
│   │   └── settings/agents/page.tsx    # Task 6
│   └── layout.tsx
├── components/
│   ├── employees/                      # Tasks 3-5
│   ├── agents/                         # Tasks 6-7
│   ├── ui/                             # ShadcnUI components
│   └── ErrorBoundary.tsx               # Task 9
├── lib/
│   ├── api-client.ts                   # Task 2 (generated)
│   ├── api.ts                          # Task 2 (types from OpenAPI)
│   ├── auth.ts                         # Task 2
│   └── utils.ts
├── tests/
│   ├── e2e/                            # Task 8
│   └── setup.ts
├── package.json
├── playwright.config.ts                # Task 8
└── go.mod                              # Web module in monorepo

docs/
├── wireframes/                         # Tasks 3-7
│   ├── employee-list.png
│   ├── employee-detail.png
│   ├── employee-form.png
│   ├── org-agent-configs.png
│   └── employee-agent-overrides.png
├── DESIGN_SYSTEM.md                    # Task 9
├── SPRINT_PLAN_v0.3.md                 # This sprint plan
└── GITHUB_TASKS_v0.3.md                # GitHub tasks

.github/workflows/
└── e2e-tests.yml                       # Task 8 (CI/CD)
```

---

**Document Version:** 1.0
**Created By:** Tech Lead
**Last Updated:** 2025-11-01
**Status:** Ready for Sprint Start

---

## How to Use This Document

1. **For Tech Lead:** Coordinate task assignments, review progress daily
2. **For Backend Agent:** Start with Task 1 immediately
3. **For Frontend Agent:** Start with Task 2, then follow sequence
4. **For Product Team:** Review business metrics, provide demo script feedback
5. **For Stakeholders:** Track sprint progress, attend retrospective

**Next Action:** Create GitHub project tasks from [GITHUB_TASKS_v0.3.md](./GITHUB_TASKS_v0.3.md)
