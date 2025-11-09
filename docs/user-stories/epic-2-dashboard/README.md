# Epic 2: Dashboard & Navigation

## Overview

This epic covers the admin dashboard, navigation structure, and overall UI layout for the Ubik Enterprise platform. Focus is on admin-first experience with activity monitoring, approval management, and quick access to key features.

## Stories

- **[2.1 Admin Dashboard](./2.1-admin-dashboard.md)** 🚧 In Progress
  - Pending approvals overview
  - Activity timeline
  - Quick stats (employees, agents, usage)
  - Quick actions navigation

## Planned Stories
- **2.2 Navigation & Layout** - Sidebar/header navigation structure
- **2.3 User Menu** - Profile, settings, logout dropdown
- **2.4 Breadcrumbs** - Page navigation breadcrumbs

## API Endpoints

- `GET /api/v1/approvals/pending` - Get pending approval requests
- `GET /api/v1/activity-logs` - Get activity log entries
- `GET /api/v1/organizations/current/stats` - Get org-level statistics
- `GET /api/v1/employees` - Get employee count
- `GET /api/v1/agents` - Get agent count
- `GET /api/v1/usage-records` - Get usage/cost data

## UI Pages

- `/dashboard` - Admin dashboard (main landing page after login)

## Design Principles

- **Admin-first:** Dashboard optimized for managers/approvers
- **Activity-focused:** Highlight what needs attention (pending approvals)
- **Quick access:** One-click navigation to key admin tasks
- **Real-time data:** Auto-refresh every 60 seconds
- **Responsive:** Mobile-friendly layout
- **Empty states:** Encouraging messages when no data
- **Error handling:** Graceful degradation if API fails

## Dependencies

- All backend APIs must be implemented
- Activity logging system operational
- Approval workflow functional

## Notes

- Regular employees (non-admin) may see different dashboard content or be redirected to `/agents` in future iterations
- Current focus is on admin experience only
