import type { Role } from './types';

/**
 * Get list of all roles in the organization
 * Uses Next.js API route which handles auth token forwarding
 */
export async function getRoles(): Promise<Role[]> {
  const response = await fetch('/api/roles', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch roles');
  }

  const data = await response.json();
  return data.roles || [];
}
