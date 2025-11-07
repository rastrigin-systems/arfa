import { test, expect } from '@playwright/test';

test.describe('Signup Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/signup');
  });

  test('should display signup form with all fields', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Ubik Enterprise' })).toBeVisible();
    await expect(page.getByText('Create Your Organization Account')).toBeVisible();

    // Check all form fields are present
    await expect(page.getByLabel('Full Name')).toBeVisible();
    await expect(page.getByLabel(/^Email/)).toBeVisible();
    await expect(page.getByLabel('Organization Name')).toBeVisible();
    await expect(page.getByLabel('Organization Slug')).toBeVisible();
    await expect(page.getByLabel(/^Password$/)).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();

    // Check submit button
    await expect(page.getByRole('button', { name: 'Create Account' })).toBeVisible();

    // Check link to login
    await expect(page.getByRole('link', { name: 'Sign in' })).toBeVisible();
  });

  test('should show password strength indicator when typing password', async ({ page }) => {
    const passwordInput = page.getByLabel(/^Password$/);

    // Type a weak password
    await passwordInput.fill('weak');
    await expect(page.getByText('Weak')).toBeVisible();

    // Type a medium password
    await passwordInput.fill('Password1');
    await expect(page.getByText('Medium')).toBeVisible();

    // Type a strong password
    await passwordInput.fill('Password123!');
    await expect(page.getByText('Strong')).toBeVisible();

    // Check requirements checklist is visible
    await expect(page.getByText('At least 8 characters')).toBeVisible();
    await expect(page.getByText('One uppercase letter')).toBeVisible();
    await expect(page.getByText('One number')).toBeVisible();
    await expect(page.getByText('One special character')).toBeVisible();
  });

  test('should convert org slug to lowercase', async ({ page }) => {
    const slugInput = page.getByLabel('Organization Slug');

    await slugInput.fill('MyOrganization');

    // Should be converted to lowercase
    await expect(slugInput).toHaveValue('myorganization');
  });

  test('should be responsive on mobile viewport', async ({ page }) => {
    // Set to mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await expect(page.getByRole('heading', { name: 'Ubik Enterprise' })).toBeVisible();

    // All form fields should be visible
    await expect(page.getByLabel('Full Name')).toBeVisible();
    await expect(page.getByLabel(/^Email/)).toBeVisible();
    await expect(page.getByLabel('Organization Name')).toBeVisible();
    await expect(page.getByLabel('Organization Slug')).toBeVisible();
    await expect(page.getByLabel(/^Password$/)).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create Account' })).toBeVisible();
  });

  test('should be responsive on tablet viewport', async ({ page }) => {
    // Set to tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });

    await expect(page.getByRole('heading', { name: 'Ubik Enterprise' })).toBeVisible();

    // All form fields should be visible
    await expect(page.getByLabel('Full Name')).toBeVisible();
    await expect(page.getByLabel(/^Email/)).toBeVisible();
    await expect(page.getByLabel('Organization Name')).toBeVisible();
    await expect(page.getByLabel('Organization Slug')).toBeVisible();
    await expect(page.getByLabel(/^Password$/)).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create Account' })).toBeVisible();
  });

  test('should be responsive on desktop viewport', async ({ page }) => {
    // Set to desktop viewport
    await page.setViewportSize({ width: 1920, height: 1080 });

    await expect(page.getByRole('heading', { name: 'Ubik Enterprise' })).toBeVisible();

    // All form fields should be visible
    await expect(page.getByLabel('Full Name')).toBeVisible();
    await expect(page.getByLabel(/^Email/)).toBeVisible();
    await expect(page.getByLabel('Organization Name')).toBeVisible();
    await expect(page.getByLabel('Organization Slug')).toBeVisible();
    await expect(page.getByLabel(/^Password$/)).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Create Account' })).toBeVisible();
  });

  test('should have proper keyboard navigation', async ({ page }) => {
    // Tab through all form fields
    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Full Name')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByLabel(/^Email/)).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Organization Name')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Organization Slug')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByLabel(/^Password$/)).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByLabel('Confirm Password')).toBeFocused();

    await page.keyboard.press('Tab');
    await expect(page.getByRole('button', { name: 'Create Account' })).toBeFocused();
  });
});
