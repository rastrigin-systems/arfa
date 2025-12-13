# Employee Detail Page - Configuration Hierarchy Display

## Purpose

Replace the simple "Agents" card on the Employee Detail page with a comprehensive configuration hierarchy display that shows how agent configurations are inherited and merged across organization, team, and employee levels.

## User Stories

### Primary User: Admin

**As an admin, I want to:**

1. **Understand configuration inheritance**
   - See organization-level base configurations
   - See team-level overrides that apply to this employee
   - See employee-level personal overrides
   - Understand the final resolved configuration

2. **Troubleshoot configuration issues**
   - Quickly identify which level a setting comes from
   - See why an employee has certain settings
   - Verify that expected configurations are applied
   - Debug conflicts or unexpected behavior

3. **Make informed decisions**
   - Decide whether to override at employee level or modify team/org config
   - Understand the impact of changing configs at different levels
   - See the complete configuration picture in one view

## Current vs New Design

### Current Design (Limited)
```
[Card: Agents]
├── Employee-level configs only
├── No context on inheritance
└── No visibility into org/team configs
```

### New Design (Comprehensive)
```
[Card: Configuration Hierarchy]
├── Organization Configs (base layer)
├── Team Configs (overrides)
├── Employee Configs (personal overrides)
└── Resolved Configs (final merged result)
```

---

## Desktop Layout (1024px+)

### Full Page View

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Ubik Enterprise                                                      [Search] [John Doe ▾] [Menu]   │
├─────────────────────────────────────────────────────────────────────────────────────────────────────┤
│  Sidebar Nav                                                                                          │
│  ├─ Dashboard                                                                                         │
│  ├─ Agents                 ┌──────────────────────────────────────────────────────────────────────┐ │
│  ├─ Configurations         │  Employees / John Doe                                                │ │
│  ├─ Teams                  │                                                                       │ │
│  └─ Employees ◄ ACTIVE     │  Employee Details                                [Edit] [Delete]     │ │
│                            │                                                                       │ │
│                            │  ┌─────────────────────────────────────────────────────────────────┐ │ │
│                            │  │ Basic Information                                               │ │ │
│                            │  │ Name: John Doe                Email: john@acme.com              │ │ │
│                            │  │ Status: Active                Team: Engineering                  │ │ │
│                            │  └─────────────────────────────────────────────────────────────────┘ │ │
│                            │                                                                       │ │
│                            │  ┌─────────────────────────────────────────────────────────────────┐ │ │
│                            │  │ Configuration Hierarchy                                         │ │ │
│                            │  │ View how agent configurations are inherited and merged          │ │ │
│                            │  │                                                                  │ │ │
│                            │  │ ┌───────────────────────────────────────────────────────────┐  │ │ │
│                            │  │ │ Organization Configs (Base)                   2 configured │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ Claude Code                                    [Enabled]   │  │ │ │
│                            │  │ │ model: claude-opus-4.5                                     │  │ │ │
│                            │  │ │ temperature: 0.7, max_tokens: 4096                         │  │ │ │
│                            │  │ │                                          [View Full Config] │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ GitHub Copilot                                 [Enabled]   │  │ │ │
│                            │  │ │ model: gpt-4                                               │  │ │ │
│                            │  │ │ auto_complete: true                                        │  │ │ │
│                            │  │ │                                          [View Full Config] │  │ │ │
│                            │  │ └───────────────────────────────────────────────────────────┘  │ │ │
│                            │  │                                                                  │ │ │
│                            │  │ ┌───────────────────────────────────────────────────────────┐  │ │ │
│                            │  │ │ Team Configs (Engineering Team Overrides)  1 configured   │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ Claude Code                                    [Enabled]   │  │ │ │
│                            │  │ │ Overrides: temperature: 0.9 ⬆                              │  │ │ │
│                            │  │ │ (Higher temperature for creative coding)                   │  │ │ │
│                            │  │ │                                          [View Full Config] │  │ │ │
│                            │  │ └───────────────────────────────────────────────────────────┘  │ │ │
│                            │  │                                                                  │ │ │
│                            │  │ ┌───────────────────────────────────────────────────────────┐  │ │ │
│                            │  │ │ Employee Configs (John's Overrides)        1 configured   │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ Claude Code                                    [Enabled]   │  │ │ │
│                            │  │ │ Overrides: max_tokens: 8192 ⬆                              │  │ │ │
│                            │  │ │ Reason: "Working on large refactoring tasks"               │  │ │ │
│                            │  │ │ Last synced: 2025-01-15 14:30 UTC                          │  │ │ │
│                            │  │ │                                [View] [Edit] [Remove]      │  │ │ │
│                            │  │ └───────────────────────────────────────────────────────────┘  │ │ │
│                            │  │                                                                  │ │ │
│                            │  │ ┌───────────────────────────────────────────────────────────┐  │ │ │
│                            │  │ │ Resolved Configs (Final Merged)        2 agents available │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ Claude Code                                    [Enabled]   │  │ │ │
│                            │  │ │ model: claude-opus-4.5 (org)                               │  │ │ │
│                            │  │ │ temperature: 0.9 (team override)                           │  │ │ │
│                            │  │ │ max_tokens: 8192 (employee override)                       │  │ │ │
│                            │  │ │                                          [View Full Config] │  │ │ │
│                            │  │ ├───────────────────────────────────────────────────────────┤  │ │ │
│                            │  │ │ GitHub Copilot                                 [Enabled]   │  │ │ │
│                            │  │ │ model: gpt-4 (org)                                         │  │ │ │
│                            │  │ │ auto_complete: true (org)                                  │  │ │ │
│                            │  │ │                                          [View Full Config] │  │ │ │
│                            │  │ └───────────────────────────────────────────────────────────┘  │ │ │
│                            │  └─────────────────────────────────────────────────────────────────┘ │ │
│                            │                                                                       │ │
│                            │  ┌─────────────────────────────────────────────────────────────────┐ │ │
│                            │  │ MCP Servers                                                     │ │ │
│                            │  │ No MCP servers configured                                       │ │ │
│                            │  └─────────────────────────────────────────────────────────────────┘ │ │
└─────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## Component Specifications

### Configuration Hierarchy Card

**Container:**
- Background: White (#FFFFFF)
- Border: 1px solid #E5E7EB
- Border radius: 8px
- Padding: 24px
- Shadow: 0 1px 3px rgba(0,0,0,0.1)

**Card Header:**
- Title: "Configuration Hierarchy" (H2, 1.5rem, font-semibold)
- Description: "View how agent configurations are inherited and merged" (0.875rem, text-muted-foreground)
- Margin bottom: 24px

### Section Structure

Each of the 4 sections follows this pattern:

```
┌─────────────────────────────────────────────────────────────┐
│ Section Title                            {count} configured │
├─────────────────────────────────────────────────────────────┤
│ [Agent cards...]                                            │
└─────────────────────────────────────────────────────────────┘
```

**Section Header:**
- Background: #F9FAFB
- Border: 1px solid #E5E7EB
- Padding: 12px 16px
- Font: Inter Medium 14px
- Color: #111827
- Count badge: #6B7280, 12px

**Section Types:**

1. **Organization Configs (Base)**
   - Icon: 🏢 or organization icon
   - Shows: Full base configuration
   - Read-only display

2. **Team Configs (Overrides)**
   - Icon: 👥 or team icon
   - Shows: Only overridden fields
   - Team name in header: "Team Configs ({team_name} Overrides)"
   - Indicates which org values are overridden

3. **Employee Configs (Overrides)**
   - Icon: 👤 or user icon
   - Shows: Only employee-specific overrides
   - Includes override reason and last sync
   - Editable: [Edit] [Remove] buttons

4. **Resolved Configs (Final Merged)**
   - Icon: ✓ or check icon
   - Shows: Complete merged configuration
   - Source annotation for each field (org/team/employee)
   - This is what the employee actually receives

---

## Agent Card Anatomy

### Organization Config Card
```
┌───────────────────────────────────────────────────────────┐
│ Claude Code                                    [Enabled]   │
│ model: claude-opus-4.5                                     │
│ temperature: 0.7, max_tokens: 4096                         │
│                                          [View Full Config] │
└───────────────────────────────────────────────────────────┘
```

**Structure:**
- Agent name: H3 (1.125rem, font-semibold)
- Status badge: Right-aligned, 28px height
  - Enabled: Green bg #D1FAE5, text #065F46
  - Disabled: Gray bg #F3F4F6, text #6B7280
- Config preview: First 2-3 key-value pairs, mono font (Inter Mono 12px)
- Action button: Secondary button (outlined), right-aligned

---

### Team Config Card (Override)
```
┌───────────────────────────────────────────────────────────┐
│ Claude Code                                    [Enabled]   │
│ Overrides: temperature: 0.9 ⬆                              │
│ (Higher temperature for creative coding)                   │
│                                          [View Full Config] │
└───────────────────────────────────────────────────────────┘
```

**Structure:**
- "Overrides:" prefix in bold
- Override values with ⬆ indicator (shows it's overriding org value)
- Optional explanation in parentheses (muted color)
- Only shows fields that differ from org config

---

### Employee Config Card (Override)
```
┌───────────────────────────────────────────────────────────┐
│ Claude Code                                    [Enabled]   │
│ Overrides: max_tokens: 8192 ⬆                              │
│ Reason: "Working on large refactoring tasks"               │
│ Last synced: 2025-01-15 14:30 UTC                          │
│                                [View] [Edit] [Remove]      │
└───────────────────────────────────────────────────────────┘
```

**Structure:**
- Override fields with ⬆ indicator
- Override reason (if provided)
- Last sync timestamp (muted)
- Action buttons: [View] [Edit] [Remove]
  - View: Opens full config modal (read-only)
  - Edit: Opens edit modal
  - Remove: Confirmation dialog, deletes employee override

---

### Resolved Config Card (Final)
```
┌───────────────────────────────────────────────────────────┐
│ Claude Code                                    [Enabled]   │
│ model: claude-opus-4.5 (org)                               │
│ temperature: 0.9 (team override)                           │
│ max_tokens: 8192 (employee override)                       │
│                                          [View Full Config] │
└───────────────────────────────────────────────────────────┘
```

**Structure:**
- All fields shown with source annotation
- Source labels in parentheses:
  - (org) - Gray #6B7280
  - (team override) - Blue #3B82F6
  - (employee override) - Purple #7C3AED
- Shows complete merged configuration
- This is what CLI/agent receives

---

## State Variations

### Empty States

#### No Organization Configs
```
┌───────────────────────────────────────────────────────────┐
│ Organization Configs (Base)                  0 configured │
├───────────────────────────────────────────────────────────┤
│                                                            │
│              No organization-level agent configs           │
│                                                            │
│    Contact your admin to enable agents for this org       │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

#### No Team Configs (Employee not on team)
```
┌───────────────────────────────────────────────────────────┐
│ Team Configs                                 0 configured │
├───────────────────────────────────────────────────────────┤
│                                                            │
│              Employee not assigned to a team               │
│                                                            │
│         Only organization configs will be applied          │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

#### No Team Configs (Team has none)
```
┌───────────────────────────────────────────────────────────┐
│ Team Configs (Engineering Team)              0 configured │
├───────────────────────────────────────────────────────────┤
│                                                            │
│          No team-level overrides for this team             │
│                                                            │
│    Using organization defaults for all agents              │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

#### No Employee Overrides
```
┌───────────────────────────────────────────────────────────┐
│ Employee Configs (John's Overrides)         0 configured │
├───────────────────────────────────────────────────────────┤
│                                                            │
│        No personal overrides for this employee             │
│                                                            │
│                      [+ Add Override]                      │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

---

### Loading State
```
┌───────────────────────────────────────────────────────────┐
│ Configuration Hierarchy                                   │
│                                                            │
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                     │
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                              │
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                          │
│                                                            │
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                        │
│ ▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓                              │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

- Skeleton loaders for 4 sections
- Shimmer animation
- No data displayed until loaded

---

### Error State
```
┌───────────────────────────────────────────────────────────┐
│ Configuration Hierarchy                                   │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ⚠️ Failed to load configuration hierarchy                 │
│                                                            │
│  We couldn't retrieve the agent configurations.           │
│  Please try again.                                         │
│                                                            │
│                        [Retry]                             │
│                                                            │
└───────────────────────────────────────────────────────────┘
```

---

## Expanded Config Modal

When user clicks [View Full Config], show full JSON in modal:

```
┌─────────────────────────────────────────────────────────────────┐
│  Claude Code Configuration (Organization Level)           [✕]   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Status: [Enabled]                                              │
│                                                                  │
│  Full Configuration:                                             │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │ {                                                          │ │
│  │   "model": "claude-opus-4.5",                              │ │
│  │   "temperature": 0.7,                                      │ │
│  │   "max_tokens": 4096,                                      │ │
│  │   "system_prompt": "You are an expert developer...",      │ │
│  │   "tools": ["filesystem", "bash", "browser"],             │ │
│  │   "safety_mode": "standard"                               │ │
│  │ }                                                          │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  Created: 2025-01-10 09:00 UTC                                  │
│  Last updated: 2025-01-12 16:45 UTC                             │
│                                                                  │
│                                                   [Close]        │
└─────────────────────────────────────────────────────────────────┘
```

**Modal specs:**
- Width: 640px
- Max-height: 80vh
- Syntax-highlighted JSON (code editor)
- Read-only for org/team configs
- Editable for employee overrides

---

## Interactions

### 1. View Full Config
```
User clicks [View Full Config] on any agent card
↓
Modal opens with complete JSON configuration
↓
Shows all fields (not just preview)
↓
User clicks [Close] or [✕] to dismiss
```

### 2. Edit Employee Override
```
User clicks [Edit] on employee config card
↓
Modal opens with current override values pre-filled
↓
User modifies configuration
↓
User clicks [Save Changes]
↓
API PATCH /employees/{id}/agent-configs/{config_id}
↓
Success: Modal closes, card updates
↓
Toast notification: "✓ Configuration updated"
```

### 3. Remove Employee Override
```
User clicks [Remove] on employee config card
↓
Confirmation dialog appears:
┌─────────────────────────────────────────────────┐
│ Remove Employee Override?                       │
│                                                 │
│ This will remove John's personal override for  │
│ Claude Code. Organization and team configs     │
│ will still apply.                               │
│                                                 │
│              [Cancel]  [Remove Override]        │
└─────────────────────────────────────────────────┘
↓
User clicks [Remove Override]
↓
API DELETE /employees/{id}/agent-configs/{config_id}
↓
Success: Card disappears, resolved config updates
↓
Toast notification: "✓ Override removed"
```

### 4. Add Employee Override
```
User clicks [+ Add Override] in empty employee section
↓
Modal opens: Create Employee Agent Config
↓
Step 1: Select agent from dropdown
↓
Step 2: Configure override values (only override what's needed)
↓
User clicks [Create Override]
↓
API POST /employees/{id}/agent-configs
↓
Success: New card appears in employee section
↓
Resolved configs section updates
↓
Toast notification: "✓ Override created"
```

---

## Responsive Behavior

### Desktop (1024px+)
- All 4 sections stacked vertically
- Full width available for config details
- Side-by-side buttons in action row

### Tablet (768px - 1023px)
- Same vertical stack
- Slightly reduced padding
- Config preview may wrap to more lines

### Mobile (320px - 767px)
- Full-screen modals
- Stacked buttons (vertical)
- Collapsible sections (accordion pattern):
  ```
  [▾] Organization Configs (2 configured)
  [▸] Team Configs (1 configured)
  [▸] Employee Configs (1 configured)
  [▾] Resolved Configs (2 agents)
  ```
- Only one section expanded at a time
- Tap header to expand/collapse

---

## Accessibility

### Keyboard Navigation
- **Tab:** Move between sections and cards
- **Enter/Space:** Expand config card or activate button
- **Escape:** Close modal/dialog
- **Arrow keys:** Navigate within config JSON in modal

### ARIA Labels
```html
<section aria-label="Organization agent configurations">
  <h3 id="org-configs-heading">Organization Configs (Base)</h3>
  <div aria-labelledby="org-configs-heading">
    <article
      role="article"
      aria-label="Claude Code configuration - Enabled"
    >
      ...
    </article>
  </div>
</section>

<button
  aria-label="View full configuration for Claude Code at organization level"
>
  View Full Config
</button>

<button
  aria-label="Edit employee override for Claude Code"
>
  Edit
</button>
```

### Screen Reader Announcements
- "Configuration hierarchy loaded. 2 organization configs, 1 team config, 1 employee override."
- "Viewing Claude Code organization configuration"
- "Employee override removed. Resolved configuration updated."
- "Override created for Claude Code. Merged configuration now includes employee settings."

### Focus Management
- Focus trap in modals
- Return focus to trigger button after modal close
- Visible focus ring: 2px solid #3B82F6, 2px offset

### Color Contrast
- All text: WCAG AA compliant (4.5:1 minimum)
- Source labels: 4.5:1 contrast ratio
- Status badges: 4.8:1 (enabled), 4.6:1 (disabled)
- Override indicators (⬆): Color + shape (not color alone)

---

## API Integration

### Endpoints Required

```
GET /employees/{employee_id}/agent-configs/hierarchy
```

**Response Schema:**
```json
{
  "organization_configs": [
    {
      "id": "cfg_org_123",
      "agent_id": "agent_abc",
      "agent_name": "Claude Code",
      "agent_type": "claude-code",
      "agent_provider": "anthropic",
      "config": {
        "model": "claude-opus-4.5",
        "temperature": 0.7,
        "max_tokens": 4096
      },
      "is_enabled": true,
      "created_at": "2025-01-10T09:00:00Z",
      "updated_at": "2025-01-12T16:45:00Z"
    }
  ],
  "team_configs": [
    {
      "id": "cfg_team_456",
      "team_id": "team_xyz",
      "team_name": "Engineering",
      "agent_id": "agent_abc",
      "agent_name": "Claude Code",
      "config_override": {
        "temperature": 0.9
      },
      "override_reason": "Higher temperature for creative coding",
      "is_enabled": true,
      "created_at": "2025-01-11T10:00:00Z",
      "updated_at": "2025-01-11T10:00:00Z"
    }
  ],
  "employee_configs": [
    {
      "id": "cfg_emp_789",
      "employee_id": "emp_john",
      "agent_id": "agent_abc",
      "agent_name": "Claude Code",
      "config_override": {
        "max_tokens": 8192
      },
      "override_reason": "Working on large refactoring tasks",
      "is_enabled": true,
      "sync_token": "tok_xyz789",
      "last_synced_at": "2025-01-15T14:30:00Z",
      "created_at": "2025-01-13T08:00:00Z",
      "updated_at": "2025-01-13T08:00:00Z"
    }
  ],
  "resolved_configs": [
    {
      "agent_id": "agent_abc",
      "agent_name": "Claude Code",
      "agent_type": "claude-code",
      "provider": "anthropic",
      "config": {
        "model": "claude-opus-4.5",
        "temperature": 0.9,
        "max_tokens": 8192
      },
      "config_sources": {
        "model": "organization",
        "temperature": "team",
        "max_tokens": "employee"
      },
      "system_prompt": "Combined prompts from org + team + employee",
      "is_enabled": true
    }
  ]
}
```

**Existing Endpoints to Use:**
```
GET /organizations/current/agent-configs
GET /teams/{team_id}/agent-configs
GET /employees/{employee_id}/agent-configs
GET /employees/{employee_id}/agent-configs/resolved

POST /employees/{employee_id}/agent-configs
PATCH /employees/{employee_id}/agent-configs/{config_id}
DELETE /employees/{employee_id}/agent-configs/{config_id}
```

---

## Design System Usage

### Colors
- **Primary:** Blue #3B82F6 (source labels, links)
- **Success:** Green #10B981 (enabled badges)
- **Info:** Purple #7C3AED (employee override source)
- **Muted:** Gray #6B7280 (org source, timestamps)
- **Border:** Gray #E5E7EB
- **Background:** White #FFFFFF, Light #F9FAFB

### Typography
- **Section heading:** H3 1.125rem (18px) font-semibold
- **Agent name:** H4 1rem (16px) font-medium
- **Config preview:** Inter Mono 12px
- **Body text:** 0.875rem (14px)
- **Muted text:** 0.75rem (12px) text-muted-foreground

### Spacing
- Card padding: 16px
- Section spacing: 24px (space-y-6)
- Config item spacing: 8px (space-y-2)
- Button spacing: 8px gap

### Components (shadcn/ui)
- **Card:** CardHeader, CardContent, CardDescription
- **Badge:** Status indicators (Enabled/Disabled)
- **Button:** Primary, Secondary, Ghost variants
- **Dialog:** Full config modal
- **Alert Dialog:** Remove confirmation
- **Skeleton:** Loading state
- **Toast:** Success/error notifications
- **Code:** JSON display with syntax highlighting

---

## Implementation Notes

### Phase 1: Basic Display
1. Create ConfigurationHierarchy component
2. Fetch data from hierarchy endpoint
3. Display 4 sections with basic cards
4. Show config preview (first 3 fields)
5. Loading and empty states

### Phase 2: View Full Config
1. Implement modal with full JSON display
2. Syntax highlighting for JSON
3. Show metadata (created, updated timestamps)

### Phase 3: Employee Overrides
1. Add/Edit/Remove employee overrides
2. Form validation
3. API integration
4. Optimistic updates

### Phase 4: Polish
1. Source annotations in resolved configs
2. Override indicators (⬆)
3. Responsive mobile view (accordion)
4. Accessibility audit
5. Keyboard navigation

---

## Testing Strategy

### Unit Tests
- Component renders with mock data
- Empty states display correctly
- Override indicators show for team/employee configs
- Source annotations correct in resolved configs

### Integration Tests
- Fetch hierarchy data from API
- View full config modal
- Create employee override
- Edit employee override
- Remove employee override
- Error handling (API failures)

### E2E Tests (Playwright)
- Navigate to employee detail page
- Verify all 4 sections load
- Expand full config modal
- Create new employee override
- Edit existing override
- Remove override with confirmation

### Accessibility Tests
- Keyboard navigation through sections
- Screen reader announcements
- Color contrast ratios
- Focus management in modals
- ARIA attribute validation

---

## Files Created

- `/docs/wireframes/employee-config-hierarchy.md` (this file)

---

## Next Steps

1. **Review wireframes** with product-strategist - Validate user stories
2. **Review with tech-lead** - Validate API design and data flow
3. **Create GitHub issue** - Implementation ticket
4. **Hand off to frontend-developer** - Begin implementation

---

**Wireframes Status:** ✅ Complete - Ready for Review

**Estimated Implementation:** 1-2 sprints
- Sprint 1: Basic hierarchy display + view full config
- Sprint 2: Employee override management + polish
