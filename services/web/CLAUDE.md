# Web UI Development Guide

**You are working on the Ubik Web UI** - the Next.js admin panel for the Ubik Enterprise platform.

---

## Quick Context

**What is this service?**
Next.js 14 admin panel providing web-based management interface for organizations, employees, teams, and AI agent configurations.

**Key capabilities:**
- JWT authentication with httpOnly cookies
- Dashboard for central management hub
- Employee, team, organization management
- Agent configuration UI
- Approval workflows
- Analytics and usage tracking
- Dark mode support

**Tech stack:** Next.js 14 (App Router), TypeScript, Tailwind CSS, shadcn/ui, React Query

---

## Essential Commands

Run `make` to see available commands (from services/web/ or repository root).

---

## Code Generation

**CRITICAL: Regenerate types after OpenAPI spec changes**

### Source Files (Edit These)
- `../../platform/api-spec/spec.yaml` - OpenAPI specification

### Generated Files (Never Edit)
- `lib/api/schema.ts` - TypeScript types (NOT committed to git)

### Workflow

```bash
# 1. Edit OpenAPI spec
vim ../../platform/api-spec/spec.yaml

# 2. Regenerate API code (from root)
cd ../.. && make generate-api

# 3. Regenerate TypeScript types
cd services/web
npm run generate:api

# 4. Use types in components
import type { components } from '@/lib/api/schema'
type Employee = components['schemas']['Employee']
```

**IMPORTANT:** TypeScript types are NOT committed to git. CI/CD regenerates them automatically.

---

## Architecture

### Next.js 14 App Router

```
app/
├── (auth)/            # Auth routes (login, register)
│   └── login/
├── (dashboard)/       # Protected dashboard routes
│   ├── dashboard/     # Main dashboard
│   ├── employees/     # Employee management (planned)
│   ├── agents/        # Agent configuration (planned)
│   └── layout.tsx     # Dashboard layout with auth
├── globals.css        # Global styles
└── layout.tsx         # Root layout
```

### Request Flow

```
Browser → Middleware (auth check) → Server Component → API Client → API Server
                                         ↓
                                   Client Component (interactive)
```

### Component Patterns

**Server Components (default):**
- Fetch data on server
- No JavaScript sent to client
- Better performance, SEO
- Use for static content, data fetching

**Client Components (interactive):**
- Add `'use client'` directive
- Use React hooks, event handlers
- Required for forms, buttons, interactivity

**Example:**
```typescript
// app/dashboard/page.tsx (Server Component)
export default async function DashboardPage() {
  const data = await fetchData() // Runs on server
  return <DashboardView data={data} />
}

// components/login-form.tsx (Client Component)
'use client'
export function LoginForm() {
  const [email, setEmail] = useState('')
  // Interactive form logic
}
```

---

## API Integration

### Type-Safe API Client

```typescript
// lib/api/client.ts
import createClient from 'openapi-fetch'
import type { paths } from './schema'

export const apiClient = createClient<paths>({
  baseUrl: process.env.NEXT_PUBLIC_API_URL,
})

// Usage in components
const { data, error } = await apiClient.GET('/api/v1/employees', {
  params: { query: { org_id: orgID } }
})
```

### Authentication Flow

1. User visits `/` → middleware redirects based on auth state
2. Login at `/login` → server action calls API
3. JWT stored in httpOnly cookie
4. Middleware protects dashboard routes
5. Dashboard fetches user info from `/auth/me`
6. Logout clears cookie and redirects

**Server action example:**
```typescript
// app/(dashboard)/actions.ts
'use server'
export async function logout() {
  cookies().delete('auth_token')
  redirect('/login')
}
```

---

## UI Development Workflow

**CRITICAL:** User stories first, THEN wireframes, THEN implementation

### Mandatory Workflow

1. **Define user stories FIRST:**
   - What user needs to accomplish
   - Who the user is
   - Why they need this feature

2. **Request wireframes from product-designer agent:**
   - Provide user stories
   - Location: `../../docs/wireframes/`

3. **Wait for wireframes approval:**
   - Review wireframes
   - Understand user flows

4. **Implement UI matching wireframes:**
   - Use shadcn/ui components
   - Follow Tailwind CSS conventions
   - Maintain accessibility standards

5. **NEVER implement without user stories + wireframes:**
   - ❌ No ad-hoc UI design
   - ✅ Always follow approved process

**See [../../.claude/agents/product-designer.md](../../.claude/agents/product-designer.md) for designer agent.**

---

## Testing Strategy

**CRITICAL: ALWAYS follow strict TDD (Test-Driven Development)**

### TDD Workflow (Mandatory)
1. ✅ Write failing test FIRST
2. ✅ Implement minimal code to pass test
3. ✅ Refactor with tests passing
4. ❌ NEVER write implementation before tests

### Test Types

**Unit Tests** (Vitest):
- Test components, utilities
- Mock API calls
- Fast execution (<1s)
- Located alongside code or in `__tests__/`

**E2E Tests** (Playwright):
- Test complete user workflows
- Real browser automation
- Test authentication, navigation
- Located in `tests/e2e/`
- Require API server running

### Running Tests

```bash
# Unit tests
npm test              # Run once
npm run test:watch    # Watch mode
npm run test:ui       # UI mode

# E2E tests (requires API running)
npm run test:e2e      # Run all E2E tests
npm run test:e2e -- --ui  # UI mode
npm run test:e2e -- --debug  # Debug mode
```

### Test Patterns

**Component testing:**
```typescript
// LoginForm.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'

describe('LoginForm', () => {
  it('validates email format', () => {
    render(<LoginForm />)
    const input = screen.getByLabelText('Email')
    fireEvent.change(input, { target: { value: 'invalid' } })
    expect(screen.getByText('Invalid email')).toBeInTheDocument()
  })
})
```

**E2E testing:**
```typescript
// tests/e2e/auth.spec.ts
test('user can login', async ({ page }) => {
  await page.goto('http://localhost:3000/login')
  await page.fill('input[name="email"]', 'user@example.com')
  await page.fill('input[name="password"]', 'password123')
  await page.click('button[type="submit"]')
  await expect(page).toHaveURL('/dashboard')
})
```

**See [../../docs/TESTING.md](../../docs/TESTING.md) for complete testing guide.**

---


## Styling

### Tailwind CSS

```typescript
// Use Tailwind utility classes
<div className="flex items-center justify-between p-4 bg-white dark:bg-gray-800">
  <h1 className="text-2xl font-bold">Dashboard</h1>
</div>
```

### shadcn/ui Components

```typescript
// Import from components/ui
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'

// Use with Tailwind classes
<Button variant="default" size="lg">Click Me</Button>
```

### Dark Mode

Implemented using `next-themes`:
- System preference detection
- Manual toggle in header
- Persists across sessions
- No flash on page load

```typescript
// Use dark mode classes
<div className="bg-white dark:bg-gray-900">
  <p className="text-gray-900 dark:text-gray-100">Content</p>
</div>
```

---

## Common Tasks

### Adding New Page

1. **Define user stories**
2. **Request wireframes from product-designer agent**
3. **Wait for approval**
4. **Create route:**
   ```typescript
   // app/(dashboard)/employees/page.tsx
   export default function EmployeesPage() {
     return <div>Employees</div>
   }
   ```
5. **Add to navigation**
6. **Write tests**
7. **Implement UI matching wireframes**

### Adding Form

1. **Use React Hook Form + Zod:**
   ```typescript
   import { useForm } from 'react-hook-form'
   import { z } from 'zod'

   const schema = z.object({
     email: z.string().email(),
     password: z.string().min(8),
   })

   const { register, handleSubmit } = useForm({
     resolver: zodResolver(schema),
   })
   ```

2. **Use shadcn/ui form components**
3. **Add validation**
4. **Test all cases**

### Calling API

```typescript
// Server Component (preferred)
export default async function Page() {
  const { data } = await apiClient.GET('/api/v1/employees')
  return <EmployeeList employees={data} />
}

// Client Component (interactive)
'use client'
export function EmployeeForm() {
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (data) => {
    setLoading(true)
    const { error } = await apiClient.POST('/api/v1/employees', { body: data })
    setLoading(false)
  }
}
```

---

## Common Pitfalls

### 1. Server/Client Component Confusion

```typescript
// ❌ BAD - Using useState in Server Component
export default function Page() {
  const [count, setCount] = useState(0)  // ERROR!
}

// ✅ GOOD - Add 'use client'
'use client'
export default function Page() {
  const [count, setCount] = useState(0)  // OK
}
```

### 2. Missing User Stories/Wireframes

```typescript
// ❌ BAD - Implementing UI without user stories + wireframes
// Just making it up as I go...

// ✅ GOOD - Following user stories → wireframes → implementation
// Based on user story + docs/wireframes/employees-list.png
```

### 3. Stale Types

```bash
# ✅ Regenerate after API spec changes
npm run generate:api
```

### 4. API Not Running

```bash
# ✅ Start API server (from root)
make dev-api

# ✅ Verify API
curl http://localhost:8080/api/v1/health
```

**See [../../docs/DEBUGGING.md](../../docs/DEBUGGING.md) for debugging strategies.**

---

## Environment Variables

Create `.env.local`:

```bash
# API backend URL
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

**IMPORTANT:** `NEXT_PUBLIC_*` variables are exposed to browser.

---

## Docker & Deployment

### Local Docker Testing

```bash
# From repository root
docker build -f services/web/Dockerfile -t ubik-web .

# Test container
docker run -p 3000:3000 \
  -e NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1 \
  ubik-web

# Verify
curl http://localhost:3000
```

### Docker Compose

```bash
# Start all services (from root)
docker compose up

# Web UI available at http://localhost:3000
```

---

## Related Documentation

- [../../CLAUDE.md](../../CLAUDE.md) - Monorepo overview
- [../../docs/TESTING.md](../../docs/TESTING.md) - Testing guide
- [../../docs/DEV_WORKFLOW.md](../../docs/DEV_WORKFLOW.md) - PR workflow
- [../../.claude/agents/product-designer.md](../../.claude/agents/product-designer.md) - Product designer agent
- [../api/CLAUDE.md](../api/CLAUDE.md) - API development
- [../cli/CLAUDE.md](../cli/CLAUDE.md) - CLI development
