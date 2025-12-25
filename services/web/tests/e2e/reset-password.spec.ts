import { test, expect } from './fixtures';

const VALID_TOKEN = 'valid-reset-token-123';
const EXPIRED_TOKEN = 'expired-reset-token-456';
const USED_TOKEN = 'used-reset-token-789';
const INVALID_TOKEN = 'invalid-token-000';

test.describe('Reset Password Flow', () => {
  test('should validate token on page load - valid token', async ({ page }) => {
    // Mock token verification API - valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          valid: true,
        }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Should show reset password form
    await expect(page.getByRole('heading', { name: 'Arfa' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Reset Password' })).toBeVisible();
    await expect(page.getByLabel('New Password')).toBeVisible();
    await expect(page.getByLabel('Confirm Password')).toBeVisible();
    await expect(page.getByRole('button', { name: /reset password/i })).toBeVisible();
  });

  test('should show error for expired token', async ({ page }) => {
    // Mock token verification API - expired token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          valid: false,
          message: 'Token has expired',
        }),
      });
    });

    await page.goto(`/reset-password/${EXPIRED_TOKEN}`);

    // Should show expired error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/this reset link has expired/i)).toBeVisible();
    await expect(page.getByText(/password reset links are valid for 1 hour/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /request new reset link/i })).toBeVisible();

    // Form should not be visible
    await expect(page.getByLabel('New Password')).not.toBeVisible();
  });

  test('should show error for already-used token', async ({ page }) => {
    // Mock token verification API - used token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          valid: false,
          message: 'Token has already been used',
        }),
      });
    });

    await page.goto(`/reset-password/${USED_TOKEN}`);

    // Should show used error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/this reset link has already been used/i)).toBeVisible();
    await expect(page.getByText(/password reset links can only be used once/i)).toBeVisible();
  });

  test('should show error for invalid token', async ({ page }) => {
    // Mock token verification API - invalid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          valid: false,
          message: 'Token not found',
        }),
      });
    });

    await page.goto(`/reset-password/${INVALID_TOKEN}`);

    // Should show invalid error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/this reset link has expired or is invalid/i)).toBeVisible();
  });

  test('should display password strength indicator', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Type weak password
    await page.getByLabel('New Password').fill('weak');

    // Should show strength indicator
    await expect(page.getByText(/password strength/i)).toBeVisible();
    await expect(page.getByText(/weak/i)).toBeVisible();

    // Should show requirements checklist
    await expect(page.getByText(/at least 8 characters/i)).toBeVisible();
    await expect(page.getByText(/one uppercase letter/i)).toBeVisible();
    await expect(page.getByText(/one number/i)).toBeVisible();
    await expect(page.getByText(/one special character/i)).toBeVisible();
  });

  test('should validate password confirmation match in real-time', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Type password
    await page.getByLabel('New Password').fill('StrongPass123!');

    // Type mismatched confirmation
    await page.getByLabel('Confirm Password').fill('StrongPass123');

    // Should show mismatch error
    await expect(page.getByText(/passwords do not match/i)).toBeVisible();

    // Confirm password field should have error styling
    const confirmInput = page.getByLabel('Confirm Password');
    await expect(confirmInput).toHaveClass(/border-destructive/);
  });

  test('should successfully reset password and redirect to login', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    // Mock reset password API
    await page.route('**/api/v1/auth/reset-password', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Password reset successful',
        }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Fill in new password
    await page.getByLabel('New Password').fill('NewStrongPass123!');
    await page.getByLabel('Confirm Password').fill('NewStrongPass123!');

    // Submit form
    await page.getByRole('button', { name: /reset password/i }).click();

    // Should show loading state
    await expect(page.getByRole('button', { name: /resetting/i })).toBeVisible();

    // Should show success message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/password reset successful/i)).toBeVisible();
    await expect(page.getByText(/redirecting to login/i)).toBeVisible();

    // Should redirect to login after 2 seconds
    await page.waitForTimeout(2100);
    await expect(page).toHaveURL(/\/login/);
  });

  test('should reject weak password', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    // Mock reset password API - weak password error
    await page.route('**/api/v1/auth/reset-password', (route) => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Password does not meet requirements',
        }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Try to submit with weak password
    await page.getByLabel('New Password').fill('weak');
    await page.getByLabel('Confirm Password').fill('weak');
    await page.getByRole('button', { name: /reset password/i }).click();

    // Should show error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/an error occurred/i)).toBeVisible();
  });

  test('should handle server errors gracefully', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    // Mock server error
    await page.route('**/api/v1/auth/reset-password', (route) => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          message: 'Internal server error',
        }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Fill form
    await page.getByLabel('New Password').fill('StrongPass123!');
    await page.getByLabel('Confirm Password').fill('StrongPass123!');
    await page.getByRole('button', { name: /reset password/i }).click();

    // Should show error message
    await expect(page.getByRole('alert')).toBeVisible();
    await expect(page.getByText(/an error occurred/i)).toBeVisible();
  });

  test('should link to forgot-password page on token error', async ({ page }) => {
    // Mock expired token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({ valid: false, message: 'Token expired' }),
      });
    });

    await page.goto(`/reset-password/${EXPIRED_TOKEN}`);

    // Click "Request new reset link" button
    await page.getByRole('button', { name: /request new reset link/i }).click();

    // Should navigate to forgot-password page
    await expect(page).toHaveURL(/\/forgot-password/);
  });

  test('accessibility: form should be keyboard navigable', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Wait for form to load
    await page.waitForSelector('[name="new_password"]');

    // Tab through form elements
    await page.keyboard.press('Tab'); // Focus new password
    await expect(page.getByLabel('New Password')).toBeFocused();

    await page.keyboard.press('Tab'); // Focus confirm password
    await expect(page.getByLabel('Confirm Password')).toBeFocused();

    await page.keyboard.press('Tab'); // Focus submit button
    await expect(page.getByRole('button', { name: /reset password/i })).toBeFocused();
  });

  test('accessibility: should have proper ARIA labels', async ({ page }) => {
    // Mock valid token
    await page.route('**/api/v1/auth/verify-reset-token*', (route) => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ valid: true }),
      });
    });

    await page.goto(`/reset-password/${VALID_TOKEN}`);

    // Check password inputs have proper aria attributes
    const newPasswordInput = page.getByLabel('New Password');
    await expect(newPasswordInput).toHaveAttribute('aria-required', 'true');
    await expect(newPasswordInput).toHaveAttribute('aria-label', 'New password');

    const confirmPasswordInput = page.getByLabel('Confirm Password');
    await expect(confirmPasswordInput).toHaveAttribute('aria-required', 'true');
    await expect(confirmPasswordInput).toHaveAttribute('aria-label', 'Confirm new password');
  });
});
