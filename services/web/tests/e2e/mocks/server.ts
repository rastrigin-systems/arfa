/**
 * MSW Server for Node.js (Server Components)
 *
 * This server intercepts API calls made from Next.js server components
 * during E2E tests, allowing tests to run without a real backend.
 */

import { setupServer } from 'msw/node';
import { handlers } from './handlers';

export const server = setupServer(...handlers);
