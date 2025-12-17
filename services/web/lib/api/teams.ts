import { getErrorMessage } from './errors';
import type { components } from './schema';

// Use schema types for API responses
type SchemaTeam = components['schemas']['Team'];

// Re-export for backwards compatibility
export type Team = SchemaTeam;

/**
 * Get list of all teams in the organization
 * Uses Next.js API route which handles auth token forwarding
 */
export async function getTeams(): Promise<Team[]> {
  const response = await fetch('/api/teams', {
    credentials: 'include',
  });

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(getErrorMessage(errorData, 'Failed to fetch teams'));
  }

  const data = await response.json();
  return data.teams || [];
}
