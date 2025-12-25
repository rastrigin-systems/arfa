# Arfa Web UI

Next.js 14 admin panel for the Arfa Enterprise platform.

## Overview

The Web UI provides:
- **Authentication & Sessions:** JWT-based authentication with httpOnly cookies
- **Dashboard:** Central hub for managing AI agents and configurations
- **Employee Management:** View and manage employee accounts (planned)
- **Agent Configuration:** Configure AI agents (Claude Code, Cursor, etc.) (planned)
- **MCP Configuration:** Manage Model Context Protocol servers (planned)
- **Team Management:** Organize employees into teams (planned)
- **Approval Workflows:** Request and approve access to agents/MCPs (planned)
- **Analytics:** Usage tracking and cost monitoring (planned)
- **Dark Mode:** System-aware theme with manual toggle

## Quick Start

### Prerequisites
- Node.js 18+
- npm (or pnpm)
- API backend running on `http://localhost:8080`

### Setup

```bash
# From project root
make db-up          # Start PostgreSQL
make dev-api        # Start API server (separate terminal)

# Install and run web UI
cd services/web
npm install         # Install dependencies
npm run generate:api  # Generate API types from OpenAPI spec
npm run dev         # Start development server (http://localhost:3000)
```

### Development

```bash
# Run development server
npm run dev

# Type checking
npm run type-check

# Linting
npm run lint

# Unit tests
npm test

# E2E tests with Playwright
npm run test:e2e

# Generate API types (after OpenAPI spec changes)
npm run generate:api
```

### Build & Deployment

```bash
# Build for production
npm run build

# Start production server
npm start

# Docker build (from project root)
docker build -f services/web/build/Dockerfile -t arfa-web .
```

## Project Structure

```
services/web/
├── app/                      # Next.js 14 App Router
│   ├── (auth)/
│   │   └── login/            # Login page
│   ├── (dashboard)/
│   │   ├── dashboard/        # Main dashboard
│   │   ├── layout.tsx        # Dashboard layout with auth
│   │   └── actions.ts        # Server actions (logout)
│   ├── globals.css           # Global styles
│   └── layout.tsx            # Root layout
├── components/               # React components
│   ├── ui/                   # shadcn/ui components
│   ├── dashboard-header.tsx  # Header with logout
│   ├── theme-provider.tsx    # Theme context
│   └── theme-toggle.tsx      # Dark mode toggle
├── lib/                      # Libraries and utilities
│   ├── api/
│   │   ├── schema.ts         # Generated API types (auto-generated)
│   │   └── client.ts         # API client wrapper
│   ├── auth.ts               # Auth helpers
│   └── utils.ts              # Utility functions
├── tests/
│   └── e2e/                  # Playwright E2E tests
├── build/                    # Deployment configurations
│   └── Dockerfile            # Production Docker build
├── scripts/                  # Service-specific scripts (currently empty)
├── docs/                     # Service-specific docs (currently empty)
├── middleware.ts             # Auth middleware (route protection)
├── next.config.js            # Next.js configuration
├── tailwind.config.ts        # Tailwind CSS config
├── playwright.config.ts      # Playwright test config
├── vitest.config.ts          # Vitest test config
└── package.json              # Dependencies and scripts
```

## Technology Stack

- **Framework:** Next.js 14 (App Router, React Server Components)
- **Language:** TypeScript (strict mode)
- **Styling:** Tailwind CSS
- **UI Components:** shadcn/ui (Radix UI primitives)
- **API Client:** openapi-fetch (type-safe, generated from OpenAPI spec)
- **State Management:** React Query (@tanstack/react-query)
- **Forms:** React Hook Form + Zod validation
- **Testing:** Vitest (unit/integration), Playwright (E2E)
- **Theme:** next-themes (dark mode support)

## Features

### Current Features ✅
- JWT authentication with httpOnly cookies
- Protected dashboard routes with middleware
- Dark mode support (system-aware + manual toggle)
- Type-safe API client (auto-generated from OpenAPI spec)
- Responsive design (mobile, tablet, desktop)
- Accessibility (WCAG AA compliant)
- E2E tests with Playwright

### Planned Features 🚧
- Employee management UI (Issue #4)
- Agent configuration UI (Issue #5)
- MCP server management UI (Issue #6)
- Team management UI (Issue #7)
- Approval workflow UI (Issue #8)
- Analytics dashboard (Issue #9)

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

The API client is auto-generated from the OpenAPI specification:

**Generate types:**
```bash
npm run generate:api
```

This reads `../../platform/api-spec/spec.yaml` and generates `lib/api/schema.ts`.

**Example usage:**
```typescript
import { apiClient } from '@/lib/api/client';

// Login
const { data, error } = await apiClient.POST('/auth/login', {
  body: {
    email: 'user@example.com',
    password: 'password123',
  },
});

// Get current employee (with auth)
const { data } = await apiClient.GET('/auth/me', {
  headers: {
    Authorization: `Bearer ${token}`,
  },
});
```

**Key endpoints:**
- `POST /api/v1/auth/register` - Employee registration
- `POST /api/v1/auth/login` - Login and get JWT
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current employee info
- See full API docs at `http://localhost:8080/api/docs`

## Environment Variables

Create `.env.local`:

```bash
# API backend URL
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Testing

### Unit & Integration Tests

```bash
# Run Vitest tests
npm test

# Watch mode
npm run test:ui
```

### E2E Tests

```bash
# Run Playwright tests
npm run test:e2e

# Run with UI
npm run test:e2e -- --ui

# Debug mode
npm run test:e2e -- --debug
```

**Test coverage:**
- Login form validation
- Authentication flow
- Protected route redirection
- Logout functionality
- Keyboard navigation
- Accessibility (ARIA labels)

**Note:** Some tests require the API backend to be running. Tests that require backend integration are skipped by default.

## Accessibility

All UI components follow WCAG AA standards:

- ✅ Semantic HTML elements (`<nav>`, `<main>`, `<button>`)
- ✅ Proper ARIA labels and roles
- ✅ Keyboard navigation support
- ✅ Focus visible indicators
- ✅ Screen reader friendly error messages
- ✅ Sufficient color contrast
- ✅ Responsive text sizing

**Tools used:**
- axe DevTools for automated accessibility testing
- Manual keyboard navigation testing
- Screen reader testing (VoiceOver, NVDA)

## Dark Mode

Implemented using `next-themes`:

- ✅ System preference detection
- ✅ Manual toggle in dashboard header
- ✅ Persists across sessions (localStorage)
- ✅ No flash on page load (SSR-safe)
- ✅ Smooth transitions

## Build & Deployment

### Local Build

```bash
# Build production bundle
npm run build

# Start production server (port 3000)
npm start
```

### Docker Build

**From project root:**
```bash
# Build Docker image
docker build -f services/web/build/Dockerfile -t arfa-web .

# Run container (port 8080)
docker run -p 8080:8080 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 \
  arfa-web
```

**Dockerfile details:**
- Multi-stage build (deps → builder → production)
- Uses Next.js standalone output for minimal image size
- Runs as non-root user (nextjs:nodejs)
- Optimized for GCP Cloud Run

### GCP Cloud Run

See `../../docs/GCP_DEPLOYMENT_GUIDE.md` for deployment instructions.

**Quick deploy:**
```bash
# Build and push to Artifact Registry
gcloud builds submit --config=cloudbuild-web.yaml

# Deploy to Cloud Run
gcloud run deploy arfa-web \
  --image=us-central1-docker.pkg.dev/$PROJECT_ID/arfa-images/web:latest \
  --region=us-central1 \
  --allow-unauthenticated
```

## Development Workflow

### Standard Development

```bash
# Start API backend (terminal 1)
cd /path/to/arfa
make dev-api

# Start web UI (terminal 2)
cd services/web
npm run dev

# Visit http://localhost:3000
```

### Using Docker Compose (Full Stack)

```bash
# From project root
docker-compose up

# Access services:
# - Web UI: http://localhost:3000
# - API: http://localhost:8080
# - PostgreSQL: localhost:5432
```

### Code Generation

**After changing OpenAPI spec:**
```bash
# From project root
make generate-api

# Then regenerate web types
cd services/web
npm run generate:api
```

## Troubleshooting

### Port already in use

If port 3000 is in use:

```bash
# Option 1: Change port in package.json
"dev": "next dev -p 3001"

# Option 2: Set PORT environment variable
PORT=3001 npm run dev
```

### API connection errors

Ensure the API backend is running on `http://localhost:8080`:

```bash
# From project root
make dev-api
```

Check `.env.local` has correct API URL:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

### Type errors after OpenAPI spec changes

Regenerate types:
```bash
npm run generate:api
```

If errors persist, delete and reinstall:
```bash
rm -rf node_modules .next
npm install
npm run generate:api
npm run build
```

### Tests failing

**E2E tests require API backend:**
```bash
# Terminal 1: Start API
cd /path/to/arfa
make dev-api

# Terminal 2: Run E2E tests
cd services/web
npm run test:e2e
```

**Unit tests should work standalone:**
```bash
npm test
```

## Architecture Patterns

### Server vs Client Components

**Server Components (default):**
- Fetch data on server
- No JavaScript sent to client
- Better performance, SEO

```typescript
// app/dashboard/page.tsx
export default async function DashboardPage() {
  const data = await fetchData() // Runs on server
  return <DashboardView data={data} />
}
```

**Client Components (interactive):**
- Use hooks, event handlers
- Add `'use client'` directive

```typescript
// components/login-form.tsx
'use client'
export function LoginForm() {
  const [email, setEmail] = useState('')
  // Interactive form logic
}
```

### API Client Pattern

**Centralized client:**
```typescript
// lib/api/client.ts
import createClient from 'openapi-fetch'
import type { paths } from './schema'

export const apiClient = createClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_URL,
})
```

**Usage in components:**
```typescript
const { data, error } = await apiClient.GET('/auth/me')
```

### Authentication

**Middleware protects routes:**
```typescript
// middleware.ts
export function middleware(request: NextRequest) {
  const token = request.cookies.get('auth_token')
  if (!token && !isPublicPath(request.nextUrl.pathname)) {
    return NextResponse.redirect(new URL('/login', request.url))
  }
}
```

**Server actions for mutations:**
```typescript
// app/(dashboard)/actions.ts
'use server'
export async function logout() {
  cookies().delete('auth_token')
  redirect('/login')
}
```

## Contributing

### Development Guidelines

1. **Follow TDD:** Write tests before implementation
2. **Type Safety:** Use TypeScript strictly (no `any`)
3. **Accessibility:** WCAG AA compliance required
4. **Responsive:** Test mobile, tablet, desktop
5. **Performance:** Minimize client-side JS, use Server Components
6. **Code Style:** Run `npm run lint` before committing

### PR Checklist

- [ ] Tests passing (`npm test` + `npm run test:e2e`)
- [ ] Type checking passing (`npm run type-check`)
- [ ] Linting passing (`npm run lint`)
- [ ] Accessibility verified (keyboard nav, screen reader)
- [ ] Responsive design tested (mobile/tablet/desktop)
- [ ] Dark mode tested
- [ ] Production build works (`npm run build`)

## Next Steps

**Current:** Issue #12 - Web UI Foundation & Authentication ✅

**Upcoming:**
- Issue #4: Employee management UI
- Issue #5: Agent configuration UI
- Issue #6: MCP server management UI
- Issue #7: Team management UI
- Issue #8: Approval workflow UI
- Issue #9: Analytics dashboard

See `../../docs/IMPLEMENTATION_ROADMAP.md` for detailed roadmap.

## Resources

- **Next.js Docs:** https://nextjs.org/docs
- **shadcn/ui:** https://ui.shadcn.com
- **Tailwind CSS:** https://tailwindcss.com
- **React Query:** https://tanstack.com/query
- **Playwright:** https://playwright.dev
- **OpenAPI Spec:** `../../platform/api-spec/spec.yaml`
- **API Docs:** http://localhost:8080/api/docs (when API running)
