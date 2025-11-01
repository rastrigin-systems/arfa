/**
 * Authentication Test Helpers
 *
 * Utilities for managing authentication state in E2E tests
 */

import { Page } from '@playwright/test';
import { mockAuthTokens, mockAuthUser } from '../fixtures';

/**
 * Set authentication cookies to simulate logged-in state
 *
 * @param page - Playwright Page object
 * @param userOverrides - Optional overrides for user data
 */
export async function setAuthCookies(
  page: Page,
  userOverrides?: Partial<typeof mockAuthUser>
) {
  const user = { ...mockAuthUser, ...userOverrides };

  // Set authentication cookie (adjust based on your auth implementation)
  await page.context().addCookies([
    {
      name: 'ubik_session',
      value: mockAuthTokens.access_token,
      domain: 'localhost',
      path: '/',
      httpOnly: true,
      secure: false,
      sameSite: 'Lax',
    },
  ]);

  // Set user data in localStorage (if applicable)
  await page.evaluate((userData) => {
    localStorage.setItem('ubik_user', JSON.stringify(userData));
  }, user);
}

/**
 * Clear authentication state
 *
 * @param page - Playwright Page object
 */
export async function clearAuthCookies(page: Page) {
  await page.context().clearCookies();
  await page.evaluate(() => {
    localStorage.removeItem('ubik_user');
    sessionStorage.clear();
  });
}

/**
 * Login via UI form
 *
 * @param page - Playwright Page object
 * @param email - User email
 * @param password - User password
 */
export async function loginViaUI(
  page: Page,
  email: string = 'alice@acme.com',
  password: string = 'password123'
) {
  await page.goto('/login');
  await page.getByLabel('Email').fill(email);
  await page.getByLabel('Password').fill(password);
  await page.getByRole('button', { name: /login/i }).click();

  // Wait for redirect to dashboard
  await page.waitForURL('/dashboard', { timeout: 10000 });
}

/**
 * Check if user is authenticated
 *
 * @param page - Playwright Page object
 * @returns True if authenticated, false otherwise
 */
export async function isAuthenticated(page: Page): Promise<boolean> {
  const cookies = await page.context().cookies();
  return cookies.some((cookie) => cookie.name === 'ubik_session');
}
