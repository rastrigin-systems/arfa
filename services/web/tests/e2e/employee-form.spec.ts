import { test, expect } from './fixtures';

test.describe('Employee Form - Create', () => {
  test.beforeEach(async ({ mockEmployees }) => {
    // Navigate to create employee page
    await mockEmployees.goto('/employees/new');
  });

  test('should display create employee form', async ({ mockEmployees }) => {
    // Check page heading
    await expect(mockEmployees.getByRole('heading', { name: /create employee|new employee/i, level: 1 })).toBeVisible();

    // Check form fields exist
    await expect(mockEmployees.locator('input[name="email"]')).toBeVisible();
    await expect(mockEmployees.locator('input[name="full_name"]')).toBeVisible();
    await expect(mockEmployees.locator('select[name="role_id"]')).toBeVisible();

    // Team is optional
    const teamField = mockEmployees.locator('select[name="team_id"]');
    if (await teamField.count() > 0) {
      await expect(teamField).toBeVisible();
    }

    // Check submit and cancel buttons
    await expect(mockEmployees.getByRole('button', { name: /create|save|submit/i })).toBeVisible();
    await expect(mockEmployees.getByRole('button', { name: /cancel/i })).toBeVisible();
  });

  test('should show validation errors for empty required fields', async ({ mockEmployees }) => {
    // Try to submit without filling fields
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Check for validation messages
    await expect(mockEmployees.getByText(/email is required|email.*required/i)).toBeVisible();
    await expect(mockEmployees.getByText(/name is required|name.*required/i)).toBeVisible();
    await expect(mockEmployees.getByText(/role is required|role.*required/i)).toBeVisible();
  });

  test('should show validation error for invalid email format', async ({ mockEmployees }) => {
    // Fill with invalid email
    await mockEmployees.locator('input[name="email"]').fill('invalid-email');
    await mockEmployees.locator('input[name="email"]').blur();

    // Check for email validation error
    await expect(mockEmployees.getByText(/invalid email|email.*valid/i)).toBeVisible();
  });

  test('should show validation error for short name', async ({ mockEmployees }) => {
    // Fill with too short name
    await mockEmployees.locator('input[name="full_name"]').fill('A');
    await mockEmployees.locator('input[name="full_name"]').blur();

    // Check for name validation error (min 2 chars based on common validation)
    await expect(mockEmployees.getByText(/name.*least.*2|name too short/i)).toBeVisible();
  });

  test.skip('should create employee with valid data', async ({ mockEmployees }) => {
    // Skip - requires backend API integration
    // Fill form with valid data
    await mockEmployees.locator('input[name="email"]').fill('newuser@example.com');
    await mockEmployees.locator('input[name="full_name"]').fill('New Test User');

    // Select role (first available option)
    await mockEmployees.locator('select[name="role_id"]').selectOption({ index: 1 });

    // Submit form
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show success message
    await expect(mockEmployees.getByText(/employee created|created successfully/i)).toBeVisible({ timeout: 5000 });

    // Should redirect to employee detail or list page
    await expect(mockEmployees).toHaveURL(/\/employees\/[a-f0-9-]+|\/employees$/);
  });

  test.skip('should show temporary password after creation', async ({ mockEmployees }) => {
    // Skip - requires backend API integration
    await mockEmployees.locator('input[name="email"]').fill('newuser2@example.com');
    await mockEmployees.locator('input[name="full_name"]').fill('Another User');
    await mockEmployees.locator('select[name="role_id"]').selectOption({ index: 1 });
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show temporary password in success message or modal
    await expect(mockEmployees.getByText(/temporary password|password:/i)).toBeVisible({ timeout: 5000 });

    // Should have a way to copy the password
    const copyButton = mockEmployees.getByRole('button', { name: /copy.*password/i });
    if (await copyButton.count() > 0) {
      await expect(copyButton).toBeVisible();
    }
  });

  test('should cancel and return to employees list', async ({ mockEmployees }) => {
    // Click cancel button
    await mockEmployees.getByRole('button', { name: /cancel/i }).click();

    // Should navigate back to employees list
    await expect(mockEmployees).toHaveURL(/\/employees$/);
  });

  test.skip('should handle duplicate email error', async ({ mockEmployees }) => {
    // Skip - requires backend API with existing employee
    await mockEmployees.locator('input[name="email"]').fill('existing@example.com');
    await mockEmployees.locator('input[name="full_name"]').fill('Test User');
    await mockEmployees.locator('select[name="role_id"]').selectOption({ index: 1 });
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show duplicate email error
    await expect(mockEmployees.getByText(/email.*already.*exists|duplicate.*email/i)).toBeVisible({ timeout: 5000 });
  });

  test('accessibility: form should be keyboard navigable', async ({ mockEmployees }) => {
    // Tab through form fields
    await mockEmployees.keyboard.press('Tab');
    const emailField = mockEmployees.locator('input[name="email"]');
    await expect(emailField).toBeFocused();

    await mockEmployees.keyboard.press('Tab');
    const nameField = mockEmployees.locator('input[name="full_name"]');
    await expect(nameField).toBeFocused();
  });

  test('accessibility: form fields should have proper labels', async ({ mockEmployees }) => {
    // Email field should have label
    const emailField = mockEmployees.locator('input[name="email"]');
    const emailLabelId = await emailField.getAttribute('aria-labelledby');
    const emailLabel = await emailField.getAttribute('aria-label');
    expect(emailLabelId || emailLabel).toBeTruthy();

    // Name field should have label
    const nameField = mockEmployees.locator('input[name="full_name"]');
    const nameLabelId = await nameField.getAttribute('aria-labelledby');
    const nameLabel = await nameField.getAttribute('aria-label');
    expect(nameLabelId || nameLabel).toBeTruthy();
  });

  test('accessibility: validation errors should be announced', async ({ mockEmployees }) => {
    // Submit form with errors
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Error messages should have role="alert" or aria-live
    const errorMessages = mockEmployees.locator('[role="alert"]');
    await expect(errorMessages.first()).toBeVisible();
  });
});

test.describe('Employee Form - Edit', () => {
  const testEmployeeId = '00000000-0000-0000-0000-000000000001'; // Mock ID

  test.beforeEach(async ({ mockEmployees }) => {
    // Navigate to edit employee page
    await mockEmployees.goto(`/employees/${testEmployeeId}/edit`);
  });

  test('should display edit employee form', async ({ mockEmployees }) => {
    // Check page heading
    await expect(mockEmployees.getByRole('heading', { name: /edit employee|update employee/i, level: 1 })).toBeVisible();

    // Check form fields exist
    await expect(mockEmployees.locator('input[name="email"]')).toBeVisible();
    await expect(mockEmployees.locator('input[name="full_name"]')).toBeVisible();
    await expect(mockEmployees.locator('select[name="role_id"]')).toBeVisible();

    // Check submit and cancel buttons
    await expect(mockEmployees.getByRole('button', { name: /update|save/i })).toBeVisible();
    await expect(mockEmployees.getByRole('button', { name: /cancel/i })).toBeVisible();
  });

  test.skip('should load existing employee data', async ({ mockEmployees }) => {
    // Skip - requires backend API
    // Wait for form to load
    await mockEmployees.waitForLoadState('networkidle');

    // Email field should be populated (and possibly disabled)
    const emailField = mockEmployees.locator('input[name="email"]');
    const emailValue = await emailField.inputValue();
    expect(emailValue).toBeTruthy();
    expect(emailValue).toContain('@');

    // Name field should be populated
    const nameField = mockEmployees.locator('input[name="full_name"]');
    const nameValue = await nameField.inputValue();
    expect(nameValue).toBeTruthy();
    expect(nameValue.length).toBeGreaterThan(0);
  });

  test.skip('should update employee with modified data', async ({ mockEmployees }) => {
    // Skip - requires backend API
    // Wait for form to load
    await mockEmployees.waitForLoadState('networkidle');

    // Modify name
    const nameField = mockEmployees.locator('input[name="full_name"]');
    await nameField.clear();
    await nameField.fill('Updated Name');

    // Submit form
    await mockEmployees.getByRole('button', { name: /update|save/i }).click();

    // Should show success message
    await expect(mockEmployees.getByText(/employee updated|updated successfully/i)).toBeVisible({ timeout: 5000 });

    // Should redirect back to employee detail
    await expect(mockEmployees).toHaveURL(/\/employees\/[a-f0-9-]+$/);
  });

  test('should cancel and return to employee detail', async ({ mockEmployees }) => {
    // Click cancel button
    await mockEmployees.getByRole('button', { name: /cancel/i }).click();

    // Should navigate back to employee detail
    await expect(mockEmployees).toHaveURL(/\/employees\/[a-f0-9-]+$/);
  });

  test.skip('should show 404 for non-existent employee', async ({ mockEmployees }) => {
    // Skip - requires backend API
    const nonExistentId = '99999999-9999-9999-9999-999999999999';
    await mockEmployees.goto(`/employees/${nonExistentId}/edit`);

    // Should show 404 or not found message
    await expect(mockEmployees.getByText(/not found|employee.*not.*exist/i)).toBeVisible({ timeout: 5000 });
  });

  test.skip('should disable email field on edit', async ({ mockEmployees }) => {
    // Skip - requires backend API
    // Email should typically be read-only on edit
    await mockEmployees.waitForLoadState('networkidle');

    const emailField = mockEmployees.locator('input[name="email"]');
    const isDisabled = await emailField.isDisabled();
    const isReadOnly = await emailField.getAttribute('readonly');

    // Email should be either disabled or readonly
    expect(isDisabled || isReadOnly).toBeTruthy();
  });

  test('accessibility: edit form should be keyboard navigable', async ({ mockEmployees }) => {
    // Tab through form fields
    await mockEmployees.keyboard.press('Tab');

    // Should be able to tab to focusable elements
    const focusedElement = mockEmployees.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });

  test('should show validation errors on update', async ({ mockEmployees }) => {
    // Clear required field
    const nameField = mockEmployees.locator('input[name="full_name"]');
    await nameField.clear();
    await nameField.blur();

    // Should show validation error
    await expect(mockEmployees.getByText(/name is required|name.*required/i)).toBeVisible();
  });
});

test.describe('Employee Form - Loading States', () => {
  test.skip('should show loading indicator while submitting', async ({ mockEmployees }) => {
    // Skip - requires backend API with slow response
    await mockEmployees.goto('/employees/new');

    await mockEmployees.locator('input[name="email"]').fill('slow@example.com');
    await mockEmployees.locator('input[name="full_name"]').fill('Slow Test');
    await mockEmployees.locator('select[name="role_id"]').selectOption({ index: 1 });

    // Submit form
    await mockEmployees.getByRole('button', { name: /create|save|submit/i }).click();

    // Should show loading state (button disabled or spinner)
    const submitButton = mockEmployees.getByRole('button', { name: /creating|saving|submitting/i });
    await expect(submitButton).toBeVisible();

    // OR button should be disabled
    const originalButton = mockEmployees.getByRole('button', { name: /create|save|submit/i });
    await expect(originalButton).toBeDisabled();
  });
});

test.describe('Employee Form - Responsive Design', () => {
  test('should display form on mobile viewport', async ({ mockEmployees }) => {
    // Set mobile viewport
    await mockEmployees.setViewportSize({ width: 375, height: 667 });
    await mockEmployees.goto('/employees/new');

    // Form should be visible and usable
    await expect(mockEmployees.locator('input[name="email"]')).toBeVisible();
    await expect(mockEmployees.locator('input[name="full_name"]')).toBeVisible();
    await expect(mockEmployees.getByRole('button', { name: /create|save|submit/i })).toBeVisible();
  });

  test('should display form on tablet viewport', async ({ mockEmployees }) => {
    // Set tablet viewport
    await mockEmployees.setViewportSize({ width: 768, height: 1024 });
    await mockEmployees.goto('/employees/new');

    // Form should be visible and well-formatted
    await expect(mockEmployees.locator('input[name="email"]')).toBeVisible();
    await expect(mockEmployees.locator('input[name="full_name"]')).toBeVisible();
  });
});
