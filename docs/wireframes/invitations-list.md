# Invitations List Page Wireframe

**Route:** `/dashboard/invitations`
**Access:** Authenticated (requires session)
**Permissions:** Admin/Approver only
**Components:** Table, Badge, Button, Tabs, Alert, DropdownMenu
**Layout:** Full dashboard layout with sidebar and main content

---

## Page Purpose

Dedicated page for viewing and managing all invitations (pending, accepted, expired, cancelled). Provides detailed invitation history, bulk actions, filtering by status, and analytics. This is an alternative/enhanced view compared to the Invitations tab in Team Management.

---

## Visual Layout (All Invitations View)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [☰] Ubik Enterprise                      🔔 Notifications    👤 John Smith │
├────────┬────────────────────────────────────────────────────────────────────┤
│        │                                                                     │
│  Home  │  📨 Invitation Management                                          │
│        │                                                                     │
│  Teams │  ┌───────────────────────────────────────────────────────────┐   │
│        │  │  Track and manage all employee invitations                │   │
│ Agents │  └───────────────────────────────────────────────────────────┘   │
│        │                                                                     │
│ Config │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│ Logs   │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │ │
│        │  │  │ Total Sent   │  │ Pending      │  │ Accepted         │  │ │
│Settings│  │  │              │  │              │  │                  │  │ │
│        │  │  │     27       │  │      8       │  │       15         │  │ │
│        │  │  └──────────────┘  └──────────────┘  └──────────────────┘  │ │
│        │  │                                                               │ │
│        │  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────────┐  │ │
│        │  │  │ Expired      │  │ Cancelled    │  │ Acceptance Rate  │  │ │
│        │  │  │              │  │              │  │                  │  │ │
│        │  │  │      3       │  │      1       │  │      56%         │  │ │
│        │  │  └──────────────┘  └──────────────┘  └──────────────────┘  │ │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
│        │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│        │  │  ┌─────┬────────┬─────────┬─────────┬──────────┐            │ │
│        │  │  │ All │Pending │Accepted │ Expired │Cancelled │            │ │
│        │  │  └─────┴────────┴─────────┴─────────┴──────────┘            │ │
│        │  │                                                               │ │
│        │  │  🔍 Search by email...        [Filter ▼]  [✚ Invite →]     │ │
│        │  │                                                               │ │
│        │  │  ┌───────────────────────────────────────────────────────┐  │ │
│        │  │  │ ☐  Email         Invited By   Role   Status   Actions │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 sarah@acme.com                                  │  │ │
│        │  │  │     John Smith    Member   [⏱ Pending]         [⋮]   │  │ │
│        │  │  │     Engineering                                       │  │ │
│        │  │  │     Sent 2 days ago · Expires in 5 days               │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 mike@acme.com                                   │  │ │
│        │  │  │     John Smith    Member   [⏱ Pending]         [⋮]   │  │ │
│        │  │  │     Sales                                             │  │ │
│        │  │  │     Sent 5 days ago · Expires in 2 days               │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 lisa@acme.com                                   │  │ │
│        │  │  │     Alice Brown   Approver [⚠ Expiring Soon]   [⋮]   │  │ │
│        │  │  │     Design                                            │  │ │
│        │  │  │     Sent 6 days ago · Expires tomorrow                │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 tom@acme.com                                    │  │ │
│        │  │  │     Bob Wilson    Member   [✓ Accepted]         [⋮]   │  │ │
│        │  │  │     Engineering                                       │  │ │
│        │  │  │     Accepted 1 week ago                               │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 old@company.com                                 │  │ │
│        │  │  │     John Smith    Member   [✗ Expired]          [⋮]   │  │ │
│        │  │  │     Sales                                             │  │ │
│        │  │  │     Expired 3 days ago                                │  │ │
│        │  │  └───────────────────────────────────────────────────────┘  │ │
│        │  │                                                               │ │
│        │  │  [☐ Select all]  [↻ Resend Selected]  [✗ Cancel Selected]   │ │
│        │  │                                                               │ │
│        │  │  Showing 5 of 27 invitations         [1][2][3][Next >]       │ │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
└────────┴────────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Pending Tab - Filtered View)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ [☰] Ubik Enterprise                      🔔 Notifications    👤 John Smith │
├────────┬────────────────────────────────────────────────────────────────────┤
│        │                                                                     │
│  Home  │  📨 Invitation Management                                          │
│        │                                                                     │
│  Teams │  ┌───────────────────────────────────────────────────────────┐   │
│        │  │  Track and manage all employee invitations                │   │
│ Agents │  └───────────────────────────────────────────────────────────┘   │
│        │                                                                     │
│ Config │  [Statistics Cards]                                                │
│        │                                                                     │
│ Logs   │  ┌─────────────────────────────────────────────────────────────┐ │
│        │  │                                                               │ │
│Settings│  │  ┌─────┬────────┬─────────┬─────────┬──────────┐            │ │
│        │  │  │ All │Pending │Accepted │ Expired │Cancelled │            │ │
│        │  │  └─────┴────────┴─────────┴─────────┴──────────┘            │ │
│        │  │           (8)                                                 │ │
│        │  │                                                               │ │
│        │  │  🔍 Search...        [Sort: Expires Soon ▼]  [✚ Invite →]   │ │
│        │  │                                                               │ │
│        │  │  ⚠️ 3 invitations expiring within 24 hours [Resend All]     │ │
│        │  │                                                               │ │
│        │  │  ┌───────────────────────────────────────────────────────┐  │ │
│        │  │  │ ☐  Email         Invited By   Role   Status   Actions │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 lisa@acme.com                                   │  │ │
│        │  │  │     Alice Brown   Approver [⚠ Expires Tomorrow] [⋮]  │  │ │
│        │  │  │     Design                                            │  │ │
│        │  │  │     Sent 6 days ago                                   │  │ │
│        │  │  │     🔗 https://app.ubik.io/accept?token=...  [Copy]   │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 carlos@acme.com                                 │  │ │
│        │  │  │     John Smith    Member   [⚠ Expires Tomorrow] [⋮]  │  │ │
│        │  │  │     Engineering                                       │  │ │
│        │  │  │     Sent 6 days ago                                   │  │ │
│        │  │  │     🔗 https://app.ubik.io/accept?token=...  [Copy]   │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 priya@acme.com                                  │  │ │
│        │  │  │     Bob Wilson    Member   [⚠ Expires Tomorrow] [⋮]  │  │ │
│        │  │  │     Sales                                             │  │ │
│        │  │  │     Sent 6 days ago                                   │  │ │
│        │  │  │     🔗 https://app.ubik.io/accept?token=...  [Copy]   │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ☐  📧 sarah@acme.com                                  │  │ │
│        │  │  │     John Smith    Member   [Expires in 5 days]  [⋮]  │  │ │
│        │  │  │     Engineering                                       │  │ │
│        │  │  │     Sent 2 days ago                                   │  │ │
│        │  │  ├───────────────────────────────────────────────────────┤  │ │
│        │  │  │ ... (4 more pending invitations)                      │  │ │
│        │  │  └───────────────────────────────────────────────────────┘  │ │
│        │  │                                                               │ │
│        │  │  [☐ Select all]  [↻ Resend Selected]  [✗ Cancel Selected]   │ │
│        │  │                                                               │ │
│        │  └─────────────────────────────────────────────────────────────┘ │
│        │                                                                     │
└────────┴────────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Invitation Details Modal)

```
                        ┌──────────────────────────────────┐
                        │ ✕                                 │
                        │  Invitation Details               │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Status                      │ │
                        │  │ [⏱ Pending]                 │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Recipient                   │ │
                        │  │ sarah@acme.com              │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Role & Team                 │ │
                        │  │ Member · Engineering        │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Invited By                  │ │
                        │  │ John Smith                  │ │
                        │  │ john@acme.com               │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Timeline                    │ │
                        │  │ Sent: Dec 8, 2024 10:00 AM  │ │
                        │  │ Expires: Dec 15, 2024       │ │
                        │  │ (In 5 days)                 │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Invitation Link             │ │
                        │  │ https://app.ubik.io/        │ │
                        │  │ accept?token=abc123...      │ │
                        │  │                [Copy Link]  │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌─────────────────────────────┐ │
                        │  │ Personal Message            │ │
                        │  │ "Hi Sarah, welcome to the   │ │
                        │  │  team! Looking forward to   │ │
                        │  │  working with you."         │ │
                        │  └─────────────────────────────┘ │
                        │                                   │
                        │  ┌──────────┐  ┌──────────────┐ │
                        │  │ Close    │  │ Resend Email │ │
                        │  └──────────┘  └──────────────┘ │
                        └──────────────────────────────────┘
```

---

## Component Breakdown

### Page Header
- **Title:** "📨 Invitation Management"
- **Description:** Brief explanation
- **Action Button:** "✚ Invite Employee" (primary, top-right)

### Statistics Dashboard (6 Cards)

**Metrics:**
1. **Total Sent**
   - Count of all invitations ever sent
   - Icon: 📨

2. **Pending**
   - Active pending invitations
   - Icon: ⏱
   - Color: Yellow

3. **Accepted**
   - Successfully accepted invitations
   - Icon: ✅
   - Color: Green

4. **Expired**
   - Invitations past expiration date
   - Icon: ⌛
   - Color: Red

5. **Cancelled**
   - Manually cancelled invitations
   - Icon: ✗
   - Color: Gray

6. **Acceptance Rate**
   - Accepted / (Accepted + Expired + Cancelled) %
   - Icon: 📊
   - Color: Blue
   - Shows trend arrow (↑ or ↓)

**Layout:**
- 3 cards per row on desktop
- Stack vertically on mobile
- Click card to filter table by that status

### Status Tabs

**Tabs:**
1. **All** (default) - Show all invitations
2. **Pending (8)** - Active invitations with count badge
3. **Accepted** - Successfully accepted
4. **Expired** - Past expiration
5. **Cancelled** - Manually cancelled

**Component:** shadcn/ui Tabs
**Behavior:**
- Changes table filter
- Updates URL param (`?status=pending`)
- Badge shows count for pending

### Toolbar

**Components:**
1. **Search Input**
   - Placeholder: "Search by email..."
   - Live search (debounced 300ms)
   - Clear button (×)

2. **Sort Dropdown** (Pending tab only)
   - Options:
     - Expires Soon (default for Pending)
     - Recently Sent
     - Oldest First
   - Icon shows current sort

3. **Filter Dropdown**
   - Filter by:
     - Role (Member, Approver, Admin)
     - Team
     - Inviter
   - Multiple selections
   - Active filter badge count

4. **Invite Button**
   - Label: "✚ Invite Employee"
   - Primary style
   - Opens invite modal (same as Team Management)

### Alert Banner (Conditional)

**Shown when:** 3+ invitations expiring within 24 hours (on Pending tab)

**Content:**
- Icon: ⚠️
- Message: "X invitations expiring within 24 hours"
- Action: "Resend All" button
- Dismissable (×)
- Color: Yellow/warning

**Behavior:**
- Appears at top of table
- Click "Resend All" → Confirmation dialog → Resend all expiring invitations
- Dismiss → Hide until next page load

### Invitations Table

**Columns:**
1. **Checkbox** (for bulk actions)
2. **Email** (with envelope icon)
   - Primary identifier
   - Bold font
3. **Invited By**
   - Inviter's name
   - Secondary text
4. **Role**
   - Badge with role name
5. **Team**
   - Team name or "No team"
   - Shown below email on mobile
6. **Status**
   - Badge with color coding:
     - Pending: Yellow ⏱
     - Expiring Soon: Red ⚠ (< 24 hours)
     - Accepted: Green ✓
     - Expired: Red ✗
     - Cancelled: Gray ✗
7. **Details**
   - Sent date
   - Expiration info or acceptance date
8. **Actions**
   - Dropdown menu (⋮)

**Features:**
- Sortable: Sent Date, Expiration, Email
- Expandable rows (optional):
  - Click row to expand
  - Shows: Invitation link, personal message, detailed timeline
- Bulk selection via checkboxes
- Color-coded expiration warnings
- Copy invitation link button (visible on hover or expand)

**Row States:**
- Default: White background
- Hover: Light gray
- Selected: Light blue
- Expiring Soon: Light red background

### Bulk Actions Bar

**Shown when:** 1+ rows selected

**Components:**
- "X selected" label
- "Select all" link (if not all selected)
- "Deselect all" link (if selections exist)

**Actions:**
1. **Resend Selected**
   - Icon: ↻
   - Confirmation: "Resend X invitations?"
   - Only for pending/expired

2. **Cancel Selected**
   - Icon: ✗
   - Confirmation: "Cancel X invitations?"
   - Destructive style
   - Only for pending

**Position:** Sticky at bottom of viewport (on mobile) or above pagination (desktop)

### Action Menu (Per Row)

**Actions (vary by status):**

**For Pending:**
1. Copy Invitation Link
2. View Details (opens modal)
3. Resend Invitation
4. Cancel Invitation (destructive)

**For Accepted:**
1. View Details
2. View Employee Profile

**For Expired:**
1. View Details
2. Resend (creates new invitation)

**For Cancelled:**
1. View Details

### Pagination

- Show 10/20/50 per page (user configurable)
- Page numbers (1, 2, 3, ..., N)
- Prev/Next buttons
- "Showing X-Y of Z invitations"
- Keyboard shortcuts: ← → for prev/next

---

## User Flows

### View Invitations by Status
1. User lands on page (default: All tab)
2. User sees statistics dashboard
3. User clicks "Pending (8)" tab
4. Table filters to show 8 pending invitations
5. Alert banner appears: "3 invitations expiring within 24 hours"
6. Table sorted by expiration (urgent first)
7. User sees expiring invitations at top with red badges

### Bulk Resend Expiring Invitations
1. User is on Pending tab
2. Alert banner shown: "3 expiring within 24 hours [Resend All]"
3. User clicks "Resend All"
4. Confirmation dialog:
   - Title: "Resend 3 expiring invitations?"
   - List of emails
   - Actions: Cancel | Resend All
5. User confirms
6. Loading state
7. On success:
   - Toast: "3 invitations resent successfully"
   - Alert banner dismisses
   - "Sent" dates update in table
   - Expiration dates extend +7 days

### Cancel Multiple Invitations
1. User selects 3 pending invitations via checkboxes
2. Bulk actions bar appears: "3 selected"
3. User clicks "✗ Cancel Selected"
4. Confirmation dialog:
   - Title: "Cancel 3 invitations?"
   - Warning: "Recipients will no longer be able to use their invitation links"
   - List of emails
   - Actions: Keep Invitations | Cancel Invitations (red)
5. User confirms
6. On success:
   - Toast: "3 invitations cancelled"
   - Rows move to Cancelled tab
   - If on Pending tab, rows disappear
   - Statistics update (-3 Pending, +3 Cancelled)

### View Invitation Details
1. User clicks actions menu (⋮) for invitation
2. User selects "View Details"
3. Modal opens showing:
   - Status badge
   - Recipient email
   - Role & team
   - Inviter info
   - Timeline (sent, expires/accepted)
   - Full invitation link
   - Personal message (if any)
4. User clicks "Copy Link" to copy invitation URL
5. Toast: "Invitation link copied to clipboard"
6. User clicks "Resend Email" (if pending/expired)
7. Confirmation, then success toast

### Sort Pending by Expiration
1. User is on Pending tab
2. User clicks sort dropdown (default: "Expires Soon")
3. Dropdown shows options:
   - ✓ Expires Soon
   - Recently Sent
   - Oldest First
4. User selects "Recently Sent"
5. Table re-sorts, most recent at top
6. URL updates: `?status=pending&sort=recent`

### Filter by Team
1. User clicks "Filter" dropdown
2. User expands "Team" section
3. User checks "Engineering" and "Sales"
4. Filter applies immediately
5. Table shows only invitations for those teams
6. Filter badge shows "2 filters"
7. User can clear all filters via "Clear" link in dropdown

---

## API Integration

### Get Invitations (with filters)
**Endpoint:** `GET /invitations?page=1&limit=10&status=pending&sort=expires_asc&search=sarah&team=eng&role=member`

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
      "token": "secure-token",
      "invitation_url": "https://app.ubik.io/accept-invite?token=secure-token",
      "message": "Welcome to the team!",
      "expires_at": "2024-12-15T23:59:59Z",
      "created_at": "2024-12-08T10:00:00Z",
      "updated_at": "2024-12-08T10:00:00Z"
    }
  ],
  "total": 8,
  "page": 1,
  "limit": 10,
  "statistics": {
    "total_sent": 27,
    "pending": 8,
    "accepted": 15,
    "expired": 3,
    "cancelled": 1,
    "acceptance_rate": 56.0
  }
}
```

### Bulk Resend Invitations
**Endpoint:** `POST /invitations/bulk/resend`

**Request:**
```json
{
  "invitation_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Response (200 OK):**
```json
{
  "resent_count": 3,
  "invitations": [
    {
      "id": "uuid1",
      "email": "sarah@acme.com",
      "expires_at": "2024-12-22T23:59:59Z",
      "resent_at": "2024-12-15T10:00:00Z"
    }
  ]
}
```

### Bulk Cancel Invitations
**Endpoint:** `POST /invitations/bulk/cancel`

**Request:**
```json
{
  "invitation_ids": ["uuid1", "uuid2", "uuid3"]
}
```

**Response (200 OK):**
```json
{
  "cancelled_count": 3,
  "message": "3 invitations cancelled successfully"
}
```

### Get Statistics
**Endpoint:** `GET /invitations/statistics`

**Response (200 OK):**
```json
{
  "total_sent": 27,
  "pending": 8,
  "accepted": 15,
  "expired": 3,
  "cancelled": 1,
  "acceptance_rate": 56.0,
  "expiring_soon": 3,
  "trend": {
    "sent_last_7_days": 5,
    "accepted_last_7_days": 2
  }
}
```

---

## State Management

### URL State (Query Params)
- `?status=pending` - Active tab
- `?sort=expires_asc` - Sort order
- `?search=sarah` - Search query
- `?team=eng&role=member` - Filters
- `?page=2` - Pagination

**Benefits:**
- Shareable URLs
- Browser back/forward works
- State persists on refresh

### Table State
```typescript
type TableState = {
  selectedRows: Set<string>;
  expandedRows: Set<string>;
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
};
```

### Filter State
```typescript
type FilterState = {
  search: string;
  status: InvitationStatus | 'all';
  sort: 'expires_asc' | 'expires_desc' | 'recent' | 'oldest';
  teams: string[];
  roles: string[];
  inviters: string[];
};
```

---

## Accessibility (WCAG AA)

### Keyboard Navigation
- Tab through all interactive elements
- Space to select checkbox
- Enter to open action menu
- Arrow keys in table (optional)
- Keyboard shortcuts:
  - `/` to focus search
  - `←` `→` for pagination
  - `Esc` to close modal/menu

### Screen Reader Support
- Table has caption: "Invitations list"
- Row count announced
- Status badges have aria-labels
- Bulk actions bar announced
- Statistics cards have proper labels
- Alert banner is live region

### Visual Design
- High contrast (4.5:1)
- Clear focus indicators
- Status badges with icons
- Expiration warnings with text + color
- Large touch targets (44px)

---

## Responsive Design

### Mobile (< 640px)
- Stack statistics cards (2 per row)
- Hide some table columns (show: Email, Status, Actions)
- Expandable rows show hidden details
- Bulk actions bar fixed at bottom
- Full-screen modals
- Simplified filters (drawer)

### Tablet (640px - 1024px)
- 3 stats cards per row
- Show most table columns
- Inline modals

### Desktop (> 1024px)
- 6 stats cards in 2 rows (3 per row) or single row
- All table columns visible
- Hover states
- Larger modals

---

## Implementation Notes

### Technologies
- **Framework:** Next.js 14 (App Router)
- **Table:** TanStack Table v8
- **URL State:** nuqs or next/navigation
- **Data Fetching:** React Query
- **Styling:** Tailwind CSS
- **Components:** shadcn/ui

### URL State Management
```typescript
import { useQueryState } from 'nuqs';

const [status, setStatus] = useQueryState('status', {
  defaultValue: 'all'
});

const [sort, setSort] = useQueryState('sort', {
  defaultValue: 'expires_asc'
});

// Changes update URL automatically
setStatus('pending'); // → ?status=pending
```

### Bulk Action Optimistic Updates
```typescript
const bulkResend = useMutation({
  mutationFn: api.bulkResendInvitations,
  onMutate: async (ids) => {
    await queryClient.cancelQueries(['invitations']);
    const previous = queryClient.getQueryData(['invitations']);

    // Optimistically update expiration dates
    queryClient.setQueryData(['invitations'], old => ({
      ...old,
      invitations: old.invitations.map(inv =>
        ids.includes(inv.id)
          ? {
              ...inv,
              expires_at: addDays(new Date(), 7),
              created_at: new Date()
            }
          : inv
      )
    }));

    return { previous };
  },
  onError: (err, ids, context) => {
    queryClient.setQueryData(['invitations'], context.previous);
  },
  onSettled: () => {
    queryClient.invalidateQueries(['invitations']);
  }
});
```

### Statistics Polling
```typescript
// Poll statistics every 60 seconds
const { data: stats } = useQuery({
  queryKey: ['invitation-statistics'],
  queryFn: api.getInvitationStatistics,
  refetchInterval: 60000,
  staleTime: 30000
});
```

---

## Testing Checklist

### Unit Tests
- [ ] URL state management
- [ ] Filter logic
- [ ] Sort logic
- [ ] Expiration calculations
- [ ] Bulk selection logic

### Integration Tests
- [ ] Load invitations by status
- [ ] Search and filter
- [ ] Sort invitations
- [ ] Bulk resend
- [ ] Bulk cancel
- [ ] View details modal
- [ ] Statistics accuracy

### E2E Tests (Playwright)
- [ ] Tab navigation
- [ ] Search and filter
- [ ] Bulk actions flow
- [ ] Copy invitation link
- [ ] Resend invitation
- [ ] Cancel invitation
- [ ] Expiring soon alert

### Accessibility Tests
- [ ] Keyboard navigation
- [ ] Screen reader compatibility
- [ ] Focus management
- [ ] Color contrast
- [ ] Touch target sizes

---

## Performance Considerations

### Optimizations
1. **Virtual Scrolling** - For large invitation lists (100+)
2. **Debounced Search** - 300ms delay
3. **Optimistic Updates** - Instant UI feedback
4. **Statistics Caching** - 30s stale time
5. **Pagination** - Limit results per page
6. **Column Virtualization** - On mobile, hide columns

### Caching Strategy
```typescript
// Aggressive caching for statistics
{
  staleTime: 30000,  // 30 seconds
  cacheTime: 300000  // 5 minutes
}

// Fresh data for invitations list
{
  staleTime: 5000,   // 5 seconds
  cacheTime: 60000   // 1 minute
}
```

---

## Related Pages
- **Previous:** `/dashboard/teams` (Team Management)
- **Related:** `/dashboard/employees/{id}` (Employee Profile)
- **Related:** `/accept-invite?token=xyz` (Invitation Acceptance)

---

## Design System References
- **shadcn/ui Table:** https://ui.shadcn.com/docs/components/table
- **shadcn/ui Tabs:** https://ui.shadcn.com/docs/components/tabs
- **shadcn/ui Badge:** https://ui.shadcn.com/docs/components/badge
- **shadcn/ui Alert:** https://ui.shadcn.com/docs/components/alert
- **TanStack Table:** https://tanstack.com/table/v8

---

## Notes

**Why Separate Page vs Tab?**
- More screen space for detailed view
- Advanced filtering and sorting
- Bulk actions more prominent
- Statistics dashboard
- Can still be accessed via tab in Team Management for quick view

**Why Show All Statuses?**
- Historical record keeping
- Audit trail for compliance
- Identify patterns (why invitations expire?)
- Learn from successful invitations

**Future Enhancements:**
- Export invitations to CSV
- Analytics dashboard (charts, trends)
- Invitation templates
- Scheduled reminders before expiration
- Webhook notifications
- API for external integrations
