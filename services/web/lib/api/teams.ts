import type { Team } from './types';

/**
 * Get list of all teams in the organization
 * Uses Next.js API route which handles auth token forwarding
 */
export async function getTeams(): Promise<Team[]> {
  const response = await fetch('/api/teams', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to fetch teams');
  }

  const data = await response.json();
  return data.teams || [];
}
