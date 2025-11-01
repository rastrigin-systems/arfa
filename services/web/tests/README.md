# E2E Testing Infrastructure

**Comprehensive Playwright E2E test suite for Ubik Enterprise web dashboard**

---

## Overview

This directory contains the E2E testing infrastructure for the Ubik Enterprise web application, built with Playwright.

### Current Status

**Test Coverage:**
- ✅ Authentication flows (11 passing tests)
- ⏳ Employee management (47 tests - pending UI completion)
- ⏳ Agent configuration (39 tests - pending UI completion)
- ⏳ Employee forms (11 tests - pending UI completion)

**Total:** 108 E2E tests (11 passing, 50 skipped, 47 failing due to incomplete UI)

---

## Directory Structure

```
tests/
├── e2e/                      # E2E test files
│   ├── auth.spec.ts          # Authentication tests ✅
│   ├── employees-list.spec.ts # Employee list tests
│   ├── employee-detail.spec.ts # Employee detail tests
│   ├── employee-form.spec.ts  # Employee form tests
│   └── agent-configs.spec.ts  # Agent configuration tests
│
├── fixtures/                 # Test fixtures (mock data)
│   ├── auth.ts               # Authentication fixtures
│   ├── employees.ts          # Employee fixtures
│   ├── agents.ts             # Agent fixtures
│   └── index.ts              # Fixture exports
│
├── helpers/                  # Test utilities
│   ├── auth.ts               # Auth helpers (login, cookies, etc.)
│   ├── api-mock.ts           # API mocking utilities
│   └── index.ts              # Helper exports
│
└── README.md                 # This file
```

---

## Quick Start

### Prerequisites

1. Node.js 20+ installed
2. Web application dependencies installed (`npm install`)
3. API server running (for integration tests)

### Running Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run tests in UI mode (interactive)
npm run test:e2e -- --ui

# Run specific test file
npm run test:e2e tests/e2e/auth.spec.ts

# Run tests in headed mode (see browser)
npm run test:e2e -- --headed

# Run tests in debug mode
npm run test:e2e -- --debug

# Generate test report
npx playwright show-report
```

### Running Tests in CI

Tests automatically run on GitHub Actions for:
- Pull requests touching `services/web/**`
- Pushes to `main` branch

See `.github/workflows/web-e2e.yml` for CI configuration.

---

## Writing Tests

### Basic Test Structure

```typescript
import { test, expect } from '@playwright/test';

test.describe('Feature Name', () => {
  test.beforeEach(async ({ page }) => {
    // Setup: Navigate to page, set auth, etc.
    await page.goto('/dashboard/feature');
  });

  test('should display feature correctly', async ({ page }) => {
    // Arrange
    const heading = page.getByRole('heading', { name: /feature/i });

    // Act & Assert
    await expect(heading).toBeVisible();
  });
});
```

### Using Fixtures

```typescript
import { test, expect } from '@playwright/test';
import { mockEmployees } from '../fixtures';
import { mockGetEmployees } from '../helpers';

test('should display employee list', async ({ page }) => {
  // Mock API response
  await mockGetEmployees(page, mockEmployees);

  // Navigate to page
  await page.goto('/dashboard/employees');

  // Verify employees are displayed
  await expect(page.getByText('Alice Johnson')).toBeVisible();
  await expect(page.getByText('Bob Smith')).toBeVisible();
});
```

### Using Authentication Helpers

```typescript
import { test, expect } from '@playwright/test';
import { setAuthCookies, loginViaUI } from '../helpers';

test('authenticated user can access dashboard', async ({ page }) => {
  // Option 1: Set auth cookies directly (faster)
  await setAuthCookies(page);
  await page.goto('/dashboard');

  // Option 2: Login via UI (more realistic)
  await loginViaUI(page, 'alice@acme.com', 'password123');
});
```

### Mocking API Responses

```typescript
import { test, expect } from '@playwright/test';
import { mockGetEmployees, mockAPIError } from '../helpers';
import { mockEmployees } from '../fixtures';

test('should handle API success', async ({ page }) => {
  await mockGetEmployees(page, mockEmployees);
  await page.goto('/dashboard/employees');
  // Assertions...
});

test('should handle API error', async ({ page }) => {
  await mockAPIError(page, '**/api/v1/employees*', 500, 'Server Error');
  await page.goto('/dashboard/employees');
  await expect(page.getByText(/error/i)).toBeVisible();
});
```

---

## Best Practices

### 1. Test Pyramid Principle

- **E2E Tests** (~10%): Critical user workflows only
- **Integration Tests** (~30%): Component interactions, API calls
- **Unit Tests** (~60%): Business logic, utilities, helpers

**E2E tests should focus on:**
- ✅ Critical user journeys (login → create employee → assign agent)
- ✅ Cross-page workflows
- ✅ Real browser interactions
- ❌ NOT every button click or form validation (use unit tests)

### 2. Test Independence

Each test should:
- ✅ Be runnable in isolation
- ✅ Clean up after itself
- ✅ Not depend on other tests
- ✅ Use unique test data

```typescript
// ✅ GOOD - Independent test
test('should create employee', async ({ page }) => {
  await setAuthCookies(page);
  await page.goto('/dashboard/employees/new');
  await page.getByLabel('Email').fill(`test-${Date.now()}@acme.com`);
  // ...
});

// ❌ BAD - Depends on previous test
test('should edit employee created in previous test', async ({ page }) => {
  // This will fail if run in isolation!
});
```

### 3. Stable Selectors

Prefer in this order:
1. ✅ **Role-based**: `getByRole('button', { name: /submit/i })`
2. ✅ **Label-based**: `getByLabel('Email')`
3. ✅ **Test IDs**: `getByTestId('employee-card')`
4. ❌ **CSS selectors**: `.btn-primary` (fragile!)

```typescript
// ✅ GOOD - Semantic, stable selectors
await page.getByRole('button', { name: /create employee/i }).click();
await page.getByLabel('Email').fill('alice@acme.com');
await page.getByTestId('employee-table').isVisible();

// ❌ BAD - Fragile selectors
await page.locator('.btn-primary').click();
await page.locator('#email-input').fill('alice@acme.com');
await page.locator('div > table').isVisible();
```

### 4. Explicit Waits

Always wait for conditions, don't use arbitrary timeouts:

```typescript
// ✅ GOOD - Wait for specific condition
await expect(page.getByText('Employee created')).toBeVisible();
await page.waitForURL('/dashboard/employees');

// ❌ BAD - Arbitrary timeout
await page.waitForTimeout(2000);
```

### 5. Page Object Pattern (for complex pages)

For frequently tested pages, consider page objects:

```typescript
// tests/pages/EmployeeListPage.ts
export class EmployeeListPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/dashboard/employees');
  }

  async searchEmployees(query: string) {
    await this.page.getByPlaceholder(/search/i).fill(query);
  }

  async getEmployeeCount() {
    return await this.page.locator('[data-testid="employee-row"]').count();
  }
}

// In test file
test('should search employees', async ({ page }) => {
  const employeeList = new EmployeeListPage(page);
  await employeeList.goto();
  await employeeList.searchEmployees('Alice');
  expect(await employeeList.getEmployeeCount()).toBeGreaterThan(0);
});
```

### 6. Accessibility Testing

Include accessibility checks:

```typescript
test('should be keyboard navigable', async ({ page }) => {
  await page.goto('/dashboard/employees');

  // Tab through form
  await page.keyboard.press('Tab');
  await expect(page.getByLabel('Email')).toBeFocused();

  await page.keyboard.press('Tab');
  await expect(page.getByLabel('Password')).toBeFocused();
});

test('should have proper ARIA labels', async ({ page }) => {
  await page.goto('/login');

  const emailInput = page.getByLabel('Email');
  await expect(emailInput).toHaveAttribute('aria-required', 'true');
});
```

---

## Debugging Tests

### Visual Debugging

```bash
# Run in headed mode (see browser)
npm run test:e2e -- --headed

# Run in debug mode (step through)
npm run test:e2e -- --debug

# Run with UI mode (interactive)
npm run test:e2e -- --ui
```

### Trace Viewer

When tests fail, Playwright captures traces:

```bash
# Show trace for failed test
npx playwright show-trace test-results/.../trace.zip
```

### Screenshots & Videos

On failure, Playwright automatically captures:
- Screenshot: `test-results/.../screenshot.png`
- Video: `test-results/.../video.webm`
- Trace: `test-results/.../trace.zip`

### Console Logs

Check browser console:

```typescript
test('should not have console errors', async ({ page }) => {
  const errors: string[] = [];

  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      errors.push(msg.text());
    }
  });

  await page.goto('/dashboard');

  expect(errors).toHaveLength(0);
});
```

---

## CI/CD Integration

### GitHub Actions Workflow

Tests run automatically on:
- ✅ Pull requests to `main`
- ✅ Pushes to `main`

**Features:**
- Parallel execution (2 shards)
- Retry on failure (2 retries)
- HTML report upload
- Screenshot/video upload on failure

See `.github/workflows/web-e2e.yml` for configuration.

### Local CI Simulation

```bash
# Run tests as they would run in CI
CI=true npm run test:e2e
```

---

## Troubleshooting

### Tests timeout

**Cause:** Elements not found, slow page load, network issues

**Solutions:**
1. Increase timeout: `test.setTimeout(60000)`
2. Use better selectors
3. Add explicit waits
4. Check network tab for slow API calls

### Tests flaky (pass/fail randomly)

**Cause:** Race conditions, timing issues, animation delays

**Solutions:**
1. Use `waitForLoadState('networkidle')`
2. Wait for specific conditions, not arbitrary timeouts
3. Disable animations in test mode
4. Use `test.retry(3)` for inherently flaky tests

### Cannot find element

**Cause:** Wrong selector, element not rendered, timing issue

**Solutions:**
1. Use Playwright Inspector: `--debug`
2. Check trace viewer
3. Verify element exists in browser DevTools
4. Add `await page.pause()` to inspect state

### Mock data not working

**Cause:** Route not matching, timing issue, wrong URL pattern

**Solutions:**
1. Log matched routes: `page.on('request', req => console.log(req.url()))`
2. Use `**` for wildcards: `**/api/v1/employees*`
3. Set up mocks before navigation

---

## Further Reading

- [Playwright Documentation](https://playwright.dev)
- [Best Practices Guide](https://playwright.dev/docs/best-practices)
- [API Reference](https://playwright.dev/docs/api/class-test)
- [Testing Library Principles](https://testing-library.com/docs/guiding-principles)

---

**Questions?** Check existing tests in `tests/e2e/` for examples!
