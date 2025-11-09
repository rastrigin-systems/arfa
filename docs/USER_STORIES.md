# Ubik Enterprise - User Stories

**Status:** In Progress
**Last Updated:** 2025-11-09
**Version:** 0.1.0

## Overview

This document contains user stories for the Ubik Enterprise AI Agent Management Platform. Stories are organized by epic and reflect the **currently implemented** system, not future wireframes.

## How to Use This Document

- Stories follow the format: "As a [role], I want to [action] so that [benefit]"
- Each story includes acceptance criteria in Given/When/Then format
- Stories are linked to implemented features and API endpoints
- Priority: P0 (Critical), P1 (High), P2 (Medium), P3 (Low)
- Status: ✅ Implemented, 🚧 In Progress, 📋 Planned

---

## Epics

1. [Authentication & Onboarding](#epic-1-authentication--onboarding)
2. [Dashboard & Navigation](#epic-2-dashboard--navigation)
3. [Organization Management](#epic-3-organization-management)
4. [Employee Management](#epic-4-employee-management)
5. [Team Management](#epic-5-team-management)
6. [Agent Configuration](#epic-6-agent-configuration)
7. [MCP Server Management](#epic-7-mcp-server-management)
8. [Approvals & Requests](#epic-8-approvals--requests)
9. [Usage & Analytics](#epic-9-usage--analytics)
10. [CLI Integration](#epic-10-cli-integration)

---

## Epic 1: Authentication & Onboarding

### Story 1.1: User Registration (New Organization)

**As a** new user
**I want to** create an account with my organization details
**So that** I can start managing AI agents for my company

**Priority:** P0 (Critical)
**Status:** ✅ Implemented
**Endpoints:** `POST /api/v1/auth/register`
**UI:** `/signup`

**Acceptance Criteria:**

```gherkin
Given I am on the signup page
When I enter my full name, email, organization name, organization slug, and password
And I click "Create Account"
Then a new organization is created with me as the owner
And I am redirected to the dashboard
And I receive a session token

Given I am typing an organization slug
When I enter a slug that already exists (e.g., "acme")
Then I see "This slug is already taken" appear in real-time (500ms debounce)
And the slug input shows a red error state
And a red X icon appears next to the input

Given I am typing an organization slug
When I enter a unique slug (e.g., "my-company-2025")
Then I see "✓" appear indicating the slug is available
And the slug input shows a green success state

Given I try to submit the form with an invalid password
When I click "Create Account" with a password missing uppercase/number/special char
Then I see inline validation errors for password requirements
And the form does not submit
```

**Implementation Notes:**
- First user in an organization automatically gets "owner" role
- Organization slug is validated in real-time (500ms debounce)
- Password requires: 8+ chars, uppercase, number, special character
- Session token stored in httpOnly cookie

---

---

### Story 1.2: User Login

**As a** registered user
**I want to** log in with my email and password
**So that** I can access my organization's AI agent management dashboard

**Priority:** P0 (Critical)
**Status:** 🚧 In Progress (Login works, auto-redirect & explicit logout needed)
**Endpoints:** `POST /api/v1/auth/login`, `POST /api/v1/auth/logout`
**UI:** `/login`, logout button in navigation

**Acceptance Criteria:**

```gherkin
Given I am on the login page
When I enter a valid email and password
And I click "Login"
Then I am authenticated and receive a session token
And I am redirected to the dashboard (/dashboard)
And my session token is stored in an httpOnly cookie

Given I am on the login page
When I click "Login" with invalid credentials
Then I see an error message "Invalid credentials"
And I remain on the login page
And no session token is created

Given I am submitting the login form
When I click "Login"
Then the button shows "Logging in..." while the request is processing
And the button is disabled during submission
And the button returns to "Login" after completion

Given I am on the login page and don't have an account
When I click "Sign up"
Then I am redirected to the signup page (/signup)

Given I am already authenticated (have valid session)
When I navigate to /login
Then I am automatically redirected to /dashboard
And I see my dashboard without having to log in again

Given I want to log in as a different user
When I am currently logged in
Then I must first click "Logout" from the navigation menu
And my session is cleared
And I am redirected to /login
And I can now log in with different credentials

Given I forgot my password
When I am on the login page
Then I see a "Forgot password?" link below the login button
And clicking it takes me to /forgot-password (see Issue #155)
```

**Implementation Notes:**
- Email and password are required fields
- Session token stored in httpOnly cookie named `ubik_token`
- No "remember me" option (session-based only)
- **Auto-redirect:** Authenticated users visiting /login are redirected to /dashboard
- **Explicit logout required:** Users must logout before switching accounts
- **Forgotten password:** Link to /forgot-password flow (tracked in Issue #155)
- Form uses React Server Actions with `useFormState` and `useFormStatus`
- Logout button must be easily discoverable in navigation UI

---

### Story 1.3: User Logout

**As a** logged-in user
**I want to** log out of my account
**So that** I can end my session securely or switch to a different account

**Priority:** P0 (Critical)
**Status:** ✅ Implemented
**Endpoints:** `POST /api/v1/auth/logout`
**UI:** Logout button in navigation/user menu

**Acceptance Criteria:**

```gherkin
Given I am logged in and on any page
When I look at the navigation bar
Then I see a user menu or logout button
And the logout option is clearly visible and accessible

Given I am logged in
When I click "Logout" from the navigation menu
Then my session is immediately terminated
And my session token (ubik_token cookie) is cleared
And I am redirected to the login page (/login)
And I see a success message "Logged out successfully"

Given I have just logged out
When I try to navigate to a protected page (e.g., /dashboard)
Then I am redirected to the login page
And I see a message "Please log in to continue"

Given I have just logged out
When I click the browser back button
Then I cannot access any protected pages
And I am redirected to the login page
And my session remains terminated

Given I want to switch accounts
When I click "Logout"
Then I am taken to the login page
And I can log in with different credentials
And a new session is created for the new user
```

**Implementation Notes:**
- Logout button should be in the top navigation bar (likely in a user menu)
- `POST /api/v1/auth/logout` clears the `ubik_token` httpOnly cookie
- Session is deleted from the database
- No confirmation dialog needed (immediate logout)
- After logout, all protected routes should redirect to /login
- Client-side should clear any cached user data

---

### Story 1.4: Password Reset (Forgot Password)

**As a** user who forgot my password
**I want to** request a password reset link via email
**So that** I can regain access to my account without contacting support

**Priority:** P2 (Medium)
**Status:** 📋 Planned (Issue #155)
**Endpoints:** `POST /api/v1/auth/forgot-password`, `POST /api/v1/auth/reset-password`, `GET /api/v1/auth/verify-reset-token`
**UI:** `/forgot-password`, `/reset-password/[token]`

**Acceptance Criteria:**

```gherkin
Given I am on the login page
When I click "Forgot password?"
Then I am redirected to /forgot-password
And I see a form asking for my email address

Given I am on the forgot password page
When I enter my registered email address
And I click "Send reset link"
Then I see a success message "Password reset link sent to your email"
And I receive an email with a password reset link
And the link is valid for 1 hour

Given I am on the forgot password page
When I enter an email that doesn't exist in the system
And I click "Send reset link"
Then I see the same success message (for security)
And no email is sent
And no error reveals that the email doesn't exist

Given I receive a password reset email
When I click the reset link
Then I am redirected to /reset-password/[token]
And I see a form to enter a new password
And the form requires password confirmation

Given I am on the reset password page with a valid token
When I enter a new password and confirmation that match
And the password meets requirements (8+ chars, uppercase, number, special)
And I click "Reset password"
Then my password is updated
And I see a success message "Password reset successful"
And I am redirected to /login
And I can log in with my new password

Given I am on the reset password page with an expired token
When I try to submit the form
Then I see an error "This reset link has expired"
And I see a link to request a new reset link

Given I am on the reset password page with an already-used token
When I try to submit the form
Then I see an error "This reset link has already been used"
And I see a link to request a new reset link

Given I am on the reset password page
When I enter passwords that don't match
And I click "Reset password"
Then I see an error "Passwords do not match"
And the form does not submit

Given I try to reuse a password reset link
When the link has already been used once
Then the link is invalid
And I cannot reset my password again with the same link
```

**Implementation Notes:**
- **Database:** Requires `password_reset_tokens` table with columns:
  - `id` (UUID, PK)
  - `employee_id` (UUID, FK to employees)
  - `token` (VARCHAR, unique, cryptographically random)
  - `expires_at` (TIMESTAMP, 1 hour from creation)
  - `used_at` (TIMESTAMP, nullable, marks token as used)
  - `created_at` (TIMESTAMP)
- **Security:**
  - Tokens must be cryptographically random (256-bit minimum)
  - Tokens are single-use (marked as used after password reset)
  - Tokens expire after 1 hour
  - Rate limiting: Max 3 reset requests per email per hour
  - Generic success message for non-existent emails (prevents email enumeration)
- **Email:**
  - Requires email service integration (e.g., SendGrid, AWS SES)
  - Email should include reset link: `https://app.ubik.com/reset-password/[TOKEN]`
  - Email should clearly state the 1-hour expiration
- **Password validation:**
  - Same requirements as registration: 8+ chars, uppercase, number, special character
  - Password confirmation must match
- **User experience:**
  - Clear error messages for expired/used tokens
  - Link back to forgot-password page if token is invalid
  - Auto-redirect to login after successful reset

**Related Issues:**
- Issue #155 - Full implementation spec

---

## Epic 2: Dashboard & Navigation

### Story 2.1: Admin Dashboard (Overview & Activity)

**As an** admin or manager
**I want to** see an overview of pending approvals and recent organization activity
**So that** I can quickly identify what needs my attention and stay informed

**Priority:** P1 (High)
**Status:** 🚧 In Progress (Basic shell exists, needs implementation)
**Endpoints:** `GET /api/v1/approvals/pending`, `GET /api/v1/activity-logs`
**UI:** `/dashboard`

**Acceptance Criteria:**

```gherkin
Given I am an admin and I log in successfully
When I am redirected to /dashboard
Then I see a welcome message with my name "Welcome back, [Name]!"
And I see the page title "Dashboard"

Given I am viewing the dashboard
When the page loads
Then I see four main sections in this order:
  1. Pending Approvals (top priority)
  2. Activity Timeline (recent org events)
  3. Quick Stats (org metrics)
  4. Quick Actions (navigation links)

# Pending Approvals Section
Given there are 3 pending approval requests
When I view the Pending Approvals section
Then I see "Pending Approvals (3)" as the section title
And I see a list of up to 5 most recent pending requests
And each request shows:
  - Employee name who made the request
  - Request type (e.g., "Agent Request" or "MCP Server Request")
  - Requested item name (e.g., "Claude Code")
  - Time since request (e.g., "2 days ago")
And I see a "View All Requests" button at the bottom
And clicking it takes me to /approvals

Given there are no pending approvals
When I view the Pending Approvals section
Then I see "Pending Approvals (0)" as the section title
And I see "No pending approvals" message
And I see "All caught up!" encouraging text
And I do NOT see "View All Requests" button

# Activity Timeline Section
Given there are recent activities in my organization
When I view the Activity Timeline section
Then I see "Activity Timeline" as the section title
And I see a list of up to 10 most recent activity log entries
And each entry shows:
  - Activity icon (🟢 create, 🔵 update, 🟡 approve, 🔴 delete)
  - Activity description (e.g., "Sarah approved agent for Bob")
  - Timestamp (e.g., "5 minutes ago", "2 hours ago", "Yesterday")
And I see a "View All Activity" button at the bottom
And clicking it takes me to /activity

Given there is no recent activity
When I view the Activity Timeline section
Then I see "Activity Timeline" as the section title
And I see "No recent activity" message
And I see "Activity will appear here as your team uses Ubik"

# Quick Stats Section
Given I am viewing the dashboard
When I look at the Quick Stats section
Then I see four stat cards displayed horizontally:
  1. "Active Employees" with count (e.g., "12")
  2. "Agents in Use" with count (e.g., "8")
  3. "Pending Requests" with count (e.g., "3")
  4. "This Month Usage" with cost (e.g., "$324")
And each stat card shows the metric name and value clearly
And the values are fetched from the API in real-time

# Quick Actions Section
Given I am viewing the dashboard
When I look at the Quick Actions section
Then I see "Quick Actions" as the section title
And I see four action buttons:
  - "Manage Employees" (links to /employees)
  - "View Teams" (links to /teams)
  - "Agent Catalog" (links to /agents)
  - "Usage Reports" (links to /usage)
And each button is clearly labeled and clickable

# Loading States
Given I just landed on the dashboard
When the data is still loading from the API
Then I see skeleton loaders for each section
And I see a loading spinner in the Pending Approvals section
And I see a loading spinner in the Activity Timeline section
And I see "Loading..." placeholders for Quick Stats

# Error States
Given the API fails to load pending approvals
When I view the dashboard
Then I see an error message "Unable to load pending approvals"
And I see a "Retry" button
And other sections still load successfully (fail gracefully)
```

**Implementation Notes:**
- **Current State:** Basic shell exists (welcome message + placeholder cards)
- **Needs Rework:** Replace placeholder cards with functional sections
- **API Endpoints:**
  - `GET /api/v1/approvals/pending?limit=5` - Get pending approval requests
  - `GET /api/v1/activity-logs?limit=10` - Get recent activity log entries
  - `GET /api/v1/organizations/current/stats` - Get org-level stats
- **Data Refresh:** Auto-refresh every 60 seconds (optional polling)
- **Responsive Design:** Stack sections vertically on mobile, side-by-side on desktop
- **Icons:** Use appropriate icons for activities (see acceptance criteria)
- **Time Formatting:** Use relative time (e.g., "2 hours ago", "Yesterday", "Last week")
- **Access Control:** This page is for admins/managers only (employees should see different content or be redirected)
- **Empty States:** Show encouraging messages when there's no data
- **Error Handling:** Fail gracefully - if one section errors, others still work

**Related Endpoints:**
- `GET /api/v1/employees` (for employee count)
- `GET /api/v1/agents` (for agents in use count)
- `GET /api/v1/usage-records` (for monthly usage cost)

---

**Next story to add:** Story 2.2 - Navigation & Layout (Sidebar/Header)

Is Story 2.1 approved?
