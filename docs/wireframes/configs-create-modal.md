# Configs Page - Create Configuration Modal

## Modal - Step 1: Agent Selection

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Select Agent *                                             │
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ Choose an agent...                                   ▾│ │
                    │  └───────────────────────────────────────────────────────┘ │
                    │                                                             │
                    │  Assign To *                                                │
                    │  ┌───┐  Organization                                        │
                    │  │ ○ │  Apply to entire organization                        │
                    │  └───┘                                                      │
                    │  ┌───┐  Team                                                │
                    │  │ ○ │  Apply to specific team                              │
                    │  └───┘                                                      │
                    │  ┌───┐  Employee                                            │
                    │  │ ○ │  Apply to specific employee                          │
                    │  └───┘                                                      │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                 [Cancel]  [Next: Configure] │
                    └─────────────────────────────────────────────────────────────┘
```

## Modal - Agent Dropdown Expanded

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Select Agent *                                             │
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ Choose an agent...                                   ▾│ │
                    │  ├───────────────────────────────────────────────────────┤ │
                    │  │ Claude Code                                           │ │
                    │  │ Advanced AI pair programming                          │ │
                    │  ├───────────────────────────────────────────────────────┤ │
                    │  │ Cursor                                                │ │
                    │  │ AI-powered code editor                                │ │
                    │  ├───────────────────────────────────────────────────────┤ │
                    │  │ Windsurf                                              │ │
                    │  │ Collaborative AI development                          │ │
                    │  └───────────────────────────────────────────────────────┘ │
                    │                                                             │
                    │  Assign To *                                                │
```

## Modal - Team Selected

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Select Agent *                                             │
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ Claude Code                                          ▾│ │
                    │  └───────────────────────────────────────────────────────┘ │
                    │                                                             │
                    │  Assign To *                                                │
                    │  ┌───┐  Organization                                        │
                    │  │ ○ │  Apply to entire organization                        │
                    │  └───┘                                                      │
                    │  ┌───┐  Team                                                │
                    │  │ ● │  Apply to specific team                              │
                    │  └───┘                                                      │
                    │       ┌───────────────────────────────────────────────────┐ │
                    │       │ Select team...                                   ▾│ │
                    │       └───────────────────────────────────────────────────┘ │
                    │  ┌───┐  Employee                                            │
                    │  │ ○ │  Apply to specific employee                          │
                    │  └───┘                                                      │
                    │                                                             │
                    │                                                             │
                    │                                 [Cancel]  [Next: Configure] │
                    └─────────────────────────────────────────────────────────────┘
```

## Modal - Employee Selected

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Select Agent *                                             │
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ Cursor                                               ▾│ │
                    │  └───────────────────────────────────────────────────────┘ │
                    │                                                             │
                    │  Assign To *                                                │
                    │  ┌───┐  Organization                                        │
                    │  │ ○ │  Apply to entire organization                        │
                    │  └───┘                                                      │
                    │  ┌───┐  Team                                                │
                    │  │ ○ │  Apply to specific team                              │
                    │  └───┘                                                      │
                    │  ┌───┐  Employee                                            │
                    │  │ ● │  Apply to specific employee                          │
                    │  └───┘                                                      │
                    │       ┌───────────────────────────────────────────────────┐ │
                    │       │ Search employee...                               ▾│ │
                    │       └───────────────────────────────────────────────────┘ │
                    │                                                             │
                    │                                 [Cancel]  [Next: Configure] │
                    └─────────────────────────────────────────────────────────────┘
```

## Modal - Step 2: Configuration

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Claude Code → Team: Backend                                │
                    │                                                             │
                    │  Configuration                                              │
                    │  ┌─────────────────────────────────────────┐ [Toggle JSON] │
                    │  │                                         │               │
                    │  │  Model *                                │               │
                    │  │  ┌────────────────────────────────────┐ │               │
                    │  │  │ claude-sonnet-4.5               ▾ │ │               │
                    │  │  └────────────────────────────────────┘ │               │
                    │  │                                         │               │
                    │  │  Temperature                            │               │
                    │  │  ┌────────────────────────────────────┐ │               │
                    │  │  │ 0.7                                │ │               │
                    │  │  └────────────────────────────────────┘ │               │
                    │  │  [────────●──────────────────────]  0-1 │               │
                    │  │                                         │               │
                    │  │  Max Tokens                             │               │
                    │  │  ┌────────────────────────────────────┐ │               │
                    │  │  │ 4096                               │ │               │
                    │  │  └────────────────────────────────────┘ │               │
                    │  │                                         │               │
                    │  │  ☑ Enable custom prompts                │               │
                    │  │  ☐ Enable MCP tools                     │               │
                    │  │  ☑ Auto-save workspace                  │               │
                    │  │                                         │               │
                    │  └─────────────────────────────────────────┘               │
                    │                                                             │
                    │  Status                                                     │
                    │  ┌────┐  Enabled                                            │
                    │  │ ON │  Configuration is active                            │
                    │  └────┘                                                     │
                    │                                                             │
                    │                           [← Back]  [Cancel]  [Create]     │
                    └─────────────────────────────────────────────────────────────┘
```

## Modal - JSON View

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │  Claude Code → Team: Backend                                │
                    │                                                             │
                    │  Configuration                                              │
                    │  ┌─────────────────────────────────────────┐ [Toggle Form] │
                    │  │  1  {                                   │               │
                    │  │  2    "model": "claude-sonnet-4.5",     │               │
                    │  │  3    "temperature": 0.7,               │               │
                    │  │  4    "max_tokens": 4096,               │               │
                    │  │  5    "custom_prompts": true,           │               │
                    │  │  6    "mcp_enabled": false,             │               │
                    │  │  7    "auto_save": true                 │               │
                    │  │  8  }                                   │               │
                    │  │  9                                      │               │
                    │  │ 10                                      │               │
                    │  │ 11                                      │               │
                    │  │ 12                                      │               │
                    │  │ 13                                      │               │
                    │  │ 14                                      │               │
                    │  │ 15                                      │               │
                    │  └─────────────────────────────────────────┘               │
                    │  ✓ Valid JSON                                               │
                    │                                                             │
                    │  Status                                                     │
                    │  ┌────┐  Enabled                                            │
                    │  │ ON │  Configuration is active                            │
                    │  └────┘                                                     │
                    │                                                             │
                    │                           [← Back]  [Cancel]  [Create]     │
                    └─────────────────────────────────────────────────────────────┘
```

## Modal - Validation Error

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                            [X] │
                    ├─────────────────────────────────────────────────────────────┤
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ ⚠ Please select an agent and assignment target       │ │
                    │  └───────────────────────────────────────────────────────┘ │
                    │                                                             │
                    │  Select Agent *                                             │
                    │  ┌───────────────────────────────────────────────────────┐ │
                    │  │ Choose an agent...                                   ▾│ │ ← Red border
                    │  └───────────────────────────────────────────────────────┘ │
                    │  Please select an agent                                     │
                    │                                                             │
                    │  Assign To *                                                │
                    │  ┌───┐  Organization                                        │
                    │  │ ○ │  Apply to entire organization                        │
                    │  └───┘                                                      │
```

## Modal - Loading State

```
                    ┌─────────────────────────────────────────────────────────────┐
                    │  Create Agent Configuration                                 │
                    ├─────────────────────────────────────────────────────────────┤
                    │                                                             │
                    │                                                             │
                    │                        ⟳ Creating configuration...          │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    │                                                             │
                    └─────────────────────────────────────────────────────────────┘
```

## Component Specifications

### Modal Container
- **Width:** 640px
- **Max Height:** 90vh
- **Background:** White (#FFFFFF)
- **Border Radius:** 12px
- **Shadow:** 0 20px 25px -5px rgba(0, 0, 0, 0.1)
- **Padding:** 32px
- **Overlay:** rgba(0, 0, 0, 0.5)

### Modal Header
- **Font:** Inter Semibold 20px
- **Color:** #111827
- **Border Bottom:** 1px solid #E5E7EB
- **Padding Bottom:** 16px
- **Close Button:** 24x24px, hover bg #F3F4F6

### Breadcrumb (Step 2)
- **Font:** Inter Regular 14px
- **Color:** #6B7280
- **Separator:** → (arrow)
- **Margin Bottom:** 24px

### Form Labels
- **Font:** Inter Medium 14px
- **Color:** #374151
- **Required Indicator:** Red asterisk (*) #EF4444
- **Margin Bottom:** 8px

### Dropdowns
- **Height:** 44px
- **Border:** 1px solid #D1D5DB
- **Border Radius:** 6px
- **Font:** Inter Regular 14px
- **Padding:** 12px 16px
- **Hover Border:** #9CA3AF
- **Focus Border:** 2px solid #3B82F6
- **Dropdown Icon:** Chevron down, 16x16px, #6B7280

### Dropdown Options
- **Height:** 56px (with description)
- **Padding:** 12px 16px
- **Hover Background:** #F3F4F6
- **Selected Background:** #EFF6FF
- **Description Font:** Inter Regular 12px, #6B7280

### Radio Buttons
- **Size:** 20x20px
- **Border:** 2px solid #D1D5DB
- **Selected:** Blue fill #3B82F6
- **Label Font:** Inter Medium 14px
- **Description Font:** Inter Regular 13px, #6B7280
- **Spacing:** 16px between options
- **Touch Target:** 44x44px minimum

### Text Inputs
- **Height:** 44px
- **Border:** 1px solid #D1D5DB
- **Border Radius:** 6px
- **Font:** Inter Regular 14px
- **Padding:** 12px 16px
- **Error Border:** 2px solid #EF4444

### Range Slider
- **Track Height:** 4px
- **Track Color:** #E5E7EB
- **Fill Color:** #3B82F6
- **Thumb Size:** 16x16px
- **Thumb Color:** #3B82F6
- **Labels:** Min/Max at ends, current value above thumb

### Checkboxes
- **Size:** 20x20px
- **Border:** 2px solid #D1D5DB
- **Checked:** Blue fill #3B82F6, white checkmark
- **Label Font:** Inter Regular 14px
- **Spacing:** 12px between items
- **Touch Target:** 44x44px minimum

### Toggle Switch
- **Width:** 48px
- **Height:** 28px
- **Background:** #D1D5DB (off), #3B82F6 (on)
- **Knob:** 24x24px, white, shadow
- **Label Font:** Inter Medium 14px
- **Description Font:** Inter Regular 13px, #6B7280

### JSON Editor
- **Height:** 320px
- **Background:** #F9FAFB
- **Border:** 1px solid #E5E7EB
- **Border Radius:** 6px
- **Font:** JetBrains Mono 13px
- **Line Numbers:** #9CA3AF, right-aligned
- **Padding:** 16px
- **Validation Indicator:** Green checkmark ✓ or Red X ✗ below editor

### Buttons
- **Height:** 44px
- **Padding:** 12px 24px
- **Border Radius:** 6px
- **Font:** Inter Medium 14px
- **Gap:** 12px between buttons

- **Primary (Create/Next):**
  - Background: #3B82F6
  - Color: White
  - Hover: #2563EB
  - Disabled: #93C5FD, not clickable

- **Secondary (Back):**
  - Background: White
  - Border: 1px solid #D1D5DB
  - Color: #374151
  - Hover: #F9FAFB

- **Tertiary (Cancel):**
  - Background: Transparent
  - Color: #6B7280
  - Hover: #F3F4F6

### Error Banner
- **Background:** #FEE2E2
- **Border:** 1px solid #FCA5A5
- **Border Radius:** 6px
- **Padding:** 12px 16px
- **Icon:** Warning ⚠ #DC2626
- **Font:** Inter Regular 14px, #991B1B
- **Margin Bottom:** 16px

### Error Messages
- **Font:** Inter Regular 12px
- **Color:** #DC2626
- **Margin Top:** 4px
- **Icon:** None (text only)

## User Flow

### Create Configuration Flow
1. Click "+ New Configuration" button
2. Modal opens with Step 1
3. Select agent from dropdown
4. Select assignment type (org/team/employee)
5. If team/employee, select from dropdown
6. Click "Next: Configure"
7. Configure settings (form or JSON)
8. Toggle enabled status
9. Click "Create"
10. Loading state shows
11. Modal closes, table refreshes with new config

### Validation Points
- **Step 1 → Step 2:** Agent and assignment must be selected
- **Step 2 → Create:** JSON must be valid (if in JSON mode)
- **Create:** Server validates configuration against agent schema

## Accessibility

### Keyboard Navigation
- Tab order: Agent dropdown → Assignment radios → Team/Employee dropdown (if visible) → Next/Cancel buttons
- Tab order (Step 2): Form fields → Toggle JSON → Status toggle → Back/Cancel/Create buttons
- Enter: Submit form if all valid
- Escape: Close modal (confirm if changes made)
- Arrow keys: Navigate radio buttons

### ARIA Labels
- Modal: `role="dialog"` `aria-labelledby="modal-title"` `aria-modal="true"`
- Agent dropdown: `aria-label="Select AI agent"` `aria-required="true"`
- Assignment radios: `role="radiogroup"` `aria-label="Select assignment target"`
- Team/Employee dropdown: `aria-label="Select {team|employee}"` (conditional)
- JSON editor: `aria-label="Configuration JSON editor"` `aria-invalid="true|false"`
- Toggle: `role="switch"` `aria-checked="true|false"` `aria-label="Enable configuration"`
- Close button: `aria-label="Close modal"`

### Focus Management
- Focus trap: Can't tab outside modal
- Initial focus: Agent dropdown
- After step change: Focus first input of new step
- After close: Return focus to "+ New Configuration" button
- Error focus: Move to first invalid field

### Screen Reader Announcements
- Step change: "Now on configure step"
- Validation error: "Error: {message}"
- Loading: "Creating configuration, please wait"
- Success: "Configuration created successfully"
- Invalid JSON: "JSON syntax error on line {N}"

### Color Contrast
- All text: WCAG AA (4.5:1)
- Form borders: 3:1 minimum
- Error text on error background: 4.8:1
- Radio/checkbox states: Non-color indicators (checkmark, dot)
