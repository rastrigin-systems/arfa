import createClient from 'openapi-fetch';
import type { paths } from './schema';

const API_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api/v1';

export const apiClient = createClient<paths>({ baseUrl: API_URL });

// Helper to set auth token
export function setAuthToken(token: string) {
  apiClient.use({
    onRequest({ request }) {
      request.headers.set('Authorization', `Bearer ${token}`);
      return request;
    },
  });
}

// Helper to clear auth token
export function clearAuthToken() {
  apiClient.eject();
}
