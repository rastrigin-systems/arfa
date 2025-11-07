# Team Management Page Wireframe

**Route:** `/dashboard/teams` or `/dashboard/employees`
**Access:** Authenticated (requires session)
**Permissions:** View (all employees), Invite (admin/approver only)
**Components:** Table, Dialog, Form, Button, Badge, Tabs, Alert
**Layout:** Full dashboard layout with sidebar and main content

---

## Page Purpose

Central hub for managing organization employees and sending invitations. Admins can view all employees, send invitations to new team members, track invitation status, and manage team assignments. Combines employee list, invitation management, and quick actions in one interface.

---

## Visual Layout (Main View)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [☰] Ubik Enterprise                      🔔 Notifications    👤 John Smith │
├────────┬────────────────────────────────────────────────────────────────────┤
│        │                                                                     │
│  Home  │  👥 Team Management                                                │
│        │                                                                     │
│  Teams │  ┌───────────────────────────────────────────────────────────┐   │
│        │  │  Manage your organization's employees and invitations     │   │
│ Agents │  └───────────────────────────────────────────────────────────┘   │
│        │                                                                     │
│ Config │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│ Logs   │  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────┐ │ │
│        │  │  │ Total Employees  │  │ Active           │  │ Pending    │ │ │
│Settings│  │  │                  │  │                  │  │ Invites    │ │ │
│        │  │  │       12         │  │       10         │  │      3     │ │ │
│        │  │  └──────────────────┘  └──────────────────┘  └────────────┘ │ │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
│        │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│        │  │  ┌──────────┬────────────────┐  🔍 Search... [Filter ▼]     │ │
│        │  │  │ Employees│ Invitations(3) │              [✚ Invite →]    │ │
│        │  │  ├──────────┴────────────────┴───────────────────────────┐  │ │
│        │  │  │                                                         │  │ │
│        │  │  │  ┌───────────────────────────────────────────────────┐ │  │ │
│        │  │  │  │ Name            Email           Role    Team  Status│ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 👤 John Smith   john@acme.com  Admin   Eng   Active│ │  │ │
│        │  │  │  │    Joined 6 months ago                      [⋮]   │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 👤 Jane Doe     jane@acme.com  Member  Eng   Active│ │  │ │
│        │  │  │  │    Joined 3 months ago                      [⋮]   │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 👤 Bob Wilson   bob@acme.com   Member  Sales Active│ │  │ │
│        │  │  │  │    Joined 1 month ago                       [⋮]   │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 👤 Alice Brown  alice@acme.com Approver Design Act.│ │  │ │
│        │  │  │  │    Joined 2 weeks ago                       [⋮]   │ │  │ │
│        │  │  │  └───────────────────────────────────────────────────┘ │  │ │
│        │  │  │                                                         │  │ │
│        │  │  │  Showing 4 of 12 employees          [1][2][3][Next >]  │  │ │
│        │  │  │                                                         │  │ │
│        │  │  └─────────────────────────────────────────────────────────┘  │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
└────────┴────────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Invitations Tab)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [☰] Ubik Enterprise                      🔔 Notifications    👤 John Smith │
├────────┬────────────────────────────────────────────────────────────────────┤
│        │                                                                     │
│  Home  │  👥 Team Management                                                │
│        │                                                                     │
│  Teams │  ┌───────────────────────────────────────────────────────────┐   │
│        │  │  Manage your organization's employees and invitations     │   │
│ Agents │  └───────────────────────────────────────────────────────────┘   │
│        │                                                                     │
│ Config │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│ Logs   │  │  ┌──────────────────┐  ┌──────────────────┐  ┌────────────┐ │ │
│        │  │  │ Total Employees  │  │ Active           │  │ Pending    │ │ │
│Settings│  │  │                  │  │                  │  │ Invites    │ │ │
│        │  │  │       12         │  │       10         │  │      3     │ │ │
│        │  │  └──────────────────┘  └──────────────────┘  └────────────┘ │ │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
│        │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│        │  │  ┌──────────┬────────────────┐  🔍 Search... [Filter ▼]     │ │
│        │  │  │ Employees│ Invitations(3) │              [✚ Invite →]    │ │
│        │  │  ├──────────┴────────────────┴───────────────────────────┐  │ │
│        │  │  │                                                         │  │ │
│        │  │  │  ┌───────────────────────────────────────────────────┐ │  │ │
│        │  │  │  │ Email         Invited By    Role   Team  Status   │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 📧 sarah@acme.com                                 │ │  │ │
│        │  │  │  │    John Smith   Member  Eng   [⏱ Pending]        │ │  │ │
│        │  │  │  │    Sent 2 days ago · Expires in 5 days    [⋮]    │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 📧 mike@acme.com                                  │ │  │ │
│        │  │  │  │    John Smith   Member  Sales [⏱ Pending]        │ │  │ │
│        │  │  │  │    Sent 5 days ago · Expires in 2 days    [⋮]    │ │  │ │
│        │  │  │  ├───────────────────────────────────────────────────┤ │  │ │
│        │  │  │  │ 📧 lisa@acme.com                                  │ │  │ │
│        │  │  │  │    Alice Brown  Approver Design [⚠ Expiring]     │ │  │ │
│        │  │  │  │    Sent 6 days ago · Expires tomorrow     [⋮]    │ │  │ │
│        │  │  │  └───────────────────────────────────────────────────┘ │  │ │
│        │  │  │                                                         │  │ │
│        │  │  │  Showing 3 of 3 pending invitations                     │  │ │
│        │  │  │                                                         │  │ │
│        │  │  └─────────────────────────────────────────────────────────┘  │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
└────────┴────────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Invite Employee Modal)

```
                        ┌───────────────────────────────┐
                        │ ✕                              │
                        │  Invite Employee to Join      │
                        │                                │
                        │  Invite a new employee to your │
                        │  organization via email.       │
                        │                                │
                        │  ┌──────────────────────────┐ │
                        │  │ Email Address *          │ │
                        │  ├──────────────────────────┤ │
                        │  │ colleague@company.com    │ │
                        │  └──────────────────────────┘ │
                        │                                │
                        │  ┌──────────────────────────┐ │
                        │  │ Role *                   │ │
                        │  ├──────────────────────────┤ │
                        │  │ Member              [▼]  │ │
                        │  └──────────────────────────┘ │
                        │  Options: Member, Approver,   │
                        │           Administrator       │
                        │                                │
                        │  ┌──────────────────────────┐ │
                        │  │ Team (Optional)          │ │
                        │  ├──────────────────────────┤ │
                        │  │ Engineering         [▼]  │ │
                        │  └──────────────────────────┘ │
                        │  Options: None, Engineering,  │
                        │           Sales, Design, etc. │
                        │                                │
                        │  ┌──────────────────────────┐ │
                        │  │ Personal Message (Optional)│
                        │  ├──────────────────────────┤ │
                        │  │ Hi! I'd like to invite   │ │
                        │  │ you to join our team...  │ │
                        │  │                          │ │
                        │  └──────────────────────────┘ │
                        │  Max 500 characters           │
                        │                                │
                        │  ℹ️ They'll receive an email  │
                        │     with a secure link to     │
                        │     accept the invitation and │
                        │     set up their account.     │
                        │                                │
                        │  ┌────────┐  ┌──────────────┐│
                        │  │ Cancel │  │ Send Invite  ││
                        │  └────────┘  └──────────────┘│
                        └───────────────────────────────┘
```

---

## Visual Layout (Employee Actions Menu)

```
                        ┌───────────────────────┐
                        │ View Profile          │
                        ├───────────────────────┤
                        │ Edit Details          │
                        ├───────────────────────┤
                        │ Change Team           │
                        ├───────────────────────┤
                        │ Change Role           │
                        ├───────────────────────┤
                        │ View Agent Configs    │
                        ├───────────────────────┤
                        │ View Activity         │
                        ├───────────────────────┤
                        │ Deactivate Employee   │
                        └───────────────────────┘
```

---

## Visual Layout (Invitation Actions Menu)

```
                        ┌───────────────────────┐
                        │ Copy Invitation Link  │
                        ├───────────────────────┤
                        │ Resend Invitation     │
                        ├───────────────────────┤
                        │ View Details          │
                        ├───────────────────────┤
                        │ Cancel Invitation     │
                        └───────────────────────┘
```

---

## Component Breakdown

### Page Header
- **Title:** "👥 Team Management"
- **Description:** Brief explanation of page purpose
- **Styling:** Large heading, subtle background card

### Summary Cards (Dashboard Metrics)
- **Total Employees:** Count of all active employees
- **Active:** Employees with "active" status
- **Pending Invites:** Count of pending invitations

**Styling:**
- Cards in row (stack on mobile)
- Large numbers
- Icons for each metric
- Subtle borders and shadows

### Tab Navigation
- **Tabs:**
  1. Employees (default)
  2. Invitations (with count badge)

**Component:** shadcn/ui Tabs
**Behavior:** Switches content below, updates URL param (`?tab=invitations`)

### Toolbar (Above Table)
- **Search Input:**
  - Placeholder: "Search by name or email..."
  - Live search (debounced 300ms)
  - Clear button (×) when text entered

- **Filter Dropdown:**
  - Filter by: Role, Team, Status
  - Multiple selections allowed
  - Show active filter count badge

- **Invite Button:**
  - Label: "✚ Invite Employee"
  - Style: Primary button
  - Visibility: Admin/Approver only
  - Action: Opens invite modal

### Employees Table

**Columns:**
1. **Name** (with avatar)
   - Full name
   - Join date below
2. **Email**
3. **Role**
   - Badge with role name
4. **Team**
   - Team name or "No team"
5. **Status**
   - Badge: Active (green), Inactive (gray), Suspended (red)
6. **Actions**
   - Dropdown menu (⋮)

**Features:**
- Sortable columns (Name, Email, Role, Team)
- Row click opens employee detail (optional)
- Hover highlights row
- Responsive: Stack columns on mobile

**Pagination:**
- Show 10 per page (configurable)
- Page numbers + Next/Prev
- "Showing X of Y employees"

### Invitations Table

**Columns:**
1. **Email** (with envelope icon)
2. **Invited By**
   - Inviter's name
3. **Role**
   - Badge
4. **Team**
   - Team name or "No team"
5. **Status**
   - Badge: Pending (yellow), Expired (red), Accepted (green), Cancelled (gray)
6. **Details**
   - Sent date
   - Expiration countdown
7. **Actions**
   - Dropdown menu (⋮)

**Features:**
- Filter by status
- Sort by sent date, expiration
- Color-coded expiration warnings:
  - Green: > 3 days
  - Yellow: 1-3 days
  - Red: < 1 day (urgent)

### Invite Employee Modal

**Form Fields:**

1. **Email Address (Required)**
   - Type: email input
   - Validation: Valid email format, not already invited/registered
   - Error messages:
     - "Invalid email format"
     - "This email is already registered"
     - "An invitation is already pending for this email"

2. **Role (Required)**
   - Type: Select dropdown
   - Options:
     - Member (default)
     - Approver
     - Administrator
   - Descriptions shown on hover/select

3. **Team (Optional)**
   - Type: Select dropdown
   - Options: List of org teams + "No team"
   - Default: "No team"

4. **Personal Message (Optional)**
   - Type: Textarea
   - Max length: 500 characters
   - Character counter
   - Placeholder: "Add a personal note to your invitation (optional)"

**Helper Text:**
- Explains invitation email will be sent
- Shows invitation expires in 7 days

**Actions:**
- **Cancel:** Close modal, discard changes
- **Send Invite:** Submit form, create invitation

**States:**
- Default: Form empty
- Validation errors: Inline errors per field
- Submitting: Loading spinner on button
- Success: Toast notification + close modal

### Employee Actions Menu

**Actions:**
1. **View Profile** → Navigate to `/dashboard/employees/{id}`
2. **Edit Details** → Open edit modal or navigate to edit page
3. **Change Team** → Quick select dropdown
4. **Change Role** → Quick select dropdown (admin only)
5. **View Agent Configs** → Navigate to configs page
6. **View Activity** → Navigate to activity logs
7. **Deactivate Employee** → Confirmation dialog (destructive action)

**Permissions:**
- View Profile: All users
- Edit/Change: Admin/Approver only
- Deactivate: Admin only

### Invitation Actions Menu

**Actions:**
1. **Copy Invitation Link** → Copy magic link to clipboard
2. **Resend Invitation** → Send email again (same token)
3. **View Details** → Show invitation details in modal
4. **Cancel Invitation** → Confirmation dialog (sets status to "cancelled")

**Permissions:**
- All actions: Admin/Approver only
- Cancel: Only for pending invitations

---

## User Flows

### Happy Path: Invite New Employee
1. Admin clicks "✚ Invite Employee" button
2. Modal opens with empty form
3. Admin enters email: "sarah@acme.com"
4. Admin selects role: "Member"
5. Admin selects team: "Engineering"
6. Admin adds personal message (optional)
7. Admin clicks "Send Invite"
8. Button shows loading state
9. On success:
   - Modal closes
   - Success toast: "Invitation sent to sarah@acme.com"
   - Invitations tab badge updates (+1)
   - If on Invitations tab, new invitation appears in table

### Search and Filter Employees
1. User enters "john" in search box
2. After 300ms debounce, table filters to show matching employees
3. User clicks "Filter" dropdown
4. User selects "Engineering" team
5. Table now shows only Engineering team members named John
6. User clears search
7. Table shows all Engineering team members

### View Pending Invitations
1. User clicks "Invitations (3)" tab
2. Table switches to show pending invitations
3. User sees 3 pending invitations with expiration warnings
4. One invitation shows "⚠ Expiring" (expires tomorrow)
5. User hovers over warning to see tooltip

### Resend Invitation
1. User navigates to Invitations tab
2. User clicks actions menu (⋮) for pending invitation
3. User selects "Resend Invitation"
4. Confirmation toast: "Resending invitation to sarah@acme.com..."
5. On success:
   - Toast: "Invitation email resent successfully"
   - "Sent" date updates in table

### Cancel Invitation
1. User clicks actions menu (⋮) for pending invitation
2. User selects "Cancel Invitation"
3. Confirmation dialog appears:
   - Title: "Cancel invitation?"
   - Message: "This will permanently cancel the invitation to sarah@acme.com. They will not be able to use this link."
   - Actions: "Keep Invitation" (default) | "Cancel Invitation" (destructive)
4. User confirms cancellation
5. On success:
   - Toast: "Invitation cancelled"
   - Invitation status changes to "Cancelled" (gray badge)
   - Invitation moves to bottom or separate cancelled section

### Change Employee Team
1. User clicks actions menu (⋮) for employee
2. User selects "Change Team"
3. Inline dropdown appears (or modal)
4. User selects new team: "Sales"
5. User confirms (if modal)
6. On success:
   - Toast: "Employee moved to Sales team"
   - Team column updates

---

## API Integration

### Get Employees
**Endpoint:** `GET /employees?page=1&limit=10&search=john&team=eng&role=member&status=active`

**Response (200 OK):**
```json
{
  "employees": [
    {
      "id": "uuid",
      "full_name": "John Smith",
      "email": "john@acme.com",
      "role": {
        "id": "uuid",
        "name": "Administrator"
      },
      "team": {
        "id": "uuid",
        "name": "Engineering"
      },
      "status": "active",
      "created_at": "2024-06-01T10:00:00Z"
    }
  ],
  "total": 12,
  "page": 1,
  "limit": 10
}
```

### Get Invitations
**Endpoint:** `GET /invitations?page=1&limit=10&status=pending`

**Response (200 OK):**
```json
{
  "invitations": [
    {
      "id": "uuid",
      "email": "sarah@acme.com",
      "role": {
        "id": "uuid",
        "name": "Member"
      },
      "team": {
        "id": "uuid",
        "name": "Engineering"
      },
      "inviter": {
        "id": "uuid",
        "full_name": "John Smith",
        "email": "john@acme.com"
      },
      "status": "pending",
      "expires_at": "2024-12-17T23:59:59Z",
      "created_at": "2024-12-10T10:00:00Z"
    }
  ],
  "total": 3,
  "page": 1,
  "limit": 10
}
```

### Create Invitation
**Endpoint:** `POST /invitations`

**Request:**
```json
{
  "email": "sarah@acme.com",
  "role_id": "uuid",
  "team_id": "uuid",
  "message": "Hi Sarah, welcome to the team!"
}
```

**Response (201 Created):**
```json
{
  "invitation": {
    "id": "uuid",
    "email": "sarah@acme.com",
    "token": "secure-token",
    "invitation_url": "https://app.ubik.io/accept-invite?token=secure-token",
    "expires_at": "2024-12-17T23:59:59Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` - Validation errors
- `409 Conflict` - Email already registered or invitation pending
- `403 Forbidden` - User doesn't have permission to invite

### Resend Invitation
**Endpoint:** `POST /invitations/{id}/resend`

**Response (200 OK):**
```json
{
  "message": "Invitation email resent successfully",
  "resent_at": "2024-12-10T14:30:00Z"
}
```

### Cancel Invitation
**Endpoint:** `DELETE /invitations/{id}`

**Response (200 OK):**
```json
{
  "message": "Invitation cancelled successfully"
}
```

### Update Employee Team
**Endpoint:** `PATCH /employees/{id}`

**Request:**
```json
{
  "team_id": "uuid"
}
```

**Response (200 OK):**
```json
{
  "employee": {
    "id": "uuid",
    "team": {
      "id": "uuid",
      "name": "Sales"
    }
  }
}
```

---

## State Management

### Table State
```typescript
type TableState = {
  tab: 'employees' | 'invitations';
  search: string;
  filters: {
    role?: string[];
    team?: string[];
    status?: string[];
  };
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
  sort: {
    column: string;
    direction: 'asc' | 'desc';
  };
};
```

### Modal State
```typescript
type ModalState = {
  isOpen: boolean;
  mode: 'invite' | 'edit' | 'view' | null;
  data?: Employee | Invitation;
};
```

### Query Management
- Use React Query for API calls
- Automatic refetching on tab switch
- Optimistic updates for quick actions
- Cache invalidation on mutations

---

## Accessibility (WCAG AA)

### Keyboard Navigation
- Tab through all interactive elements
- Arrow keys to navigate table rows (optional)
- Space/Enter to open actions menu
- Escape to close modal/menu
- Search box accessible via /

### Screen Reader Support
- Table has proper headers
- Row counts announced: "Showing 4 of 12 employees"
- Status badges have aria-labels
- Actions menu labeled properly
- Modal title and description
- Form labels and errors

### Visual Design
- High contrast (4.5:1)
- Clear focus indicators
- Status badges with icons (not color alone)
- Expiration warnings with text (not color alone)
- Large touch targets (44px)

---

## Responsive Design

### Mobile (< 640px)
- Stack summary cards vertically
- Hide some table columns (show essential: Name, Role, Actions)
- Full-width search and buttons
- Actions menu full-screen drawer
- Modal full-screen

### Tablet (640px - 1024px)
- 2-column summary cards
- Show most table columns
- Slightly larger touch targets

### Desktop (> 1024px)
- 3-column summary cards
- Full table with all columns
- Hover states
- Inline actions

---

## Implementation Notes

### Technologies
- **Framework:** Next.js 14 (App Router)
- **Table:** shadcn/ui Table + TanStack Table
- **Modal:** shadcn/ui Dialog
- **Dropdown:** shadcn/ui DropdownMenu
- **Tabs:** shadcn/ui Tabs
- **Forms:** React Hook Form + Zod
- **Data Fetching:** React Query
- **Styling:** Tailwind CSS

### Search Debouncing
```typescript
import { useDebouncedValue } from '@/hooks/useDebouncedValue';

const [search, setSearch] = useState('');
const debouncedSearch = useDebouncedValue(search, 300);

// Use debouncedSearch in API call
useQuery(['employees', debouncedSearch], () =>
  api.getEmployees({ search: debouncedSearch })
);
```

### Optimistic Updates
```typescript
const cancelInvitation = useMutation({
  mutationFn: api.cancelInvitation,
  onMutate: async (invitationId) => {
    // Cancel outgoing refetches
    await queryClient.cancelQueries(['invitations']);

    // Snapshot previous value
    const previous = queryClient.getQueryData(['invitations']);

    // Optimistically update
    queryClient.setQueryData(['invitations'], old =>
      old.map(inv =>
        inv.id === invitationId
          ? { ...inv, status: 'cancelled' }
          : inv
      )
    );

    return { previous };
  },
  onError: (err, invitationId, context) => {
    // Rollback on error
    queryClient.setQueryData(['invitations'], context.previous);
  },
  onSettled: () => {
    // Refetch to sync with server
    queryClient.invalidateQueries(['invitations']);
  }
});
```

---

## Testing Checklist

### Unit Tests
- [ ] Search debouncing
- [ ] Filter logic
- [ ] Pagination calculations
- [ ] Sort logic
- [ ] Form validation

### Integration Tests
- [ ] Load employees list
- [ ] Search employees
- [ ] Filter by role/team
- [ ] Create invitation
- [ ] Resend invitation
- [ ] Cancel invitation
- [ ] Change employee team
- [ ] Pagination

### E2E Tests (Playwright)
- [ ] Full invite flow
- [ ] Search and filter
- [ ] Tab switching
- [ ] Cancel invitation flow
- [ ] Actions menu interactions
- [ ] Modal form submission

### Accessibility Tests
- [ ] Keyboard navigation
- [ ] Screen reader compatibility
- [ ] Focus management
- [ ] Color contrast

---

## Related Pages
- **Previous:** `/onboarding` (Onboarding Wizard)
- **Related:** `/dashboard/employees/{id}` (Employee Detail)
- **Related:** `/dashboard/invitations` (Full Invitations Page - if separate)

---

## Design System References
- **shadcn/ui Table:** https://ui.shadcn.com/docs/components/table
- **shadcn/ui Dialog:** https://ui.shadcn.com/docs/components/dialog
- **shadcn/ui DropdownMenu:** https://ui.shadcn.com/docs/components/dropdown-menu
- **shadcn/ui Tabs:** https://ui.shadcn.com/docs/components/tabs
- **shadcn/ui Badge:** https://ui.shadcn.com/docs/components/badge
- **TanStack Table:** https://tanstack.com/table/v8

---

## Notes

**Why Combine Employees & Invitations?**
- Related functionality (team management)
- Reduces navigation complexity
- Common use case: invite while viewing team
- Easy comparison of team composition

**Why Tabs Instead of Separate Pages?**
- Faster switching (no full page reload)
- Maintains filter/search state
- Visual connection between concepts
- Better UX for admin workflows

**Future Enhancements:**
- Bulk actions (invite multiple, change teams)
- CSV import for bulk invitations
- Invitation templates
- Team-based permissions (team leads can invite to their team only)
- Activity timeline per employee
- Usage analytics per employee
