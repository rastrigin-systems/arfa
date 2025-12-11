# Teams & Employees UI Implementation Plan

## Overview

Building a hierarchical Teams and Employees management interface for Ubik Enterprise admin panel.

## Phase 1: Foundation - Navigation & Sidebar

### Goals
- Add proper navigation with sidebar for organization hierarchy
- Create a sidebar component with Teams/Employees sections
- Update dashboard layout to accommodate sidebar

### Components to Create
1. `components/sidebar/Sidebar.tsx` - Main sidebar component
2. `components/sidebar/TeamsList.tsx` - Collapsible teams list with employee counts
3. `components/sidebar/NavItem.tsx` - Reusable navigation item

### Navigation Structure
```
Dashboard
├── Teams (collapsible)
│   ├── Engineering (5)
│   ├── Design (3)
│   └── + New Team
├── Employees
├── Agents
├── Activity Logs
└── Settings
```

## Phase 2: Teams List Page

### Goals
- Display all teams with member count
- Allow creating new teams
- Click team to see details

### Routes
- `/teams` - Teams list
- `/teams/new` - Create team form
- `/teams/[id]` - Team detail with members

### Components
1. `app/(dashboard)/teams/page.tsx` - Teams list page
2. `components/teams/TeamCard.tsx` - Team card with stats
3. `components/teams/TeamForm.tsx` - Create/edit team form

## Phase 3: Team Detail with Employees

### Goals
- Show team details with member list
- Add/remove employees from team
- Link to employee details

### Components
1. `app/(dashboard)/teams/[id]/page.tsx` - Team detail
2. `components/teams/TeamMemberList.tsx` - Member list for team

## Phase 4: Enhanced Employees Page

### Goals
- Show employees with team grouping
- Filter by team
- Quick actions (edit, assign team)

### Enhancements
1. Add team filter to existing EmployeeTable
2. Add "Assign to Team" action
3. Add bulk actions support

## API Endpoints Used

### Teams
- `GET /api/v1/teams` - List all teams
- `POST /api/v1/teams` - Create team
- `GET /api/v1/teams/{id}` - Get team details
- `PATCH /api/v1/teams/{id}` - Update team
- `DELETE /api/v1/teams/{id}` - Delete team

### Employees
- `GET /api/v1/employees` - List employees (has team_id, team_name)
- `PATCH /api/v1/employees/{id}` - Update employee (assign team)

## Implementation Order

1. **Step 1**: Create sidebar with navigation
2. **Step 2**: Add Teams list page
3. **Step 3**: Add Team detail page
4. **Step 4**: Enhance employees list with team filter
5. **Step 5**: Add team assignment to employees

## File Structure

```
services/web/
├── app/(dashboard)/
│   ├── layout.tsx          # Update with sidebar
│   ├── teams/
│   │   ├── page.tsx        # Teams list
│   │   ├── new/
│   │   │   └── page.tsx    # Create team
│   │   └── [id]/
│   │       ├── page.tsx    # Team detail
│   │       └── edit/
│   │           └── page.tsx # Edit team
│   └── employees/          # Existing - enhance
├── components/
│   ├── layout/
│   │   └── Sidebar.tsx     # New sidebar
│   └── teams/
│       ├── TeamCard.tsx    # Team card component
│       ├── TeamForm.tsx    # Create/edit form
│       └── TeamMemberList.tsx # Member list
└── lib/
    └── api/
        └── teams.ts        # Team API functions
```

## Progress Tracking

- [ ] Step 1: Sidebar navigation
- [ ] Step 2: Teams list page
- [ ] Step 3: Team detail page
- [ ] Step 4: Enhanced employees with team filter
- [ ] Step 5: Team assignment feature
