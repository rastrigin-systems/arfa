# Organization Agent Configs Page Wireframe

## Overview
Page for viewing and managing organization-wide agent configurations. Shows which agents are configured and enabled for the organization, with ability to configure new agents or edit existing ones.

## Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ Dashboard > Agent Configuration                                      │
├─────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  ╔═══════════════════════════════════════════════════════════╗     │
│  ║ 🏢 Organization Agent Configuration                        ║     │
│  ╚═══════════════════════════════════════════════════════════╝     │
│                                                                       │
│  Manage AI agents available to your organization                    │
│                                                                       │
│  ┌─────────────────────────────────────────────────────────┐       │
│  │ [Available Agents] [Organization Configs] [Team Configs] │       │
│  └─────────────────────────────────────────────────────────┘       │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ Organization Configured Agents (3)                             │ │
│  ├───────────────────────────────────────────────────────────────┤ │
│  │                                                                 │ │
│  │ ┌─────────────────────────────────────────────────────────┐   │ │
│  │ │ 🤖 Claude Code                              ✓ ENABLED    │   │ │
│  │ │ AI coding assistant with full context                    │   │ │
│  │ │                                                           │   │ │
│  │ │ Model: claude-sonnet-4.5                                 │   │ │
│  │ │ API Key: sk-ant-api03-••••••••••••••                    │   │ │
│  │ │ Last Updated: 2025-10-30 14:23                          │   │ │
│  │ │                                                           │   │ │
│  │ │                [🔧 Edit] [❌ Disable] [🗑️ Remove]        │   │ │
│  │ └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                 │ │
│  │ ┌─────────────────────────────────────────────────────────┐   │ │
│  │ │ 🤖 Cursor                                    ✓ ENABLED    │   │ │
│  │ │ Intelligent code editor with AI pair programming         │   │ │
│  │ │                                                           │   │ │
│  │ │ Model: gpt-4-turbo                                       │   │ │
│  │ │ API Key: sk-••••••••••••••••                           │   │ │
│  │ │ Last Updated: 2025-10-29 10:15                          │   │ │
│  │ │                                                           │   │ │
│  │ │                [🔧 Edit] [❌ Disable] [🗑️ Remove]        │   │ │
│  │ └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                 │ │
│  │ ┌─────────────────────────────────────────────────────────┐   │ │
│  │ │ 🤖 Windsurf                                 ⚠️ DISABLED   │   │ │
│  │ │ Flow-state programming with AI collaboration             │   │ │
│  │ │                                                           │   │ │
│  │ │ Status: Temporarily disabled                             │   │ │
│  │ │ Last Updated: 2025-10-28 16:42                          │   │ │
│  │ │                                                           │   │ │
│  │ │                [🔧 Edit] [✓ Enable] [🗑️ Remove]         │   │ │
│  │ └─────────────────────────────────────────────────────────┘   │ │
│  │                                                                 │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                       │
│  ┌───────────────────────────────────────────────────────────────┐ │
│  │ Available Agents (10)                                          │ │
│  ├───────────────────────────────────────────────────────────────┤ │
│  │                                                                 │ │
│  │ 🔍 Search agents...                        🏷️ [All Categories ▾]│ │
│  │                                                                 │ │
│  │ ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │ │
│  │ │ 🤖 Cody      │  │ 🤖 Continue  │  │ 🤖 Supermaven│         │ │
│  │ │ Code AI      │  │ Autocomplete │  │ Fast AI      │         │ │
│  │ │ assistant    │  │              │  │              │         │ │
│  │ │              │  │              │  │              │         │ │
│  │ │ [Configure]  │  │ [Configure]  │  │ [Configure]  │         │ │
│  │ └──────────────┘  └──────────────┘  └──────────────┘         │ │
│  │                                                                 │ │
│  └───────────────────────────────────────────────────────────────┘ │
│                                                                       │
└─────────────────────────────────────────────────────────────────────┘
```

## Tab Structure

### Tab 1: Available Agents
- Shows all agents from `agent_catalog` table
- Grid/card view of available agents
- Each card shows:
  - Agent icon/logo
  - Agent name
  - Short description
  - Category badge
  - "Configure" button
- Click "Configure" → Opens agent configuration modal
- Search and filter by category

### Tab 2: Organization Configs (Current)
- Shows configured agents for current org (`org_agent_configs` table)
- List view of configured agents
- Each row shows:
  - Agent name and icon
  - Status badge (Enabled/Disabled)
  - Key configuration fields (model, API key masked)
  - Last updated timestamp
  - Action buttons (Edit, Enable/Disable, Remove)
- Empty state if no agents configured:
  ```
  No agents configured yet
  [Browse Available Agents →]
  ```

### Tab 3: Team Configs
- Shows team-specific overrides (`team_agent_configs` table)
- Select team from dropdown
- Shows inherited org configs with team overrides highlighted
- Similar layout to Organization Configs tab

## Agent Config Card (Expanded)

```
┌─────────────────────────────────────────────────────────────┐
│ 🤖 Claude Code                                 ✓ ENABLED     │
│ AI coding assistant with full context and debugging          │
├─────────────────────────────────────────────────────────────┤
│ Configuration                                                 │
│                                                               │
│ Model:         claude-sonnet-4.5                             │
│ API Key:       sk-ant-api03-••••••••••••••                  │
│ Max Tokens:    4096                                          │
│ Temperature:   0.7                                           │
│                                                               │
│ Enabled Tools: filesystem, git, http (3 of 12)              │
│                                                               │
│ Policies:      Rate Limit (100 req/day)                     │
│                Cost Limit ($50/month)                        │
│                Path Restrictions (/project/*)                │
│                                                               │
│ Last Updated:  2025-10-30 14:23 by admin@company.com        │
│                                                               │
│ Teams Using:   Engineering (45), Product (12), Design (8)   │
│                                                               │
│                [🔧 Edit] [❌ Disable] [🗑️ Remove]           │
└─────────────────────────────────────────────────────────────┘
```

## Empty States

### No Agents Configured
```
┌───────────────────────────────────────────┐
│                                           │
│          📦 No agents configured          │
│                                           │
│  Configure AI agents to make them         │
│  available to your organization           │
│                                           │
│        [Browse Available Agents]          │
│                                           │
└───────────────────────────────────────────┘
```

### Search No Results
```
┌───────────────────────────────────────────┐
│   🔍 No agents found matching "copilot"   │
│                                           │
│        Try different search terms         │
│                                           │
└───────────────────────────────────────────┘
```

## Interactions

### Configure Agent (New)
1. Click "Configure" button on available agent card
2. Modal opens (see `agent-configure-modal.md`)
3. Enter API key and config
4. Click "Save Configuration"
5. Agent appears in Organization Configs tab with "Enabled" status
6. Success toast: "Claude Code configured successfully"

### Edit Agent Config
1. Click "Edit" button on configured agent
2. Modal opens with existing config pre-filled
3. Modify fields
4. Click "Save Changes"
5. Config updated in list
6. Success toast: "Configuration updated"

### Disable Agent
1. Click "Disable" button
2. Confirmation modal: "Disable Claude Code for organization?"
3. Click "Disable"
4. Status changes to "Disabled"
5. Agent still visible but greyed out
6. Employees cannot use agent until re-enabled

### Remove Agent
1. Click "Remove" button
2. Confirmation modal:
   ```
   Remove Claude Code?

   This will delete the organization configuration.
   Teams and employees using this agent will lose access.

   45 employees across 3 teams are using this agent.

   [Cancel] [Remove Configuration]
   ```
3. Click "Remove Configuration"
4. Agent removed from list
5. Returns to Available Agents

### Enable Agent (Re-enable)
1. Click "Enable" button on disabled agent
2. Status changes to "Enabled" immediately
3. Success toast: "Claude Code enabled"

## Filter/Search

### Search Bar
- Real-time search across:
  - Agent name
  - Description
  - Category
- Debounced (300ms)

### Category Filter
- Dropdown with categories from `mcp_categories` table:
  - All Categories
  - Code Assistants
  - Productivity
  - Communication
  - Data/Analytics
  - DevOps

## Permissions & Validation

### Who Can Access
- Employees with role permissions: `manage_agents = true`
- Typically: Admins, IT Managers

### Validation Rules
- API key format validation (provider-specific)
- JSON config syntax validation
- Duplicate agent check (one config per agent per org)
- Required fields check

## API Endpoints

### List Organization Agent Configs
```
GET /api/v1/organizations/current/agent-configs
Response: [
  {
    "id": "uuid",
    "agent_id": "uuid",
    "agent": {
      "id": "uuid",
      "name": "Claude Code",
      "icon": "...",
      "description": "..."
    },
    "config": { "api_key": "sk-...", "model": "..." },
    "is_enabled": true,
    "updated_at": "2025-10-30T14:23:00Z"
  }
]
```

### Create Agent Config
```
POST /api/v1/organizations/current/agent-configs
Body: {
  "agent_id": "uuid",
  "config": { "api_key": "...", "model": "..." },
  "is_enabled": true
}
```

### Update Agent Config
```
PATCH /api/v1/organizations/current/agent-configs/{id}
Body: {
  "config": { "api_key": "..." },
  "is_enabled": false
}
```

### Delete Agent Config
```
DELETE /api/v1/organizations/current/agent-configs/{id}
```

### List Available Agents
```
GET /api/v1/agents?category=code-assistants&search=claude
Response: [
  {
    "id": "uuid",
    "name": "Claude Code",
    "description": "...",
    "category": "code-assistants",
    "icon": "...",
    "is_configured": true  // For current org
  }
]
```

## Responsive Behavior

### Desktop (>1024px)
- Full layout as shown above
- Cards in 3-column grid
- Side-by-side panels

### Tablet (768-1024px)
- 2-column grid for agent cards
- Stacked panels

### Mobile (<768px)
- Single column layout
- Collapsed filters (expandable)
- Full-width agent cards
- Bottom sheet for modals

## Accessibility

- **Keyboard Navigation**: Tab through cards, Enter to configure
- **Screen Readers**: Announce agent status changes
- **Focus Management**: Return focus to trigger after modal close
- **ARIA Labels**:
  - `role="tabpanel"` for tab content
  - `aria-label="Enable Claude Code"` on action buttons
  - `aria-live="polite"` for status updates

## User Flows

### Flow 1: Configure First Agent
1. Admin logs in
2. Navigates to Agent Configuration
3. Sees empty Organization Configs tab
4. Clicks "Browse Available Agents"
5. Switches to Available Agents tab
6. Searches for "Claude"
7. Clicks "Configure" on Claude Code
8. Modal opens
9. Enters API key
10. Clicks "Save Configuration"
11. Switches back to Organization Configs tab
12. Sees Claude Code listed as Enabled

### Flow 2: Disable Agent Temporarily
1. Admin views Organization Configs
2. Sees Claude Code is Enabled
3. Needs to disable due to budget constraints
4. Clicks "Disable"
5. Confirms in modal
6. Status changes to Disabled
7. Employees see "Agent not available" in CLI
8. Later, clicks "Enable" to restore

### Flow 3: Update API Key
1. API key rotated by security team
2. Admin clicks "Edit" on agent
3. Modal opens with existing config
4. Updates API key field
5. JSON editor auto-updates
6. Clicks "Save Changes"
7. New key takes effect immediately
8. Employees continue using agent without interruption

## Technical Notes

### Database Tables
- `agent_catalog` - All available agents
- `org_agent_configs` - Organization-level configs
- `team_agent_configs` - Team-level overrides
- `employee_agent_configs` - Employee-level instances

### State Management
- Fetch org configs on page load
- Cache agent catalog (rarely changes)
- Optimistic updates for enable/disable
- Refetch after create/update/delete

### Security
- API keys masked in UI (show last 4 chars)
- API keys encrypted at rest in database
- HTTPS required for all API calls
- JWT authentication for endpoints

## Implementation Priority

**Phase 1: MVP**
- Organization Configs tab (list view)
- Configure modal (basic fields)
- Enable/Disable/Remove actions
- Available Agents tab (simple grid)

**Phase 2: Enhanced**
- Edit config modal
- Search and filters
- Team Configs tab
- Advanced config fields

**Phase 3: Polish**
- Usage statistics per agent
- Agent recommendations
- Bulk operations
- Import/Export configs
