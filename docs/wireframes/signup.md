# Signup Page Wireframe

**Route:** `/signup`
**Access:** Public (unauthenticated)
**Components:** Form, Input, Button, Alert, Card
**Layout:** Centered card on full-page background

---

## Page Purpose

Combined registration and organization creation form. New users create their account and organization in a single step. This is the entry point for all new customers to the platform.

---

## Visual Layout

```
┌─────────────────────────────────────────────────────────────────┐
│                         UBIK ENTERPRISE                          │
│                     AI Agent Management Platform                 │
│                                                                   │
│   ┌────────────────────────────────────────────────────────┐   │
│   │                                                          │   │
│   │                    Create Your Account                   │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ Account Information                               │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Email Address *                                   │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ you@company.com                              │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  Full Name *                                       │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ John Smith                                   │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  Password *                                        │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ ●●●●●●●●                           [👁]       │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │  [Strength: ▓▓▓▓░░░░ Medium]                      │ │   │
│   │   │  ℹ️ Must be at least 8 characters with mix of     │ │   │
│   │   │     letters, numbers, and symbols                 │ │   │
│   │   │                                                    │ │   │
│   │   │  Confirm Password *                                │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ ●●●●●●●●                           [👁]       │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ Organization Details                              │ │   │
│   │   ├──────────────────────────────────────────────────┤ │   │
│   │   │                                                    │ │   │
│   │   │  Organization Name *                               │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ Acme Corporation                             │ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │                                                    │ │   │
│   │   │  Organization Slug *                               │ │   │
│   │   │  ┌──────────────────────────────────────────────┐ │ │   │
│   │   │  │ acme-corp                          [✓ Available]│ │ │   │
│   │   │  └──────────────────────────────────────────────┘ │ │   │
│   │   │  🔗 Your workspace URL:                           │ │   │
│   │   │     https://acme-corp.ubik.io                     │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │ ☐ I agree to the Terms of Service and            │ │   │
│   │   │   Privacy Policy                                  │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │   ┌──────────────────────────────────────────────────┐ │   │
│   │   │          Create Account & Organization           │ │   │
│   │   └──────────────────────────────────────────────────┘ │   │
│   │                                                          │   │
│   │             Already have an account? Log in              │   │
│   │                                                          │   │
│   └────────────────────────────────────────────────────────┘   │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Component Breakdown

### Layout Container
- **Component:** Full-page centered layout with branded background
- **Styling:** Gradient or subtle pattern background, card centered vertically and horizontally
- **Responsive:** Card full-width on mobile (with padding), max-width 500px on desktop

### Header Section
- **Logo:** Ubik Enterprise logo with tagline
- **Title:** "Create Your Account" (h1)
- **Styling:** Centered, clear typography hierarchy

### Form Sections

#### Account Information Section
- **Container:** Card with border and subtle shadow
- **Fields:**
  1. **Email Address**
     - Type: email input
     - Validation: Valid email format, unique in system
     - Error messages: "Invalid email format", "This email is already registered"
     - Auto-focus on page load

  2. **Full Name**
     - Type: text input
     - Validation: 2-100 characters, letters/spaces/hyphens only
     - Error messages: "Name must be between 2 and 100 characters"

  3. **Password**
     - Type: password input with toggle visibility button
     - Validation:
       - Minimum 8 characters
       - Must contain: uppercase, lowercase, number, special character
     - Real-time strength indicator:
       - Weak (red): < 8 chars or missing requirements
       - Medium (yellow): 8+ chars, 2-3 requirements met
       - Strong (green): 8+ chars, all requirements met
       - Very Strong (dark green): 12+ chars, all requirements met
     - Helper text below input with requirements

  4. **Confirm Password**
     - Type: password input with toggle visibility button
     - Validation: Must match password field
     - Real-time validation: Shows checkmark when matches
     - Error message: "Passwords do not match"

#### Organization Details Section
- **Container:** Separate card section with visual separation
- **Fields:**
  1. **Organization Name**
     - Type: text input
     - Validation: 3-100 characters, required
     - Error messages: "Organization name must be between 3 and 100 characters"
     - Auto-generates slug on blur (if slug empty)

  2. **Organization Slug**
     - Type: text input with real-time validation
     - Validation:
       - 3-50 characters
       - Lowercase letters, numbers, hyphens only
       - Must start with letter
       - Cannot end with hyphen
       - Unique in system (real-time check via debounced API call)
     - Visual indicators:
       - ⏳ Checking... (while API call in progress)
       - ✓ Available (green, slug is unique)
       - ✗ Already taken (red, slug exists)
     - Preview text: "Your workspace URL: https://{slug}.ubik.io"
     - Error messages:
       - "Slug must be 3-50 characters"
       - "Slug can only contain lowercase letters, numbers, and hyphens"
       - "Slug must start with a letter"
       - "This slug is already taken"

#### Terms of Service
- **Component:** Checkbox with linked text
- **Validation:** Must be checked to submit
- **Links:**
  - Terms of Service (opens in new tab)
  - Privacy Policy (opens in new tab)
- **Error message:** "You must agree to the Terms of Service and Privacy Policy"

### Action Buttons

#### Create Account Button
- **Type:** Primary button (full-width on mobile, auto-width on desktop)
- **Label:** "Create Account & Organization"
- **States:**
  - Default: Blue/primary color, enabled
  - Hover: Darker shade
  - Disabled: Gray, cursor not-allowed (when form invalid or submitting)
  - Loading: Show spinner + "Creating account..." text
- **Behavior:** Submits form, shows loading state, handles errors

#### Login Link
- **Type:** Text link
- **Label:** "Already have an account? Log in"
- **Behavior:** Navigates to `/login` page

---

## Field Validation Rules

### Client-Side Validation (Real-time)

| Field | Validation Rules | Error Timing |
|-------|------------------|--------------|
| Email | Valid email format | On blur |
| Full Name | 2-100 chars, letters/spaces/hyphens | On blur |
| Password | 8+ chars, complexity requirements | On change (for strength), on blur (for errors) |
| Confirm Password | Matches password | On change |
| Org Name | 3-100 chars | On blur |
| Org Slug | 3-50 chars, format, uniqueness | On change (debounced 500ms for uniqueness) |
| Terms Checkbox | Must be checked | On submit |

### Server-Side Validation (On Submit)

- Email uniqueness (double-check)
- Org slug uniqueness (double-check)
- Password complexity
- All required fields present
- Rate limiting (5 registrations/hour per IP)

---

## API Integration

### Endpoint: `POST /auth/register`

**Request Body:**
```json
{
  "email": "john@acme.com",
  "password": "SecurePass123!",
  "full_name": "John Smith",
  "org_name": "Acme Corporation",
  "org_slug": "acme-corp"
}
```

**Success Response (201 Created):**
```json
{
  "employee": {
    "id": "uuid",
    "email": "john@acme.com",
    "full_name": "John Smith",
    "role": {
      "id": "uuid",
      "name": "admin"
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
      {"field": "email", "message": "Invalid email format"},
      {"field": "org_slug", "message": "Slug is already taken"}
    ]
  }
  ```
- `409 Conflict` - Email already registered
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

### Slug Availability Check: `GET /auth/check-slug?slug={slug}`

**Success Response (200 OK):**
```json
{
  "available": true
}
```

**Conflict Response (409 Conflict):**
```json
{
  "available": false
}
```

---

## User Interactions & Flows

### Happy Path
1. User lands on `/signup`
2. User enters email, full name, password
3. Password strength indicator updates in real-time
4. User confirms password (checkmark appears when match)
5. User enters organization name
6. Organization slug is auto-generated from org name
7. User modifies slug (optional)
8. Slug availability is checked in real-time (✓ Available shown)
9. User checks Terms of Service checkbox
10. User clicks "Create Account & Organization"
11. Button shows loading state
12. On success:
    - Session token is stored
    - User is redirected to `/onboarding` wizard

### Error Scenarios

#### Email Already Registered
1. User enters email that's already in system
2. On blur, inline error appears: "This email is already registered"
3. Alternative: Server returns 409 on submit
4. Error alert shown: "An account with this email already exists. Please log in or use a different email."
5. Provide "Go to Login" link in error

#### Slug Already Taken
1. User enters slug
2. After 500ms debounce, API check is made
3. ✗ Icon shown with "Already taken" message
4. User must choose different slug before submitting

#### Password Mismatch
1. User enters password
2. User enters different confirm password
3. Real-time error shown: "Passwords do not match"
4. Confirm password field has red border
5. Submit button remains disabled

#### Terms Not Accepted
1. User fills form but doesn't check terms
2. User clicks submit
3. Checkbox field is highlighted with error
4. Focus moves to checkbox
5. Error message shown: "You must agree to the Terms of Service"

#### Network Error
1. User submits form
2. API request fails (timeout, network error)
3. Loading state stops
4. Error alert shown: "Unable to create account. Please check your connection and try again."
5. Form fields remain filled
6. User can retry submission

#### Rate Limit Exceeded
1. User submits form multiple times
2. Server returns 429
3. Error alert shown: "Too many registration attempts. Please try again in 1 hour."
4. Form is disabled for cooldown period

---

## Accessibility (WCAG AA)

### Keyboard Navigation
- Tab order: Email → Full Name → Password → Show Password → Confirm Password → Show Password → Org Name → Org Slug → Terms Checkbox → Submit Button → Login Link
- Enter key submits form when valid
- Escape key clears form (with confirmation)

### Screen Reader Support
- Form labeled as "Registration Form"
- Each input has associated label
- Required fields announced as "required"
- Error messages associated with inputs via `aria-describedby`
- Password strength announced via live region
- Slug availability announced via live region
- Loading state announced: "Creating your account, please wait"

### Visual Design
- Clear focus indicators on all interactive elements
- High contrast text (4.5:1 minimum)
- Error messages in red with icon (not color alone)
- Success indicators with checkmark icon (not color alone)
- Large touch targets (44px minimum on mobile)

### Form Labels
- All inputs have visible labels (not just placeholders)
- Labels remain visible when input is focused/filled
- Helper text provides guidance
- Error messages are specific and actionable

---

## Responsive Design

### Mobile (< 640px)
- Card takes full width with 16px padding
- Single column layout
- Full-width inputs
- Full-width button
- Stack password strength indicator below input
- Reduce logo size
- Simplify header

### Tablet (640px - 1024px)
- Card width: 90% max 500px
- Maintain single column
- Larger touch targets
- More spacing between sections

### Desktop (> 1024px)
- Card width: 500px fixed
- Centered on page
- Larger typography
- More generous spacing
- Show org slug preview more prominently

---

## Implementation Notes

### Technologies
- **Framework:** Next.js 14 (App Router)
- **Form Handling:** React Hook Form
- **Validation:** Zod schema
- **Components:** shadcn/ui (Form, Input, Button, Card, Alert, Checkbox)
- **Styling:** Tailwind CSS
- **State Management:** React Hook Form state + React Query for API calls

### Form Schema (Zod)
```typescript
const signupSchema = z.object({
  email: z.string().email('Invalid email format'),
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
  org_name: z.string()
    .min(3, 'Organization name must be at least 3 characters')
    .max(100, 'Organization name must be less than 100 characters'),
  org_slug: z.string()
    .min(3, 'Slug must be at least 3 characters')
    .max(50, 'Slug must be less than 50 characters')
    .regex(/^[a-z][a-z0-9-]*[a-z0-9]$/, 'Slug must start with letter, contain only lowercase letters, numbers, hyphens'),
  terms_accepted: z.boolean().refine(val => val === true, {
    message: 'You must agree to the Terms of Service'
  })
}).refine(data => data.password === data.confirm_password, {
  message: 'Passwords do not match',
  path: ['confirm_password']
});
```

### Debounced Slug Check
- Use `useDebounce` hook or `lodash.debounce`
- Debounce delay: 500ms
- Cancel pending requests on unmount
- Show loading spinner during check

### Auto-Generate Slug
- Triggered on org_name blur (if org_slug is empty)
- Transform: lowercase, replace spaces with hyphens, remove special chars
- Example: "Acme Corporation!" → "acme-corporation"

### Password Strength Calculation
```typescript
function calculatePasswordStrength(password: string): PasswordStrength {
  let score = 0;
  if (password.length >= 8) score++;
  if (password.length >= 12) score++;
  if (/[A-Z]/.test(password)) score++;
  if (/[a-z]/.test(password)) score++;
  if (/[0-9]/.test(password)) score++;
  if (/[^A-Za-z0-9]/.test(password)) score++;

  if (score < 3) return 'weak';
  if (score < 5) return 'medium';
  if (score < 6) return 'strong';
  return 'very-strong';
}
```

---

## Testing Checklist

### Unit Tests
- [ ] Form validation rules (all fields)
- [ ] Password strength calculation
- [ ] Slug auto-generation
- [ ] Error message display
- [ ] Accessibility attributes

### Integration Tests
- [ ] Successful registration flow
- [ ] Email already registered error
- [ ] Slug already taken error
- [ ] Password mismatch error
- [ ] Terms not accepted error
- [ ] Network error handling
- [ ] Rate limit handling

### E2E Tests (Playwright)
- [ ] Complete registration journey
- [ ] Form field interactions
- [ ] Real-time validation
- [ ] Slug availability check
- [ ] Error recovery flows
- [ ] Redirect to onboarding after success

### Accessibility Tests
- [ ] Keyboard navigation
- [ ] Screen reader announcements
- [ ] Focus management
- [ ] Color contrast
- [ ] Touch target sizes

---

## Related Pages
- **Previous:** None (entry point)
- **Next:** `/onboarding` (Onboarding Wizard)
- **Alternative:** `/login` (Login Page)

---

## Design System References
- **shadcn/ui Form:** https://ui.shadcn.com/docs/components/form
- **shadcn/ui Input:** https://ui.shadcn.com/docs/components/input
- **shadcn/ui Button:** https://ui.shadcn.com/docs/components/button
- **shadcn/ui Card:** https://ui.shadcn.com/docs/components/card
- **shadcn/ui Alert:** https://ui.shadcn.com/docs/components/alert
- **shadcn/ui Checkbox:** https://ui.shadcn.com/docs/components/checkbox
