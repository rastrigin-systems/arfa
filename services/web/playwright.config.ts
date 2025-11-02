import { defineConfig, devices } from '@playwright/test';

/**
 * See https://playwright.dev/docs/test-configuration.
 */
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  // Reduced from 2 to 1 - with mocked APIs, tests should be more reliable
  retries: process.env.CI ? 1 : 0,
  // Increased from 1 to 4 in CI - mocked APIs allow safe parallelization
  workers: process.env.CI ? 4 : undefined,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:3000',
    // Changed from 'on-first-retry' to 'retain-on-failure'
    // Only keep traces when tests actually fail, not on first retry
    trace: 'retain-on-failure',
    // Run headless in CI, headed locally for debugging
    headless: !!process.env.CI,
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  /* Run your local dev server before starting the tests */
  webServer: {
    // Use production build for more realistic testing
    // 'npm run build && npm start' instead of 'npm run dev'
    command: 'npm run build && npm start',
    url: 'http://localhost:3000',
    reuseExistingServer: !process.env.CI,
    // Reduced from 120000 to 60000 - build should be cached in CI
    timeout: 60000,
    // Pass E2E_TEST env var to enable MSW in Next.js server
    env: {
      E2E_TEST: 'true',
    },
  },
});
