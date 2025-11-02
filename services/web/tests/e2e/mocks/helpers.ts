/**
 * MSW Test Helpers
 *
 * Utilities for customizing MSW handlers during tests
 */

import { http, HttpResponse, delay } from 'msw';
import { server } from './server';

const API_BASE = 'http://localhost:8080/api/v1';

/**
 * Temporarily override a handler for the duration of a test
 *
 * @example
 * await withMockHandler(
 *   http.get(`${API_BASE}/agents`, async () => {
 *     await delay(2000); // Simulate slow network
 *     return HttpResponse.json({ agents: [] });
 *   }),
 *   async () => {
 *     // Test code runs here with the overridden handler
 *     await page.goto('/agents');
 *   }
 * );
 */
export async function withMockHandler<T>(
  handler: ReturnType<typeof http.get> | ReturnType<typeof http.post>,
  testFn: () => Promise<T>
): Promise<T> {
  // Add temporary handler
  server.use(handler);

  try {
    // Run test
    return await testFn();
  } finally {
    // Reset to default handlers after test
    server.resetHandlers();
  }
}

/**
 * Create a slow network handler (useful for testing loading states)
 */
export function createSlowHandler(
  method: 'get' | 'post' | 'patch' | 'delete',
  path: string,
  response: unknown,
  delayMs: number = 2000
) {
  const httpMethod = http[method];
  return httpMethod(path, async () => {
    await delay(delayMs);
    return HttpResponse.json(response);
  });
}

/**
 * Create an empty response handler (useful for testing empty states)
 */
export function createEmptyHandler(
  method: 'get' | 'post' | 'patch' | 'delete',
  path: string
) {
  const httpMethod = http[method];
  return httpMethod(path, () => {
    return HttpResponse.json({ agents: [], employees: [], teams: [], configs: [], total: 0 });
  });
}

/**
 * Create an error response handler (useful for testing error states)
 */
export function createErrorHandler(
  method: 'get' | 'post' | 'patch' | 'delete',
  path: string,
  status: number = 500,
  message: string = 'Internal server error'
) {
  const httpMethod = http[method];
  return httpMethod(path, () => {
    return HttpResponse.json(
      { error: 'server_error', message },
      { status }
    );
  });
}
