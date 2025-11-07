import { useQuery } from '@tanstack/react-query';
import { getRoles } from '../api/roles';
import type { Role } from '../api/types';

/**
 * Hook to fetch list of all roles
 */
export function useRoles() {
  return useQuery<Role[]>({
    queryKey: ['roles'],
    queryFn: getRoles,
  });
}
