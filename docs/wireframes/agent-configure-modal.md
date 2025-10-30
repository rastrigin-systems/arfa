# Agent Configuration Modal Wireframe

## Overview
Modal dialog that appears when clicking "Configure" button on an agent card in the Available Agents tab.

## Layout

```
┌─────────────────────────────────────────────────────────┐
│ Configure Claude Code                              [X]  │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  Agent Settings                                          │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ Agent Name                                      │    │
│  │ Claude Code (read-only)                        │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ API Key *                                       │    │
│  │ sk-ant-api03-...                               │    │
│  └────────────────────────────────────────────────┘    │
│  [💡] Enter your Anthropic API key from               │
│      https://console.anthropic.com                     │
│                                                          │
│  ☑ Enable this agent for organization                  │
│                                                          │
│  Advanced Configuration (JSON)                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ {                                               │    │
│  │   "api_key": "sk-ant-api03-...",               │    │
│  │   "model": "claude-sonnet-4.5",                │    │
│  │   "max_tokens": 4096                           │    │
│  │ }                                               │    │
│  │                                                 │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
│                           [Cancel]  [Save Configuration] │
└─────────────────────────────────────────────────────────┘
```

## Behavior

### Opening Modal
- Click "Configure" button on agent card
- Check if org config already exists
  - If exists: Load existing config, populate form
  - If new: Show empty form with `is_enabled = true` default

### Fields
1. **Agent Name** - Read-only, shows agent display name
2. **API Key** - Password input (required)
   - Masked input: `••••••••••••••••`
   - Validation: Must start with `sk-ant-`
   - Help text with link to Anthropic console
3. **Enable Toggle** - Checkbox (default: checked)
4. **JSON Editor** - Textarea with monospace font
   - Auto-populated with: `{ "api_key": "<entered-key>", ... }`
   - Advanced users can add custom fields
   - Real-time JSON validation

### Actions
- **Cancel**: Close modal, discard changes
- **Save Configuration**:
  - Validate API key format
  - Validate JSON syntax
  - POST `/organizations/current/agent-configs` (if new)
  - PATCH `/organizations/current/agent-configs/{id}` (if editing)
  - Show success message
  - Refresh Organization Configs tab
  - Close modal

### Validation
- API key required (unless unchecking enable)
- JSON must be valid
- API key in JSON must match API Key field

## Integration Points

### API Endpoints
- `POST /api/v1/organizations/current/agent-configs`
  ```json
  {
    "agent_id": "uuid",
    "config": {
      "api_key": "sk-ant-api03-...",
      "model": "claude-sonnet-4.5"
    },
    "is_enabled": true
  }
  ```

- `PATCH /api/v1/organizations/current/agent-configs/{id}`
  ```json
  {
    "config": { "api_key": "..." },
    "is_enabled": false
  }
  ```

### Database
- Table: `org_agent_configs`
- Config JSONB structure:
  ```json
  {
    "api_key": "sk-ant-api03-...",
    "model": "claude-sonnet-4.5",
    "max_tokens": 4096
  }
  ```

## User Flow

1. Admin views Available Agents tab
2. Clicks "Configure" on Claude Code agent
3. Modal opens with empty form
4. Enters API key: `sk-ant-api03-xyz123...`
5. JSON editor auto-updates with key
6. Checks "Enable this agent"
7. Clicks "Save Configuration"
8. Success message: "Agent configured successfully"
9. Modal closes
10. Organization Configs tab now shows Claude Code as configured
11. Employees can now use CLI without manual API key setup
