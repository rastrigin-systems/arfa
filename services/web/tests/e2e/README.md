# E2E Testing with MSW (Mock Service Worker)

This directory contains end-to-end tests for the Arfa Enterprise web application using Playwright and MSW for API mocking.

## Architecture

### Problem Solved

Next.js Server Components make API calls **server-side** during SSR, which cannot be intercepted by Playwright's client-side `page.route()` mocking. This caused tests to fail with `ECONNREFUSED` errors when trying to reach the real backend API at `http://localhost:8080`.

###Solution: MSW with Next.js Instrumentation

We use **MSW (Mock Service Worker)** to intercept API calls at the network level in both:
1. **Browser** (client-side fetch calls)
2. **Node.js** (server-side fetch calls in Next.js)

MSW starts automatically when the Next.js server boots up during E2E tests via the `instrumentation.ts` hook.

## File Structure

```
tests/e2e/
├── README.md (this file)
├── fixtures/
│   ├── index.ts          # Playwright fixtures (mockApi, mockEmployees, mockAgents)
│   ├── apiMocks.ts       # Legacy Playwright route mocks (deprecated)
│   └── mockData.ts       # Mock API response data
├── mocks/
│   ├── handlers.ts       # MSW request handlers
│   ├── server.ts         # MSW server instance
│   └── helpers.ts        # Test helpers for custom MSW behaviors
├── *.spec.ts             # Test files
└── playwright.config.ts  # Playwright configuration
```

## How It Works

### 1. Next.js Instrumentation (`instrumentation.ts`)

```typescript
export async function register() {
  if (process.env.E2E_TEST === 'true') {
    const { server } = await import('./tests/e2e/mocks/server');
    server.listen({ onUnhandledRequest: 'warn' });
  }
}
```

When `E2E_TEST=true`, MSW starts in the Next.js server process and intercepts all fetch calls.

### 2. Playwright Config (`playwright.config.ts`)

```typescript
webServer: {
  command: 'npm run build && npm start',
  env: {
    E2E_TEST: 'true',  // Enables MSW
  },
}
```

### 3. MSW Handlers (`mocks/handlers.ts`)

All API endpoints are mocked:
- `/api/v1/auth/*` - Authentication
- `/api/v1/employees/*` - Employee CRUD
- `/api/v1/agents/*` - Agent catalog
- `/api/v1/organizations/current/agent-configs` - Agent configs
- etc.

### 4. Test Fixtures (`fixtures/index.ts`)

Tests use fixtures that automatically set up mocking and authentication:

```typescript
test('my test', async ({ mockEmployees }) => {
  // mockEmployees fixture includes:
  // - MSW handlers for employee endpoints
  // - Mock authenticated session
  // - Ready to navigate to authenticated pages
  await mockEmployees.goto('/dashboard/employees');
  // ...
});
```

## Available Fixtures

| Fixture | Use Case | Includes |
|---------|----------|----------|
| `mockApi` | Full API mocking | All endpoints mocked |
| `mockAuth` | Auth endpoints only | Login/logout mocked |
| `mockEmployees` | Employee features | Employee + auth + teams + roles |
| `mockAgents` | Agent features | Agents + auth + configs |
| `authenticatedPage` | Authenticated session | Full API + mock session cookie |

## Running Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run specific test file
npx playwright test agent-configs.spec.ts

# Run in headed mode (see browser)
npx playwright test --headed

# Run in debug mode
npx playwright test --debug

# Run tests matching pattern
npx playwright test --grep "should render page"
```

## Writing Tests

### Basic Test

```typescript
import { test, expect } from './fixtures';

test('should display employees', async ({ mockEmployees }) => {
  await mockEmployees.goto('/dashboard/employees');

  // Assertions
  await expect(mockEmployees.getByRole('heading', { name: 'Employees' })).toBeVisible();
});
```

### Testing Edge Cases

For testing slow networks, empty states, or errors, use MSW helpers:

```typescript
import { test, expect } from './fixtures';
import { withMockHandler, createEmptyHandler } from './mocks/helpers';
import { http } from 'msw';

test('should show empty state', async ({ mockEmployees }) => {
  await withMockHandler(
    createEmptyHandler('get', 'http://localhost:8080/api/v1/employees'),
    async () => {
      await mockEmployees.goto('/dashboard/employees');
      await expect(mockEmployees.getByText(/no employees/i)).toBeVisible();
    }
  );
});
```

## Troubleshooting

### Tests failing with `ECONNREFUSED`

**Cause**: MSW is not starting or E2E_TEST env var is not set.

**Fix**: Check that:
1. `instrumentation.ts` exists in project root
2. `next.config.js` has `experimental.instrumentationHook: true`
3. Playwright config passes `E2E_TEST: 'true'` to webServer

### Tests failing with redirect loops

**Cause**: Authentication is not properly mocked.

**Fix**: Use `mockEmployees`, `mockAgents`, or `authenticatedPage` fixtures instead of plain `page`.

### Server-side API calls not mocked

**Cause**: MSW handlers don't match the request URL.

**Fix**: Check that MSW handlers in `mocks/handlers.ts` use the exact API base URL (`http://localhost:8080/api/v1`).

## Known Limitations

1. **Playwright route mocking conflicts**: Tests using `page.route()` for client-side edge cases will fail. Use MSW helpers instead.
2. **Static generation**: Pages with `generateStaticParams` may not use MSW. Convert to dynamic rendering for tests.
3. **Response time**: MSW adds ~50-100ms latency compared to real API.

## Migration from Playwright Route Mocking

**Before (Playwright routes):**
```typescript
await page.route('**/api/v1/employees', async (route) => {
  await route.fulfill({
    status: 200,
    body: JSON.stringify({ employees: [] }),
  });
});
```

**After (MSW):**
```typescript
import { withMockHandler, createEmptyHandler } from './mocks/helpers';

await withMockHandler(
  createEmptyHandler('get', 'http://localhost:8080/api/v1/employees'),
  async () => {
    // Test code
  }
);
```

## References

- [MSW Documentation](https://mswjs.io/)
- [Playwright Documentation](https://playwright.dev/)
- [Next.js Instrumentation](https://nextjs.org/docs/app/building-your-application/optimizing/instrumentation)
- [Testing Library](https://testing-library.com/docs/queries/about/)
