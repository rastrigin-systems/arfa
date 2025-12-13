# Agents Page Redesign - Quick Reference

**One-page visual summary for developers**

---

## Card Design (Final Specs)

```
┌─────────────────────────────────────┐
│ Agent Name (H3, 20px, bold)         │ ← Primary focus
│ ─────────────────────────────────── │ ← 1px separator (gray-200)
│ Type: IDE Agent (12px, uppercase)   │ ← Badge (blue-100 bg)
│                                     │
│ AI-powered CLI coding assistant     │ ← Description (14px, 2 lines max)
│ for developers                      │   gray-600, line-clamp-2
│                                     │ ← 16px spacing
│ [◉ Enabled]                         │ ← Toggle (44x24px, green-600)
│ [Configure →]                       │ ← Button (44px height, outlined)
└─────────────────────────────────────┘

Dimensions: 280px width × 220px height
Padding: 24px (p-6)
Border: 1px solid gray-200
Radius: 12px (rounded-lg)
Shadow: shadow-md, hover:shadow-lg
```

---

## Grid Layout

```
Desktop (1024px+):
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Card 1    │ │   Card 2    │ │   Card 3    │
└─────────────┘ └─────────────┘ └─────────────┘
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│   Card 4    │ │   Card 5    │ │   Card 6    │
└─────────────┘ └─────────────┘ └─────────────┘

grid-cols-3, gap-6


Tablet (768px-1023px):
┌─────────────┐ ┌─────────────┐
│   Card 1    │ │   Card 2    │
└─────────────┘ └─────────────┘
┌─────────────┐ ┌─────────────┐
│   Card 3    │ │   Card 4    │
└─────────────┘ └─────────────┘

grid-cols-2, gap-6


Mobile (320px-767px):
┌─────────────┐
│   Card 1    │
└─────────────┘
┌─────────────┐
│   Card 2    │
└─────────────┘
┌─────────────┐
│   Card 3    │
└─────────────┘

grid-cols-1, gap-4
```

---

## Component Code Snippets

### Agent Card Component

```tsx
interface AgentCardProps {
  agent: {
    id: string
    name: string
    type: string
    description: string
    enabled: boolean
  }
  onToggle: (agentId: string, enabled: boolean) => void
  onConfigure: (agentId: string) => void
}

export function AgentCard({ agent, onToggle, onConfigure }: AgentCardProps) {
  return (
    <div className="bg-white border border-gray-200 rounded-lg p-6 shadow-md hover:shadow-lg transition-shadow">
      {/* Title */}
      <h3 className="text-xl font-semibold text-gray-900 mb-2">
        {agent.name}
      </h3>

      {/* Separator */}
      <div className="h-px bg-gray-200 mb-3" />

      {/* Badge */}
      <span className="inline-block px-2 py-1 bg-blue-100 text-blue-600 text-xs font-medium uppercase rounded-sm mb-2">
        {agent.type}
      </span>

      {/* Description */}
      <p className="text-sm text-gray-600 line-clamp-2 mb-4">
        {agent.description}
      </p>

      {/* Toggle */}
      <div className="flex items-center mb-2">
        <Switch
          checked={agent.enabled}
          onCheckedChange={(enabled) => onToggle(agent.id, enabled)}
          className={cn(
            "data-[state=checked]:bg-green-600",
            "data-[state=unchecked]:bg-gray-300"
          )}
          aria-label={`Toggle ${agent.name}`}
        />
        <span className="ml-2 text-sm font-medium text-gray-900">
          {agent.enabled ? 'Enabled' : 'Disabled'}
        </span>
      </div>

      {/* Configure Button */}
      {agent.enabled && (
        <Button
          variant="outline"
          className="w-full"
          onClick={() => onConfigure(agent.id)}
        >
          Configure →
        </Button>
      )}
    </div>
  )
}
```

---

## Key Interactions

### 1. Enable Agent

```tsx
async function handleToggle(agentId: string, enabled: boolean) {
  if (!enabled) {
    // Disabling - show confirmation
    const confirmed = await showConfirmDialog({
      title: `Disable ${agent.name}?`,
      description: `This will remove access for all teams and employees. Configurations will be preserved.`,
      confirmText: 'Disable Agent',
      confirmVariant: 'destructive',
    })

    if (!confirmed) return
  }

  // Optimistic update
  updateAgentLocally(agentId, { enabled })

  try {
    const response = enabled
      ? await api.post(`/agents/${agentId}/enable`)
      : await api.post(`/agents/${agentId}/disable`)

    // Success toast
    toast({
      title: `${agent.name} ${enabled ? 'enabled' : 'disabled'}`,
      variant: 'default',
    })
  } catch (error) {
    // Rollback
    updateAgentLocally(agentId, { enabled: !enabled })

    // Error toast
    toast({
      title: `Failed to ${enabled ? 'enable' : 'disable'} ${agent.name}`,
      description: error.message,
      variant: 'destructive',
    })
  }
}
```

### 2. Configure Agent

```tsx
function handleConfigure(agentId: string) {
  router.push(`/configs?agent=${agentId}`)
}
```

### 3. Search & Filter

```tsx
const [searchQuery, setSearchQuery] = useState('')
const [filters, setFilters] = useState({
  types: [],
  statuses: [],
})

const filteredAgents = useMemo(() => {
  return agents.filter(agent => {
    // Search
    if (searchQuery && !agent.name.toLowerCase().includes(searchQuery.toLowerCase())) {
      return false
    }

    // Type filter
    if (filters.types.length > 0 && !filters.types.includes(agent.type)) {
      return false
    }

    // Status filter
    if (filters.statuses.length > 0) {
      const status = agent.enabled ? 'enabled' : 'disabled'
      if (!filters.statuses.includes(status)) {
        return false
      }
    }

    return true
  })
}, [agents, searchQuery, filters])

// Debounced search
const debouncedSearch = useDebouncedCallback(
  (value: string) => setSearchQuery(value),
  300
)
```

---

## Color Reference

```css
/* Primary Actions */
--blue-600: #3B82F6;    /* Enable button, links */
--blue-500: #60A5FA;    /* Hover states */
--blue-100: #DBEAFE;    /* Badge backgrounds */

/* Success States */
--green-600: #10B981;   /* Enabled toggle */
--green-500: #34D399;   /* Hover */
--green-100: #D1FAE5;   /* Background */

/* Destructive Actions */
--red-600: #EF4444;     /* Disable button */
--red-500: #F87171;     /* Hover */
--red-100: #FEE2E2;     /* Background */

/* Neutral Palette */
--gray-900: #111827;    /* Headings */
--gray-700: #374151;    /* Body text */
--gray-600: #4B5563;    /* Secondary text */
--gray-300: #D1D5DB;    /* Borders */
--gray-200: #E5E7EB;    /* Separators */
--gray-100: #F3F4F6;    /* Backgrounds */

/* Surface */
--white: #FFFFFF;       /* Cards */
```

---

## Typography Scale

```css
/* Font Family */
font-family: 'Inter', sans-serif;

/* Sizes */
--h1: 2.25rem;     /* 36px - Page title */
--h2: 1.5rem;      /* 24px - Sections */
--h3: 1.25rem;     /* 20px - Card titles */
--body: 1rem;      /* 16px - Default */
--small: 0.875rem; /* 14px - Descriptions */
--label: 0.75rem;  /* 12px - Badges */

/* Weights */
--bold: 700;       /* Headings */
--semibold: 600;   /* Subheadings */
--medium: 500;     /* Labels */
--regular: 400;    /* Body */
```

---

## Spacing

```
Base unit: 4px (0.25rem)

Common values:
2  →  8px  (gap between toggle and label)
3  → 12px  (separator margins)
4  → 16px  (section spacing)
6  → 24px  (card padding, grid gap)
8  → 32px  (page padding)
```

---

## States

### Agent Card States

```
1. Enabled
   - Toggle: green-600 background, checked
   - Button: "Configure →" (outline variant)

2. Disabled
   - Toggle: gray-300 background, unchecked
   - Button: "Enable" (primary variant)

3. Loading (during toggle)
   - Toggle: disabled with spinner
   - Button: disabled with "Enabling..." text

4. Error
   - Toggle: returns to previous state
   - Toast: error message with retry option
```

---

## Accessibility Checklist

```
✓ Color contrast: 4.5:1 minimum (WCAG AA)
✓ Touch targets: 44x44px minimum
✓ Focus indicators: 2px outline, 4px offset
✓ Keyboard navigation: Tab order logical
✓ ARIA labels: All interactive elements
✓ Screen reader: Descriptive announcements
✓ Motion: Respects prefers-reduced-motion
```

---

## API Endpoints

```
GET  /api/v1/agents
→ Response: Agent[]

POST /api/v1/agents/{id}/enable
→ Request: { org_id: string }
→ Response: { agent_id, enabled: true, enabled_at }

POST /api/v1/agents/{id}/disable
→ Request: { org_id: string }
→ Response: { agent_id, enabled: false, disabled_at, affected_teams, affected_employees }
```

---

## File Structure

```
services/web/
├── app/(dashboard)/agents/
│   ├── page.tsx              # Main agents page
│   └── loading.tsx           # Loading skeleton
├── components/agents/
│   ├── agent-card.tsx        # Individual card
│   ├── agent-grid.tsx        # Grid container
│   ├── agent-search.tsx      # Search input
│   ├── agent-filters.tsx     # Filter controls
│   ├── disable-dialog.tsx    # Confirmation dialog
│   └── agent-skeleton.tsx    # Loading skeleton
└── lib/api/
    └── agents.ts             # API client
```

---

## Testing Checklist

### Functional
- [ ] Enable agent (disabled → enabled)
- [ ] Disable agent with confirmation
- [ ] Cancel disable confirmation
- [ ] Configure button navigates to /configs
- [ ] Search filters agents (debounced)
- [ ] Type filter works
- [ ] Status filter works
- [ ] Clear filters resets

### Responsive
- [ ] Mobile (375px): Single column
- [ ] Tablet (768px): Two columns
- [ ] Desktop (1024px): Three columns
- [ ] Touch targets meet 44x44px

### Accessibility
- [ ] Keyboard navigation works
- [ ] Screen reader announces elements
- [ ] Focus indicators visible
- [ ] Color contrast passes WCAG AA
- [ ] Zoom to 200% works

---

## Performance Targets

```
Initial Load:     < 1s (LCP)
Toggle Response:  < 100ms (perceived)
Search Filter:    < 50ms (debounced)
Bundle Size:      < 200KB
Lighthouse Score: > 90
```

---

## Related Files

- [agents-redesign-README.md](./agents-redesign-README.md) - Complete overview
- [agents-redesign-desktop.md](./agents-redesign-desktop.md) - Desktop wireframes
- [agents-redesign-mobile.md](./agents-redesign-mobile.md) - Mobile wireframes
- [agents-redesign-specs.md](./agents-redesign-specs.md) - Full design specs
- [agents-redesign-userflow.md](./agents-redesign-userflow.md) - User flows
- [agents-redesign-comparison.md](./agents-redesign-comparison.md) - Before/after

---

**Ready for implementation:** Yes
**Designer:** product-designer agent
**Issue:** #287
**Date:** 2025-12-12
