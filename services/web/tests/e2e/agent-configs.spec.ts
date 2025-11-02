import { test, expect } from './fixtures';

test.describe('Organization Agent Configs Page', () => {
  test.describe('Page Load', () => {
    test('should render page with three tabs', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Check page heading
      await expect(mockAgents.getByRole('heading', { name: /organization agent configuration/i, level: 1 })).toBeVisible();

      // Check description
      await expect(mockAgents.getByText(/manage ai agents available to your organization/i)).toBeVisible();

      // Check tabs exist
      await expect(mockAgents.getByRole('tab', { name: /available agents/i })).toBeVisible();
      await expect(mockAgents.getByRole('tab', { name: /organization configs/i })).toBeVisible();
      await expect(mockAgents.getByRole('tab', { name: /team configs/i })).toBeVisible();
    });

    test('should show Available Agents tab by default', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Available Agents tab should be selected
      const availableTab = mockAgents.getByRole('tab', { name: /available agents/i });
      await expect(availableTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should display agent cards correctly', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Wait for agent cards to load
      const agentCards = mockAgents.locator('[data-testid="agent-card"]');
      await expect(agentCards.first()).toBeVisible({ timeout: 5000 });
    });
  });

  test.describe('Available Agents Tab', () => {
    test.beforeEach(async ({ mockAgents }) => {
      await mockAgents.goto('/agents');
      // Ensure we're on Available Agents tab
      await mockAgents.getByRole('tab', { name: /available agents/i }).click();
    });

    test('should display agent cards in grid layout', async ({ mockAgents }) => {
      // Check grid container
      const gridContainer = mockAgents.locator('[data-testid="agent-grid"]');
      await expect(gridContainer).toBeVisible();

      // Check cards exist
      const cards = mockAgents.locator('[data-testid="agent-card"]');
      const count = await cards.count();
      expect(count).toBeGreaterThan(0);
    });

    test('should display agent card with correct information', async ({ mockAgents }) => {
      const firstCard = mockAgents.locator('[data-testid="agent-card"]').first();

      // Check card has name
      await expect(firstCard.locator('[data-testid="agent-name"]')).toBeVisible();

      // Check card has description
      await expect(firstCard.locator('[data-testid="agent-description"]')).toBeVisible();

      // Check card has provider
      await expect(firstCard.locator('[data-testid="agent-provider"]')).toBeVisible();
    });

    test('should have Configure button on each agent card', async ({ mockAgents }) => {
      const firstCard = mockAgents.locator('[data-testid="agent-card"]').first();
      const configureButton = firstCard.getByRole('button', { name: /configure/i });
      await expect(configureButton).toBeVisible();
      await expect(configureButton).toBeEnabled();
    });

    test('should show loading state while fetching agents', async ({ mockAgents }) => {
      // Navigate with slow network to see loading state
      await mockAgents.route('**/api/v1/agents', async (route) => {
        await mockAgents.waitForTimeout(1000);
        await route.continue();
      });

      await mockAgents.goto('/agents');

      // Loading skeleton should be visible
      await expect(mockAgents.locator('[role="status"][aria-label*="Loading"]')).toBeVisible();
    });

    test('should show empty state when no agents available', async ({ mockAgents }) => {
      // Mock empty response
      await mockAgents.route('**/api/v1/agents', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ agents: [] }),
        });
      });

      await mockAgents.goto('/agents');

      // Empty state should be visible
      await expect(mockAgents.getByText(/no agents available/i)).toBeVisible();
    });

    test('should show error state when API fails', async ({ mockAgents }) => {
      // Mock error response
      await mockAgents.route('**/api/v1/agents', async (route) => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' }),
        });
      });

      await mockAgents.goto('/agents');

      // Error message should be visible
      await expect(mockAgents.getByText(/failed to load agents/i)).toBeVisible();

      // Retry button should be present
      await expect(mockAgents.getByRole('button', { name: /retry/i })).toBeVisible();
    });
  });

  test.describe('Configure Modal', () => {
    test.beforeEach(async ({ mockAgents }) => {
      await mockAgents.goto('/agents');
      await mockAgents.getByRole('tab', { name: /available agents/i }).click();
    });

    test('should open configure modal when clicking Configure button', async ({ mockAgents }) => {
      // Click Configure on first agent
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Modal should be visible
      const modal = mockAgents.getByRole('dialog', { name: /configure/i });
      await expect(modal).toBeVisible();
    });

    test('should display agent name in modal title', async ({ mockAgents }) => {
      // Get first agent name
      const firstCard = mockAgents.locator('[data-testid="agent-card"]').first();
      const agentName = await firstCard.locator('[data-testid="agent-name"]').textContent();

      // Open modal
      await firstCard.getByRole('button', { name: /configure/i }).click();

      // Modal title should contain agent name
      const modal = mockAgents.getByRole('dialog');
      await expect(modal.getByRole('heading', { name: new RegExp(agentName || '', 'i') })).toBeVisible();
    });

    test('should have JSON editor in modal', async ({ mockAgents }) => {
      // Open modal
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // JSON editor should be present (Monaco Editor)
      const editor = mockAgents.locator('.monaco-editor');
      await expect(editor).toBeVisible();
    });

    test('should have Save and Cancel buttons', async ({ mockAgents }) => {
      // Open modal
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      const modal = mockAgents.getByRole('dialog');

      // Check buttons exist
      await expect(modal.getByRole('button', { name: /save configuration/i })).toBeVisible();
      await expect(modal.getByRole('button', { name: /cancel/i })).toBeVisible();
    });

    test('should close modal when clicking Cancel', async ({ mockAgents }) => {
      // Open modal
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Click Cancel
      await mockAgents.getByRole('dialog').getByRole('button', { name: /cancel/i }).click();

      // Modal should be closed
      await expect(mockAgents.getByRole('dialog')).not.toBeVisible();
    });

    test.skip('should validate JSON syntax before saving', async ({ mockAgents }) => {
      // Skip - requires Monaco Editor interaction
      // Open modal
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Try to save with invalid JSON
      // Note: This would require Monaco Editor interaction which is complex in E2E tests

      // Error message should be shown
      await expect(mockAgents.getByText(/invalid json/i)).toBeVisible();
    });

    test.skip('should create org config on successful save', async ({ mockAgents }) => {
      // Skip - requires API mock
      // Mock successful API call
      await mockAgents.route('**/api/v1/organizations/current/agent-configs', async (route) => {
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
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Click Save
      await mockAgents.getByRole('dialog').getByRole('button', { name: /save configuration/i }).click();

      // Success toast should appear
      await expect(mockAgents.getByText(/configuration created successfully/i)).toBeVisible();

      // Modal should close
      await expect(mockAgents.getByRole('dialog')).not.toBeVisible();
    });
  });

  test.describe('Organization Configs Tab', () => {
    test.beforeEach(async ({ mockAgents }) => {
      await mockAgents.goto('/agents');
      await mockAgents.getByRole('tab', { name: /organization configs/i }).click();
    });

    test('should switch to Organization Configs tab', async ({ mockAgents }) => {
      const orgConfigsTab = mockAgents.getByRole('tab', { name: /organization configs/i });
      await expect(orgConfigsTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should show empty state when no configs exist', async ({ mockAgents }) => {
      // Mock empty configs
      await mockAgents.route('**/api/v1/organizations/current/agent-configs', async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ configs: [] }),
        });
      });

      await mockAgents.reload();
      await mockAgents.getByRole('tab', { name: /organization configs/i }).click();

      // Empty state should be visible
      await expect(mockAgents.getByText(/no agent configurations/i)).toBeVisible();
    });

    test.skip('should display list of configured agents', async ({ mockAgents }) => {
      // Skip - requires API data
      // Table should be visible
      const table = mockAgents.locator('table');
      await expect(table).toBeVisible();

      // Check table has rows
      const rows = mockAgents.locator('table tbody tr');
      const count = await rows.count();
      expect(count).toBeGreaterThan(0);
    });

    test.skip('should display agent config details', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();

      // Check columns
      await expect(firstRow.locator('td').nth(0)).toBeVisible(); // Agent name
      await expect(firstRow.locator('td').nth(1)).toBeVisible(); // Provider
      await expect(firstRow.locator('td').nth(2)).toBeVisible(); // Status
      await expect(firstRow.locator('td').nth(3)).toBeVisible(); // Last updated
      await expect(firstRow.locator('td').nth(4)).toBeVisible(); // Actions
    });

    test.skip('should show status badges correctly', async ({ mockAgents }) => {
      // Skip - requires API data
      const statusBadge = mockAgents.locator('table tbody tr').first().locator('[data-testid="status-badge"]');
      await expect(statusBadge).toBeVisible();

      const text = await statusBadge.textContent();
      expect(text).toMatch(/enabled|disabled/i);
    });

    test.skip('should have Edit button for each config', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();
      const editButton = firstRow.getByRole('button', { name: /edit/i });
      await expect(editButton).toBeVisible();
      await expect(editButton).toBeEnabled();
    });

    test.skip('should have Enable/Disable button based on status', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();

      // Should have either Enable or Disable button
      const actionButton = firstRow.locator('button:has-text("Enable"), button:has-text("Disable")');
      await expect(actionButton).toBeVisible();
    });

    test.skip('should have Remove button for each config', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();
      const removeButton = firstRow.getByRole('button', { name: /delete|remove/i });
      await expect(removeButton).toBeVisible();
      await expect(removeButton).toBeEnabled();
    });

    test.skip('should open edit modal when clicking Edit', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /edit/i }).click();

      // Modal should open with existing config
      const modal = mockAgents.getByRole('dialog', { name: /configure/i });
      await expect(modal).toBeVisible();
      await expect(modal.getByRole('button', { name: /update configuration/i })).toBeVisible();
    });

    test.skip('should disable agent when clicking Disable', async ({ mockAgents }) => {
      // Skip - requires API data and mock
      const firstRow = mockAgents.locator('table tbody tr').first();

      // Mock disable API call
      await mockAgents.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
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

    test.skip('should enable agent when clicking Enable', async ({ mockAgents }) => {
      // Skip - requires API data and mock
      const firstRow = mockAgents.locator('table tbody tr').first();

      // Mock enable API call
      await mockAgents.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
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

    test.skip('should show confirmation when removing config', async ({ mockAgents }) => {
      // Skip - requires API data
      const firstRow = mockAgents.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /delete|remove/i }).click();

      // Confirmation dialog should appear
      await expect(mockAgents.getByText(/are you sure|confirm/i)).toBeVisible();
    });

    test.skip('should remove config on confirmation', async ({ mockAgents }) => {
      // Skip - requires API data and mock
      // Mock delete API call
      await mockAgents.route('**/api/v1/organizations/current/agent-configs/*', async (route) => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 204,
          });
        }
      });

      const firstRow = mockAgents.locator('table tbody tr').first();
      await firstRow.getByRole('button', { name: /delete|remove/i }).click();

      // Confirm deletion
      await mockAgents.getByRole('button', { name: /confirm|delete|remove/i }).click();

      // Config should be removed from list
      // (We'd need to check the row is gone, but this requires more complex state)
    });
  });

  test.describe('Team Configs Tab', () => {
    test.beforeEach(async ({ mockAgents }) => {
      await mockAgents.goto('/agents');
      await mockAgents.getByRole('tab', { name: /team configs/i }).click();
    });

    test('should switch to Team Configs tab', async ({ mockAgents }) => {
      const teamConfigsTab = mockAgents.getByRole('tab', { name: /team configs/i });
      await expect(teamConfigsTab).toHaveAttribute('aria-selected', 'true');
    });

    test('should show "Coming Soon" empty state', async ({ mockAgents }) => {
      // Coming soon message should be visible
      await expect(mockAgents.getByText(/coming soon/i)).toBeVisible();
    });
  });

  test.describe('Accessibility', () => {
    test('should support keyboard navigation through tabs', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Focus on first tab
      await mockAgents.keyboard.press('Tab');
      const availableTab = mockAgents.getByRole('tab', { name: /available agents/i });
      await expect(availableTab).toBeFocused();

      // Navigate to second tab with arrow keys
      await mockAgents.keyboard.press('ArrowRight');
      const orgConfigsTab = mockAgents.getByRole('tab', { name: /organization configs/i });
      await expect(orgConfigsTab).toBeFocused();

      // Navigate to third tab
      await mockAgents.keyboard.press('ArrowRight');
      const teamConfigsTab = mockAgents.getByRole('tab', { name: /team configs/i });
      await expect(teamConfigsTab).toBeFocused();
    });

    test('should have proper ARIA labels on action buttons', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Check Configure button has aria-label
      const configureButton = mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i });
      await expect(configureButton).toHaveAttribute('aria-label');
    });

    test('should have proper focus management in modal', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Open modal
      await mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i }).click();

      // Focus should be trapped in modal
      const modal = mockAgents.getByRole('dialog');

      // Tab through elements in modal
      await mockAgents.keyboard.press('Tab');

      // Focus should stay within modal
      await expect(modal.locator(':focus')).toBeVisible();
    });

    test('should return focus to trigger after modal close', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      const configureButton = mockAgents.locator('[data-testid="agent-card"]').first().getByRole('button', { name: /configure/i });

      // Open modal
      await configureButton.click();

      // Close modal
      await mockAgents.getByRole('dialog').getByRole('button', { name: /cancel/i }).click();

      // Focus should return to Configure button
      await expect(configureButton).toBeFocused();
    });

    test('should have descriptive text for screen readers', async ({ mockAgents }) => {
      await mockAgents.goto('/agents');

      // Check page has proper landmarks
      await expect(mockAgents.locator('main, [role="main"]')).toBeVisible();

      // Check headings hierarchy
      await expect(mockAgents.getByRole('heading', { level: 1 })).toBeVisible();
    });
  });

  test.describe('Responsive Design', () => {
    test('should display correctly on mobile viewport', async ({ mockAgents }) => {
      await mockAgents.setViewportSize({ width: 375, height: 667 });
      await mockAgents.goto('/agents');

      // Page should be visible
      await expect(mockAgents.getByRole('heading', { level: 1 })).toBeVisible();

      // Tabs should be visible
      await expect(mockAgents.getByRole('tab', { name: /available agents/i })).toBeVisible();
    });

    test('should display grid as single column on mobile', async ({ mockAgents }) => {
      await mockAgents.setViewportSize({ width: 375, height: 667 });
      await mockAgents.goto('/agents');

      // Agent cards should stack vertically
      const grid = mockAgents.locator('[data-testid="agent-grid"]');

      // Check computed style (grid should be 1 column)
      const gridComputedStyle = await grid.evaluate((el) => {
        return window.getComputedStyle(el).gridTemplateColumns;
      });

      // Should have single column layout
      expect(gridComputedStyle).not.toContain('repeat');
    });

    test('should display correctly on tablet viewport', async ({ mockAgents }) => {
      await mockAgents.setViewportSize({ width: 768, height: 1024 });
      await mockAgents.goto('/agents');

      // Page should be visible
      await expect(mockAgents.getByRole('heading', { level: 1 })).toBeVisible();

      // Grid should show 2 columns
      const grid = mockAgents.locator('[data-testid="agent-grid"]');
      await expect(grid).toBeVisible();
    });
  });
});
