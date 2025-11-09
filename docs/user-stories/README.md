# Ubik Enterprise - User Stories

**Status:** In Progress
**Last Updated:** 2025-11-09
**Version:** 0.1.0

## Overview

This directory contains user stories for the Ubik Enterprise AI Agent Management Platform. Stories are organized by epic and reflect the **currently implemented** system, not future wireframes.

## How to Use

- Stories follow the format: "As a [role], I want to [action] so that [benefit]"
- Each story includes acceptance criteria in Given/When/Then format (Gherkin)
- Stories are linked to implemented features and API endpoints
- Priority: P0 (Critical), P1 (High), P2 (Medium), P3 (Low)
- Status: ✅ Implemented, 🚧 In Progress, 📋 Planned

## Quick Navigation

### Epic 1: Authentication & Onboarding
- [1.1 User Registration](./epic-1-authentication/1.1-user-registration.md) ✅
- [1.2 User Login](./epic-1-authentication/1.2-user-login.md) 🚧
- [1.3 User Logout](./epic-1-authentication/1.3-user-logout.md) ✅
- [1.4 Password Reset](./epic-1-authentication/1.4-password-reset.md) 📋

### Epic 2: Dashboard & Navigation
- [2.1 Admin Dashboard](./epic-2-dashboard/2.1-admin-dashboard.md) 🚧

### Epic 3: Organization Management
Coming soon...

### Epic 4: Employee Management
Coming soon...

### Epic 5: Team Management
Coming soon...

### Epic 6: Agent Configuration
Coming soon...

### Epic 7: MCP Server Management
Coming soon...

### Epic 8: Approvals & Requests
Coming soon...

### Epic 9: Usage & Analytics
Coming soon...

### Epic 10: CLI Integration
Coming soon...

## Story Template

See individual epic directories for examples. Each story should include:

```markdown
# Story X.Y: Title

**As a** [role]
**I want to** [action]
**So that** [benefit]

**Priority:** P0/P1/P2/P3
**Status:** ✅ Implemented / 🚧 In Progress / 📋 Planned
**Endpoints:** API endpoints used
**UI:** UI pages/components

## Acceptance Criteria

\`\`\`gherkin
Given [context]
When [action]
Then [expected result]
\`\`\`

## Implementation Notes
- Technical details
- Edge cases
- Security considerations
```

## Contributing

When adding new stories:
1. Create file in appropriate epic directory
2. Follow naming convention: `{epic}.{story}-{slug}.md`
3. Update this README with link to new story
4. Update epic README if needed
5. Ensure acceptance criteria are precise and testable
