import { test, expect } from './fixtures';

test.describe('Employee Detail Page', () => {
  test.beforeEach(async ({ mockEmployees }) => {
    // Note: These tests assume a mock/test backend is running
    // In real environment, we'd set up test data via API
  });

  test.skip('should display employee detail page with all information', async ({ mockEmployees }) => {
    // Navigate to a specific employee detail page
    // Using a mock employee ID that should exist in test data
    await mockEmployees.goto('/employees/employee-id-123');

    // Check page title/heading
    await expect(mockEmployees.getByRole('heading', { name: /employee details/i })).toBeVisible();

    // Check employee basic information is displayed
    await expect(mockEmployees.getByText(/name/i)).toBeVisible();
    await expect(mockEmployees.getByText(/email/i)).toBeVisible();
    await expect(mockEmployees.getByText(/status/i)).toBeVisible();
    await expect(mockEmployees.getByText(/role/i)).toBeVisible();
  });

  test.skip('should display employee teams', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Check teams section
    await expect(mockEmployees.getByRole('heading', { name: /teams/i })).toBeVisible();

    // Should show team cards or "No teams" message
    const teamsSection = mockEmployees.locator('[data-testid="employee-teams"]');
    await expect(teamsSection).toBeVisible();
  });

  test.skip('should display employee MCP configurations', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Check MCPs section
    await expect(mockEmployees.getByRole('heading', { name: /mcp servers/i })).toBeVisible();

    // Should show MCP cards or "No MCPs" message
    const mcpsSection = mockEmployees.locator('[data-testid="employee-mcps"]');
    await expect(mcpsSection).toBeVisible();
  });

  test.skip('should have edit button that navigates to edit form', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Find and click edit button
    const editButton = mockEmployees.getByRole('button', { name: /edit/i });
    await expect(editButton).toBeVisible();
    await editButton.click();

    // Should navigate to edit page
    await expect(mockEmployees).toHaveURL(/\/employees\/employee-id-123\/edit/);
  });

  test.skip('should have delete button with confirmation modal', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Find and click delete button
    const deleteButton = mockEmployees.getByRole('button', { name: /delete/i });
    await expect(deleteButton).toBeVisible();
    await deleteButton.click();

    // Should show confirmation modal
    await expect(mockEmployees.getByRole('dialog')).toBeVisible();
    await expect(mockEmployees.getByText(/are you sure/i)).toBeVisible();

    // Check cancel and confirm buttons exist
    await expect(mockEmployees.getByRole('button', { name: /cancel/i })).toBeVisible();
    await expect(mockEmployees.getByRole('button', { name: /confirm|delete/i })).toBeVisible();
  });

  test.skip('should cancel delete operation when clicking cancel', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Click delete
    await mockEmployees.getByRole('button', { name: /delete/i }).click();

    // Click cancel in modal
    await mockEmployees.getByRole('button', { name: /cancel/i }).click();

    // Modal should close, still on employee detail page
    await expect(mockEmployees.getByRole('dialog')).not.toBeVisible();
    await expect(mockEmployees).toHaveURL(/\/employees\/employee-id-123/);
  });

  test.skip('should show loading state while fetching employee data', async ({ mockEmployees }) => {
    // Intercept API call and delay response
    await mockEmployees.route('**/api/v1/employees/employee-id-123', async (route) => {
      await new Promise((resolve) => setTimeout(resolve, 1000));
      await route.continue();
    });

    await mockEmployees.goto('/employees/employee-id-123');

    // Should show loading skeleton or spinner
    await expect(mockEmployees.getByRole('status')).toBeVisible();
  });

  test.skip('should show error message when employee not found', async ({ mockEmployees }) => {
    // Intercept API call and return 404
    await mockEmployees.route('**/api/v1/employees/non-existent-id', async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Employee not found' }),
      });
    });

    await mockEmployees.goto('/employees/non-existent-id');

    // Should show error message
    await expect(mockEmployees.getByRole('alert')).toContainText(/not found/i);
  });

  test.skip('should show error message when API fails', async ({ mockEmployees }) => {
    // Intercept API call and return 500
    await mockEmployees.route('**/api/v1/employees/employee-id-123', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' }),
      });
    });

    await mockEmployees.goto('/employees/employee-id-123');

    // Should show error message
    await expect(mockEmployees.getByRole('alert')).toContainText(/error|failed/i);
  });

  test('accessibility: employee detail page should be keyboard navigable', async ({ mockEmployees }) => {
    test.skip(); // Requires backend
    await mockEmployees.goto('/employees/employee-id-123');

    // Tab through interactive elements
    await mockEmployees.keyboard.press('Tab');

    // Edit button should be focusable
    await expect(mockEmployees.getByRole('button', { name: /edit/i })).toBeFocused();

    await mockEmployees.keyboard.press('Tab');

    // Delete button should be focusable
    await expect(mockEmployees.getByRole('button', { name: /delete/i })).toBeFocused();
  });

  test('accessibility: should have proper ARIA labels and roles', async ({ mockEmployees }) => {
    test.skip(); // Requires backend
    await mockEmployees.goto('/employees/employee-id-123');

    // Check main content has proper ARIA landmark
    const main = mockEmployees.locator('main');
    await expect(main).toBeVisible();

    // Check sections have proper headings
    const headings = mockEmployees.locator('h1, h2, h3');
    expect(await headings.count()).toBeGreaterThan(0);
  });

  test('accessibility: delete confirmation modal should trap focus', async ({ mockEmployees }) => {
    test.skip(); // Requires backend
    await mockEmployees.goto('/employees/employee-id-123');

    // Open delete modal
    await mockEmployees.getByRole('button', { name: /delete/i }).click();

    // Press Tab multiple times
    await mockEmployees.keyboard.press('Tab');
    await mockEmployees.keyboard.press('Tab');
    await mockEmployees.keyboard.press('Tab');

    // Focus should stay within modal
    const focusedElement = mockEmployees.locator(':focus');
    const modal = mockEmployees.getByRole('dialog');

    // Check focused element is inside modal
    expect(await modal.locator(':focus').count()).toBeGreaterThan(0);
  });

  test('accessibility: should support Escape key to close modal', async ({ mockEmployees }) => {
    test.skip(); // Requires backend
    await mockEmployees.goto('/employees/employee-id-123');

    // Open delete modal
    await mockEmployees.getByRole('button', { name: /delete/i }).click();
    await expect(mockEmployees.getByRole('dialog')).toBeVisible();

    // Press Escape
    await mockEmployees.keyboard.press('Escape');

    // Modal should close
    await expect(mockEmployees.getByRole('dialog')).not.toBeVisible();
  });

  test.skip('should display employee status badge with correct color', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Should show status badge
    const statusBadge = mockEmployees.locator('[data-testid="employee-status-badge"]');
    await expect(statusBadge).toBeVisible();

    // Badge should have appropriate styling based on status
    // Active = green, Inactive = gray, etc.
    const badgeClass = await statusBadge.getAttribute('class');
    expect(badgeClass).toBeTruthy();
  });

  test.skip('should display back button to return to employees list', async ({ mockEmployees }) => {
    await mockEmployees.goto('/employees/employee-id-123');

    // Find back button or breadcrumb
    const backButton = mockEmployees.getByRole('link', { name: /back|employees/i });
    await expect(backButton).toBeVisible();

    // Clicking should navigate back
    await backButton.click();
    await expect(mockEmployees).toHaveURL(/\/employees$/);
  });
});
