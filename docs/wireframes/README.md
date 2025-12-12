# Wireframes Directory

This directory contains wireframes, design specifications, and user flows for Ubik Enterprise UI features.

## Current Wireframes

### Agent Configurations Page

**Status:** ✅ Complete - Ready for Implementation

A unified "Configs" page that consolidates all AI agent configurations (organization, team, and employee levels) into a single, filterable view.

**Documents:**

| File | Description |
|------|-------------|
| [configs-overview.md](./configs-overview.md) | Main overview with user stories, flows, and implementation checklist |
| [configs-desktop.md](./configs-desktop.md) | Desktop view (1440px) with table layout, filters, and interactions |
| [configs-create-modal.md](./configs-create-modal.md) | 2-step creation wizard with form and JSON editor |
| [configs-empty-state.md](./configs-empty-state.md) | Empty, loading, and error states |
| [configs-mobile.md](./configs-mobile.md) | Mobile responsive view (375px) with card layout |
| [configs-design-system.md](./configs-design-system.md) | Complete design system reference with colors, typography, components |

**Quick Links:**
- [User Stories](./configs-overview.md#user-stories)
- [Component Hierarchy](./configs-overview.md#component-hierarchy)
- [User Flows](./configs-overview.md#user-flows)
- [API Integration](./configs-overview.md#api-integration)
- [Implementation Checklist](./configs-overview.md#implementation-checklist)

---

## How to Use These Wireframes

### For Frontend Developers

1. **Start with Overview:** Read [configs-overview.md](./configs-overview.md) for context
2. **Review Desktop First:** Check [configs-desktop.md](./configs-desktop.md) for main layout
3. **Study Components:** Reference [configs-design-system.md](./configs-design-system.md) for specs
4. **Implement States:** Don't forget [configs-empty-state.md](./configs-empty-state.md)
5. **Make Responsive:** Use [configs-mobile.md](./configs-mobile.md) for mobile breakpoints

### For Product Managers

1. **Validate User Stories:** Check if [user stories](./configs-overview.md#user-stories) match requirements
2. **Review Flows:** Walk through [user flows](./configs-overview.md#user-flows) diagrams
3. **Approve Scope:** Review [implementation checklist](./configs-overview.md#implementation-checklist)
4. **Track Progress:** Use checklist to monitor development

### For QA Engineers

1. **Test Cases:** Use user flows as basis for test scenarios
2. **State Coverage:** Verify all states in [empty-state.md](./configs-empty-state.md) are tested
3. **Accessibility:** Use [design system](./configs-design-system.md#accessibility) checklist
4. **Responsive:** Test all breakpoints from [mobile.md](./configs-mobile.md)

### For Backend Developers

1. **API Requirements:** See [API Integration](./configs-overview.md#api-integration) section
2. **Response Schema:** Implement schema as specified
3. **Filter Support:** Add query param support for filters/search
4. **Status Codes:** Return appropriate codes for empty/error states

---

## Wireframe Standards

All wireframes in this directory follow these conventions:

### Format

- **ASCII Art:** Primary format for quick iteration and version control
- **Markdown:** Structured documentation with specifications
- **Mermaid Diagrams:** User flows and state machines

### Structure

Each feature wireframe set includes:

1. **Overview Document:** User stories, flows, implementation plan
2. **Desktop View:** Primary desktop layout (1440px)
3. **Modal/Dialog Views:** Any overlays or pop-ups
4. **Empty/Loading/Error States:** All non-happy-path states
5. **Mobile View:** Responsive mobile layout (375px)
6. **Design System Reference:** Component specifications

### Specifications Include

- ✅ All UI states (default, hover, active, disabled, loading, error)
- ✅ Responsive breakpoints (mobile 375px, tablet 768px, desktop 1440px)
- ✅ Component dimensions (width, height, padding, margin)
- ✅ Typography (font family, size, weight, line height, color)
- ✅ Colors (hex, rgb, usage)
- ✅ Spacing (using 4px base unit)
- ✅ Accessibility (ARIA labels, keyboard navigation, focus states)
- ✅ Interactions (hover, click, drag, swipe)
- ✅ Animations (transitions, timing functions)

---

## Design System

### Core Principles

1. **Consistency:** Use shadcn/ui components, Tailwind utilities
2. **Accessibility:** WCAG 2.1 AA minimum, AAA target
3. **Responsiveness:** Mobile-first, touch-friendly (44x44px targets)
4. **Performance:** Lazy loading, skeleton states, optimistic updates
5. **Clarity:** Clear labels, helpful errors, contextual guidance

### Color Palette

- **Primary:** Blue #3B82F6
- **Success:** Green #10B981
- **Error:** Red #EF4444
- **Warning:** Amber #F59E0B
- **Neutrals:** Gray scale 50-900

### Typography

- **UI Font:** Inter
- **Code Font:** JetBrains Mono
- **Sizes:** 12px - 36px scale
- **Weights:** Regular 400, Medium 500, Semibold 600

### Spacing

- **Base Unit:** 4px
- **Common:** 8px, 16px, 24px, 32px
- **Touch Targets:** 44x44px minimum

See [configs-design-system.md](./configs-design-system.md) for complete reference.

---

## Approval Workflow

### Before Implementation

1. **Product Designer** creates wireframes
2. **Product Strategist** validates user stories and flows
3. **Tech Lead** validates API feasibility and architecture
4. **Frontend Developer** reviews component specs
5. **Project Lead** approves scope and timeline

### Sign-Off Checklist

- [ ] User stories match product requirements
- [ ] All states and edge cases covered
- [ ] API endpoints feasible
- [ ] Component specs complete
- [ ] Accessibility requirements met
- [ ] Responsive breakpoints defined
- [ ] Implementation checklist created
- [ ] Estimated timeline agreed upon

### After Approval

1. **Frontend Developer** creates GitHub issue from checklist
2. **Development** begins implementation
3. **QA Engineer** creates test plan from flows
4. **Product Designer** available for questions/clarifications

---

## Version History

### v1.0 - Agent Configurations Page

**Date:** 2025-01-15
**Status:** ✅ Complete - Ready for Review
**Designer:** Product Designer Agent
**Documents:** 6 files (overview, desktop, modal, empty, mobile, design system)

**Changes:**
- Initial wireframes for `/configs` page
- Unified org/team/employee configs view
- 2-step creation wizard
- Mobile responsive card layout
- Complete design system reference

**Next Steps:**
- Review with product-strategist
- Review with tech-lead
- Approve and create implementation issue

---

## Contributing

### Creating New Wireframes

1. **User Stories First:** Define what users need to accomplish
2. **Review Existing Patterns:** Reuse components from design system
3. **Create ASCII Wireframes:** Start with quick sketches
4. **Add Specifications:** Document all component details
5. **Cover All States:** Default, loading, empty, error, success
6. **Mobile Responsive:** Design mobile view separately
7. **Accessibility:** Include ARIA labels, keyboard nav, focus states
8. **Save to /docs/wireframes/:** Follow naming convention

### Naming Convention

```
{feature}-{view}-{variant}.md

Examples:
configs-desktop.md
configs-create-modal.md
configs-empty-state.md
configs-mobile.md
employee-list-desktop.md
employee-edit-modal.md
```

### Required Sections

Each wireframe document should include:

```markdown
# Feature Name - View Name

## Purpose
[What problem this solves]

## Wireframe
[ASCII art or visual]

## Component Specifications
[Detailed specs: dimensions, colors, fonts, spacing]

## User Flow
[How users interact with this view]

## Accessibility
[Keyboard nav, ARIA labels, focus management, screen reader support]

## Edge Cases
[Loading, empty, error states]
```

### Review Process

1. Create wireframe documents
2. Open PR with wireframes
3. Tag reviewers: product-strategist, tech-lead, frontend-developer
4. Address feedback
5. Get approval
6. Merge to main
7. Create implementation issue

---

## FAQ

**Q: Why ASCII wireframes instead of Figma?**
A: ASCII wireframes are:
- Version controlled (git diff shows changes)
- Always in sync with code (no external tool)
- Fast to iterate on
- Accessible to all team members
- Easy to reference in issues/PRs

**Q: When do we need high-fidelity mockups?**
A: For:
- Marketing pages (visual design critical)
- Client presentations
- Complex visualizations
- Unusual interactions not covered by design system

**Q: How detailed should wireframes be?**
A: Include:
- All UI states (loading, empty, error, success)
- Exact dimensions for all components
- Color values (hex), font sizes, spacing
- Accessibility requirements
- Responsive breakpoints
- User flows (happy path + errors)

**Q: What if implementation doesn't match wireframe?**
A: Options:
1. Update wireframe to match implementation (if better)
2. Update implementation to match wireframe (if intentional)
3. Discuss with product designer before deviating

**Q: How do I find component specs?**
A: Check:
1. Feature-specific design system doc (e.g., configs-design-system.md)
2. Main design system reference (if it exists)
3. shadcn/ui documentation
4. Ask product designer

---

## Resources

### Design Tools
- [ASCII Tree Generator](https://tree.nathanfriend.io/) - Create ASCII diagrams
- [Mermaid Live Editor](https://mermaid.live/) - Test Mermaid diagrams
- [Contrast Checker](https://webaim.org/resources/contrastchecker/) - WCAG compliance

### Design Systems
- [shadcn/ui](https://ui.shadcn.com/) - Component library
- [Tailwind CSS](https://tailwindcss.com/) - Utility classes
- [Heroicons](https://heroicons.com/) - Icon library

### Accessibility
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [A11y Project Checklist](https://www.a11yproject.com/checklist/)
- [ARIA Authoring Practices](https://www.w3.org/WAI/ARIA/apg/)

### Wireframe Examples
- [Current: Agent Configs](./configs-overview.md) - Reference implementation

---

**Maintained By:** Product Designer Agent
**Last Updated:** 2025-01-15
**Questions?** Tag @product-designer in GitHub issues
