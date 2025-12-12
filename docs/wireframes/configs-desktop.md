# Configs Page - Desktop View (1440px)

## Default State (Populated)

```
┌─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┐
│  Ubik Enterprise                                                                               [Search] [John Doe ▾] [Menu]  │
├─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┤
│  Sidebar Nav                │  Agent Configurations                                                                         │
│  ┌──────────────────────┐  │  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ Dashboard            │  │  │  Filters:                                                                              │  │
│  │ Organizations        │  │  │  ┌─────────────┐  ┌──────────────┐  ┌─────────────┐  ┌──────────────────────────┐    │  │
│  │ Teams                │  │  │  │ Level: All ▾│  │ Agent: All ▾ │  │ Status: All │  │ Search configurations... │    │  │
│  │ Employees            │  │  │  └─────────────┘  └──────────────┘  └─────────────┘  └──────────────────────────┘    │  │
│  │ ► Configs            │  │  └───────────────────────────────────────────────────────────────────────────────────────┘  │
│  │ Policies             │  │                                                                                              │
│  │ Tools & MCPs         │  │  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │
│  │ Activity             │  │  │                                                      [+ New Configuration]             │  │
│  │ Analytics            │  │  └───────────────────────────────────────────────────────────────────────────────────────┘  │
│  └──────────────────────┘  │                                                                                              │
│                             │  ┌───────────────────────────────────────────────────────────────────────────────────────┐  │
│                             │  │ Agent              Assigned To           Configuration                Status   Actions │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Claude Code        Organization          model: opus-4.5                [Enabled]  ⋮  │  │
│                             │  │                                          temperature: 0.7                              │  │
│                             │  │                                          max_tokens: 4096                              │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Cursor             Team: Backend         model: sonnet-4.5              [Enabled]  ⋮  │  │
│                             │  │                                          auto_save: true                               │  │
│                             │  │                                          theme: dark                                   │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Claude Code        Employee: John Doe    model: sonnet-4.5              [Enabled]  ⋮  │  │
│                             │  │                                          temperature: 0.9                              │  │
│                             │  │                                          custom_prompts: enabled                        │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Windsurf           Team: Frontend        model: opus-4.5                [Disabled] ⋮  │  │
│                             │  │                                          preview_features: true                        │  │
│                             │  │                                          workspace_sync: enabled                       │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Cursor             Employee: Jane Smith  model: opus-4.5                [Enabled]  ⋮  │  │
│                             │  │                                          auto_complete: true                           │  │
│                             │  │                                          inline_suggestions: on                        │  │
│                             │  ├───────────────────────────────────────────────────────────────────────────────────────┤  │
│                             │  │ Claude Code        Organization          model: sonnet-4.5              [Disabled] ⋮  │  │
│                             │  │                                          tools: mcp_enabled                            │  │
│                             │  │                                          safety_mode: strict                           │  │
│                             │  └───────────────────────────────────────────────────────────────────────────────────────┘  │
│                             │                                                                                              │
│                             │  Showing 6 configurations                                    [1] 2 3 Next →                 │
└─────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┘
```

## Filter Dropdowns (Expanded)

### Level Filter Expanded
```
┌─────────────────┐
│ Level: All     ▾│
├─────────────────┤
│ ✓ All           │
│   Organization  │
│   Team          │
│   Employee      │
└─────────────────┘
```

### Agent Filter Expanded
```
┌──────────────────┐
│ Agent: All      ▾│
├──────────────────┤
│ ✓ All            │
│   Claude Code    │
│   Cursor         │
│   Windsurf       │
└──────────────────┘
```

### Status Filter Expanded
```
┌─────────────────┐
│ Status: All    ▾│
├─────────────────┤
│ ✓ All           │
│   Enabled       │
│   Disabled      │
└─────────────────┘
```

## Actions Menu (Expanded)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│ Agent              Assigned To           Configuration                Status  ⋮ │
├─────────────────────────────────────────────────────────────────────────────┌───┴─────────┐
│ Claude Code        Organization          model: opus-4.5                [En │ Edit        │
│                                          temperature: 0.7                   │ Duplicate   │
│                                          max_tokens: 4096                   │ Toggle Off  │
│                                                                             │ ───────────│
│                                                                             │ Delete      │
│                                                                             └─────────────┘
```

## Row Hover State

```
┌───────────────────────────────────────────────────────────────────────────────────────┐
│ Cursor             Employee: Jane Smith  model: opus-4.5                [Enabled]  ⋮  │ ← Hover: light blue bg
│                                          auto_complete: true                           │
│                                          inline_suggestions: on                        │
└───────────────────────────────────────────────────────────────────────────────────────┘
```

## Component Specifications

### Table Row
- **Height:** 88px (3 lines of config preview)
- **Padding:** 16px vertical, 24px horizontal
- **Border:** 1px solid #E5E7EB between rows

### Agent Column
- **Width:** 15%
- **Font:** Inter Medium 14px
- **Color:** #111827

### Assigned To Column
- **Width:** 20%
- **Font:** Inter Regular 14px
- **Color:** #6B7280
- **Format:**
  - Organization (plain text)
  - Team: {name} (with "Team:" prefix)
  - Employee: {full_name} (with "Employee:" prefix)

### Configuration Column
- **Width:** 45%
- **Font:** Inter Mono 12px
- **Color:** #4B5563
- **Max Lines:** 3
- **Truncation:** Show first 3 key-value pairs, then "..."
- **Format:** key: value (one per line)

### Status Column
- **Width:** 10%
- **Badge:**
  - Enabled: Green bg #D1FAE5, text #065F46
  - Disabled: Gray bg #F3F4F6, text #6B7280
  - Height: 28px, padding: 4px 12px, radius: 12px

### Actions Column
- **Width:** 10%
- **Icon:** Three vertical dots (⋮)
- **Size:** 20x20px
- **Color:** #9CA3AF
- **Hover:** #4B5563

### Filter Section
- **Height:** 64px
- **Background:** #F9FAFB
- **Border:** 1px solid #E5E7EB, radius: 8px
- **Padding:** 16px
- **Gap:** 12px between filters

### Search Input
- **Width:** 280px
- **Height:** 40px
- **Placeholder:** "Search configurations..."
- **Icon:** Magnifying glass (left side)

### New Configuration Button
- **Style:** Primary blue (#3B82F6)
- **Height:** 40px
- **Padding:** 12px 20px
- **Font:** Inter Medium 14px
- **Icon:** Plus (+) left of text
- **Hover:** Darken to #2563EB

## Accessibility

### Keyboard Navigation
- Tab order: Filters (left to right) → Search → New Config → Table rows → Actions
- Enter on table row: Opens edit modal
- Space on status badge: Toggles enabled/disabled
- Arrow keys in dropdowns: Navigate options
- Escape: Close dropdowns/menus

### ARIA Labels
- Filter dropdowns: `aria-label="Filter by {type}"`
- Search input: `aria-label="Search configurations"`
- New Config button: `aria-label="Create new configuration"`
- Actions menu: `aria-label="Configuration actions for {agent} assigned to {entity}"`
- Table: `role="table"` with proper headers
- Status badges: `aria-label="{agent} configuration is {status}"`

### Focus Management
- Visible focus ring: 2px solid #3B82F6, offset 2px
- Focus trap in modals
- Focus returns to trigger after closing dropdown/menu
- Skip link to main content

### Color Contrast
- All text: WCAG AA compliant (4.5:1 minimum)
- Status badges: 4.8:1 (green), 4.6:1 (gray)
- Icons: 3:1 minimum (UI components)
