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

**Next story to add:** Story 1.4 - Dashboard Overview

Is Story 1.3 approved?
