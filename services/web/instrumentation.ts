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
    const { server } = await import('./tests/e2e/mocks/server');

    server.listen({
      onUnhandledRequest: 'warn',
    });

    console.log('✅ MSW server started in Next.js process for E2E tests');

    // Clean up on process exit
    process.on('SIGTERM', () => {
      server.close();
      console.log('🛑 MSW server stopped');
    });
  }
}
