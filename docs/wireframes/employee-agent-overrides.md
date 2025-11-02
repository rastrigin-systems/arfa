# Employee Agent Overrides Page Wireframe

## Overview
Page for viewing and managing employee-specific agent configuration overrides. Shows which agents are available to the employee (from org/team configs) and allows managers to create employee-specific overrides.

## Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ Dashboard > Employees > John Doe > Agent Overrides                  │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ╔═══════════════════════════════════════════════════════════╗     │
│  ║ 👤 John Doe - Agent Overrides                             ║     │
│  ╚═══════════════════════════════════════════════════════════╝     │
│                                                                       │
│  Manage agent configurations specific to this employee              │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────┐       │
│  │ [Inherited Configs] [Employee Overrides] [All Agents]    │       │
│  └─────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ Employee Agent Overrides (2)                    [+ Add Override]│ │
│  ├───────────────────────────────────────────────────────────────┤ │
│  │                                                                 │ │
│  │ ┌─────────────────────────────────────────────────────────┐   │ │
│  │ │ 🤖 Claude Code                              ✓ ENABLED    │   │ │
│  │ │ Employee-specific configuration override                 │   │ │
│  │ │                                                           │   │ │
│  │ │ Override Reason: Higher rate limit for senior engineer   │   │ │
│  │ │ Model: claude-sonnet-4.5                                 │   │ │
│  │ │ Rate Limit: 200 req/day (org default: 100)             │   │ │
│  │ │ Cost Limit: $100/month (org default: $50)              │   │ │
│  │ │ Last Updated: 2025-10-30 14:23 by manager@company.com  │   │ │
│  │ │                                                           │   │ │
│  │ │                [🔧 Edit] [❌ Disable] [🗑️ Remove]        │   │ │
│  │ └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                 │ │
│  │ ┌─────────────────────────────────────────────────────────┐   │ │
│  │ │ 🤖 Cursor                                   ⚠️ DISABLED   │   │ │
│  │ │ Employee-specific restriction                            │   │ │
│  │ │                                                           │   │ │
│  │ │ Override Reason: Security incident - temporary block     │   │ │
│  │ │ Status: Disabled (org config is enabled)                │   │ │
│  │ │ Last Updated: 2025-10-28 09:15 by security@company.com │   │ │
│  │ │                                                           │   │ │
│  │ │                [🔧 Edit] [✓ Enable] [🗑️ Remove]         │   │ │
│  │ └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                 │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

## Tab Structure

### Tab 1: Inherited Configs (Default View)
- Shows all agents available to employee through org/team configs
- Read-only view of inherited configurations
- Each card shows:
  - Agent name and icon
  - Source of config (Organization/Team)
  - Status (Enabled/Disabled)
  - Key config fields (model, rate limits)
  - "Create Override" button
- Helps managers understand what employee has access to
- Example:
  ```
  🤖 Claude Code                        ✓ ENABLED
  From: Organization Config

  Model: claude-sonnet-4.5
  Rate Limit: 100 req/day
  Cost Limit: $50/month

                              [Create Override →]
  ```

### Tab 2: Employee Overrides (Management View)
- Shows employee-specific overrides (`employee_agent_configs` table)
- List view of override configurations
- Each row shows:
  - Agent name and icon
  - Override reason/note
  - Status badge (Enabled/Disabled/Override)
  - Key override fields highlighted
  - Comparison with org/team defaults
  - Last updated timestamp and by whom
  - Action buttons (Edit, Enable/Disable, Remove)
- Empty state if no overrides:
  ```
  No employee overrides configured
  Employee is using default organization/team configurations
  [Browse Inherited Configs →]
  ```

### Tab 3: All Agents
- Shows all available agents in catalog
- Indicates which are configured (inherited or override)
- Quick access to create new overrides
- Similar to "Available Agents" in org config page

## Override Card (Expanded)

```
┌─────────────────────────────────────────────────────────────┐
│ 🤖 Claude Code                                 ✓ OVERRIDE    │
│ Employee-specific configuration with higher limits           │
├─────────────────────────────────────────────────────────────┤
│ Override Configuration                                        │
│                                                               │
│ ✏️ Override Reason:                                          │
│    "Senior engineer needs higher limits for large projects"  │
│                                                               │
│ Model:         claude-sonnet-4.5 (same as org)              │
│ Rate Limit:    200 req/day ↑ (org: 100 req/day)            │
│ Cost Limit:    $100/month ↑ (org: $50/month)               │
│ Path Restrict: /work/* (org: /project/*)                    │
│                                                               │
│ Config Source: Employee Override                             │
│ Org Config:    Enabled, uses claude-sonnet-4.5              │
│ Team Config:   No team override                              │
│                                                               │
│ Last Updated:  2025-10-30 14:23                              │
│ Updated By:    manager@company.com (Jane Smith)              │
│                                                               │
│ Audit Trail:                                                  │
│ • Created: 2025-10-20 09:00 by admin@company.com            │
│ • Updated: 2025-10-30 14:23 by manager@company.com          │
│   Reason: "Increased rate limit for Q4 sprint"              │
│                                                               │
│                [🔧 Edit] [❌ Disable] [🗑️ Remove]           │
└─────────────────────────────────────────────────────────────┘
```

## Empty States

### No Overrides
```
┌───────────────────────────────────────────┐
│                                           │
│     📋 No employee overrides              │
│                                           │
│  This employee is using default           │
│  organization and team configurations     │
│                                           │
│        [View Inherited Configs →]         │
│                                           │
└───────────────────────────────────────────┘
```

### No Inherited Configs
```
┌───────────────────────────────────────────┐
│                                           │
│     ⚠️ No agents configured               │
│                                           │
│  No agents are configured at the          │
│  organization or team level               │
│                                           │
│     [Configure Organization Agents →]     │
│                                           │
└───────────────────────────────────────────┘
```

## Interactions

### Create Override
1. Navigate to "Inherited Configs" tab
2. Find agent to override (e.g., Claude Code)
3. Click "Create Override" button
4. Modal opens (similar to agent config modal)
5. Form shows:
   - Agent name (read-only)
   - Override reason (required, text area)
   - Config fields (pre-filled with org/team defaults)
   - Visual indicators showing changes from default
6. Modify fields as needed
7. Click "Create Override"
8. Override appears in "Employee Overrides" tab
9. Success toast: "Override created for Claude Code"

### Edit Override
1. Navigate to "Employee Overrides" tab
2. Find override to edit
3. Click "Edit" button
4. Modal opens with existing override config
5. Modify fields
6. Add update reason (optional, auto-logged)
7. Click "Save Changes"
8. Config updated in list
9. Success toast: "Override updated"

### Disable Override
1. Click "Disable" button on override
2. Confirmation modal:
   ```
   Disable Claude Code override for John Doe?

   Employee will fall back to organization/team configuration.

   [Cancel] [Disable Override]
   ```
3. Click "Disable Override"
4. Status changes to "Disabled"
5. Employee uses org/team config

### Remove Override
1. Click "Remove" button
2. Confirmation modal:
   ```
   Remove Claude Code override?

   This will delete the employee-specific configuration.
   Employee will use default organization/team settings.

   Override Details:
   • Rate Limit: 200 req/day → will revert to 100 req/day
   • Cost Limit: $100/month → will revert to $50/month

   [Cancel] [Remove Override]
   ```
3. Click "Remove Override"
4. Override removed from list
5. Employee falls back to org/team config

## Comparison View (Key Feature)

When viewing/editing overrides, show clear comparison with defaults:

```
┌───────────────────────────────────────────────────────────┐
│ Configuration Comparison                                   │
├───────────────────────────────────────────────────────────┤
│                                                             │
│ Field          │ Organization │ Team        │ Employee     │
│ ─────────────  │ ──────────── │ ─────────── │ ──────────── │
│ Model          │ sonnet-4.5   │ (inherit)   │ sonnet-4.5   │
│ Rate Limit     │ 100 req/day  │ (inherit)   │ 200 req/day ↑│
│ Cost Limit     │ $50/month    │ $75/month ↑ │ $100/month ↑ │
│ Path Restrict  │ /project/*   │ (inherit)   │ /work/*      │
│                                                             │
│ Legend: ↑ Increased ↓ Decreased (inherit) Using parent     │
└───────────────────────────────────────────────────────────┘
```

## Permissions & Validation

### Who Can Access
- Employees with role permissions: `manage_employees = true` AND `manage_agents = true`
- Typically: Admins, Team Managers
- Employees cannot modify their own overrides (security)

### Validation Rules
- Override reason required (1-500 characters)
- Config must be valid JSON
- Rate/cost limits must be positive numbers
- Cannot create override for non-configured org agent
- Cannot lower security restrictions (only increase)

## API Endpoints

### List Employee Agent Configs (with inheritance)
```
GET /api/v1/employees/{employee_id}/agent-configs/resolved
Response: [
  {
    "id": "uuid",
    "agent_id": "uuid",
    "agent": { ... },
    "config": { ... },
    "is_enabled": true,
    "source": "organization" | "team" | "employee",
    "org_config": { ... },
    "team_config": { ... },
    "employee_override": { ... } | null
  }
]
```

### List Employee Overrides Only
```
GET /api/v1/employees/{employee_id}/agent-configs?overrides_only=true
Response: [
  {
    "id": "uuid",
    "employee_id": "uuid",
    "agent_id": "uuid",
    "config_override": { ... },
    "override_reason": "...",
    "is_enabled": true,
    "created_by": "uuid",
    "updated_at": "2025-10-30T14:23:00Z"
  }
]
```

### Create Employee Override
```
POST /api/v1/employees/{employee_id}/agent-configs
Body: {
  "agent_id": "uuid",
  "config_override": { "rate_limit": 200, "cost_limit": 100 },
  "override_reason": "Senior engineer needs higher limits",
  "is_enabled": true
}
```

### Update Employee Override
```
PATCH /api/v1/employees/{employee_id}/agent-configs/{config_id}
Body: {
  "config_override": { "rate_limit": 250 },
  "override_reason": "Increased for Q4 project",
  "is_enabled": true
}
```

### Delete Employee Override
```
DELETE /api/v1/employees/{employee_id}/agent-configs/{config_id}
```

## Responsive Behavior

### Desktop (>1024px)
- Full layout with side-by-side comparison
- 3-column comparison table
- Expanded override cards

### Tablet (768-1024px)
- Stacked comparison view
- 2-column table (org + employee)
- Collapsed override details

### Mobile (<768px)
- Single column layout
- Tabbed comparison (swipe between org/team/employee)
- Bottom sheet for modals
- Simplified override cards

## Accessibility

- **Keyboard Navigation**: Tab through configs, Enter to edit
- **Screen Readers**: Announce override status and changes
- **Focus Management**: Return focus after modal close
- **ARIA Labels**:
  - `role="region"` for comparison table
  - `aria-label="Employee override for Claude Code"` on cards
  - `aria-live="polite"` for status updates
  - `aria-describedby` for comparison indicators

## User Flows

### Flow 1: Create Override for Senior Engineer
1. Manager navigates to employee detail page
2. Clicks "Agent Overrides" tab
3. Views "Inherited Configs" tab
4. Sees Claude Code is enabled (org config)
5. Clicks "Create Override"
6. Modal opens with pre-filled org config
7. Enters override reason: "Senior engineer needs higher limits"
8. Increases rate limit: 100 → 200 req/day
9. Increases cost limit: $50 → $100/month
10. Clicks "Create Override"
11. Override appears in "Employee Overrides" tab
12. Employee can now use higher limits

### Flow 2: Temporarily Restrict Agent Access
1. Security team identifies suspicious activity
2. Manager navigates to employee page
3. Goes to "Employee Overrides" tab
4. Finds Cursor agent
5. Clicks "Disable"
6. Confirms: "Disable Cursor for security review"
7. Agent disabled immediately
8. Employee sees "Agent not available" in CLI
9. Security reviews activity
10. Manager re-enables after review complete

### Flow 3: Audit Employee Access
1. Manager reviews employee agent usage
2. Navigates to "Employee Overrides" tab
3. Reviews all overrides
4. Checks "Audit Trail" for each override
5. Sees who created/modified overrides and when
6. Reviews override reasons
7. Identifies unnecessary overrides
8. Removes outdated overrides
9. Employee falls back to org/team defaults

## Technical Notes

### Database Tables
- `employee_agent_configs` - Employee-specific overrides
- `org_agent_configs` - Organization defaults
- `team_agent_configs` - Team overrides
- `agent_catalog` - All available agents

### Config Resolution Order
1. Employee override (highest priority)
2. Team config
3. Organization config (fallback)

### State Management
- Fetch resolved configs on page load (includes inheritance)
- Cache org/team configs (rarely change)
- Optimistic updates for enable/disable
- Refetch after create/update/delete
- Compare view recalculates on config change

### Audit Trail
- Track all override creation/updates
- Store `created_by`, `updated_by` (employee_id)
- Store `override_reason` with each change
- Immutable audit log in `activity_logs` table

## Implementation Priority

**Phase 1: MVP**
- Employee Overrides tab (list view)
- Create override modal
- Enable/Disable/Remove actions
- Basic comparison with org config

**Phase 2: Enhanced**
- Inherited Configs tab (read-only)
- All Agents tab (catalog view)
- Edit override modal
- Full comparison table (org/team/employee)

**Phase 3: Polish**
- Audit trail display
- Usage statistics per employee
- Bulk override operations
- Export override configs
- Override templates

## Success Metrics

- Managers can create employee overrides in <2 minutes
- Clear visibility into config inheritance
- 100% WCAG AA accessibility compliance
- <200ms page load time
- Zero config conflicts (validation prevents)
