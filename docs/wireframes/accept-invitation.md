# Accept Invitation Page Wireframe

**Route:** `/accept-invite?token={invitation_token}`
**Access:** Public (unauthenticated)
**Components:** Card, Form, Input, Button, Alert, Badge
**Layout:** Centered card on full-page background

---

## Page Purpose

Allows invited users to accept email invitation and join an organization. New users create their password and immediately become part of the organization. Validates invitation token, displays invitation details, and handles registration.

---

## Visual Layout (Valid Token - New User)

```
┌─────────────────────────────────────────────────────────────────┐
│                         UBIK ENTERPRISE                          │
│                     AI Agent Management Platform                 │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │                 📨 You've Been Invited!                  │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ Invitation Details                                │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Organization                                      │ │   │
│   │   │  ┌──────────────────────────────────────────┐     │ │   │
│   │   │  │  🏢 Acme Corporation                     │     │ │   │
│   │   │  │     https://acme-corp.ubik.io            │     │ │   │
│   │   │  └──────────────────────────────────────────┘     │ │   │
│   │   │                                                    │ │   │
│   │   │  Invited By                                        │ │   │
│   │   │  ┌──────────────────────────────────────────┐     │ │   │
│   │   │  │  👤 John Smith (john@acme.com)           │     │ │   │
│   │   │  │     Administrator                         │     │ │   │
│   │   │  └──────────────────────────────────────────┘     │ │   │
│   │   │                                                    │ │   │
│   │   │  Your Role                                         │ │   │
│   │   │  ┌──────────────────────────────────────────┐     │ │   │
│   │   │  │  🎯 Member                               │     │ │   │
│   │   │  │     Standard team member permissions      │     │ │   │
│   │   │  └──────────────────────────────────────────┘     │ │   │
│   │   │                                                    │ │   │
│   │   │  Team Assignment                                   │ │   │
│   │   │  ┌──────────────────────────────────────────┐     │ │   │
│   │   │  │  👥 Engineering                          │     │ │   │
│   │   │  └──────────────────────────────────────────┘     │ │   │
│   │   │                                                    │ │   │
│   │   │  Invitation Expires                                │ │   │
│   │   │  ┌──────────────────────────────────────────┐     │ │   │
│   │   │  │  ⏰ In 5 days (Dec 15, 2024)             │     │ │   │
│   │   │  └──────────────────────────────────────────┘     │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ Set Your Password                                 │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Your Email                                        │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ jane@acme.com                    [✓ Verified]│ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │  ℹ️ This email is associated with your invitation │ │   │
│   │   │                                                    │ │   │
│   │   │  Full Name *                                       │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ Jane Doe                                     │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  Password *                                        │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ ●●●●●●●●                           [👁]       │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │  [Strength: ▓▓▓▓▓▓░░ Strong]                      │ │   │
│   │   │  ℹ️ Must be at least 8 characters with mix of     │ │   │
│   │   │     letters, numbers, and symbols                 │ │   │
│   │   │                                                    │ │   │
│   │   │  Confirm Password *                                │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ ●●●●●●●●                           [👁]       │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ ☑ I agree to the Terms of Service and            │ │   │
│   │   │   Privacy Policy                                  │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │          Accept Invitation & Join Team           │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │             Already have an account? Log in              │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Invalid/Expired Token)

```
┌─────────────────────────────────────────────────────────────────┐
│                         UBIK ENTERPRISE                          │
│                     AI Agent Management Platform                 │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │                  ❌ Invalid Invitation                   │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │                                                    │ │   │
│   │   │  This invitation link is invalid or has expired.   │ │   │
│   │   │                                                    │ │   │
│   │   │  Possible reasons:                                 │ │   │
│   │   │  • The invitation has expired (7 days)             │ │   │
│   │   │  • The invitation was cancelled by the sender      │ │   │
│   │   │  • The invitation has already been accepted        │ │   │
│   │   │  • The link is malformed or incorrect              │ │   │
│   │   │                                                    │ │   │
│   │   │  ──────────────────────────────────────────────   │ │   │
│   │   │                                                    │ │   │
│   │   │  What to do next:                                  │ │   │
│   │   │                                                    │ │   │
│   │   │  1. Contact the person who invited you and         │ │   │
│   │   │     request a new invitation                       │ │   │
│   │   │                                                    │ │   │
│   │   │  2. If you already have an account, please log in  │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │      ┌──────────────────┐    ┌──────────────────────┐  │   │
│   │      │  Go to Login     │    │  Contact Support     │  │   │
│   │      └──────────────────┘    └──────────────────────┘  │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Already Accepted)

```
┌─────────────────────────────────────────────────────────────────┐
│                         UBIK ENTERPRISE                          │
│                     AI Agent Management Platform                 │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │             ✅ Invitation Already Accepted               │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │                                                    │ │   │
│   │   │  This invitation has already been accepted.        │ │   │
│   │   │                                                    │ │   │
│   │   │  If you accepted this invitation, please log in    │ │   │
│   │   │  to access your account.                           │ │   │
│   │   │                                                    │ │   │
│   │   │  If you did not accept this invitation, please     │ │   │
│   │   │  contact the organization administrator.           │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │                  ┌──────────────────┐                    │   │
│   │                  │  Go to Login     │                    │   │
│   │                  └──────────────────┘                    │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Visual Layout (Email Already Registered - MVP Limitation)

```
┌─────────────────────────────────────────────────────────────────┐
│                         UBIK ENTERPRISE                          │
│                     AI Agent Management Platform                 │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │              ⚠️ Email Already Registered                 │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │                                                    │ │   │
│   │   │  The email address for this invitation            │ │   │
│   │   │  (jane@acme.com) is already registered.           │ │   │
│   │   │                                                    │ │   │
│   │   │  In our MVP version, each email can only be       │ │   │
│   │   │  associated with one organization.                 │ │   │
│   │   │                                                    │ │   │
│   │   │  ──────────────────────────────────────────────   │ │   │
│   │   │                                                    │ │   │
│   │   │  What to do next:                                  │ │   │
│   │   │                                                    │ │   │
│   │   │  Option 1: Use a different email address          │ │   │
│   │   │  Contact the person who invited you and ask them   │ │   │
│   │   │  to send a new invitation to a different email.    │ │   │
│   │   │                                                    │ │   │
│   │   │  Option 2: Contact support                         │ │   │
│   │   │  We can help resolve this limitation for           │ │   │
│   │   │  enterprise customers.                             │ │   │
│   │   │                                                    │ │   │
│   │   │  💡 Multi-organization support is coming soon!     │ │   │
│   │   │                                                    │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │                                                          │   │
│   │      ┌──────────────────┐    ┌──────────────────────┐  │   │
│   │      │  Go to Login     │    │  Contact Support     │  │   │
│   │      └──────────────────┘    └──────────────────────┘  │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### Layout Container
- **Component:** Full-page centered layout with branded background
- **Styling:** Gradient or subtle pattern background, card centered
- **Responsive:** Card full-width on mobile, max-width 600px on desktop

### Header Section
- **Logo:** Ubik Enterprise logo with tagline
- **Title:** Changes based on invitation state
  - Valid: "📨 You've Been Invited!"
  - Invalid: "❌ Invalid Invitation"
  - Accepted: "✅ Invitation Already Accepted"
  - Email exists: "⚠️ Email Already Registered"

### Invitation Details Card (Valid Token Only)
- **Organization Info:**
  - Organization name with emoji
  - Workspace URL
  - Styling: Prominent, with icon

- **Inviter Info:**
  - Full name and email
  - Role badge
  - Styling: Secondary

- **Your Role:**
  - Role name with icon
  - Role description
  - Styling: Badge or highlight

- **Team Assignment:**
  - Team name (if assigned)
  - Or "No team assigned" (gray text)

- **Expiration:**
  - Relative time (e.g., "In 5 days")
  - Absolute date
  - Color-coded:
    - Green: > 3 days remaining
    - Yellow: 1-3 days remaining
    - Red: < 1 day remaining

### Password Form (Valid Token Only)
- **Email Field:**
  - Read-only (pre-filled from invitation)
  - Verified badge
  - Helper text explaining it's from invitation

- **Full Name:**
  - Text input
  - Validation: 2-100 characters
  - Required

- **Password:**
  - Password input with visibility toggle
  - Real-time strength indicator
  - Same validation as signup page
  - Helper text with requirements

- **Confirm Password:**
  - Password input with visibility toggle
  - Real-time match validation
  - Error message if mismatch

### Terms Checkbox
- **Component:** Checkbox with linked text
- **Validation:** Must be checked
- **Links:** Terms of Service, Privacy Policy (new tab)

### Action Buttons

#### Accept Invitation Button (Valid Token)
- **Label:** "Accept Invitation & Join Team"
- **Style:** Primary, full-width on mobile
- **States:**
  - Default: Enabled (when form valid)
  - Disabled: Gray (form invalid or terms not accepted)
  - Loading: Spinner + "Accepting invitation..."
  - Success: Checkmark + "Success! Redirecting..."

#### Alternative Action Buttons (Error States)
- **Go to Login:** Secondary button
- **Contact Support:** Secondary button or link

---

## Field Validation Rules

### Client-Side Validation

| Field | Validation Rules | Error Timing |
|-------|------------------|--------------|
| Email | Read-only (pre-filled) | N/A |
| Full Name | 2-100 chars, letters/spaces/hyphens | On blur |
| Password | 8+ chars, complexity requirements | On change (strength), on blur (errors) |
| Confirm Password | Matches password | On change |
| Terms Checkbox | Must be checked | On submit |

### Server-Side Validation

- Token validity (not expired, not used, not cancelled)
- Email uniqueness (global for MVP)
- Password complexity
- All required fields present

---

## API Integration

### Step 1: Validate Token (On Page Load)
**Endpoint:** `GET /invitations/{token}`

**Success Response (200 OK):**
```json
{
  "invitation": {
    "id": "uuid",
    "email": "jane@acme.com",
    "status": "pending",
    "expires_at": "2024-12-15T23:59:59Z",
    "organization": {
      "id": "uuid",
      "name": "Acme Corporation",
      "slug": "acme-corp"
    },
    "inviter": {
      "id": "uuid",
      "full_name": "John Smith",
      "email": "john@acme.com",
      "role": {
        "name": "Administrator"
      }
    },
    "role": {
      "id": "uuid",
      "name": "Member",
      "description": "Standard team member permissions"
    },
    "team": {
      "id": "uuid",
      "name": "Engineering"
    }
  }
}
```

**Error Responses:**
- `404 Not Found` - Invalid token
  ```json
  {
    "error": "Invitation not found",
    "code": "INVITATION_NOT_FOUND"
  }
  ```
- `410 Gone` - Expired invitation
  ```json
  {
    "error": "Invitation has expired",
    "code": "INVITATION_EXPIRED",
    "expired_at": "2024-12-08T23:59:59Z"
  }
  ```
- `409 Conflict` - Already accepted
  ```json
  {
    "error": "Invitation already accepted",
    "code": "INVITATION_ALREADY_ACCEPTED",
    "accepted_at": "2024-12-10T10:30:00Z"
  }
  ```

### Step 2: Accept Invitation
**Endpoint:** `POST /invitations/{token}/accept`

**Request Body:**
```json
{
  "full_name": "Jane Doe",
  "password": "SecurePass123!"
}
```

**Success Response (201 Created):**
```json
{
  "employee": {
    "id": "uuid",
    "email": "jane@acme.com",
    "full_name": "Jane Doe",
    "role": {
      "id": "uuid",
      "name": "Member"
    },
    "team": {
      "id": "uuid",
      "name": "Engineering"
    }
  },
  "organization": {
    "id": "uuid",
    "name": "Acme Corporation",
    "slug": "acme-corp"
  },
  "token": "jwt-session-token"
}
```

**Error Responses:**
- `400 Bad Request` - Validation errors
  ```json
  {
    "error": "Validation failed",
    "details": [
      {"field": "password", "message": "Password must be at least 8 characters"}
    ]
  }
  ```
- `409 Conflict` - Email already registered
  ```json
  {
    "error": "Email already registered",
    "code": "EMAIL_EXISTS"
  }
  ```
- `410 Gone` - Invitation expired
- `500 Internal Server Error` - Server error

---

## User Interactions & Flows

### Happy Path (New User)
1. User clicks invitation link in email
2. Page loads, validates token via API
3. Invitation details displayed
4. User enters full name and password
5. Password strength indicator updates in real-time
6. User confirms password (checkmark when match)
7. User checks Terms of Service
8. User clicks "Accept Invitation & Join Team"
9. Button shows loading state
10. On success:
    - Session token stored
    - Welcome toast: "Welcome to [Org Name]!"
    - Redirect to `/onboarding` wizard (or `/dashboard` if already onboarded)

### Error Scenarios

#### Invalid Token
1. User clicks invitation link
2. Page loads, validates token
3. API returns 404
4. Error state displayed: "Invalid Invitation"
5. Show reasons and actions
6. Provide "Go to Login" and "Contact Support" buttons

#### Expired Token
1. User clicks old invitation link
2. Page loads, validates token
3. API returns 410 Gone
4. Error state displayed with expiration date
5. Suggest contacting inviter for new link

#### Already Accepted
1. User clicks invitation link (already used)
2. Page loads, validates token
3. API returns 409 Conflict
4. Success-style message: "Already Accepted"
5. Provide "Go to Login" button

#### Email Already Registered (MVP Limitation)
1. User submits form
2. API returns 409 Conflict with EMAIL_EXISTS code
3. Error state displayed
4. Explain MVP limitation
5. Suggest alternatives:
   - Use different email
   - Contact support for enterprise solution
6. Provide "Contact Support" button

#### Password Mismatch
1. User enters password
2. User enters different confirm password
3. Real-time error: "Passwords do not match"
4. Confirm password field highlighted red
5. Submit button disabled

#### Terms Not Accepted
1. User fills form but doesn't check terms
2. User clicks submit
3. Checkbox highlighted with error
4. Focus moves to checkbox
5. Error message: "You must agree to the Terms of Service"

#### Network Error
1. User submits form
2. API request fails
3. Loading state stops
4. Error alert: "Unable to accept invitation. Please check your connection and try again."
5. Form remains filled
6. User can retry

---

## State Management

### Page States
```typescript
type PageState =
  | 'loading'           // Initial token validation
  | 'valid'             // Token valid, show form
  | 'invalid'           // Token not found
  | 'expired'           // Token expired
  | 'accepted'          // Already accepted
  | 'submitting'        // Form being submitted
  | 'error';            // Submission error

const [pageState, setPageState] = useState<PageState>('loading');
const [invitation, setInvitation] = useState<Invitation | null>(null);
const [error, setError] = useState<string | null>(null);
```

### Token Validation Flow
```typescript
useEffect(() => {
  async function validateToken() {
    setPageState('loading');
    try {
      const response = await api.validateInvitation(token);
      setInvitation(response.invitation);
      setPageState('valid');
    } catch (error) {
      if (error.status === 404) {
        setPageState('invalid');
      } else if (error.status === 410) {
        setPageState('expired');
      } else if (error.status === 409) {
        setPageState('accepted');
      } else {
        setPageState('error');
        setError(error.message);
      }
    }
  }
  validateToken();
}, [token]);
```

---

## Accessibility (WCAG AA)

### Keyboard Navigation
- Tab order: Full Name → Password → Show Password → Confirm Password → Show Password → Terms Checkbox → Submit Button → Login Link
- Enter submits form (when valid)
- Escape clears form (with confirmation)

### Screen Reader Support
- Page title announces state: "Accept Invitation to [Org Name]"
- Invitation details section labeled
- Form labeled as "Registration Form"
- Error states announced via live regions
- Loading state: "Validating invitation token, please wait"
- Success state: "Invitation accepted successfully, redirecting"

### Visual Design
- Clear focus indicators
- High contrast (4.5:1 minimum)
- Error messages with icons (not color alone)
- Success indicators with checkmark icon
- Expiration warning color-coded with text (not color alone)
- Large touch targets (44px minimum)

### Labels & Helper Text
- All inputs have visible labels
- Helper text provides guidance
- Error messages are specific and actionable
- Read-only email field clearly indicated

---

## Responsive Design

### Mobile (< 640px)
- Card full-width with padding
- Single column layout
- Full-width inputs
- Full-width button
- Stack invitation details vertically
- Reduce logo size

### Tablet (640px - 1024px)
- Card width: 90% max 600px
- Maintain single column
- Larger touch targets
- More spacing

### Desktop (> 1024px)
- Card width: 600px fixed
- Centered on page
- Larger typography
- More generous spacing
- Side-by-side layout for some details

---

## Implementation Notes

### Technologies
- **Framework:** Next.js 14 (App Router)
- **Form Handling:** React Hook Form
- **Validation:** Zod schema
- **Components:** shadcn/ui (Form, Input, Button, Card, Alert, Badge, Checkbox)
- **Styling:** Tailwind CSS
- **State Management:** React Hook Form + React Query

### URL Token Extraction
```typescript
'use client'

import { useSearchParams } from 'next/navigation';

export default function AcceptInvitationPage() {
  const searchParams = useSearchParams();
  const token = searchParams.get('token');

  if (!token) {
    return <InvalidTokenState reason="missing" />;
  }

  // Continue with token validation...
}
```

### Form Schema (Zod)
```typescript
const acceptInvitationSchema = z.object({
  full_name: z.string()
    .min(2, 'Name must be at least 2 characters')
    .max(100, 'Name must be less than 100 characters')
    .regex(/^[a-zA-Z\s-]+$/, 'Name can only contain letters, spaces, and hyphens'),
  password: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/[A-Z]/, 'Password must contain uppercase letter')
    .regex(/[a-z]/, 'Password must contain lowercase letter')
    .regex(/[0-9]/, 'Password must contain number')
    .regex(/[^A-Za-z0-9]/, 'Password must contain special character'),
  confirm_password: z.string(),
  terms_accepted: z.boolean().refine(val => val === true, {
    message: 'You must agree to the Terms of Service'
  })
}).refine(data => data.password === data.confirm_password, {
  message: 'Passwords do not match',
  path: ['confirm_password']
});
```

### Expiration Display
```typescript
function formatExpiration(expiresAt: string): {
  text: string;
  color: 'green' | 'yellow' | 'red';
} {
  const now = new Date();
  const expiry = new Date(expiresAt);
  const daysRemaining = Math.ceil((expiry.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));

  let color: 'green' | 'yellow' | 'red';
  if (daysRemaining > 3) color = 'green';
  else if (daysRemaining > 1) color = 'yellow';
  else color = 'red';

  const relativeText = daysRemaining === 1
    ? 'Tomorrow'
    : `In ${daysRemaining} days`;

  const absoluteText = expiry.toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  });

  return {
    text: `${relativeText} (${absoluteText})`,
    color
  };
}
```

---

## Testing Checklist

### Unit Tests
- [ ] Token extraction from URL
- [ ] Form validation rules
- [ ] Password strength calculation
- [ ] Expiration date formatting
- [ ] Error state rendering

### Integration Tests
- [ ] Valid token flow
- [ ] Invalid token handling
- [ ] Expired token handling
- [ ] Already accepted handling
- [ ] Email conflict handling
- [ ] Successful acceptance
- [ ] Network error handling

### E2E Tests (Playwright)
- [ ] Complete acceptance journey
- [ ] Invalid token error display
- [ ] Expired token error display
- [ ] Form field interactions
- [ ] Real-time password validation
- [ ] Terms checkbox requirement
- [ ] Redirect after success

### Accessibility Tests
- [ ] Keyboard navigation
- [ ] Screen reader announcements
- [ ] Focus management
- [ ] Color contrast
- [ ] Touch target sizes

---

## Security Considerations

### Token Security
- Tokens are single-use (marked as used after acceptance)
- Tokens expire after 7 days
- Tokens are 256-bit random (cryptographically secure)
- No predictable patterns in tokens

### Password Security
- Passwords bcrypt hashed before storage
- Complexity requirements enforced
- Strength indicator helps users choose strong passwords

### Rate Limiting
- Limit token validation attempts (prevent token scanning)
- Limit acceptance attempts per token (prevent brute force)

---

## Related Pages
- **Previous:** Email invitation link
- **Next:** `/onboarding` (Onboarding Wizard) or `/dashboard`
- **Alternative:** `/login` (if already have account)

---

## Design System References
- **shadcn/ui Form:** https://ui.shadcn.com/docs/components/form
- **shadcn/ui Input:** https://ui.shadcn.com/docs/components/input
- **shadcn/ui Button:** https://ui.shadcn.com/docs/components/button
- **shadcn/ui Card:** https://ui.shadcn.com/docs/components/card
- **shadcn/ui Alert:** https://ui.shadcn.com/docs/components/alert
- **shadcn/ui Badge:** https://ui.shadcn.com/docs/components/badge
- **shadcn/ui Checkbox:** https://ui.shadcn.com/docs/components/checkbox

---

## Notes

**Why Pre-fill Email?**
- Email comes from invitation token
- User cannot change it (tied to invitation)
- Reduces friction and errors

**Why MVP Limitation (One Email = One Org)?**
- Simplifies initial implementation
- Most common use case for first version
- Multi-org support requires users table + org_memberships
- Can add later without breaking changes

**Why 7-Day Expiration?**
- Industry standard (GitHub, Slack, etc.)
- Long enough for users to respond
- Short enough to maintain security
- Prevents old invitations from accumulating

**Future Enhancements:**
- Resend invitation functionality
- Custom expiration periods
- Multi-organization support (users table)
- Social login (OAuth)
- SSO/SAML for enterprise
