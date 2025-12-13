# Settings Pages - Design Overview

## Purpose

Provide centralized settings management for organization administrators and employees to configure organization details, user profiles, preferences, and account settings.

## User Stories

### Primary User: Admin

**As an admin, I want to:**

1. **Manage organization settings**
   - Update organization name and description
   - View subscription details (plan, usage, renewal date)
   - Understand billing period and spending limits
   - See when organization was created and last updated

2. **Manage my profile and preferences**
   - Update my display name
   - View my email, role, and team assignment
   - Choose my preferred theme (Light/Dark/System)
   - Control notification preferences
   - Update my password securely

3. **Navigate settings efficiently**
   - Access different settings sections via sidebar/tabs
   - Understand which section I'm currently viewing
   - Save or cancel changes clearly
   - See validation errors immediately

### Secondary User: Employee (Non-Admin)

**As an employee, I want to:**

1. **Manage my profile**
   - Update my display name
   - View (but not edit) my email, role, and team
   - Customize my preferences

2. **Limited access to settings**
   - Cannot view or edit organization settings
   - Can only access Profile/Preferences sections

## Wireframes Created

| File | Purpose | Viewport |
|------|---------|----------|
| [settings-layout.md](./settings-layout.md) | Layout with sidebar navigation | 1440px |
| [settings-organization.md](./settings-organization.md) | Organization settings page | 1440px |
| [settings-profile.md](./settings-profile.md) | Profile and preferences page | 1440px |

## Key Design Decisions

### 1. Sidebar vs Tabs Navigation

**Decision:** Sidebar navigation (matching main dashboard pattern)

**Rationale:**
- Consistent with main application sidebar pattern
- Allows for easy expansion of settings sections
- Clearer visual hierarchy
- Better for responsive design (can collapse to tabs on mobile)

### 2. Organization Settings Visibility

**Decision:** Show to admins only, hide from regular employees

**Rationale:**
- Organization settings require admin permissions
- Regular employees shouldn't see settings they can't access
- Reduces confusion and clutter
- Matches typical SaaS permission model

### 3. Inline Editing vs Modal Forms

**Decision:** Inline editing with save/cancel buttons

**Rationale:**
- Faster workflow (no modal popup)
- See all settings at once
- Clear save state (enabled only when changes exist)
- Follows modern SaaS patterns (Stripe, GitHub, etc.)

### 4. Subscription Info Read-Only

**Decision:** Display only, no editing

**Rationale:**
- Subscription changes handled by billing portal (future feature)
- Prevents accidental plan changes
- Shows context for usage limits
- Clear visual distinction (read-only cards vs editable forms)

### 5. Password Change Flow

**Decision:** Inline form with current + new password fields

**Rationale:**
- Common pattern users expect
- Security: requires current password
- Validation feedback immediate
- No need for separate page/modal

## Component Hierarchy

```
/settings
├── SettingsLayout
│   ├── Sidebar
│   │   ├── OrganizationLink (admin only)
│   │   ├── ProfileLink
│   │   ├── BillingLink (future, grayed out)
│   │   ├── SecurityLink (future, grayed out)
│   │   └── IntegrationsLink (future, grayed out)
│   └── ContentArea
│       └── [Dynamic route content]
└── SettingsRoutes
    ├── /settings → redirect to /settings/organization (admin) or /settings/profile (employee)
    ├── /settings/organization (admin only)
    │   ├── PageHeader
    │   ├── OrganizationInfoForm
    │   │   ├── NameInput (editable)
    │   │   ├── DescriptionTextarea (editable)
    │   │   ├── CreatedDate (read-only)
    │   │   └── UpdatedDate (read-only)
    │   ├── SubscriptionCard (read-only)
    │   │   ├── PlanBadge
    │   │   ├── UsageProgress
    │   │   ├── SpendingInfo
    │   │   └── RenewalDate
    │   └── FormActions
    │       ├── CancelButton (enabled when dirty)
    │       └── SaveButton (enabled when dirty + valid)
    └── /settings/profile
        ├── PageHeader
        ├── ProfileInfoForm
        │   ├── DisplayNameInput (editable)
        │   ├── EmailInput (read-only)
        │   ├── RoleBadge (read-only)
        │   └── TeamBadge (read-only)
        ├── AppearanceSection
        │   └── ThemeSelect (Light/Dark/System)
        ├── NotificationSection
        │   ├── EmailNotificationsSwitch
        │   ├── InAppNotificationsSwitch
        │   ├── AgentActivitySwitch
        │   └── WeeklySummarySwitch
        ├── SecuritySection
        │   └── PasswordUpdateForm
        │       ├── CurrentPasswordInput
        │       ├── NewPasswordInput
        │       ├── ConfirmPasswordInput
        │       └── UpdatePasswordButton
        └── FormActions
            ├── CancelButton (for profile/appearance/notifications)
            └── SaveButton (for profile/appearance/notifications)
```

## State Management

### Page States

1. **Loading:** Skeleton loaders for forms
2. **Idle:** No changes, Save button disabled
3. **Dirty:** Changes detected, Save/Cancel enabled
4. **Saving:** Loading state on Save button
5. **Success:** Toast notification, form reset to idle
6. **Error:** Inline validation errors, toast for API errors

### Form Validation

**Organization Settings:**
- Name: Required, 3-255 characters
- Description: Optional, max 1000 characters

**Profile Settings:**
- Display name: Required, 2-255 characters

**Password Change:**
- Current password: Required
- New password: Required, min 8 characters, must include uppercase, lowercase, number
- Confirm password: Must match new password

### Permission Checks

- Organization settings route: Check `role.permissions.manage_organization`
- Profile settings route: All authenticated users
- Render sidebar items conditionally based on permissions

## Responsive Behavior

### Desktop (1024px+)
- Sidebar always visible (200px width)
- Content area uses remaining space
- Forms max-width: 800px for readability

### Tablet (768px - 1023px)
- Sidebar collapses to icons only (64px width)
- Labels hidden, icons + tooltips
- Content area expands

### Mobile (320px - 767px)
- Sidebar becomes top tabs (horizontal scroll)
- Full-width content
- Stacked form fields
- Touch-friendly tap targets (44x44px minimum)

## Accessibility Requirements

- [ ] Keyboard navigation through all form fields
- [ ] Tab order follows visual hierarchy
- [ ] Focus indicators visible (2px ring)
- [ ] ARIA labels for icon buttons
- [ ] Error messages announced by screen readers
- [ ] Color contrast WCAG AA compliant
- [ ] Password fields have show/hide toggle
- [ ] Success/error toasts have role="alert"

## Design System Usage

**Colors:**
- Primary: Blue #3B82F6 (Save buttons, active sidebar)
- Success: Green #10B981 (Success toasts)
- Error: Red #EF4444 (Validation errors)
- Muted: Gray #6B7280 (Read-only fields, disabled state)

**Typography:**
- Page heading: H1 2.25rem (36px) font-bold
- Section heading: H2 1.5rem (24px) font-semibold
- Form label: 0.875rem (14px) font-medium
- Body text: 1rem (16px)
- Helper text: 0.875rem (14px) text-muted-foreground

**Spacing:**
- Section spacing: 32px (space-y-8)
- Form field spacing: 16px (space-y-4)
- Input padding: 8px (p-2)

**Components (shadcn/ui):**
- Form: Field, Label, Control, Description, Message
- Input: Text inputs, textarea
- Select: Theme selector
- Switch: Notification toggles
- Button: Primary, Secondary, Ghost
- Card: Subscription info
- Badge: Role, Team, Plan
- Toast: Success/error notifications
- Separator: Section dividers

## API Integration

### Endpoints Needed

**Organization Settings:**
```
GET /organizations/current
Response: { id, name, slug, plan, settings, created_at, updated_at }

PATCH /organizations/current
Body: { name?, description? }
Response: { organization }

GET /subscriptions/current
Response: { plan_type, monthly_budget_usd, current_spending_usd, billing_period_start, billing_period_end }
```

**Profile Settings:**
```
GET /auth/me
Response: { id, email, full_name, role, team, preferences }

PATCH /employees/current
Body: { full_name?, preferences? }
Response: { employee }

POST /auth/change-password
Body: { current_password, new_password }
Response: { success: true }
```

### Data Structure

**Employee Preferences JSONB:**
```json
{
  "theme": "light" | "dark" | "system",
  "notifications": {
    "email": true,
    "in_app": true,
    "agent_activity": false,
    "weekly_summary": true
  }
}
```

**Organization Settings JSONB:**
```json
{
  "description": "Company description here..."
}
```

## Future Enhancements

**Phase 2 (Future):**
- Billing section (view invoices, update payment method)
- Security section (2FA, session management, API tokens)
- Integrations section (Slack, webhooks, SSO)
- Team settings (if user is team lead)
- Audit log for settings changes

**Phase 3 (Future):**
- Organization logo upload
- Custom branding colors
- Email template customization
- Timezone and locale settings

## Related Issues

- #109 - Create settings pages (Organization, Profile)

## Notes for Implementation

1. **Start with Organization Settings:**
   - Simplest page (mostly read-only)
   - Establishes patterns for forms

2. **Then Profile Settings:**
   - More complex (multiple sections)
   - Password change flow needs careful UX

3. **Layout Last:**
   - Sidebar navigation
   - Route structure and redirects

4. **Testing Priority:**
   - Permission checks (admin vs employee)
   - Form validation (all fields)
   - Password change security flow
   - Save/cancel state management

5. **Use Existing Patterns:**
   - Follow agents page table patterns
   - Reuse form components from existing pages
   - Match toast notification style
