import { apiClient } from './client';
import type { Role } from './types';

/**
 * Get list of all roles in the organization
 */
export async function getRoles(): Promise<Role[]> {
  const { data, error } = await apiClient.GET('/roles');

  if (error) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    throw new Error((error as any).message || 'Failed to fetch roles');
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return ((data as any)?.data || []) as Role[];
}
