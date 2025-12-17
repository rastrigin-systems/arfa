import { useQuery } from '@tanstack/react-query';
import { getTeams, type Team } from '../api/teams';

/**
 * Hook to fetch list of all teams
 */
export function useTeams() {
  return useQuery<Team[]>({
    queryKey: ['teams'],
    queryFn: getTeams,
  });
}
