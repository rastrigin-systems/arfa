/**
 * Playwright Test Fixtures
 *
 * This file extends Playwright's base test with custom fixtures
 * for API mocking and authenticated sessions.
 *
 * Usage:
 *   import { test, expect } from './fixtures';
 *
 *   test('my test', async ({ mockApi }) => {
 *     // API routes are automatically mocked
 *     await mockApi.goto('/dashboard');
 *   });
 */

import { test as base, expect } from '@playwright/test';
import type { Page } from '@playwright/test';
import {
  setupApiMocks,
  setupAuthMocks,
  setupEmployeeMocks,
  mockUserSession,
} from './apiMocks';
import { mockEmployees } from './mockData';

/**
 * Extended fixtures for E2E tests
 */
type Fixtures = {
  /**
   * Page with full API mocking enabled
   * All API routes are intercepted and return mock data
   */
  mockApi: Page;

  /**
   * Page with authentication mocking only
   * Only /auth/* routes are mocked
   */
  mockAuth: Page;

  /**
   * Page with employee-related mocking
   * Includes /auth/*, /employees/*, /teams/*, /roles/*
   */
  mockEmployees: Page;

  /**
   * Page with authenticated session
   * Sets auth cookie and mocks /auth/me
   */
  authenticatedPage: Page;
};

/**
 * Extend Playwright test with custom fixtures
 */
export const test = base.extend<Fixtures>({
  /**
   * Full API mocking - all routes intercepted
   */
  mockApi: async ({ page }, use) => {
    await setupApiMocks(page);
    await use(page);
  },

  /**
   * Auth-only mocking
   */
  mockAuth: async ({ page }, use) => {
    await setupAuthMocks(page);
    await use(page);
  },

  /**
   * Employee feature mocking (with authentication)
   */
  mockEmployees: async ({ page }, use) => {
    await setupEmployeeMocks(page);
    // Add authenticated session so tests don't redirect to login
    await mockUserSession(page, mockEmployees[0]);
    await use(page);
  },

  /**
   * Authenticated session with full API mocking
   */
  authenticatedPage: async ({ page }, use) => {
    // Set up full API mocking
    await setupApiMocks(page);

    // Create authenticated session (defaults to Alice)
    await mockUserSession(page, mockEmployees[0]);

    await use(page);
  },
});

/**
 * Re-export expect for convenience
 */
export { expect };

/**
 * Export mock data and helpers for use in tests
 */
export * from './mockData';
export * from './apiMocks';
