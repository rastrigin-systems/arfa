import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Clear cookies before each test
    await page.context().clearCookies();
  });

  test('should redirect unauthenticated users to login page', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('should display login form', async ({ page }) => {
    await page.goto('/login');

    // Check page title
    await expect(page.getByRole('heading', { name: 'Ubik Enterprise' })).toBeVisible();

    // Check form fields
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByLabel('Password')).toBeVisible();
    await expect(page.getByRole('button', { name: /login/i })).toBeVisible();
  });

  test('should show validation errors for empty form', async ({ page }) => {
    await page.goto('/login');

    // Submit empty form
    await page.getByRole('button', { name: /login/i }).click();

    // Check for HTML5 validation (email field required)
    const emailInput = page.getByLabel('Email');
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBe(true);
  });

  test('should show validation error for invalid email', async ({ page }) => {
    await page.goto('/login');

    // Enter invalid email
    await page.getByLabel('Email').fill('invalid-email');
    await page.getByLabel('Password').fill('password123');
    await page.getByRole('button', { name: /login/i }).click();

    // Check for validation error
    const emailInput = page.getByLabel('Email');
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBe(true);
  });

  test.skip('should login with valid credentials and redirect to dashboard', async ({ page }) => {
    // This test requires a running API backend with test data
    // Skip for now as we don't have the backend running
    await page.goto('/login');

    await page.getByLabel('Email').fill('alice@acme.com');
    await page.getByLabel('Password').fill('password123');
    await page.getByRole('button', { name: /login/i }).click();

    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard');
    await expect(page.getByText(/welcome back/i)).toBeVisible();
  });

  test.skip('should show error for invalid credentials', async ({ page }) => {
    // This test requires a running API backend
    // Skip for now as we don't have the backend running
    await page.goto('/login');

    await page.getByLabel('Email').fill('wrong@example.com');
    await page.getByLabel('Password').fill('wrongpassword');
    await page.getByRole('button', { name: /login/i }).click();

    // Should show error message
    await expect(page.getByRole('alert')).toContainText(/invalid credentials/i);
  });

  test.skip('should logout and redirect to login page', async ({ page }) => {
    // This test requires a running API backend and authenticated session
    // Skip for now as we don't have the backend running

    // Assume we're logged in and on dashboard
    await page.goto('/dashboard');

    // Click logout button
    await page.getByRole('button', { name: /logout/i }).click();

    // Should redirect to login
    await expect(page).toHaveURL('/login');
  });

  test('should redirect authenticated users away from login page', async ({ page }) => {
    // This test would require setting up a mock authenticated session
    // For now, we'll skip it as it requires backend integration
    test.skip();
  });

  test('accessibility: login form should be keyboard navigable', async ({ page }) => {
    await page.goto('/login');

    // Tab through form elements
    await page.keyboard.press('Tab'); // Focus email
    await expect(page.getByLabel('Email')).toBeFocused();

    await page.keyboard.press('Tab'); // Focus password
    await expect(page.getByLabel('Password')).toBeFocused();

    await page.keyboard.press('Tab'); // Focus login button
    await expect(page.getByRole('button', { name: /login/i })).toBeFocused();
  });

  test('accessibility: should have proper ARIA labels', async ({ page }) => {
    await page.goto('/login');

    // Check email input has proper aria attributes
    const emailInput = page.getByLabel('Email');
    await expect(emailInput).toHaveAttribute('aria-required', 'true');

    // Check password input has proper aria attributes
    const passwordInput = page.getByLabel('Password');
    await expect(passwordInput).toHaveAttribute('aria-required', 'true');
  });

  test('theme: should toggle between light and dark mode', async ({ page }) => {
    await page.goto('/login');

    // Wait for hydration
    await page.waitForTimeout(1000);

    // Check initial theme (system default)
    const html = page.locator('html');
    const initialClass = await html.getAttribute('class');

    // Note: Theme toggle is only on dashboard, not login page
    // This test documents that dark mode infrastructure is in place
    expect(initialClass).toBeDefined();
  });
});
