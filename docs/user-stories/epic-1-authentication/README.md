# Epic 1: Authentication & Onboarding

## Overview

This epic covers the complete authentication and user onboarding flow, including registration, login, logout, and password recovery.

## Stories

- **[1.1 User Registration](./1.1-user-registration.md)** ✅ Implemented
  - New organization signup with email/password
  - Real-time slug validation
  - Auto-owner role assignment

- **[1.2 User Login](./1.2-user-login.md)** 🚧 In Progress
  - Email/password authentication
  - Auto-redirect for authenticated users
  - Explicit logout requirement
  - Forgot password link

- **[1.3 User Logout](./1.3-user-logout.md)** ✅ Implemented
  - Session termination
  - Cookie clearing
  - Redirect to login

- **[1.4 Password Reset](./1.4-password-reset.md)** 📋 Planned
  - Email-based password reset
  - Token expiration and single-use
  - Security best practices

## API Endpoints

- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/forgot-password` - Request password reset (planned)
- `POST /api/v1/auth/reset-password` - Reset password with token (planned)
- `GET /api/v1/auth/verify-reset-token` - Verify reset token (planned)

## UI Pages

- `/signup` - Registration page
- `/login` - Login page
- `/forgot-password` - Password reset request (planned)
- `/reset-password/[token]` - Password reset form (planned)

## Dependencies

- JWT authentication with httpOnly cookies
- Session management in PostgreSQL
- Email service integration (for password reset - planned)

## Related Issues

- Issue #155 - Password reset flow implementation
