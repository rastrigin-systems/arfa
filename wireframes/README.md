# Ubik Enterprise - UI Wireframes

ASCII wireframes for the Ubik Enterprise AI Agent Management Platform web interface.

## Overview

These wireframes map directly to the API endpoints defined in `openapi/spec.yaml` and the database schema in `schema.sql`. Each page shows:
- Page layout and navigation
- UI elements and interactions
- Related API endpoints
- Data structures

## Wireframe Index

### Authentication & Profile
- **01-login.txt** - Login page
  - Endpoint: `POST /auth/login`

- **12-profile.txt** - User profile and personal settings
  - Endpoint: `GET /auth/me`

### Core Pages
- **02-dashboard.txt** - Main dashboard with overview statistics
  - Endpoints: `GET /auth/me`, `GET /employees`, `GET /teams`, `GET /organizations/current`

### Organization Management
- **07-organization-settings.txt** - Organization settings and configuration
  - Endpoints: `GET /organizations/current`, `PATCH /organizations/current`

### Employee Management
- **03-employees-list.txt** - List all employees with filters
  - Endpoint: `GET /employees`

- **04-employee-detail.txt** - View/edit individual employee
  - Endpoints: `GET /employees/{id}`, `PATCH /employees/{id}`, `DELETE /employees/{id}`

- **14-create-employee.txt** - Create new employee form
  - Endpoint: `POST /employees`

### Team Management
- **05-teams-list.txt** - List all teams
  - Endpoint: `GET /teams`

- **06-team-detail.txt** - View/edit team with members and agent configs
  - Endpoints: `GET /teams/{id}`, `PATCH /teams/{id}`, `GET /teams/{id}/agent-configs`

### Roles & Permissions
- **11-roles-list.txt** - View available roles and permissions
  - Endpoint: `GET /roles`

### Agent Management
- **08-agent-catalog.txt** - View available AI agents
  - Endpoint: `GET /agents`

- **09-org-agent-configs.txt** - Organization-level agent configurations
  - Endpoints: `GET /organizations/current/agent-configs`, `POST /organizations/current/agent-configs`

- **10-employee-agent-configs.txt** - Employee-level agent configuration overrides
  - Endpoints: `GET /employees/{id}/agent-configs`, `POST /employees/{id}/agent-configs`

- **13-team-agent-config-form.txt** - Create team-level agent configuration override
  - Endpoint: `POST /teams/{team_id}/agent-configs`

- **15-resolved-agent-configs.txt** - View fully resolved agent configs (org → team → employee merge)
  - Endpoint: `GET /employees/{id}/agent-configs/resolved`

## Configuration Hierarchy

The platform uses a 3-level configuration hierarchy:

```
Organization Level (base config)
    ↓
Team Level (overrides org config)
    ↓
Employee Level (overrides team/org config)
    ↓
Resolved Config (final merged configuration for CLI sync)
```

## Navigation Structure

```
Dashboard
├── Employees
│   ├── List Employees
│   ├── Create Employee
│   ├── Employee Detail
│   └── Employee Agent Configs
│       └── Resolved Configs
├── Teams
│   ├── List Teams
│   └── Team Detail
│       └── Team Agent Configs
├── Agents
│   ├── Agent Catalog
│   └── Organization Agent Configs
├── Settings
│   ├── Organization Settings
│   └── Roles
└── Profile
    └── My Profile
```

## Key Features Shown

### Multi-Tenancy
- Organization scoping on all pages
- Organization name in header: `[Acme Corp]`
- All data filtered by org_id

### Role-Based Access
- Different actions available based on role (Admin, Approver, Member)
- Edit/Delete buttons shown conditionally

### Agent Configuration
- Three levels: Organization, Team, Employee
- Override system with deep merge
- Resolved configs for CLI sync
- Last sync tracking

### Data Display Patterns
- List pages with pagination
- Detail pages with tabs/sections
- Forms with validation
- JSON editors for configuration
- Search and filter controls

## Design Principles

1. **Simplicity** - Clean ASCII layout, easy to understand
2. **Consistency** - Same header/nav on every page
3. **Context** - Show current org/user in header
4. **Actions** - Buttons clearly labeled with primary actions
5. **Documentation** - API endpoints shown at bottom of each page

## Implementation Notes

### Technology Stack (Planned)
- **Frontend**: Next.js 14, React, TypeScript
- **UI Library**: shadcn/ui, Tailwind CSS
- **API Client**: Generated from OpenAPI spec
- **State**: React Query for server state

### Authentication
- JWT tokens stored in httpOnly cookies
- Token refresh on API calls
- Automatic redirect to login on 401

### Data Fetching
- React Query for caching and loading states
- Optimistic updates for mutations
- Real-time updates via polling or WebSocket (future)

## Next Steps

1. Convert wireframes to React components
2. Implement API client from OpenAPI spec
3. Add form validation with Zod
4. Implement role-based route guards
5. Add real-time updates
6. Implement MCP configuration pages (not yet in wireframes)
7. Add approval workflow UI (not yet in wireframes)
8. Add usage analytics dashboard

## Related Documentation

- **API Spec**: `../openapi/spec.yaml`
- **Database Schema**: `../schema.sql`
- **ERD Diagram**: `../docs/ERD.md`
- **Project Documentation**: `../CLAUDE.md`
