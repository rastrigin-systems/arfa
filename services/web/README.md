# Arfa Web UI

Next.js 14 admin panel for the Arfa platform.

## Quick Start

```bash
# From project root
make db-up              # Start PostgreSQL
make dev-api            # Start API server (separate terminal)

# Install and run web UI
cd services/web
pnpm install            # Install dependencies
pnpm run generate:api   # Generate API types from OpenAPI spec
pnpm run dev            # Start dev server (http://localhost:3000)
```

## Commands

```bash
pnpm run dev            # Development server
pnpm run build          # Production build
pnpm run type-check     # TypeScript check
pnpm run lint           # ESLint
pnpm test               # Vitest unit tests
pnpm run test:e2e       # Playwright E2E tests
pnpm run generate:api   # Regenerate API types
```

## Project Structure

```
services/web/
├── app/                    Next.js 14 App Router
│   ├── (auth)/login/       Login page
│   ├── (dashboard)/        Protected dashboard routes
│   └── layout.tsx          Root layout
├── components/
│   ├── ui/                 shadcn/ui components
│   └── ...                 Feature components
├── lib/
│   ├── api/
│   │   ├── schema.ts       Generated API types (auto-generated)
│   │   └── client.ts       API client wrapper
│   └── utils.ts            Utilities
├── tests/e2e/              Playwright tests
├── middleware.ts           Auth middleware
└── package.json
```

## Tech Stack

- Next.js 14 (App Router, React Server Components)
- TypeScript (strict mode)
- Tailwind CSS + shadcn/ui
- openapi-fetch (type-safe API client)
- React Query, React Hook Form + Zod
- Vitest + Playwright

## API Integration

Types are auto-generated from OpenAPI spec:

```bash
pnpm run generate:api   # After spec changes
```

Usage:
```typescript
import { apiClient } from '@/lib/api/client';

const { data, error } = await apiClient.POST('/auth/login', {
  body: { email: 'user@example.com', password: 'password123' },
});
```

## Environment Variables

Create `.env.local`:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## Authentication Flow

1. User visits `/` -> redirects to `/login` or `/dashboard`
2. Login calls `POST /auth/login`, stores JWT in httpOnly cookie
3. Middleware protects routes, redirects unauthenticated users
4. Logout clears cookie and redirects to `/login`

## Docker

```bash
# From project root
docker build -f services/web/build/Dockerfile -t arfa-web .
docker run -p 3000:3000 -e NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 arfa-web
```

## Documentation

- [Architecture](../../docs/architecture/overview.md)
- [Testing](../../docs/development/testing.md)
- [Contributing](../../docs/development/contributing.md)
- API Docs: http://localhost:8080/api/docs (when API running)
