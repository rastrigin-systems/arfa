import { test, expect } from './fixtures';

test.describe('Forgot Password Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/forgot-password');
  });

  test('should display forgot password form', async ({ page }) => {
    // Check page titles
    await expect(page.getByRole('heading', { name: 'Arfa' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Forgot Password' })).toBeVisible();

    // Check description text
    await expect(
      page.getByText("Enter your email address and we'll send you a link to reset your password")
    ).toBeVisible();

    // Check form elements
    await expect(page.getByLabel('Email')).toBeVisible();
    await expect(page.getByRole('button', { name: /send reset link/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /back to login/i })).toBeVisible();
  });

  test('should show validation error for empty email', async ({ page }) => {
    // Submit empty form
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Check for HTML5 validation (email field required)
    const emailInput = page.getByLabel('Email');
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBe(true);
  });

  test('should show validation error for invalid email format', async ({ page }) => {
    // Enter invalid email
    await page.getByLabel('Email').fill('invalid-email');
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Check for HTML5 validation
    const emailInput = page.getByLabel('Email');
    const isInvalid = await emailInput.evaluate((el: HTMLInputElement) => !el.validity.valid);
    expect(isInvalid).toBe(true);
  });

  test('should submit email and show generic success message', async ({ page }) => {
    // Mock the forgot-password API endpoint
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Password reset link sent to your email',
        }),
      });
    });

    // Fill in email and submit
    await page.getByLabel('Email').fill('user@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Wait for button text to change to "Sending..."
    await expect(page.getByRole('button', { name: /sending/i })).toBeVisible();

    // Check success message appears
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/password reset link sent to your email/i)).toBeVisible();
    await expect(
      page.getByText(/if an account exists with this email, you will receive/i)
    ).toBeVisible();
    await expect(page.getByText(/please check your spam folder/i)).toBeVisible();

    // Form should be disabled after success
    const emailInput = page.getByLabel('Email');
    await expect(emailInput).toBeDisabled();
  });

  test('should show same success message for non-existent email (security)', async ({ page }) => {
    // Mock API to return success even for non-existent email (security - no enumeration)
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Password reset link sent to your email',
        }),
      });
    });

    // Submit with non-existent email
    await page.getByLabel('Email').fill('nonexistent@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Should still show success message (prevents email enumeration)
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/password reset link sent to your email/i)).toBeVisible();
  });

  test('should handle network errors gracefully', async ({ page }) => {
    // Mock API to return network error
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Internal server error',
        }),
      });
    });

    // Fill in email and submit
    await page.getByLabel('Email').fill('user@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Should show error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/an error occurred/i)).toBeVisible();
  });

  test('should navigate back to login page', async ({ page }) => {
    // Click "Back to login" link
    await page.getByRole('link', { name: /back to login/i }).click();

    // Should navigate to login page
    await expect(page).toHaveURL(/\/login/);
  });

  test('accessibility: form should be keyboard navigable', async ({ page }) => {
    // Tab through form elements
    await page.keyboard.press('Tab'); // Focus email
    await expect(page.getByLabel('Email')).toBeFocused();

    await page.keyboard.press('Tab'); // Focus submit button
    await expect(page.getByRole('button', { name: /send reset link/i })).toBeFocused();

    await page.keyboard.press('Tab'); // Focus back to login link
    await expect(page.getByRole('link', { name: /back to login/i })).toBeFocused();
  });

  test('accessibility: should have proper ARIA labels', async ({ page }) => {
    // Check email input has proper aria attributes
    const emailInput = page.getByLabel('Email');
    await expect(emailInput).toHaveAttribute('aria-required', 'true');
    await expect(emailInput).toHaveAttribute('aria-label', 'Email address');
  });

  test('accessibility: success message should announce to screen readers', async ({ page }) => {
    // Mock the API
    await page.route('**/api/v1/auth/forgot-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Password reset link sent to your email',
        }),
      });
    });

    // Submit form
    await page.getByLabel('Email').fill('user@example.com');
    await page.getByRole('button', { name: /send reset link/i }).click();

    // Check success alert has proper ARIA attributes
    const alert = page.getByRole('alert');
    await expect(alert).toHaveAttribute('aria-live', 'polite');
  });
});
