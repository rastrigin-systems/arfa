import { useQuery } from '@tanstack/react-query';
import { getRoles, type Role } from '../api/roles';

/**
 * Hook to fetch list of all roles
 */
export function useRoles() {
  return useQuery<Role[]>({
    queryKey: ['roles'],
    queryFn: getRoles,
  });
}
