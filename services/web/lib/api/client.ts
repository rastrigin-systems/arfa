import createClient from 'openapi-fetch';
import type { paths } from './schema';

const API_URL = process.env.API_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api/v1';

// Use relative URLs for client-side requests to leverage Next.js rewrites
// This ensures cookies are forwarded through the proxy
const isBrowser = typeof window !== 'undefined';
const baseUrl = isBrowser ? '/api/v1' : API_URL;

export const apiClient = createClient<paths>({
  baseUrl,
  // Include credentials to send cookies with requests
  credentials: 'include',
});

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
