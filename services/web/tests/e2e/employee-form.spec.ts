import { test, expect } from '@playwright/test';

test.describe('Employee Form - Create', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to create employee page
    await page.goto('/employees/new');
  });

  test('should display create employee form', async ({ page }) => {
    // Check page heading
    await expect(page.getByRole('heading', { name: /create employee|new employee/i, level: 1 })).toBeVisible();

    // Check form fields exist
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="full_name"]')).toBeVisible();
    await expect(page.locator('select[name="role_id"]')).toBeVisible();

    // Team is optional
    const teamField = page.locator('select[name="team_id"]');
    if (await teamField.count() > 0) {
      await expect(teamField).toBeVisible();
    }

    // Check submit and cancel buttons
    await expect(page.getByRole('button', { name: /create|save|submit/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();
  });

  test('should show validation errors for empty required fields', async ({ page }) => {
    // Try to submit without filling fields
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Check for validation messages
    await expect(page.getByText(/email is required|email.*required/i)).toBeVisible();
    await expect(page.getByText(/name is required|name.*required/i)).toBeVisible();
    await expect(page.getByText(/role is required|role.*required/i)).toBeVisible();
  });

  test('should show validation error for invalid email format', async ({ page }) => {
    // Fill with invalid email
    await page.locator('input[name="email"]').fill('invalid-email');
    await page.locator('input[name="email"]').blur();

    // Check for email validation error
    await expect(page.getByText(/invalid email|email.*valid/i)).toBeVisible();
  });

  test('should show validation error for short name', async ({ page }) => {
    // Fill with too short name
    await page.locator('input[name="full_name"]').fill('A');
    await page.locator('input[name="full_name"]').blur();

    // Check for name validation error (min 2 chars based on common validation)
    await expect(page.getByText(/name.*least.*2|name too short/i)).toBeVisible();
  });

  test.skip('should create employee with valid data', async ({ page }) => {
    // Skip - requires backend API integration
    // Fill form with valid data
    await page.locator('input[name="email"]').fill('newuser@example.com');
    await page.locator('input[name="full_name"]').fill('New Test User');

    // Select role (first available option)
    await page.locator('select[name="role_id"]').selectOption({ index: 1 });

    // Submit form
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show success message
    await expect(page.getByText(/employee created|created successfully/i)).toBeVisible({ timeout: 5000 });

    // Should redirect to employee detail or list page
    await expect(page).toHaveURL(/\/employees\/[a-f0-9-]+|\/employees$/);
  });

  test.skip('should show temporary password after creation', async ({ page }) => {
    // Skip - requires backend API integration
    await page.locator('input[name="email"]').fill('newuser2@example.com');
    await page.locator('input[name="full_name"]').fill('Another User');
    await page.locator('select[name="role_id"]').selectOption({ index: 1 });
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show temporary password in success message or modal
    await expect(page.getByText(/temporary password|password:/i)).toBeVisible({ timeout: 5000 });

    // Should have a way to copy the password
    const copyButton = page.getByRole('button', { name: /copy.*password/i });
    if (await copyButton.count() > 0) {
      await expect(copyButton).toBeVisible();
    }
  });

  test('should cancel and return to employees list', async ({ page }) => {
    // Click cancel button
    await page.getByRole('button', { name: /cancel/i }).click();

    // Should navigate back to employees list
    await expect(page).toHaveURL(/\/employees$/);
  });

  test.skip('should handle duplicate email error', async ({ page }) => {
    // Skip - requires backend API with existing employee
    await page.locator('input[name="email"]').fill('existing@example.com');
    await page.locator('input[name="full_name"]').fill('Test User');
    await page.locator('select[name="role_id"]').selectOption({ index: 1 });
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show duplicate email error
    await expect(page.getByText(/email.*already.*exists|duplicate.*email/i)).toBeVisible({ timeout: 5000 });
  });

  test('accessibility: form should be keyboard navigable', async ({ page }) => {
    // Tab through form fields
    await page.keyboard.press('Tab');
    const emailField = page.locator('input[name="email"]');
    await expect(emailField).toBeFocused();

    await page.keyboard.press('Tab');
    const nameField = page.locator('input[name="full_name"]');
    await expect(nameField).toBeFocused();
  });

  test('accessibility: form fields should have proper labels', async ({ page }) => {
    // Email field should have label
    const emailField = page.locator('input[name="email"]');
    const emailLabelId = await emailField.getAttribute('aria-labelledby');
    const emailLabel = await emailField.getAttribute('aria-label');
    expect(emailLabelId || emailLabel).toBeTruthy();

    // Name field should have label
    const nameField = page.locator('input[name="full_name"]');
    const nameLabelId = await nameField.getAttribute('aria-labelledby');
    const nameLabel = await nameField.getAttribute('aria-label');
    expect(nameLabelId || nameLabel).toBeTruthy();
  });

  test('accessibility: validation errors should be announced', async ({ page }) => {
    // Submit form with errors
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Error messages should have role="alert" or aria-live
    const errorMessages = page.locator('[role="alert"]');
    await expect(errorMessages.first()).toBeVisible();
  });
});

test.describe('Employee Form - Edit', () => {
  const testEmployeeId = '00000000-0000-0000-0000-000000000001'; // Mock ID

  test.beforeEach(async ({ page }) => {
    // Navigate to edit employee page
    await page.goto(`/employees/${testEmployeeId}/edit`);
  });

  test('should display edit employee form', async ({ page }) => {
    // Check page heading
    await expect(page.getByRole('heading', { name: /edit employee|update employee/i, level: 1 })).toBeVisible();

    // Check form fields exist
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="full_name"]')).toBeVisible();
    await expect(page.locator('select[name="role_id"]')).toBeVisible();

    // Check submit and cancel buttons
    await expect(page.getByRole('button', { name: /update|save/i })).toBeVisible();
    await expect(page.getByRole('button', { name: /cancel/i })).toBeVisible();
  });

  test.skip('should load existing employee data', async ({ page }) => {
    // Skip - requires backend API
    // Wait for form to load
    await page.waitForLoadState('networkidle');

    // Email field should be populated (and possibly disabled)
    const emailField = page.locator('input[name="email"]');
    const emailValue = await emailField.inputValue();
    expect(emailValue).toBeTruthy();
    expect(emailValue).toContain('@');

    // Name field should be populated
    const nameField = page.locator('input[name="full_name"]');
    const nameValue = await nameField.inputValue();
    expect(nameValue).toBeTruthy();
    expect(nameValue.length).toBeGreaterThan(0);
  });

  test.skip('should update employee with modified data', async ({ page }) => {
    // Skip - requires backend API
    // Wait for form to load
    await page.waitForLoadState('networkidle');

    // Modify name
    const nameField = page.locator('input[name="full_name"]');
    await nameField.clear();
    await nameField.fill('Updated Name');

    // Submit form
    await page.getByRole('button', { name: /update|save/i }).click();

    // Should show success message
    await expect(page.getByText(/employee updated|updated successfully/i)).toBeVisible({ timeout: 5000 });

    // Should redirect back to employee detail
    await expect(page).toHaveURL(/\/employees\/[a-f0-9-]+$/);
  });

  test('should cancel and return to employee detail', async ({ page }) => {
    // Click cancel button
    await page.getByRole('button', { name: /cancel/i }).click();

    // Should navigate back to employee detail
    await expect(page).toHaveURL(/\/employees\/[a-f0-9-]+$/);
  });

  test.skip('should show 404 for non-existent employee', async ({ page }) => {
    // Skip - requires backend API
    const nonExistentId = '99999999-9999-9999-9999-999999999999';
    await page.goto(`/employees/${nonExistentId}/edit`);

    // Should show 404 or not found message
    await expect(page.getByText(/not found|employee.*not.*exist/i)).toBeVisible({ timeout: 5000 });
  });

  test.skip('should disable email field on edit', async ({ page }) => {
    // Skip - requires backend API
    // Email should typically be read-only on edit
    await page.waitForLoadState('networkidle');

    const emailField = page.locator('input[name="email"]');
    const isDisabled = await emailField.isDisabled();
    const isReadOnly = await emailField.getAttribute('readonly');

    // Email should be either disabled or readonly
    expect(isDisabled || isReadOnly).toBeTruthy();
  });

  test('accessibility: edit form should be keyboard navigable', async ({ page }) => {
    // Tab through form fields
    await page.keyboard.press('Tab');

    // Should be able to tab to focusable elements
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });

  test('should show validation errors on update', async ({ page }) => {
    // Clear required field
    const nameField = page.locator('input[name="full_name"]');
    await nameField.clear();
    await nameField.blur();

    // Should show validation error
    await expect(page.getByText(/name is required|name.*required/i)).toBeVisible();
  });
});

test.describe('Employee Form - Loading States', () => {
  test.skip('should show loading indicator while submitting', async ({ page }) => {
    // Skip - requires backend API with slow response
    await page.goto('/employees/new');

    await page.locator('input[name="email"]').fill('slow@example.com');
    await page.locator('input[name="full_name"]').fill('Slow Test');
    await page.locator('select[name="role_id"]').selectOption({ index: 1 });

    // Submit form
    await page.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show loading state (button disabled or spinner)
    const submitButton = page.getByRole('button', { name: /creating|saving|submitting/i });
    await expect(submitButton).toBeVisible();

    // OR button should be disabled
    const originalButton = page.getByRole('button', { name: /create|save|submit/i });
    await expect(originalButton).toBeDisabled();
  });
});

test.describe('Employee Form - Responsive Design', () => {
  test('should display form on mobile viewport', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/employees/new');

    // Form should be visible and usable
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="full_name"]')).toBeVisible();
    await expect(page.getByRole('button', { name: /create|save|submit/i })).toBeVisible();
  });

  test('should display form on tablet viewport', async ({ page }) => {
    // Set tablet viewport
    await page.setViewportSize({ width: 768, height: 1024 });
    await page.goto('/employees/new');

    // Form should be visible and well-formatted
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="full_name"]')).toBeVisible();
  });
});
