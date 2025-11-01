# Sprint Plan: v0.3.0 - Web UI & Beta Launch
**Sprint Duration:** 2 weeks (November 4-15, 2025)
**Version Target:** v0.3.0
**Primary Goal:** Enable beta launch with functional web UI

---

## Executive Summary

### Current Status (v0.2.0)
- **API Platform:** 39 endpoints, 144+ tests passing, 73-88% coverage
- **CLI Client:** Complete (79 tests, Docker integration, interactive mode)
- **Database:** 20 tables + 3 views, hierarchical config system
- **Architecture:** Monorepo with Go workspace, clean module boundaries

### Strategic Context
**Business Priority:** Beta launch by February 1, 2025
**Critical Dependencies:**
- Landing page needs functional Web UI demo
- Demo video requires v0.3.0 Web UI
- Beta customer outreach needs production-ready platform

### Sprint Objective
Deliver **minimum viable Web UI** to enable:
1. Beta customer signup and onboarding
2. Product demo video recording
3. Internal testing by January 15, 2025

---

## Technical Feasibility Assessment

### 🟢 HIGH FEASIBILITY - Recommended for Sprint
1. **Fix Failing Integration Test** (Blocker)
2. **Web UI Foundation** (New Module)
3. **Core Employee Management UI** (MVP)
4. **Agent Configuration UI** (Core Value Prop)

### 🟡 MEDIUM FEASIBILITY - Conditional
5. **MCP Configuration UI** (If time permits)
6. **System Prompts UI** (Defer to v0.4.0)
7. **Approval Workflows UI** (Defer to v0.4.0)

### 🔴 LOW FEASIBILITY - Defer
- Analytics Dashboard (no backend APIs yet)
- Usage Tracking UI (requires telemetry infrastructure)
- Advanced Policy Management (complex UX)

---

## Sprint Backlog (Prioritized)

### PHASE 1: Foundation & Blockers (Day 1-2)

#### Task 1.1: Fix Employee Creation Integration Test ⚠️ BLOCKER
**Status:** CRITICAL - Blocking all employee endpoints
**Owner:** backend-api agent
**Effort:** 2-4 hours
**Priority:** P0 (Start immediately)

**Problem:**
```
TestCreateEmployee_Integration_Success failing:
- response.Status is empty (expected "active")
- response.RoleId is nil UUID (expected valid UUID)
- Nil pointer dereference on line 557
```

**Root Cause Analysis:**
- Handler not properly mapping DB model to API response
- Missing fields in employee creation response
- Potential issue with role_id foreign key validation

**Acceptance Criteria:**
- ✅ All 144+ existing tests pass
- ✅ Employee creation returns complete API.Employee object
- ✅ Status field populated correctly
- ✅ RoleId field populated correctly
- ✅ No nil pointer dereferences
- ✅ Test coverage maintained at >80%

**Implementation Steps:**
1. Write failing test for complete response mapping
2. Debug CreateEmployee handler response construction
3. Verify DB query returns all required fields
4. Add response field validation
5. Run full test suite
6. Update COVERAGE_ANALYSIS.md

**Dependencies:** None (blocking everything else)
**Risk:** LOW - Well-defined bug fix
**Estimated Story Points:** 3

---

#### Task 1.2: Initialize Web UI Module (Next.js 14)
**Owner:** frontend-web agent
**Effort:** 4-6 hours
**Priority:** P0

**Objectives:**
- Set up Next.js 14 project in `services/web/`
- Configure TypeScript, Tailwind CSS, ShadcnUI
- Establish project structure and conventions
- Wire up authentication with API

**Deliverables:**
```
services/web/
├── app/
│   ├── (auth)/
│   │   ├── login/page.tsx
│   │   └── layout.tsx
│   ├── (dashboard)/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── employees/
│   ├── api/
│   └── layout.tsx
├── components/
│   ├── ui/ (shadcn components)
│   ├── forms/
│   └── layouts/
├── lib/
│   ├── api-client.ts (typed API client)
│   ├── auth.ts (session management)
│   └── utils.ts
├── public/
├── package.json
├── tsconfig.json
├── tailwind.config.ts
└── next.config.js
```

**Acceptance Criteria:**
- ✅ Next.js 14 App Router working
- ✅ TypeScript strict mode enabled
- ✅ Tailwind CSS + ShadcnUI installed
- ✅ API client generated from OpenAPI spec (using openapi-typescript)
- ✅ Login page functional (connects to /auth/login)
- ✅ Protected dashboard route with middleware
- ✅ Logout functionality working
- ✅ Dark mode toggle (optional but easy)

**Dependencies:**
- Requires Task 1.1 complete (API must be stable)

**Risk:** LOW - Standard Next.js setup
**Estimated Story Points:** 5

**Technical Decisions:**
1. **OpenAPI Client Generation:**
```bash
npm install openapi-typescript openapi-fetch
npx openapi-typescript ../../shared/openapi/spec.yaml -o lib/api.ts
```

2. **Authentication Pattern:**
```typescript
// lib/api-client.ts
import createClient from 'openapi-fetch';
import type { paths } from './api';

export const apiClient = createClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    Authorization: `Bearer ${getToken()}`
  }
});
```

3. **Session Management:**
- NextAuth.js OR custom JWT in httpOnly cookie
- Recommendation: Custom JWT (simpler, matches API design)

---

### PHASE 2: Employee Management UI (Day 3-5)

#### Task 2.1: Employee List Page with Table
**Owner:** frontend-web agent
**Effort:** 6-8 hours
**Priority:** P1

**User Story:**
As an admin, I want to see a list of all employees in my organization so I can manage team access.

**Wireframe Required:** `docs/wireframes/employee-list.png`

**Features:**
- Server-side rendered table (React Server Components)
- Client-side filtering (status, team)
- Pagination (20 per page)
- Search by name/email
- Sortable columns (name, email, team, status, created date)
- Action buttons (View, Edit, Deactivate)

**API Endpoints Used:**
- `GET /employees` (with query params)
- `GET /teams` (for filter dropdown)

**Components to Build:**
```typescript
components/
├── employees/
│   ├── EmployeeTable.tsx (Server Component)
│   ├── EmployeeFilters.tsx (Client Component)
│   ├── EmployeeRow.tsx
│   └── EmployeeActions.tsx
└── ui/
    ├── data-table.tsx (generic table component)
    ├── badge.tsx (for status)
    └── dropdown.tsx (for actions)
```

**Acceptance Criteria:**
- ✅ Table displays all employees from current org
- ✅ Filtering by status works (active, inactive, invited)
- ✅ Filtering by team works
- ✅ Search by name/email works (client-side)
- ✅ Pagination works (20 per page)
- ✅ Sorting works (at least by name, created_at)
- ✅ Click "View" navigates to employee detail
- ✅ Status badges color-coded (green=active, gray=inactive, yellow=invited)
- ✅ Responsive design (mobile-friendly)
- ✅ Loading states handled
- ✅ Empty state handled ("No employees found")
- ✅ Error states handled (API errors)

**TDD Requirements:**
- Write Playwright E2E test FIRST:
```typescript
// tests/e2e/employees-list.spec.ts
test('displays employee list with filtering', async ({ page }) => {
  await page.goto('/employees');
  await expect(page.locator('table')).toBeVisible();
  await expect(page.locator('tbody tr')).toHaveCount(5); // seed data

  // Filter by status
  await page.selectOption('[name="status"]', 'active');
  await expect(page.locator('tbody tr')).toHaveCount(3);
});
```

**Dependencies:** Task 1.2 (Web UI foundation)
**Risk:** LOW - Standard CRUD UI
**Estimated Story Points:** 8

---

#### Task 2.2: Employee Detail Page
**Owner:** frontend-web agent
**Effort:** 4-6 hours
**Priority:** P1

**User Story:**
As an admin, I want to view an employee's full profile including assigned agents and MCP servers.

**Wireframe Required:** `docs/wireframes/employee-detail.png`

**Features:**
- Employee profile card (name, email, team, role, status)
- Assigned agents section (cards with config preview)
- Assigned MCP servers section (cards with config preview)
- Edit button (navigates to edit page)
- Delete/Deactivate button (with confirmation modal)
- Activity log (last synced, last login)

**API Endpoints Used:**
- `GET /employees/{id}`
- `GET /employees/{id}/agent-configs/resolved` (hierarchical config)
- `GET /employees/{id}/mcp-configs`

**Components to Build:**
```typescript
components/
├── employees/
│   ├── EmployeeProfile.tsx
│   ├── EmployeeAgents.tsx
│   ├── EmployeeMCPs.tsx
│   ├── EmployeeActivity.tsx
│   └── DeactivateDialog.tsx
└── ui/
    ├── card.tsx
    ├── dialog.tsx (confirmation)
    └── tabs.tsx (for sections)
```

**Acceptance Criteria:**
- ✅ Profile displays all employee fields correctly
- ✅ Agent configurations displayed with resolved config (org → team → employee merge)
- ✅ MCP configurations displayed
- ✅ Edit button navigates to `/employees/{id}/edit`
- ✅ Delete button shows confirmation dialog
- ✅ Deactivation works (soft delete via PATCH status=inactive)
- ✅ Activity log shows last_synced_at timestamp
- ✅ Breadcrumb navigation (Employees → {name})
- ✅ Loading states for async data
- ✅ 404 page if employee not found

**Dependencies:** Task 2.1
**Risk:** MEDIUM - Hierarchical config resolution needs testing
**Estimated Story Points:** 6

---

#### Task 2.3: Create/Edit Employee Form
**Owner:** frontend-web agent
**Effort:** 6-8 hours
**Priority:** P1

**User Story:**
As an admin, I want to create and edit employee accounts so I can onboard new team members.

**Wireframe Required:** `docs/wireframes/employee-form.png`

**Features:**
- Form with validation (email, password, name, team, role)
- Password strength indicator (on create)
- Team dropdown (populated from API)
- Role dropdown (populated from API)
- Status toggle (active/inactive)
- Auto-generate password option
- Send invitation email (future)

**Form Fields:**
```typescript
interface EmployeeFormData {
  email: string;          // Required, email validation
  full_name: string;      // Required
  password?: string;      // Required on create, optional on edit
  team_id?: string;       // Optional (dropdown)
  role_id: string;        // Required (dropdown)
  status: 'active' | 'inactive' | 'invited';
}
```

**API Endpoints Used:**
- `POST /employees` (create)
- `PATCH /employees/{id}` (update)
- `GET /teams` (dropdown)
- `GET /roles` (dropdown)

**Validation Rules:**
- Email: Valid format, unique per org
- Password: Min 8 chars, 1 uppercase, 1 lowercase, 1 number, 1 special
- Full name: Min 2 chars, max 255
- Team: Must exist in org (optional field)
- Role: Must exist (required)

**Acceptance Criteria:**
- ✅ Form validation works (client + server)
- ✅ Create employee works (POST /employees)
- ✅ Edit employee works (PATCH /employees/{id})
- ✅ Team dropdown populated from API
- ✅ Role dropdown populated from API
- ✅ Password strength indicator shows (weak/medium/strong)
- ✅ Password not required on edit (only if changing)
- ✅ Success message on save
- ✅ Error handling (duplicate email, invalid data)
- ✅ Redirect to employee detail on success
- ✅ Cancel button returns to list

**TDD Requirements:**
```typescript
// tests/e2e/employee-form.spec.ts
test('creates employee with valid data', async ({ page }) => {
  await page.goto('/employees/new');
  await page.fill('[name="email"]', 'newuser@example.com');
  await page.fill('[name="full_name"]', 'New User');
  await page.fill('[name="password"]', 'SecureP@ss123');
  await page.selectOption('[name="role_id"]', '...role-uuid...');
  await page.click('button[type="submit"]');

  await expect(page).toHaveURL(/\/employees\/[a-f0-9-]+/);
  await expect(page.locator('text=Employee created successfully')).toBeVisible();
});

test('shows validation errors for invalid email', async ({ page }) => {
  await page.goto('/employees/new');
  await page.fill('[name="email"]', 'invalid-email');
  await page.blur('[name="email"]');
  await expect(page.locator('text=Invalid email format')).toBeVisible();
});
```

**Dependencies:** Task 2.2
**Risk:** MEDIUM - Password validation, duplicate email handling
**Estimated Story Points:** 8

---

### PHASE 3: Agent Configuration UI (Day 6-8)

#### Task 3.1: Organization Agent Configs Page
**Owner:** frontend-web agent
**Effort:** 8-10 hours
**Priority:** P0 (Core Value Prop)

**User Story:**
As an org admin, I want to configure which AI agents are available to my organization so I can control costs and tool access.

**Wireframe Required:** `docs/wireframes/org-agent-configs.png`

**Features:**
- List of all agents from catalog
- Toggle to enable/disable agent for org
- Configure org-level defaults (model, temperature, max_tokens)
- JSON editor for advanced config
- Policy assignment (path restrictions, rate limits)
- Preview resolved config

**API Endpoints Used:**
- `GET /agents` (catalog)
- `GET /organizations/current/agent-configs` (org configs)
- `POST /organizations/current/agent-configs` (enable agent)
- `PATCH /organizations/current/agent-configs/{id}` (update config)
- `DELETE /organizations/current/agent-configs/{id}` (disable agent)

**UI Sections:**
1. **Agent Catalog** (grid of cards)
   - Agent name, provider, description, logo
   - "Enable for Org" button (if not enabled)
   - "Configure" button (if enabled)

2. **Active Org Configs** (table)
   - Agent name
   - Config preview (model, temperature)
   - Policies applied
   - Edit/Delete actions

3. **Config Editor Modal**
   - Agent details (read-only)
   - Config JSON editor with Monaco Editor
   - Policy selector (checkboxes)
   - System prompt input (org-level)
   - Save/Cancel buttons

**Acceptance Criteria:**
- ✅ Displays all agents from catalog
- ✅ Displays current org agent configs
- ✅ "Enable Agent" creates org_agent_configs entry with default config
- ✅ Config editor shows current config in JSON
- ✅ Saving config updates org_agent_configs via PATCH
- ✅ Deleting config removes agent from org (soft delete)
- ✅ Validation: Invalid JSON shows error
- ✅ Validation: Required fields checked (model, etc.)
- ✅ Preview shows what employees will inherit
- ✅ Success/error toast notifications

**TDD Requirements:**
```typescript
// tests/e2e/org-agent-configs.spec.ts
test('enables agent for organization', async ({ page }) => {
  await page.goto('/settings/agents');
  await expect(page.locator('text=Claude Code')).toBeVisible();

  // Enable agent
  await page.click('button:has-text("Enable for Org")');
  await page.fill('[name="model"]', 'claude-3-5-sonnet-20241022');
  await page.click('button:has-text("Save")');

  // Verify in active configs
  await expect(page.locator('table tbody tr:has-text("Claude Code")')).toBeVisible();
});

test('updates agent config with JSON editor', async ({ page }) => {
  await page.goto('/settings/agents');
  await page.click('button:has-text("Configure")');

  // Edit JSON (Monaco Editor)
  await page.locator('.monaco-editor').click();
  await page.keyboard.type('{"model": "claude-3-5-sonnet-20241022", "temperature": 0.7}');
  await page.click('button:has-text("Save")');

  await expect(page.locator('text=Configuration updated')).toBeVisible();
});
```

**Dependencies:** Task 1.2 (Web UI foundation)
**Risk:** HIGH - Complex JSON editing, hierarchical config
**Estimated Story Points:** 13

**Technical Notes:**
- Use Monaco Editor for JSON editing: `@monaco-editor/react`
- Validate JSON schema against agent's expected config
- Show diff preview: org default → team override → employee override

---

#### Task 3.2: Employee Agent Overrides Page
**Owner:** frontend-web agent
**Effort:** 6-8 hours
**Priority:** P1

**User Story:**
As an admin, I want to assign specific agent overrides to individual employees so they can customize their experience.

**Wireframe Required:** `docs/wireframes/employee-agent-overrides.png`

**Features:**
- View employee's inherited agents (from org/team)
- Add employee-specific overrides
- Edit override config (partial config, merges with org/team)
- Remove override (revert to org/team defaults)
- Preview final resolved config

**UI Components:**
1. **Inherited Agents Section** (cards)
   - Agent name, source (org or team)
   - Config preview (read-only)
   - "Add Override" button

2. **Employee Overrides Section** (cards)
   - Agent name
   - Override config preview
   - Edit/Remove buttons
   - Badge showing "Override Active"

3. **Add/Edit Override Modal**
   - Base config display (org + team merged, read-only)
   - Override fields (temperature, max_tokens, etc.)
   - Preview merged config
   - Save/Cancel

**API Endpoints Used:**
- `GET /employees/{id}/agent-configs/resolved` (inherited configs)
- `POST /employees/{id}/agent-configs` (create override)
- `PATCH /employees/{id}/agent-configs/{config_id}` (update override)
- `DELETE /employees/{id}/agent-configs/{config_id}` (remove override)

**Acceptance Criteria:**
- ✅ Displays inherited agents from org/team
- ✅ Shows source of each config (org, team, or employee override)
- ✅ "Add Override" creates employee_agent_configs entry
- ✅ Override editor shows base config + override fields
- ✅ Preview shows final merged config (org → team → employee)
- ✅ Saving override updates employee_agent_configs
- ✅ Removing override soft deletes employee_agent_configs
- ✅ Validation: Override must be valid partial config
- ✅ Cannot override if agent not enabled at org/team level
- ✅ Success/error notifications

**TDD Requirements:**
```typescript
// tests/e2e/employee-agent-overrides.spec.ts
test('adds employee agent override', async ({ page, employeeId }) => {
  await page.goto(`/employees/${employeeId}`);
  await page.click('text=Agents');

  // Verify inherited agent visible
  await expect(page.locator('text=Claude Code (from org)')).toBeVisible();

  // Add override
  await page.click('button:has-text("Add Override")');
  await page.fill('[name="temperature"]', '0.9');
  await page.click('button:has-text("Save Override")');

  // Verify override applied
  await expect(page.locator('text=Claude Code (override active)')).toBeVisible();
  await expect(page.locator('text=temperature: 0.9')).toBeVisible();
});
```

**Dependencies:** Task 3.1 (Org configs must exist)
**Risk:** MEDIUM - Config merging logic, partial updates
**Estimated Story Points:** 8

---

### PHASE 4: Testing & Polish (Day 9-10)

#### Task 4.1: E2E Test Suite with Playwright
**Owner:** frontend-web agent + backend-api agent
**Effort:** 6-8 hours
**Priority:** P1

**Objectives:**
- Comprehensive E2E tests for all user flows
- Seed test database with realistic data
- Automated test runs in CI/CD

**Test Scenarios:**
1. **Authentication Flow**
   - Login with valid credentials
   - Login with invalid credentials
   - Logout
   - Protected route access (redirect to login)

2. **Employee Management Flow**
   - List employees
   - Filter by status, team
   - Create employee
   - Edit employee
   - Deactivate employee
   - Search employees

3. **Agent Configuration Flow**
   - Enable agent for org
   - Configure org-level defaults
   - Assign agent to employee
   - Add employee override
   - View resolved config
   - Disable agent (verify cascade to employees)

4. **Error Handling**
   - Network errors
   - 401 Unauthorized
   - 404 Not Found
   - 500 Server Error
   - Form validation errors

**Acceptance Criteria:**
- ✅ All user flows covered with E2E tests
- ✅ Tests run against real PostgreSQL + API
- ✅ Seed data script for test database
- ✅ All tests passing (green)
- ✅ Test coverage >80% for critical paths
- ✅ Tests run in CI/CD (GitHub Actions)
- ✅ Test reports generated (HTML, JSON)

**Test Infrastructure:**
```typescript
// tests/setup.ts
import { test as base } from '@playwright/test';
import { execSync } from 'child_process';

const test = base.extend({
  async seedDatabase({ page }, use) {
    // Seed test data
    execSync('make db-reset && make db-seed-test');
    await use(page);
    // Cleanup after test
  }
});
```

**Dependencies:** Tasks 2.1, 2.2, 2.3, 3.1, 3.2 (all UI complete)
**Risk:** MEDIUM - Flaky tests, timing issues
**Estimated Story Points:** 8

---

#### Task 4.2: UI/UX Polish & Accessibility
**Owner:** frontend-web agent
**Effort:** 4-6 hours
**Priority:** P2

**Objectives:**
- Consistent design system (colors, spacing, typography)
- Accessibility compliance (WCAG 2.1 AA)
- Loading states and skeletons
- Error boundaries
- Responsive design (mobile, tablet, desktop)

**Deliverables:**
1. **Design System Documentation**
   - Color palette (primary, secondary, success, error, warning)
   - Typography scale (headings, body, captions)
   - Spacing scale (4px base unit)
   - Component variants

2. **Accessibility Audit**
   - Keyboard navigation works (tab order)
   - Screen reader labels (aria-labels)
   - Color contrast >4.5:1
   - Focus indicators visible
   - Form labels associated

3. **Loading States**
   - Skeleton loaders for tables
   - Spinner for async actions
   - Optimistic UI updates (where safe)
   - Progress indicators for multi-step flows

4. **Error Handling**
   - Error boundary component
   - Toast notifications (success, error, warning)
   - Inline form validation
   - Friendly error messages

**Acceptance Criteria:**
- ✅ All pages responsive (mobile, tablet, desktop)
- ✅ Keyboard navigation works everywhere
- ✅ Screen reader testing passes (basic)
- ✅ Color contrast meets WCAG AA
- ✅ Loading states for all async operations
- ✅ Error boundaries catch React errors
- ✅ Toast notifications consistent
- ✅ No console errors or warnings

**Tools:**
- axe DevTools (accessibility testing)
- Lighthouse (performance + accessibility)
- React Testing Library (component testing)

**Dependencies:** All UI tasks complete
**Risk:** LOW - Standard polish work
**Estimated Story Points:** 5

---

## Sprint Capacity & Velocity

### Team Capacity (2 weeks)
- **Tech Lead (you):** 20 hours (coordination, architecture, reviews)
- **backend-api agent:** 40 hours (API fixes, endpoint additions)
- **frontend-web agent:** 60 hours (Web UI development)
- **Total:** 120 hours

### Story Point Allocation
| Task | Owner | Story Points | Hours |
|------|-------|--------------|-------|
| 1.1 Fix Integration Test | backend-api | 3 | 4 |
| 1.2 Web UI Foundation | frontend-web | 5 | 6 |
| 2.1 Employee List Page | frontend-web | 8 | 8 |
| 2.2 Employee Detail Page | frontend-web | 6 | 6 |
| 2.3 Employee Form | frontend-web | 8 | 8 |
| 3.1 Org Agent Configs | frontend-web | 13 | 12 |
| 3.2 Employee Overrides | frontend-web | 8 | 8 |
| 4.1 E2E Test Suite | frontend-web + backend-api | 8 | 10 |
| 4.2 UI/UX Polish | frontend-web | 5 | 6 |
| **TOTAL** | | **64** | **68** |

### Velocity Projection
- **Sprint 1 Velocity:** 64 story points (estimated)
- **Stretch Goal:** Add MCP UI (if >20% time remaining)
- **Minimum Viable:** Tasks 1.1, 1.2, 2.1, 2.2, 2.3, 3.1 (MVP for demo)

---

## Risk Assessment & Mitigation

### HIGH RISK 🔴
**Risk:** Hierarchical config resolution UI too complex
**Impact:** Users confused, bugs in config merging
**Mitigation:**
- Start with simple UI (no deep merging in v0.3.0)
- Show read-only inherited config + editable overrides
- Defer full JSON merge preview to v0.4.0
- Add tooltips explaining config inheritance

### MEDIUM RISK 🟡
**Risk:** E2E tests flaky or slow
**Impact:** CI/CD unreliable, slows development
**Mitigation:**
- Use testcontainers for isolated test DB
- Implement retry logic for flaky tests
- Run tests in parallel (Playwright workers)
- Mock external APIs (if any)

**Risk:** Integration test blocker delays sprint start
**Impact:** Web UI waits for API fix
**Mitigation:**
- Fix test on Day 1 (highest priority)
- Parallel work: UI foundation can start while fixing
- Use API mock server if needed (openapi-mock)

### LOW RISK 🟢
**Risk:** Scope creep (too many features)
**Impact:** Sprint overcommitted, features incomplete
**Mitigation:**
- Strict MVP definition (no feature additions)
- Weekly sprint review (cut scope if behind)
- Defer MCP UI, system prompts, approvals to v0.4.0

---

## Dependencies & Sequence

### Critical Path
```
Day 1-2:  Task 1.1 (Test Fix) → BLOCKS → Task 1.2 (Web UI Foundation)
Day 3-5:  Task 1.2 → BLOCKS → Task 2.1 (Employee List)
          Task 2.1 → Task 2.2 (Detail) → Task 2.3 (Form)
Day 6-8:  Task 1.2 → Task 3.1 (Org Configs)
          Task 3.1 → Task 3.2 (Employee Overrides)
Day 9-10: All UI tasks → Task 4.1 (E2E Tests) → Task 4.2 (Polish)
```

### Parallel Work Opportunities
- **Day 3-5:** Task 2.1 and Task 3.1 can be parallel (different devs)
- **Day 6-8:** Task 2.3 and Task 3.2 can be parallel
- **Day 9:** E2E tests + Polish can overlap

---

## Definition of Done (DoD)

### For Each Task
- ✅ Code reviewed (self-review + tech lead approval)
- ✅ All tests passing (unit + integration + E2E)
- ✅ Test coverage >80% for new code
- ✅ No console errors or warnings
- ✅ Wireframe matches implementation (or updated)
- ✅ API types match OpenAPI spec
- ✅ Documentation updated (if architectural change)
- ✅ Accessibility checklist passed (keyboard, screen reader)
- ✅ Responsive design verified (mobile, tablet, desktop)

### For Sprint (v0.3.0 Release)
- ✅ All P0 and P1 tasks complete
- ✅ E2E test suite passing (>80% coverage)
- ✅ No critical bugs (P0/P1)
- ✅ Performance: Page load <2s, API response <200ms
- ✅ Security: No exposed secrets, HTTPS enforced
- ✅ Demo video can be recorded
- ✅ Beta landing page can link to working app
- ✅ Internal team can dogfood (use the platform)

---

## Success Metrics

### Technical Metrics
- **Test Coverage:** >80% overall (frontend + backend)
- **API Response Time:** <100ms (p95)
- **Web UI Load Time:** <2s (p95)
- **E2E Tests Passing:** 100% (no flaky tests)
- **Accessibility Score:** >90 (Lighthouse)
- **Performance Score:** >85 (Lighthouse)

### Business Metrics (Post-Sprint)
- **Internal Dogfooding:** 5+ team members using daily
- **Demo Video:** Recorded by January 20, 2025
- **Beta Signups:** Landing page ready to capture leads
- **Beta Launch Readiness:** Pass pre-launch checklist (security, performance, UX)

---

## Post-Sprint Retrospective

### Questions to Answer
1. Did we achieve the sprint goal (MVP Web UI)?
2. What blockers slowed us down?
3. Were story point estimates accurate?
4. What technical debt did we incur?
5. What should we do differently in v0.4.0 sprint?

### Lessons Learned (To Be Filled Post-Sprint)
- TBD

---

## Next Sprint Preview (v0.4.0)

### Deferred Features
- MCP Configuration UI (Task 5.1-5.3)
- System Prompts UI (Task 6.1)
- Approval Workflows UI (Task 7.1-7.3)
- Usage Analytics Dashboard (Task 8.1)
- Advanced Policy Management (Task 9.1)

### Estimated Effort
- v0.4.0 Sprint: 6-8 weeks (larger scope)
- Focus: Complete beta feature set + polish
- Target: Production-ready by January 31, 2025

---

**Sprint Plan Version:** 1.0
**Created By:** Tech Lead
**Last Updated:** 2025-11-01
**Status:** DRAFT (Pending Product Strategy Review)
