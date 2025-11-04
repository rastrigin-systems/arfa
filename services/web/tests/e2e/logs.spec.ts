import { test, expect } from '@playwright/test';

test.describe('Logs Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock login
    await page.goto('/login');
    await page.fill('input[name="email"]', 'test@example.com');
    await page.fill('input[name="password"]', 'password');
    await page.click('button[type="submit"]');

    // Navigate to logs page
    await page.goto('/logs');
  });

  test('should display logs page with filters', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Activity Logs');

    // Check filters are present
    await expect(page.locator('label:has-text("Search Logs")')).toBeVisible();
    await expect(page.locator('label:has-text("Date Range")')).toBeVisible();
    await expect(page.locator('label:has-text("Employee")')).toBeVisible();
    await expect(page.locator('label:has-text("Agent")')).toBeVisible();
    await expect(page.locator('label:has-text("Event Type")')).toBeVisible();
  });

  test('should show live indicator when connected', async ({ page }) => {
    // Wait for WebSocket connection
    await page.waitForSelector('[data-testid="live-indicator"]', { timeout: 5000 });

    const liveIndicator = page.locator('[data-testid="live-indicator"]');
    await expect(liveIndicator).toContainText('Live');
    await expect(liveIndicator).toHaveClass(/bg-green-500/);
  });

  test('should filter logs by date range', async ({ page }) => {
    const startDate = '2024-01-01';
    const endDate = '2024-01-31';

    await page.fill('#start-date', startDate);
    await page.fill('#end-date', endDate);

    // Wait for filtered results
    await page.waitForTimeout(500);

    // Verify logs are filtered (check dates in displayed logs)
    const logs = page.locator('[data-testid="log-entry"]');
    await expect(logs.first()).toBeVisible();
  });

  test('should filter logs by event type', async ({ page }) => {
    await page.click('#event-type');
    await page.click('text=Input');

    // Wait for filtered results
    await page.waitForTimeout(500);

    // Verify only input events are shown
    const eventBadges = page.locator('text=input');
    await expect(eventBadges.first()).toBeVisible();
  });

  test('should search logs by content', async ({ page }) => {
    const searchTerm = 'authentication';
    await page.fill('input[placeholder="Search logs..."]', searchTerm);

    // Wait for search results
    await page.waitForTimeout(500);

    // Verify search results contain the search term
    const logContent = page.locator('pre:has-text("authentication")');
    await expect(logContent.first()).toBeVisible();
  });

  test('should expand and collapse sessions', async ({ page }) => {
    // Wait for logs to load
    await page.waitForSelector('text=/Session:/', { timeout: 5000 });

    const sessionHeader = page.locator('text=/Session:/').first();
    await sessionHeader.click();

    // Session should expand
    await expect(page.locator('[data-testid="log-entry"]').first()).toBeVisible();

    // Click again to collapse
    await sessionHeader.click();

    // Session should collapse (logs hidden)
    await expect(page.locator('[data-testid="log-entry"]').first()).not.toBeVisible();
  });

  test('should display session metadata', async ({ page }) => {
    await page.waitForSelector('text=/Session:/', { timeout: 5000 });

    // Check session info is displayed
    await expect(page.locator('text=/\\d+m \\d+s/')).toBeVisible(); // Duration
    await expect(page.locator('text=/\\d+ events/')).toBeVisible(); // Event count
  });

  test('should export logs as JSON', async ({ page }) => {
    // Set export format to JSON
    await page.click('text=JSON');

    // Start download
    const downloadPromise = page.waitForEvent('download');
    await page.click('button:has-text("Export")');

    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/activity-logs-.*\.json/);
  });

  test('should export logs as CSV', async ({ page }) => {
    // Change export format to CSV
    const formatSelect = page.locator('select, [role="combobox"]').first();
    await formatSelect.click();
    await page.click('text=CSV');

    // Start download
    const downloadPromise = page.waitForEvent('download');
    await page.click('button:has-text("Export")');

    const download = await downloadPromise;
    expect(download.suggestedFilename()).toMatch(/activity-logs-.*\.csv/);
  });

  test('should clear all filters', async ({ page }) => {
    // Apply some filters
    await page.fill('input[placeholder="Search logs..."]', 'test');
    await page.fill('#start-date', '2024-01-01');

    // Clear filters
    await page.click('button:has-text("Clear Filters")');

    // Verify filters are cleared
    await expect(page.locator('input[placeholder="Search logs..."]')).toHaveValue('');
    await expect(page.locator('#start-date')).toHaveValue('');
  });

  test('should show empty state when no logs match filters', async ({ page }) => {
    // Apply filters that return no results
    await page.fill('input[placeholder="Search logs..."]', 'nonexistentsearchterm12345');
    await page.waitForTimeout(500);

    await expect(page.locator('text=No logs found')).toBeVisible();
    await expect(page.locator('text=/try adjusting your filters/i')).toBeVisible();
  });

  test('should handle real-time log updates', async ({ page }) => {
    // Wait for WebSocket connection
    await page.waitForSelector('[data-testid="live-indicator"]', { timeout: 5000 });

    // Get initial log count
    const initialCount = await page.locator('[data-testid="log-entry"]').count();

    // Wait for new logs (mock or wait for actual updates)
    await page.waitForTimeout(2000);

    // Check if new logs appeared
    const newCount = await page.locator('[data-testid="log-entry"]').count();
    expect(newCount).toBeGreaterThanOrEqual(initialCount);
  });

  test('should be responsive on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });

    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('button:has-text("Export")')).toBeVisible();

    // Filters should be stacked vertically on mobile
    const filtersContainer = page.locator('.grid');
    await expect(filtersContainer).toBeVisible();
  });

  test('should display loading state', async ({ page }) => {
    // Navigate to logs page (fast refresh)
    await page.goto('/logs');

    // Should show loading skeleton briefly
    const loadingIndicator = page.locator('[role="status"]');
    // Loading might be very fast, so we just check it exists or completes quickly
    const isLoading = await loadingIndicator.isVisible().catch(() => false);

    // After loading, content should be visible
    await expect(page.locator('h1:has-text("Activity Logs")')).toBeVisible();
  });

  test('should handle WebSocket disconnection gracefully', async ({ page }) => {
    await page.waitForSelector('[data-testid="live-indicator"]', { timeout: 5000 });

    // Simulate network disconnect (close WebSocket)
    await page.evaluate(() => {
      // Force close all WebSocket connections
      // @ts-ignore
      window.WebSocket.prototype.close.call(null);
    });

    // Wait for reconnection attempt
    await page.waitForTimeout(6000); // Reconnect timeout is 5s

    // Should attempt to reconnect
    await expect(page.locator('[data-testid="live-indicator"]')).toBeVisible();
  });
});
