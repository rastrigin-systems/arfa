import { useQuery } from '@tanstack/react-query';
import { getTeams } from '../api/teams';
import type { Team } from '../api/types';

/**
 * Hook to fetch list of all teams
 */
export function useTeams() {
  return useQuery<Team[]>({
    queryKey: ['teams'],
    queryFn: getTeams,
  });
}
