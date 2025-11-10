# Wireframes

This directory contains wireframes for all UI pages.

**Last Updated:** 2025-11-09

---

## 📁 Directory Structure

```
wireframes/
├── epic-1-authentication/          # Authentication & onboarding flows
│   ├── 1.1-user-registration.md    # Signup page
│   ├── 1.2-user-login.md           # Login page
│   ├── 1.4-forgot-password.md      # Password reset request
│   └── 1.4-reset-password.md       # Password reset form
│
├── epic-2-dashboard/               # Admin dashboard
│   └── 2.1-admin-dashboard.md      # Main dashboard with approvals, activity, stats
│
└── [Legacy Wireframes]             # Older wireframes from previous work
    ├── signup.md
    ├── onboarding-wizard.md
    ├── team-management.md
    ├── invitations-list.md
    ├── accept-invitation.md
    ├── org-agent-configs.md
    ├── employee-agent-overrides.md
    └── agent-configure-modal.md
```

---

## 📚 Wireframe Index

### Epic 1: Authentication & Onboarding

| Story | Page | File | Status | Priority |
|-------|------|------|--------|----------|
| 1.1 | `/signup` | [1.1-user-registration.md](./epic-1-authentication/1.1-user-registration.md) | ✅ Implemented | P0 |
| 1.2 | `/login` | [1.2-user-login.md](./epic-1-authentication/1.2-user-login.md) | 🚧 Needs Implementation | P0 |
| 1.4 | `/forgot-password` | [1.4-forgot-password.md](./epic-1-authentication/1.4-forgot-password.md) | 📋 Planned | P2 |
| 1.4 | `/reset-password/[token]` | [1.4-reset-password.md](./epic-1-authentication/1.4-reset-password.md) | 📋 Planned | P2 |

### Epic 2: Dashboard

| Story | Page | File | Status | Priority |
|-------|------|------|--------|----------|
| 2.1 | `/dashboard` | [2.1-admin-dashboard.md](./epic-2-dashboard/2.1-admin-dashboard.md) | 🚧 In Progress | P1 |

---

## 📋 Wireframe Format

All new wireframes (Epic 1+) follow a comprehensive markdown format including:

**Structure:**
- ✅ ASCII visual layouts
- ✅ Component specifications (sizes, spacing, colors)
- ✅ All UI states (default, loading, error, success, empty)
- ✅ Complete user flows (happy path + error scenarios)
- ✅ Validation rules (client-side and server-side)
- ✅ Accessibility annotations (keyboard nav, ARIA, screen readers)
- ✅ Technical implementation notes
- ✅ API endpoint specifications
- ✅ Design decisions with rationale
- ✅ Responsive behavior (mobile, tablet, desktop)

**Example Structure:**
```markdown
# Wireframe: [Page Name]
- Layout Overview (ASCII wireframe)
- Component Specifications
- States & Interactions
- User Flows
- Validation Rules
- Accessibility
- Technical Implementation Notes
- Design Decisions
- Responsive Behavior
```

---

## 🎨 Design System

**Colors:**
- Primary: Blue (#3B82F6)
- Success: Green (#10B981)
- Warning: Amber (#F59E0B)
- Error: Red (#EF4444)

**Typography:**
- Font: Inter
- H1: 32px, bold
- H2: 24px, semi-bold
- Body: 16px, regular

**Components:**
- Based on shadcn/ui
- Tailwind CSS utilities
- WCAG AA accessible

---

## 📐 Naming Convention

**New Format (Epic-based):**
```
epic-{number}-{name}/{story-number}-{page-name}.md
```

**Examples:**
- `epic-1-authentication/1.2-user-login.md`
- `epic-2-dashboard/2.1-admin-dashboard.md`

**Legacy Format:**
```
page-name.md
```

**Examples:**
- `signup.md`
- `team-management.md`

---

## 🔄 Workflow

### For New Pages:

1. **Request Wireframe**
   - Consult **product-designer** agent
   - Provide user story reference

2. **Review Wireframe**
   - Check all states are covered
   - Verify accessibility requirements
   - Validate responsive behavior

3. **Implement UI**
   - Follow wireframe exactly
   - Implement all specified states
   - Meet accessibility standards

4. **Update if Needed**
   - Document any design changes
   - Update wireframe if design evolves

### For Existing Pages:

1. **Update Wireframe First**
   - Request updated wireframe from product-designer
   - Document what's changing

2. **Review & Approve**
   - Verify changes make sense
   - Check impact on user flows

3. **Implement Changes**
   - Follow updated wireframe
   - Update all affected states

---

## ✅ Required For

Wireframes are **MANDATORY** for:
- ✅ New page creation
- ✅ Existing page modifications
- ✅ New UI components
- ✅ Layout changes
- ✅ New user flows

**Do NOT implement UI without wireframes!**

---

## 🛠 Tools

**Current Approach:**
- Markdown documents with ASCII wireframes
- Detailed specifications in markdown
- Version controlled in Git

**Alternative Tools (if needed):**
- Figma, Sketch, Adobe XD (visual design)
- Balsamiq, Wireframe.cc (wireframe tools)
- Screenshots with annotations

---

## 📖 Using These Wireframes

### For Frontend Developers:
1. Read complete wireframe specification
2. Implement all states (not just happy path)
3. Follow accessibility requirements exactly
4. Use specified component sizes/spacing
5. Test all user flows
6. Ask questions if unclear

### For Reviewers:
1. Verify all states implemented
2. Check accessibility (keyboard, screen readers)
3. Test responsive behavior
4. Validate error handling
5. Ensure design system consistency

### For Product Team:
1. Use to understand feature scope
2. Validate user flows match requirements
3. Review error messages and empty states
4. Confirm acceptance criteria coverage

---

## 🤔 Questions or Feedback?

For wireframe questions:
1. Check wireframe document first (includes design decisions)
2. Review related user story for business context
3. Ask **product-designer** agent for clarifications
4. Consult **tech-lead** for technical feasibility

---

## 📦 Legacy Wireframes

Older wireframes exist in the root directory:
- `signup.md` - Original signup wireframe (superseded by 1.1)
- `onboarding-wizard.md` - Multi-step onboarding
- `team-management.md` - Team creation and management
- `invitations-list.md` - Team invitations list
- `accept-invitation.md` - Invitation acceptance flow
- `org-agent-configs.md` - Organization-level agent configs
- `employee-agent-overrides.md` - Employee-specific overrides
- `agent-configure-modal.md` - Agent configuration modal

These remain for reference but may not follow the new comprehensive format.

---

**Maintained by:** Product Designer Agent
**Repository:** ubik-enterprise
**Issue:** #158
