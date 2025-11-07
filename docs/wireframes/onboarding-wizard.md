# Onboarding Wizard Wireframe

**Route:** `/onboarding`
**Access:** Authenticated (requires session)
**Components:** Stepper, Card, Button, Form, Alert
**Layout:** Centered wizard with progress indicator

---

## Page Purpose

4-step guided setup wizard for new users/organizations. Helps admins configure their organization, set up teams, understand agent management, and complete initial setup. Users can skip and return later.

---

## Wizard Flow

**Steps:**
1. **Welcome** - Organization overview and next steps
2. **Setup Teams** - Create initial team structure (optional)
3. **Agent Configuration** - Learn about agent management
4. **Complete** - Summary and next actions

---

## Visual Layout (Step 1: Welcome)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo] Ubik Enterprise                           [Skip Tour →] │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│        ①────────○────────○────────○                              │
│      Welcome  Teams   Agents  Complete                           │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │         👋 Welcome to Ubik Enterprise!                  │   │
│   │                                                          │   │
│   │     Let's get your AI agent management platform         │   │
│   │                    set up in 4 easy steps               │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 📊 Your Organization                              │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Name: Acme Corporation                            │ │   │
│   │   │  Workspace: https://acme-corp.ubik.io             │ │   │
│   │   │  Plan: Free Trial (14 days remaining)             │ │   │
│   │   │  Your Role: Administrator                          │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ ✅ What You Can Do                                │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  • Invite team members to your organization       │ │   │
│   │   │  • Configure AI agents (Claude, Cursor, etc.)     │ │   │
│   │   │  • Set usage policies and permissions             │ │   │
│   │   │  • Track team usage and costs                     │ │   │
│   │   │  • Manage MCP server configurations               │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 🎯 What's Next                                    │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  In the next few steps, we'll help you:           │ │   │
│   │   │                                                    │ │   │
│   │   │  1. Create your first team (optional)             │ │   │
│   │   │  2. Learn about agent configuration               │ │   │
│   │   │  3. Get ready to invite your team                 │ │   │
│   │   │                                                    │ │   │
│   │   │  This will only take 2-3 minutes!                 │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │                      ┌──────────────────┐               │   │
│   │                      │  Let's Get Started│               │   │
│   │                      └──────────────────┘               │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Step 2: Setup Teams)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo] Ubik Enterprise                           [Skip Tour →] │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│        ○────────②────────○────────○                              │
│      Welcome  Teams   Agents  Complete                           │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │              👥 Organize Your Team                      │   │
│   │                                                          │   │
│   │   Teams help you organize employees and manage          │   │
│   │   agent configurations by group.                         │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 💡 Why Create Teams?                              │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  • Group employees by department or function      │ │   │
│   │   │  • Apply agent configs to entire teams            │ │   │
│   │   │  • Set team-wide usage policies                   │ │   │
│   │   │  • Track team-level analytics                     │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ Create Your First Team (Optional)                 │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Team Name *                                       │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ Engineering                                  │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  Description                                       │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ Software development and engineering team    │ │ │   │
│   │   │  │                                              │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  ┌──────────────────────────────┐                 │ │   │
│   │   │  │  ✚ Create Team              │                 │ │   │
│   │   │  └──────────────────────────────┘                 │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 📋 Existing Teams (0)                             │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  No teams created yet.                             │ │   │
│   │   │  You can always create teams later from the        │ │   │
│   │   │  Teams page.                                       │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │      ┌────────┐                  ┌──────────────────┐   │   │
│   │      │ ← Back │                  │  Next: Agents →  │   │   │
│   │      └────────┘                  └──────────────────┘   │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Step 3: Agent Configuration)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo] Ubik Enterprise                           [Skip Tour →] │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│        ○────────○────────③────────○                              │
│      Welcome  Teams   Agents  Complete                           │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │            🤖 Manage AI Agent Configurations             │   │
│   │                                                          │   │
│   │   Control which AI agents your team can use and how     │   │
│   │   they're configured.                                    │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 🎯 How Agent Management Works                     │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  1. Organization Level                             │ │   │
│   │   │     → Set base configurations for all agents      │ │   │
│   │   │                                                    │ │   │
│   │   │  2. Team Level (Optional)                          │ │   │
│   │   │     → Override configs for specific teams         │ │   │
│   │   │                                                    │ │   │
│   │   │  3. Employee Level (Optional)                      │ │   │
│   │   │     → Custom configs for individual employees     │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 🔧 Available Agents                               │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  ┌──────────────────────────────────────────┐    │ │   │
│   │   │  │ 🤖 Claude Code                            │    │ │   │
│   │   │  │ Anthropic's official CLI for Claude       │    │ │   │
│   │   │  │ Status: Available                         │    │ │   │
│   │   │  │                    [Configure Org-wide →] │    │ │   │
│   │   │  └──────────────────────────────────────────┘    │ │   │
│   │   │                                                    │ │   │
│   │   │  ┌──────────────────────────────────────────┐    │ │   │
│   │   │  │ 💻 Cursor                                 │    │ │   │
│   │   │  │ AI-powered code editor                    │    │ │   │
│   │   │  │ Status: Available                         │    │ │   │
│   │   │  │                    [Configure Org-wide →] │    │ │   │
│   │   │  └──────────────────────────────────────────┘    │ │   │
│   │   │                                                    │ │   │
│   │   │  ┌──────────────────────────────────────────┐    │ │   │
│   │   │  │ 🌊 Windsurf                               │    │ │   │
│   │   │  │ AI-native IDE by Codeium                  │    │ │   │
│   │   │  │ Status: Available                         │    │ │   │
│   │   │  │                    [Configure Org-wide →] │    │ │   │
│   │   │  └──────────────────────────────────────────┘    │ │   │
│   │   │                                                    │ │   │
│   │   │  💡 You can configure agents later from the       │ │   │
│   │   │     Agent Catalog page.                           │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │      ┌────────┐                  ┌──────────────────┐   │   │
│   │      │ ← Back │                  │  Next: Complete →│   │   │
│   │      └────────┘                  └──────────────────┘   │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Step 4: Complete)

```
┌─────────────────────────────────────────────────────────────────┐
│  [Logo] Ubik Enterprise                                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│        ○────────○────────○────────④                              │
│      Welcome  Teams   Agents  Complete                           │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │              🎉 You're All Set!                          │   │
│   │                                                          │   │
│   │     Your organization is ready to use Ubik Enterprise    │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ ✅ Setup Summary                                  │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Organization: Acme Corporation                    │ │   │
│   │   │  Your Role: Administrator                          │ │   │
│   │   │  Teams Created: 1                                  │ │   │
│   │   │  Agents Configured: 0 (you can do this later)      │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 🚀 Next Steps                                     │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  1. Invite Your Team                               │ │   │
│   │   │     Add employees to your organization             │ │   │
│   │   │     ┌─────────────────────────────────┐            │ │   │
│   │   │     │  Go to Team Management →        │            │ │   │
│   │   │     └─────────────────────────────────┘            │ │   │
│   │   │                                                    │ │   │
│   │   │  2. Configure AI Agents                            │ │   │
│   │   │     Set up Claude Code, Cursor, and more           │ │   │
│   │   │     ┌─────────────────────────────────┐            │ │   │
│   │   │     │  Go to Agent Catalog →          │            │ │   │
│   │   │     └─────────────────────────────────┘            │ │   │
│   │   │                                                    │ │   │
│   │   │  3. Install CLI Client                             │ │   │
│   │   │     Your team needs the Ubik CLI to sync configs   │ │   │
│   │   │     ┌─────────────────────────────────┐            │ │   │
│   │   │     │  View Installation Guide →      │            │ │   │
│   │   │     └─────────────────────────────────┘            │ │   │
│   │   │                                                    │ │   │
│   │   │  4. Explore Analytics                              │ │   │
│   │   │     Track usage and costs across your organization │ │   │
│   │   │     ┌─────────────────────────────────┐            │ │   │
│   │   │     │  Go to Dashboard →              │            │ │   │
│   │   │     └─────────────────────────────────┘            │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ 📚 Resources                                      │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  • Documentation and Guides                        │ │   │
│   │   │  • Video Tutorials                                 │ │   │
│   │   │  • Community Support                               │ │   │
│   │   │  • Contact Support                                 │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │                ┌───────────────────────────┐             │   │
│   │                │  Go to Dashboard         │             │   │
│   │                └───────────────────────────┘             │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### Progress Stepper
- **Component:** Custom stepper or shadcn/ui breadcrumb adapted
- **States:**
  - Completed: Filled circle with checkmark (green)
  - Current: Filled circle with number (blue)
  - Upcoming: Empty circle (gray)
- **Labels:** Step names below circles
- **Responsive:** On mobile, show current step only with "Step X of 4"

### Header Bar
- **Logo:** Ubik Enterprise logo (clickable, links to dashboard)
- **Skip Button:** Text link "Skip Tour →" on right
  - Action: Shows confirmation dialog
  - Dialog: "Skip onboarding? You can access this wizard later from Settings"
  - Options: "Continue Tour" (default) | "Skip to Dashboard"

### Content Cards

Each step has:
- **Title:** Large, centered heading with emoji
- **Description:** Brief explanation of step purpose
- **Information Sections:** Cards with icons and content
- **Interactive Elements:** Forms, buttons, links as appropriate

### Navigation Buttons
- **Back Button:** Secondary style, left-aligned (except Step 1)
- **Next/Complete Button:** Primary style, right-aligned
- **States:**
  - Default: Enabled, colored
  - Hover: Darker shade
  - Loading: Spinner + "Processing..."
  - Disabled: Grayed out (when form invalid)

---

## Step Details

### Step 1: Welcome

**Purpose:** Orient user and show what they'll accomplish

**Components:**
- Organization info card (read-only)
- Benefits list
- What's next preview

**Actions:**
- Primary: "Let's Get Started" → Next step
- Secondary: "Skip Tour" → Dashboard

**No API calls required**

---

### Step 2: Setup Teams

**Purpose:** Optional team creation

**Components:**
- Why teams? Info card
- Team creation form (optional)
- List of created teams (starts empty)

**Form Fields:**
- Team Name (required, 3-100 chars)
- Description (optional, max 500 chars)

**Actions:**
- "Create Team" button (creates team, adds to list, clears form)
- "Back" → Previous step
- "Next: Agents →" → Next step

**API Call:** `POST /teams`

**Validation:**
- Team name required if form is touched
- Can proceed without creating any teams

**Success State:**
- Show created team in list below form
- Success toast: "Team '[Name]' created successfully"
- Form clears for creating another team

**Error Handling:**
- Show inline errors for validation
- Show alert for API errors
- Allow retry without losing data

---

### Step 3: Agent Configuration

**Purpose:** Educate about agent management, optional configuration

**Components:**
- How it works explanation
- Available agents list with cards
- Quick config buttons (optional)

**Agent Cards:**
- Agent icon/logo
- Agent name
- Brief description
- Status badge
- "Configure Org-wide →" link (opens modal or navigates to config page)

**Actions:**
- "Configure Org-wide" (optional, per agent)
- "Back" → Previous step
- "Next: Complete →" → Final step

**API Calls:** None required (info only)
- Optional: `POST /organizations/current/agent-configs` if user configures

**Note:** Configuration is completely optional. User can skip and configure later.

---

### Step 4: Complete

**Purpose:** Summary and next actions

**Components:**
- Setup summary card
- Next steps list with action buttons
- Resources links

**Next Steps Cards:**
1. **Invite Your Team** → `/teams` page
2. **Configure AI Agents** → `/agents` page
3. **Install CLI Client** → Documentation page
4. **Explore Analytics** → `/dashboard` page

**Actions:**
- Primary: "Go to Dashboard" → `/dashboard`
- Each next step has its own action button

**No API calls required**

---

## User Flows

### Happy Path (Full Wizard)
1. User completes signup → Redirected to `/onboarding`
2. Step 1: Reads welcome, clicks "Let's Get Started"
3. Step 2: Creates "Engineering" team, clicks "Next"
4. Step 3: Reads about agents, optionally configures one, clicks "Next"
5. Step 4: Reviews summary, clicks "Go to Dashboard"
6. Redirected to `/dashboard`
7. Session flag set: `onboarding_completed = true`

### Skip Wizard
1. User lands on Step 1
2. Clicks "Skip Tour" link
3. Confirmation dialog appears
4. User confirms "Skip to Dashboard"
5. Redirected to `/dashboard`
6. Session flag set: `onboarding_skipped = true`
7. Show toast: "You can access the onboarding wizard from Settings anytime"

### Partial Completion
1. User completes Steps 1-2
2. User closes browser tab
3. User returns later
4. On next login, redirected to `/onboarding` (remembers progress)
5. Resumes at Step 3
6. Completes wizard

### Back Navigation
1. User is on Step 3
2. Clicks "Back" button
3. Returns to Step 2 with previous data preserved
4. User can edit team form or proceed forward

---

## State Management

### Progress Tracking
- Store wizard state in session/cookies
- Track: `current_step`, `completed_steps`, `created_teams`, `configured_agents`
- Persist across page refreshes
- Clear on wizard completion or skip

### Data Persistence
- Teams created during wizard are saved immediately (via API)
- Agent configs are saved immediately (if user configures)
- No "Save Draft" needed - each action is committed

### Completion Flag
- Set `employee.preferences.onboarding_completed = true` on finish
- Set `employee.preferences.onboarding_skipped = true` on skip
- Check this flag on login to decide if redirect is needed

---

## API Integration

### Step 2: Create Team
**Endpoint:** `POST /teams`

**Request:**
```json
{
  "name": "Engineering",
  "description": "Software development team"
}
```

**Response (201):**
```json
{
  "id": "uuid",
  "org_id": "uuid",
  "name": "Engineering",
  "description": "Software development team",
  "created_at": "timestamp"
}
```

### Step 3: Configure Agent (Optional)
**Endpoint:** `POST /organizations/current/agent-configs`

**Request:**
```json
{
  "agent_id": "uuid",
  "config": {
    "settings": {...}
  }
}
```

### Wizard Completion
**Endpoint:** `PATCH /auth/me`

**Request:**
```json
{
  "preferences": {
    "onboarding_completed": true,
    "onboarding_completed_at": "timestamp"
  }
}
```

---

## Accessibility (WCAG AA)

### Keyboard Navigation
- Tab through all interactive elements in logical order
- Arrow keys to navigate stepper (optional enhancement)
- Enter key to submit forms
- Escape key to close skip confirmation dialog

### Screen Reader Support
- Stepper announces: "Step 2 of 4: Setup Teams"
- Progress announced when step changes
- Form errors announced via live regions
- Success messages announced
- Skip dialog properly labeled

### Visual Design
- High contrast for all text (4.5:1)
- Clear focus indicators
- Step numbers large and readable
- Icons supplement text (not replace)
- Touch targets 44px minimum

---

## Responsive Design

### Mobile (< 640px)
- Stepper: Show current step only + count
  ```
  ② Setup Teams
  Step 2 of 4
  ```
- Stack all content vertically
- Full-width buttons
- Reduce padding/spacing
- Smaller card sizes

### Tablet (640px - 1024px)
- Show abbreviated stepper with icons only
- Maintain single column layout
- Slightly larger cards

### Desktop (> 1024px)
- Full stepper with labels
- Centered content (max 800px wide)
- More generous spacing
- Side-by-side layout for some sections (e.g., benefits)

---

## Implementation Notes

### Technologies
- **Framework:** Next.js 14 (App Router)
- **State:** React useState for current step, React Query for API calls
- **Components:** shadcn/ui (Card, Button, Form, Input, Dialog, Alert)
- **Styling:** Tailwind CSS
- **Persistence:** Session storage or cookies for wizard state

### Wizard State Management
```typescript
type WizardState = {
  currentStep: 1 | 2 | 3 | 4;
  completedSteps: number[];
  createdTeams: Team[];
  configuredAgents: string[];
};

// Store in session storage
const [wizardState, setWizardState] = useState<WizardState>(() => {
  const saved = sessionStorage.getItem('onboarding_wizard');
  return saved ? JSON.parse(saved) : {
    currentStep: 1,
    completedSteps: [],
    createdTeams: [],
    configuredAgents: []
  };
});

// Persist on change
useEffect(() => {
  sessionStorage.setItem('onboarding_wizard', JSON.stringify(wizardState));
}, [wizardState]);
```

### Skip Confirmation Dialog
```typescript
<AlertDialog>
  <AlertDialogTrigger>Skip Tour →</AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogTitle>Skip onboarding wizard?</AlertDialogTitle>
    <AlertDialogDescription>
      You can access this wizard later from Settings > Onboarding.
      Are you sure you want to skip?
    </AlertDialogDescription>
    <AlertDialogFooter>
      <AlertDialogCancel>Continue Tour</AlertDialogCancel>
      <AlertDialogAction onClick={handleSkip}>
        Skip to Dashboard
      </AlertDialogAction>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
```

### Progress Indicator Component
```typescript
<div className="flex items-center justify-center gap-4">
  {[1, 2, 3, 4].map(step => (
    <div key={step} className="flex flex-col items-center">
      <div className={cn(
        "w-10 h-10 rounded-full flex items-center justify-center",
        step < currentStep && "bg-green-500 text-white",
        step === currentStep && "bg-blue-500 text-white",
        step > currentStep && "bg-gray-200 text-gray-500"
      )}>
        {step < currentStep ? <CheckIcon /> : step}
      </div>
      <span className="text-sm mt-2">
        {stepLabels[step]}
      </span>
    </div>
  ))}
</div>
```

---

## Testing Checklist

### Unit Tests
- [ ] Stepper renders correctly for each step
- [ ] Navigation between steps works
- [ ] Skip confirmation dialog appears
- [ ] Form validation on Step 2
- [ ] State persistence across steps

### Integration Tests
- [ ] Create team on Step 2
- [ ] Navigate back and forward preserves data
- [ ] Skip wizard redirects to dashboard
- [ ] Complete wizard sets completion flag
- [ ] API errors handled gracefully

### E2E Tests (Playwright)
- [ ] Full wizard completion flow
- [ ] Skip wizard flow
- [ ] Partial completion + resume
- [ ] Team creation during wizard
- [ ] Back/forward navigation
- [ ] Mobile responsive behavior

### Accessibility Tests
- [ ] Keyboard navigation through wizard
- [ ] Screen reader announcements
- [ ] Focus management between steps
- [ ] Dialog accessibility

---

## Related Pages
- **Previous:** `/signup` (Signup Page)
- **Next:** `/dashboard` (Dashboard)
- **Related:** `/teams` (Team Management), `/agents` (Agent Catalog)

---

## Design System References
- **shadcn/ui Card:** https://ui.shadcn.com/docs/components/card
- **shadcn/ui Button:** https://ui.shadcn.com/docs/components/button
- **shadcn/ui Form:** https://ui.shadcn.com/docs/components/form
- **shadcn/ui Dialog:** https://ui.shadcn.com/docs/components/dialog
- **shadcn/ui Alert:** https://ui.shadcn.com/docs/components/alert

---

## Notes

**Why 4 Steps?**
- Research shows 3-5 steps is optimal for onboarding
- Each step has clear purpose and value
- Users can skip without penalty
- Balances education with speed to value

**Why Optional?**
- Not all users need teams immediately
- Agent config can be complex, better done later
- Forcing completion increases abandonment
- "Progressive disclosure" pattern

**Future Enhancements:**
- Video tutorials embedded in wizard
- Interactive tour of dashboard at completion
- Personalized recommendations based on company size
- A/B test different wizard lengths
