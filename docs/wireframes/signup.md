# Signup Page Wireframe

## Page: /signup

### Layout

```
┌─────────────────────────────────────────────────────────────┐
│                                                               │
│                     Centered Card (max-w-lg)                 │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │          Ubik Enterprise                        │  │  │
│  │  │     Create Your Organization Account           │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Full Name                                       │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [John Doe                                  ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Email                                           │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [you@example.com                           ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  │ [!] Invalid email format                        │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Organization Name                               │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [Acme Corporation                          ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Organization Slug                               │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [acme                                      ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  │ ✓ acme.ubik.com is available                    │  │  │
│  │  │ [!] This slug is already taken                  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Password                                        │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [••••••••                                  ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  │                                                 │  │  │
│  │  │ Password Strength: Weak/Medium/Strong          │  │  │
│  │  │ ▓▓▓▓░░░░ (visual indicator bar)                │  │  │
│  │  │                                                 │  │  │
│  │  │ Requirements:                                   │  │  │
│  │  │ ✓ At least 8 characters                        │  │  │
│  │  │ ✗ One uppercase letter                         │  │  │
│  │  │ ✓ One number                                   │  │  │
│  │  │ ✗ One special character                        │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │ Confirm Password                                │  │  │
│  │  │ ┌─────────────────────────────────────────────┐ │  │  │
│  │  │ │ [••••••••                                  ] │ │  │  │
│  │  │ └─────────────────────────────────────────────┘ │  │  │
│  │  │ [!] Passwords do not match                      │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  ┌─────────────────────────────────────────────────┐  │  │
│  │  │        [Create Account]                         │  │  │
│  │  │     (full width button, disabled when invalid)  │  │  │
│  │  └─────────────────────────────────────────────────┘  │  │
│  │                                                         │  │
│  │  Already have an account? [Sign in]                   │  │
│  │                                                         │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

## Form Fields

### 1. Full Name
- **Type**: Text input
- **Required**: Yes
- **Validation**:
  - Not empty
  - Min 2 characters
  - Max 100 characters
- **Error Messages**:
  - "Full name is required"
  - "Full name must be at least 2 characters"

### 2. Email
- **Type**: Email input
- **Required**: Yes
- **Validation**:
  - Valid email format
  - Unique in database (server-side)
- **Error Messages**:
  - "Invalid email address"
  - "This email is already registered"

### 3. Organization Name
- **Type**: Text input
- **Required**: Yes
- **Validation**:
  - Not empty
  - Min 2 characters
  - Max 100 characters
- **Error Messages**:
  - "Organization name is required"
  - "Organization name must be at least 2 characters"

### 4. Organization Slug
- **Type**: Text input
- **Required**: Yes
- **Validation**:
  - Lowercase letters, numbers, hyphens only
  - Must start with letter
  - Min 3 characters
  - Max 50 characters
  - Unique in database (real-time check)
- **Real-time Feedback**:
  - "✓ {slug}.ubik.com is available" (green)
  - "✗ This slug is already taken" (red)
  - "⏳ Checking availability..." (loading)
- **Debounce**: 500ms
- **Error Messages**:
  - "Organization slug is required"
  - "Slug must be 3-50 characters"
  - "Slug can only contain lowercase letters, numbers, and hyphens"
  - "Slug must start with a letter"
  - "This slug is already taken"

### 5. Password
- **Type**: Password input
- **Required**: Yes
- **Validation**:
  - Min 8 characters
  - At least 1 uppercase letter
  - At least 1 lowercase letter
  - At least 1 number
  - At least 1 special character (@$!%*?&)
- **Visual Feedback**:
  - Strength indicator bar (Weak/Medium/Strong)
  - Color-coded: Red/Yellow/Green
  - Checklist of requirements (✓/✗)
- **Error Messages**:
  - "Password is required"
  - "Password must be at least 8 characters"
  - "Password must contain at least one uppercase letter"
  - "Password must contain at least one number"
  - "Password must contain at least one special character"

### 6. Confirm Password
- **Type**: Password input
- **Required**: Yes
- **Validation**:
  - Must match password field
- **Error Messages**:
  - "Please confirm your password"
  - "Passwords do not match"

## Form Behavior

### Submit Button
- **Label**: "Create Account" (default)
- **Label (loading)**: "Creating Account..." (with spinner)
- **Disabled when**:
  - Form is invalid
  - Submission in progress
  - Org slug is being checked
  - Org slug is not available

### Form Submission
1. Validate all fields client-side
2. Show loading state on button
3. Call POST /auth/register API endpoint
4. Handle responses:
   - **Success (201)**: Store session token → Redirect to /onboarding/welcome
   - **400 (Validation error)**: Show field errors
   - **409 (Duplicate email)**: Show "This email is already registered"
   - **409 (Duplicate slug)**: Show "This slug is already taken"
   - **500**: Show generic error message

### API Endpoint: POST /auth/register

**Request Body:**
```json
{
  "full_name": "John Doe",
  "email": "john@example.com",
  "org_name": "Acme Corporation",
  "org_slug": "acme",
  "password": "SecurePass123!"
}
```

**Success Response (201):**
```json
{
  "token": "jwt-token-here",
  "employee": {
    "id": "uuid",
    "email": "john@example.com",
    "full_name": "John Doe"
  },
  "organization": {
    "id": "uuid",
    "name": "Acme Corporation",
    "slug": "acme"
  }
}
```

**Error Response (400):**
```json
{
  "message": "Validation failed",
  "errors": {
    "email": ["Invalid email format"],
    "password": ["Password too weak"]
  }
}
```

### Org Slug Availability Check

**API Endpoint**: GET /auth/check-slug?slug={slug}

**Request**: `GET /auth/check-slug?slug=acme`

**Success Response (200):**
```json
{
  "available": true
}
```

**Response (200) - Not Available:**
```json
{
  "available": false
}
```

## Responsive Design

### Mobile (<768px)
- Single column layout
- Full-width card with padding
- Stacked form fields
- Touch-friendly input sizes (min 44px height)
- Password requirements list below password field

### Tablet (768px-1024px)
- Centered card with max-width
- All fields visible without scrolling

### Desktop (>1024px)
- Centered card with max-width: 512px
- Comfortable spacing
- Focus states clearly visible

## Accessibility

### Keyboard Navigation
- Tab order: Name → Email → Org Name → Org Slug → Password → Confirm Password → Submit
- Enter key submits form
- Escape key clears focus (browser default)

### Screen Reader Support
- All form labels properly associated with inputs
- ARIA attributes:
  - `aria-required="true"` on all required fields
  - `aria-invalid="true"` when field has error
  - `aria-describedby` linking to error messages
  - `role="alert"` on error messages
- Password strength indicator announced to screen readers
- Org slug availability announced with live region

### WCAG AA Compliance
- Color contrast ratio ≥ 4.5:1 for all text
- Focus indicators visible and clear
- Error messages descriptive and specific
- Form labels always visible (no placeholder-only labels)

## User Flow

1. User lands on /signup
2. User fills in full name
3. User enters email
4. User enters organization name
5. User enters organization slug → Real-time availability check starts
6. User sees availability feedback (available/taken/checking)
7. User enters password → Strength indicator updates in real-time
8. User sees password requirements checklist
9. User enters confirm password → Match validation happens
10. All validations pass → Submit button becomes enabled
11. User clicks "Create Account"
12. Loading state shown
13. Success: Redirect to /onboarding/welcome
14. Error: Show error messages, keep form data

## Error Handling

### Client-side Validation
- Show errors on blur (after user leaves field)
- Show errors on submit attempt
- Clear errors when user starts typing again

### Server-side Errors
- Duplicate email: "This email is already registered. [Sign in]"
- Duplicate slug: "This organization slug is already taken. Please choose another."
- Network error: "Unable to create account. Please check your connection and try again."
- Unknown error: "An unexpected error occurred. Please try again later."

## Link to Login
- Text: "Already have an account? Sign in"
- Link: /login
- Positioned below submit button

## Design Tokens

### Colors
- Primary: Blue (#3B82F6)
- Success: Green (#10B981)
- Error: Red (#EF4444)
- Warning: Yellow (#F59E0B)

### Typography
- Title: 3xl font size, font-bold
- Description: sm font size, text-muted-foreground
- Labels: sm font size, font-medium
- Error messages: sm font size, text-destructive

### Spacing
- Card padding: p-6
- Form fields gap: space-y-4
- Input height: 40px
- Button height: 40px
