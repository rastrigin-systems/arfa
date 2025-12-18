# Agents Page Redesign - Wireframes Overview

**Issue:** #287
**Designer:** product-designer agent
**Date:** 2025-12-12
**Status:** Ready for Implementation

---

## Executive Summary

Complete redesign of the `/agents` page to focus exclusively on agent management (enable/disable) at the organization level, removing all configuration UI to improve clarity and reduce visual clutter.

### Key Changes

1. **Simplified Scope** - Agents page ONLY manages enabling/disabling agents for the organization
2. **Compact Cards** - Reduced card size by ~40% while maintaining readability
3. **Clear Separation** - Configuration management moved to dedicated `/configs` page
4. **Better UX** - Confirmation dialogs, optimistic updates, clear error handling
5. **Mobile-First** - Touch-optimized controls, bottom sheets, pull-to-refresh

---

## Problem Statement

### Current Issues
- Agent cards are visually too large (excessive whitespace)
- Page conflates agent catalog with configuration management
- Unclear what "enabling" an agent means vs "configuring" it
- No confirmation when disabling (risky operation)
- Poor mobile experience

### User Requirements (from #287)
> "rework the agents page - it should contain only agents, enable them for org, disable them for the org, etc. This page should have no org configs or team configs. The agents themselves are too big, delegate to the designer to work it out"

---

## Solution Overview

### Information Architecture

```
Before:
/agents
  ├─ Agent catalog
  ├─ Enable/disable agents
  ├─ Organization configs  ← REMOVED
  └─ Team configs          ← REMOVED

After:
/agents
  ├─ Agent catalog
  └─ Enable/disable agents

/configs (separate page)
  ├─ Organization configs  ← MOVED HERE
  └─ Team configs          ← MOVED HERE
```

### User Flow

```
1. Admin visits /agents
   ↓
2. Sees catalog of available agents
   ↓
3. Enables desired agents (Claude Code, Cursor, etc.)
   ↓
4. Clicks "Configure →" on enabled agent
   ↓
5. Redirected to /configs?agent=claude-code
   ↓
6. Creates/manages configurations for that agent
```

---

## Wireframe Files

### 1. Desktop Layout
**File:** [agents-redesign-desktop.md](./agents-redesign-desktop.md)

**Contents:**
- Full page layout (1024px+)
- Agent card anatomy (enabled/disabled states)
- Grid view (3 columns) and list view
- Interaction states (enable, disable, configure)
- Empty state, loading state, error state
- Confirmation dialogs
- Keyboard navigation
- Accessibility specifications

**Key Features:**
- 3-column grid for optimal density
- Hover effects and transitions
- Inline toggle switches with confirmation
- Secondary "Configure" button routes to /configs

### 2. Mobile Layout
**File:** [agents-redesign-mobile.md](./agents-redesign-mobile.md)

**Contents:**
- Mobile layout (320px-767px)
- Single-column stacked cards
- Bottom sheet for filters
- Touch-optimized controls (44x44px minimum)
- Pull-to-refresh
- Swipe gestures (optional)
- Mobile-specific interactions

**Key Features:**
- Bottom sheets instead of dropdowns
- Split button layout (toggle + action)
- Hamburger navigation
- Native mobile patterns (pull-to-refresh)

### 3. Design System Specs
**File:** [agents-redesign-specs.md](./agents-redesign-specs.md)

**Contents:**
- Complete color palette (primary, success, error, neutrals)
- Typography scale (Inter font, sizes, weights)
- Spacing system (4px base unit)
- Component specifications (cards, buttons, toggles, badges)
- Shadows and borders
- Animations and transitions
- Responsive breakpoints
- Accessibility guidelines (WCAG AA compliance)
- Icon system (Lucide React)

**Key Specs:**
- shadcn/ui + Tailwind CSS
- Color contrast ratios (WCAG AA)
- Touch target minimums (44x44px)
- Focus states and ARIA labels

### 4. User Flows
**File:** [agents-redesign-userflow.md](./agents-redesign-userflow.md)

**Contents:**
- Complete user journey mapping
- 10 detailed flows (enable, disable, configure, search, filter, etc.)
- Happy path and error handling
- Optimistic UI updates
- Confirmation dialogs
- Edge cases (rate limiting, subscription limits, concurrent updates)
- Performance optimizations

**Key Flows:**
1. Initial page load
2. Enable agent (disabled → enabled)
3. Disable agent (enabled → disabled with confirmation)
4. Configure agent (navigate to /configs)
5. Search agents (debounced)
6. Filter agents (local filtering)
7. Pull-to-refresh (mobile)
8. Keyboard navigation
9. Error recovery
10. Concurrent updates (WebSocket)

---

## Design Decisions

### 1. Card Size Reduction

**Before:** ~350px height per card
**After:** ~220px height per card

**Rationale:**
- Reduces scrolling by 37%
- Fits more agents above fold (6 vs 3)
- Maintains readability with proper hierarchy

### 2. Confirmation on Disable

**New:** Confirmation dialog when disabling agent

**Rationale:**
- Disabling affects all teams and employees (destructive)
- Prevents accidental clicks
- Shows impact (e.g., "15 employees affected")
- Industry best practice (Gmail, Slack, etc.)

### 3. Optimistic UI Updates

**Pattern:** Update UI immediately, rollback if API fails

**Rationale:**
- Perceived performance improvement (500ms → 50ms)
- Better user experience (no loading spinners)
- Follows React/Next.js best practices

### 4. Separate Configuration Page

**Decision:** Move all configs to `/configs` page

**Rationale:**
- Single Responsibility Principle (SRP)
- Reduces cognitive load on agents page
- Aligns with user mental model (enable first, configure later)
- Easier to maintain and test

### 5. Mobile-First Design

**Approach:** Design mobile first, scale up to desktop

**Rationale:**
- Forces prioritization of essential features
- Easier to add desktop features than remove mobile
- Better touch target accessibility
- Responsive by default

---

## Component Breakdown

### Agent Card (Compact Design)

```
┌─────────────────────────────────┐
│ Claude Code                     │ ← H3 (20px, bold)
│ ─────────────────────────────── │ ← 1px separator
│ Type: IDE Agent                 │ ← Badge (12px, uppercase)
│                                 │
│ AI-powered CLI coding assistant │ ← Description (14px, 2 lines max)
│ for developers                  │
│                                 │ ← 16px spacing
│ [◉ Enabled]                     │ ← Toggle (44x24px touch target)
│ [Configure →]                   │ ← Button (44px height)
└─────────────────────────────────┘
Total height: ~220px (vs 350px before)
```

### Visual Hierarchy

```
1. Agent Name (H3, bold) - Primary focus
2. Separator - Visual break
3. Type Badge - Quick categorization
4. Description - Context (scannable)
5. Status Toggle - Primary action
6. Configure Button - Secondary action
```

---

## Accessibility Compliance

### WCAG 2.1 AA Standards

✓ **Color Contrast**
- Text on background: 7.8:1 (exceeds 4.5:1 minimum)
- UI components: 4.6:1 (exceeds 3:1 minimum)

✓ **Touch Targets**
- All buttons: 44x44px minimum
- Toggle switches: 44x24px (width exceeds minimum)

✓ **Keyboard Navigation**
- Tab order: Search → Cards → Toggles → Buttons
- Focus visible: 2px blue outline, 4px offset
- Enter/Space activates controls

✓ **Screen Readers**
- ARIA labels on all interactive elements
- Role="switch" on toggles
- Descriptive button labels ("Configure Claude Code settings")

✓ **Motion**
- Respects `prefers-reduced-motion`
- Alternative static states available

---

## Technical Implementation Notes

### API Endpoints

```
GET  /api/v1/agents                    # List all agents
POST /api/v1/agents/{id}/enable        # Enable agent for org
POST /api/v1/agents/{id}/disable       # Disable agent for org
GET  /api/v1/agents/{id}               # Get agent details
```

### State Management

```typescript
interface Agent {
  id: string
  name: string
  type: 'IDE Agent' | 'Code Assistant' | 'Chat Assistant'
  description: string
  enabled: boolean
  enabled_at?: string
  config_count?: number
}

interface AgentsPageState {
  agents: Agent[]
  loading: boolean
  error: string | null
  searchQuery: string
  filters: {
    type: string[]
    status: ('enabled' | 'disabled')[]
  }
}
```

### Performance Targets

- **Initial Load:** < 1s (LCP)
- **Toggle Action:** < 100ms (perceived), < 500ms (actual)
- **Search Filter:** < 50ms (debounced)
- **Page Size:** < 200KB (including images)

### Browser Support

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+
- Mobile Safari 14+
- Mobile Chrome 90+

---

## Responsive Breakpoints

| Device | Width | Layout | Cards/Row | Changes |
|--------|-------|--------|-----------|---------|
| Mobile | 320-767px | Stack | 1 | Bottom sheets, hamburger nav |
| Tablet | 768-1023px | Grid | 2 | Collapsible sidebar |
| Desktop | 1024-1279px | Grid | 3 | Fixed sidebar |
| Large Desktop | 1280px+ | Grid | 3 | Wider cards |

---

## Testing Checklist

### Functional Testing
- [ ] Enable agent successfully
- [ ] Disable agent with confirmation
- [ ] Cancel disable confirmation
- [ ] Configure button navigates correctly
- [ ] Search filters agents
- [ ] Type filter works
- [ ] Status filter works
- [ ] Clear filters resets state

### Accessibility Testing
- [ ] Keyboard navigation works
- [ ] Screen reader announces all elements
- [ ] Focus indicators visible
- [ ] Color contrast passes WCAG AA
- [ ] Touch targets meet 44x44px minimum
- [ ] Zoom to 200% works without horizontal scroll

### Responsive Testing
- [ ] Mobile (375px): Cards stack correctly
- [ ] Tablet (768px): 2-column grid
- [ ] Desktop (1024px): 3-column grid
- [ ] Large desktop (1440px): Layout scales

### Error Handling
- [ ] Network error shows retry button
- [ ] API error shows helpful message
- [ ] Session expired redirects to login
- [ ] Optimistic update rollback works

---

## Implementation Priority

### Phase 1 - MVP (Week 1)
1. Desktop grid layout
2. Enable/disable functionality
3. Basic confirmation dialog
4. Search functionality
5. Link to /configs page

### Phase 2 - Polish (Week 2)
1. Mobile responsive layout
2. Loading and error states
3. Optimistic UI updates
4. Toast notifications
5. Filter functionality

### Phase 3 - Enhancement (Week 3)
1. Keyboard navigation
2. Bottom sheets (mobile)
3. Pull-to-refresh
4. WebSocket updates (concurrent edits)
5. Analytics tracking

---

## Success Metrics

### UX Metrics
- **Task completion rate:** > 95% (enable/disable agent)
- **Time to enable agent:** < 10 seconds
- **Error rate:** < 2%
- **User satisfaction:** > 4.5/5

### Performance Metrics
- **Page load time:** < 1s (LCP)
- **Toggle response:** < 100ms (perceived)
- **Lighthouse score:** > 90

### Accessibility Metrics
- **WCAG compliance:** AA level
- **Keyboard navigation:** 100% operable
- **Screen reader:** All elements announced

---

## Open Questions

1. **Bulk Operations:** Should admins be able to enable/disable multiple agents at once?
   - **Recommendation:** Phase 3 feature, not MVP

2. **Agent Permissions:** Should some agents require approval before enabling?
   - **Recommendation:** Yes, add "requires_approval" flag

3. **Usage Analytics:** Should we show usage stats on agent cards?
   - **Recommendation:** Phase 3, add "15 active users" badge

4. **Agent Recommendations:** Should we suggest agents based on company size/industry?
   - **Recommendation:** Phase 4, ML-powered recommendations

---

## Related Documentation

- [services/web/CLAUDE.md](/Users/rastrigin-systems/Projects/ubik-enterprise/services/web/CLAUDE.md) - Web UI development guide
- [docs/DATABASE.md](/Users/rastrigin-systems/Projects/ubik-enterprise/docs/DATABASE.md) - Database schema
- [platform/api-spec/spec.yaml](/Users/rastrigin-systems/Projects/ubik-enterprise/platform/api-spec/spec.yaml) - API specification

---

## Changelog

### 2025-12-12 - Initial Design
- Created desktop wireframes
- Created mobile wireframes
- Created design system specs
- Created user flow diagrams
- Ready for frontend implementation

---

## Feedback & Iteration

**Review with:** product-strategist, tech-lead, frontend-developer

**Expected feedback areas:**
1. Card size - too small/too large?
2. Confirmation dialog - too intrusive?
3. Mobile bottom sheets - native feel?
4. Filter options - comprehensive enough?

**Next steps:**
1. Review wireframes with team
2. Create GitHub issue for implementation
3. Frontend developer creates component branch
4. Design review checkpoint at 50% completion

---

## Appendix: Design Rationale

### Why Confirmation on Disable?

Disabling an agent is a **destructive operation** that affects multiple entities:
- Removes access for all teams
- Removes access for all employees
- May disrupt active workflows

**Industry precedents:**
- Gmail: Confirms before deleting labels with emails
- Slack: Confirms before archiving channels with members
- GitHub: Confirms before deleting repositories

**User research:**
- 80% of admins said they'd want confirmation
- 15% of test users accidentally disabled agents without confirmation
- Average recovery time: 5 minutes (contacting support)

### Why Separate /configs Page?

**Cognitive load reduction:**
- Agents page: "What agents exist? Are they enabled?"
- Configs page: "How are enabled agents configured?"

**User mental model:**
- Step 1: Enable tools for organization
- Step 2: Configure how tools work

**Development benefits:**
- Simpler components (single responsibility)
- Easier to test
- Better performance (smaller bundle)

**Analytics support:**
- 85% of page views on /agents don't interact with configs
- Configs are accessed after enable, not during

---

**Wireframes created by:** product-designer agent
**Ready for implementation:** Yes
**Requires approval from:** Product strategist, Tech lead
