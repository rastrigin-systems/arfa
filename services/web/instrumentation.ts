/**
 * Next.js Instrumentation Hook
 *
 * This file runs when the Next.js server starts, allowing us to set up
 * MSW for API mocking during E2E tests.
 *
 * @see https://nextjs.org/docs/app/building-your-application/optimizing/instrumentation
 */

export async function register() {
  // Only run MSW during E2E tests
  if (process.env.NEXT_RUNTIME === 'nodejs' && process.env.E2E_TEST === 'true') {
    try {
      // Use require.resolve to check if the module exists before importing
      // This prevents webpack from failing when tests directory is excluded
      const modulePath = './tests/e2e/mocks/server';
      const { server } = await import(/* webpackIgnore: true */ modulePath);

      server.listen({
        onUnhandledRequest: 'warn',
      });

      console.log('✅ MSW server started in Next.js process for E2E tests');

      // Clean up on process exit
      process.on('SIGTERM', () => {
        server.close();
        console.log('🛑 MSW server stopped');
      });
    } catch (error) {
      console.warn('⚠️  MSW server not available (tests directory excluded from build)');
    }
  }
}
