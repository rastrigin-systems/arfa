# Ubik Enterprise Web UI

Next.js 14 web application for managing AI agent configurations.

## Features

- Next.js 14 App Router
- TypeScript with strict mode
- Tailwind CSS for styling
- shadcn/ui components
- JWT authentication with httpOnly cookies
- Protected dashboard routes
- Dark mode support
- Type-safe API client (generated from OpenAPI spec)
- E2E tests with Playwright

## Getting Started

### Prerequisites

- Node.js 18+
- npm or pnpm
- Go API backend running on `http://localhost:8080`

### Installation

```bash
# Install dependencies
npm install

# Generate API types from OpenAPI spec
npm run generate:api
```

### Development

```bash
# Start development server (http://localhost:3000)
npm run dev

# Type check
npm run type-check

# Lint
npm run lint

# Run E2E tests
npm run test:e2e
```

### Build

```bash
# Build for production
npm run build

# Start production server
npm start
```

## Project Structure

```
services/web/
├── app/
│   ├── (auth)/
│   │   └── login/              # Login page
│   ├── (dashboard)/
│   │   ├── dashboard/          # Dashboard page
│   │   ├── layout.tsx          # Dashboard layout with auth
│   │   └── actions.ts          # Server actions (logout)
│   ├── globals.css             # Global styles
│   └── layout.tsx              # Root layout
├── components/
│   ├── ui/                     # shadcn/ui components
│   ├── dashboard-header.tsx    # Header with logout
│   ├── theme-provider.tsx      # Theme context
│   └── theme-toggle.tsx        # Dark mode toggle
├── lib/
│   ├── api/
│   │   ├── schema.ts           # Generated API types
│   │   └── client.ts           # API client wrapper
│   ├── auth.ts                 # Auth helpers
│   └── utils.ts                # Utility functions
├── tests/
│   └── e2e/                    # Playwright tests
├── middleware.ts               # Auth middleware
├── next.config.js              # Next.js config
├── tailwind.config.ts          # Tailwind config
└── package.json
```

## Authentication Flow

1. User visits `/` → redirects to `/login` (if not authenticated) or `/dashboard` (if authenticated)
2. User enters credentials on `/login`
3. Server action calls `POST /auth/login` API
4. JWT token stored in httpOnly cookie
5. User redirected to `/dashboard`
6. Middleware protects all routes except `/login`
7. Dashboard layout fetches user info from `GET /auth/me`
8. Logout calls `POST /auth/logout` and clears cookie

## API Integration

API client is auto-generated from OpenAPI spec:

```typescript
import { apiClient } from '@/lib/api/client';

// Example: Login
const { data, error } = await apiClient.POST('/auth/login', {
  body: {
    email: 'user@example.com',
    password: 'password123',
  },
});

// Example: Get current employee (with auth)
const { data } = await apiClient.GET('/auth/me', {
  headers: {
    Authorization: `Bearer ${token}`,
  },
});
```

## Environment Variables

Create `.env.local`:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Testing

### E2E Tests

```bash
# Run E2E tests
npm run test:e2e

# Run with UI
npm run test:e2e -- --ui
```

E2E tests cover:
- Login form validation
- Authentication flow
- Protected route redirection
- Logout functionality
- Keyboard navigation
- Accessibility (ARIA labels)

**Note:** Some tests require the Go API backend to be running. Tests that require backend integration are skipped by default.

## Accessibility

All UI components follow WCAG AA standards:

- Semantic HTML elements
- Proper ARIA labels and roles
- Keyboard navigation support
- Focus visible indicators
- Screen reader friendly error messages

## Dark Mode

Dark mode is implemented using `next-themes`:

- System preference detection
- Manual toggle in dashboard header
- Persists across sessions
- No flash on page load

## Next Steps

This implements Issue #12 (Web UI Foundation & Authentication). Next steps:

- Issue #4: Employee management UI
- Issue #5: Agent configuration UI
- Issue #6: MCP server management UI
- Issue #7: Team management UI
- Issue #8: Approval workflow UI
- Issue #9: Analytics dashboard

## Troubleshooting

### Port already in use

If port 3000 is in use, change the port in `package.json`:

```json
"dev": "next dev -p 3001"
```

### API connection errors

Ensure the Go API backend is running on `http://localhost:8080` and accessible.

### Type errors after OpenAPI spec changes

Regenerate types:

```bash
npm run generate:api
```
