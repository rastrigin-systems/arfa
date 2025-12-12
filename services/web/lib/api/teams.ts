import { apiClient } from './client';
import type { Team } from './types';

/**
 * Get list of all teams in the organization
 */
export async function getTeams(): Promise<Team[]> {
  const { data, error } = await apiClient.GET('/teams');

  if (error) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    throw new Error((error as any).message || 'Failed to fetch teams');
  }

  // API returns { teams: [...], total: number }
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return ((data as any)?.teams || []) as Team[];
}
