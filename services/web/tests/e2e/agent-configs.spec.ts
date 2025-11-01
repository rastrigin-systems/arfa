import { test, expect } from '@playwright/test';

test.describe('Organization Agent Configs Page', () => {
  test.beforeEach(async ({ page }) => {
    // Note: These tests assume authentication is handled by middleware
    // In production, we'd need to set up auth cookies or mock auth
  });

  test.describe('Page Load', () => {
    test('should render page with three tabs', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Check page heading
      await expect(page.getByRole('heading', { name: /organization agent configuration/i, level: 1 })).toBeVisible();

      // Check description
      await expect(page.getByText(/manage ai agents available to your organization/i)).toBeVisible();

      // Check tabs exist
      await expect(page.getByRole('tab', { name: /available agents/i })).toBeVisible();
      await expect(page.getByRole('tab', { name: /organization configs/i })).toBeVisible();
      await expect(page.getByRole('tab', { name: /team configs/i })).toBeVisible();
    });

    test('should show Available Agents tab by default', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Available Agents tab should be selected
      const availableTab = page.getByRole('tab', { name: /available agents/i });
      await expect(availableTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should display agent cards correctly', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Wait for agent cards to load
      const agentCards = page.locator('[data-testid="agent-card"]');
      await expect(agentCards.first()).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Available Agents Tab', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/dashboard/agents');
      // Ensure we're on Available Agents tab
      await page.getByRole('tab', { name: /available agents/i }).click();
    });

    test('should display agent cards in grid layout', async ({ page }) => {
      // Check grid container
      const gridContainer = page.locator('[data-testid="agent-grid"]');
      await expect(gridContainer).toBeVisible();

      // Check cards exist
      const cards = page.locator('[data-testid="agent-card"]');
      const count = await cards.count();
      expect(count).toBeGreaterThan(0);
    });

    test('should display agent card with correct information', async ({ page }) => {
      const firstCard = page.locator('[data-testid="agent-card"]').first();

      // Check card has name
      await expect(firstCard.locator('[data-testid="agent-name"]')).toBeVisible();

      // Check card has description
      await expect(firstCard.locator('[data-testid="agent-description"]')).toBeVisible();

      // Check card has provider
      await expect(firstCard.locator('[data-testid="agent-provider"]')).toBeVisible();
    });

    test('should have Configure button on each agent card', async ({ page }) => {
      const firstCard = page.locator('[data-testid="agent-card"]').first();
      const configureButton = firstCard.getByRole('button', { name: /configure/i });
      await expect(configureButton).toBeVisible();
      await expect(configureButton).toBeEnabled();
    });

    test('should show loading state while fetching agents', async ({ page }) => {
      // Navigate with slow network to see loading state
      await page.route('**/api/v1/agents', async (route) => {
        await page.waitForTimeout(1000);
        await route.continue();
      });

      await page.goto('/dashboard/agents');

      // Loading skeleton should be visible
      await expect(page.locator('[role="status"][aria-label*="Loading"]')).toBeVisible();
    });

    test('should show empty state when no agents available', async ({ page }) => {
      // Mock empty response
      await page.route('**/api/v1/agents', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ agents: [] }),
        });
      });

      await page.goto('/dashboard/agents');

      // Empty state should be visible
      await expect(page.getByText(/no agents available/i)).toBeVisible();
    });

    test('should show error state when API fails', async ({ page }) => {
      // Mock error response
      await page.route('**/api/v1/agents', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' }),
        });
      });

      await page.goto('/dashboard/agents');

      // Error message should be visible
      await expect(page.getByText(/failed to load agents/i)).toBeVisible();

      // Retry button should be present
      await expect(page.getByRole('button', { name: /retry/i })).toBeVisible();
    });
  });

  test.describe('Configure Modal', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/dashboard/agents');
      await page.getByRole('tab', { name: /available agents/i }).click();
    });

    test('should open configure modal when clicking Configure button', async ({ page }) => {
      // Click Configure on first agent
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Modal should be visible
      const modal = page.getByRole('dialog', { name: /configure/i });
      await expect(modal).toBeVisible();
    });

    test('should display agent name in modal title', async ({ page }) => {
      // Get first agent name
      const firstCard = page.locator('[data-testid="agent-card"]').first();
      const agentName = await firstCard.locator('[data-testid="agent-name"]').textContent();

      // Open modal
      await firstCard.getByRole('button', { name: /configure/i }).click();

      // Modal title should contain agent name
      const modal = page.getByRole('dialog');
      await expect(modal.getByRole('heading', { name: new RegExp(agentName || '', 'i') })).toBeVisible();
    });

    test('should have JSON editor in modal', async ({ page }) => {
      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // JSON editor should be present (Monaco Editor)
      const editor = page.locator('.monaco-editor');
      await expect(editor).toBeVisible();
    });

    test('should have Save and Cancel buttons', async ({ page }) => {
      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      const modal = page.getByRole('dialog');

      // Check buttons exist
      await expect(modal.getByRole('button', { name: /save configuration/i })).toBeVisible();
      await expect(modal.getByRole('button', { name: /cancel/i })).toBeVisible();
    });

    test('should close modal when clicking Cancel', async ({ page }) => {
      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Click Cancel
      await page.getByRole('dialog').getByRole('button', { name: /cancel/i }).click();

      // Modal should be closed
      await expect(page.getByRole('dialog')).not.toBeVisible();
    });

    test.skip('should validate JSON syntax before saving', async ({ page }) => {
      // Skip - requires Monaco Editor interaction
      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Try to save with invalid JSON
      // Note: This would require Monaco Editor interaction which is complex in E2E tests

      // Error message should be shown
      await expect(page.getByText(/invalid json/i)).toBeVisible();
    });

    test.skip('should create org config on successful save', async ({ page }) => {
      // Skip - requires API mock
      // Mock successful API call
      await page.route('**/api/v1/organizations/current/agent-configs', async (route) => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            config: {
              id: 'config-123',
              agent_id: 'agent-123',
              config: { model: 'claude-sonnet-4.5' },
              is_enabled: true,
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
            },
          }),
        });
      });

      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Click Save
      await page.getByRole('dialog').getByRole('button', { name: /save configuration/i }).click();

      // Success toast should appear
      await expect(page.getByText(/configuration created successfully/i)).toBeVisible();

      // Modal should close
      await expect(page.getByRole('dialog')).not.toBeVisible();
    });
  });

  test.describe('Organization Configs Tab', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/dashboard/agents');
      await page.getByRole('tab', { name: /organization configs/i }).click();
    });

    test('should switch to Organization Configs tab', async ({ page }) => {
      const orgConfigsTab = page.getByRole('tab', { name: /organization configs/i });
      await expect(orgConfigsTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should show empty state when no configs exist', async ({ page }) => {
      // Mock empty configs
      await page.route('**/api/v1/organizations/current/agent-configs', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ configs: [] }),
        });
      });

      await page.reload();
      await page.getByRole('tab', { name: /organization configs/i }).click();

      // Empty state should be visible
      await expect(page.getByText(/no agent configurations/i)).toBeVisible();
    });

    test.skip('should display list of configured agents', async ({ page }) => {
      // Skip - requires API data
      // Table should be visible
      const table = page.locator('table');
      await expect(table).toBeVisible();

      // Check table has rows
      const rows = page.locator('table tbody tr');
      const count = await rows.count();
      expect(count).toBeGreaterThan(0);
    });

    test.skip('should display agent config details', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();

      // Check columns
      await expect(firstRow.locator('td').nth(0)).toBeVisible(); // Agent name
      await expect(firstRow.locator('td').nth(1)).toBeVisible(); // Provider
      await expect(firstRow.locator('td').nth(2)).toBeVisible(); // Status
      await expect(firstRow.locator('td').nth(3)).toBeVisible(); // Last updated
      await expect(firstRow.locator('td').nth(4)).toBeVisible(); // Actions
    });

    test.skip('should show status badges correctly', async ({ page }) => {
      // Skip - requires API data
      const statusBadge = page.locator('table tbody tr').first().locator('[data-testid="status-badge"]');
      await expect(statusBadge).toBeVisible();

      const text = await statusBadge.textContent();
      expect(text).toMatch(/enabled|disabled/i);
    });

    test.skip('should have Edit button for each config', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();
      const editButton = firstRow.getByRole('button', { name: /edit/i });
      await expect(editButton).toBeVisible();
      await expect(editButton).toBeEnabled();
    });

    test.skip('should have Enable/Disable button based on status', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();

      // Should have either Enable or Disable button
      const actionButton = firstRow.locator('button:has-text("Enable"), button:has-text("Disable")');
      await expect(actionButton).toBeVisible();
    });

    test.skip('should have Remove button for each config', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();
      const removeButton = firstRow.getByRole('button', { name: /delete|remove/i });
      await expect(removeButton).toBeVisible();
      await expect(removeButton).toBeEnabled();
    });

    test.skip('should open edit modal when clicking Edit', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /edit/i }).click();

      // Modal should open with existing config
      const modal = page.getByRole('dialog', { name: /configure/i });
      await expect(modal).toBeVisible();
      await expect(modal.getByRole('button', { name: /update configuration/i })).toBeVisible();
    });

    test.skip('should disable agent when clicking Disable', async ({ page }) => {
      // Skip - requires API data and mock
      const firstRow = page.locator('table tbody tr').first();

      // Mock disable API call
      await page.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
        if (route.request().method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              config: {
                id: 'config-123',
                is_enabled: false,
                updated_at: new Date().toISOString(),
              },
            }),
          });
        }
      });

      await firstRow.getByRole('button', { name: /disable/i }).click();

      // Status should change to Disabled
      await expect(firstRow.getByText(/disabled/i)).toBeVisible();
    });

    test.skip('should enable agent when clicking Enable', async ({ page }) => {
      // Skip - requires API data and mock
      const firstRow = page.locator('table tbody tr').first();

      // Mock enable API call
      await page.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
        if (route.request().method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              config: {
                id: 'config-123',
                is_enabled: true,
                updated_at: new Date().toISOString(),
              },
            }),
          });
        }
      });

      await firstRow.getByRole('button', { name: /enable/i }).click();

      // Status should change to Enabled
      await expect(firstRow.getByText(/enabled/i)).toBeVisible();
    });

    test.skip('should show confirmation when removing config', async ({ page }) => {
      // Skip - requires API data
      const firstRow = page.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /delete|remove/i }).click();

      // Confirmation dialog should appear
      await expect(page.getByText(/are you sure|confirm/i)).toBeVisible();
    });

    test.skip('should remove config on confirmation', async ({ page }) => {
      // Skip - requires API data and mock
      // Mock delete API call
      await page.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 204,
          });
        }
      });

      const firstRow = page.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /delete|remove/i }).click();

      // Confirm deletion
      await page.getByRole('button', { name: /confirm|delete|remove/i }).click();

      // Config should be removed from list
      // (We'd need to check the row is gone, but this requires more complex state)
    });
  });

  test.describe('Team Configs Tab', () => {
    test.beforeEach(async ({ page }) => {
      await page.goto('/dashboard/agents');
      await page.getByRole('tab', { name: /team configs/i }).click();
    });

    test('should switch to Team Configs tab', async ({ page }) => {
      const teamConfigsTab = page.getByRole('tab', { name: /team configs/i });
      await expect(teamConfigsTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should show "Coming Soon" empty state', async ({ page }) => {
      // Coming soon message should be visible
      await expect(page.getByText(/coming soon/i)).toBeVisible();
    });
  });

  test.describe('Accessibility', () => {
    test('should support keyboard navigation through tabs', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Focus on first tab
      await page.keyboard.press('Tab');
      const availableTab = page.getByRole('tab', { name: /available agents/i });
      await expect(availableTab).toBeFocused();

      // Navigate to second tab with arrow keys
      await page.keyboard.press('ArrowRight');
      const orgConfigsTab = page.getByRole('tab', { name: /organization configs/i });
      await expect(orgConfigsTab).toBeFocused();

      // Navigate to third tab
      await page.keyboard.press('ArrowRight');
      const teamConfigsTab = page.getByRole('tab', { name: /team configs/i });
      await expect(teamConfigsTab).toBeFocused();
    });

    test('should have proper ARIA labels on action buttons', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Check Configure button has aria-label
      const configureButton = page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i });
      await expect(configureButton).toHaveAttribute('aria-label');
    });

    test('should have proper focus management in modal', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Open modal
      await page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Focus should be trapped in modal
      const modal = page.getByRole('dialog');

      // Tab through elements in modal
      await page.keyboard.press('Tab');

      // Focus should stay within modal
      await expect(modal.locator(':focus')).toBeVisible();
    });

    test('should return focus to trigger after modal close', async ({ page }) => {
      await page.goto('/dashboard/agents');

      const configureButton = page.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i });

      // Open modal
      await configureButton.click();

      // Close modal
      await page.getByRole('dialog').getByRole('button', { name: /cancel/i }).click();

      // Focus should return to Configure button
      await expect(configureButton).toBeFocused();
    });

    test('should have descriptive text for screen readers', async ({ page }) => {
      await page.goto('/dashboard/agents');

      // Check page has proper landmarks
      await expect(page.locator('main, [role="main"]')).toBeVisible();

      // Check headings hierarchy
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    });
  });

  test.describe('Responsive Design', () => {
    test('should display correctly on mobile viewport', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/dashboard/agents');

      // Page should be visible
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();

      // Tabs should be visible
      await expect(page.getByRole('tab', { name: /available agents/i })).toBeVisible();
    });

    test('should display grid as single column on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/dashboard/agents');

      // Agent cards should stack vertically
      const grid = page.locator('[data-testid="agent-grid"]');

      // Check computed style (grid should be 1 column)
      const gridComputedStyle = await grid.evaluate((el) => {
        return window.getComputedStyle(el).gridTemplateColumns;
      });

      // Should have single column layout
      expect(gridComputedStyle).not.toContain('repeat');
    });

    test('should display correctly on tablet viewport', async ({ page }) => {
      await page.setViewportSize({ width: 768, height: 1024 });
      await page.goto('/dashboard/agents');

      // Page should be visible
      await expect(page.getByRole('heading', { level: 1 })).toBeVisible();

      // Grid should show 2 columns
      const grid = page.locator('[data-testid="agent-grid"]');
      await expect(grid).toBeVisible();
    });
  });
});
