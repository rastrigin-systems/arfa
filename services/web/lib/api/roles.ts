import { getErrorMessage } from './errors';
import type { components } from './schema';

// Use schema types for API responses
type SchemaRole = components['schemas']['Role'];

// Re-export for backwards compatibility
export type Role = SchemaRole;

/**
 * Get list of all roles in the organization
 * Uses Next.js API route which handles auth token forwarding
 */
export async function getRoles(): Promise<Role[]> {
  const response = await fetch('/api/roles', {
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to fetch roles'));
  }

  const data = await response.json();
  return data.roles || [];
}
