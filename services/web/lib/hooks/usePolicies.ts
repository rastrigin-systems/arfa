import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getPolicies,
  createPolicy,
  updatePolicy,
  deletePolicy,
  type ToolPolicy,
  type CreateToolPolicyRequest,
  type UpdateToolPolicyRequest,
} from '../api/policies';

/**
 * Hook to fetch list of all tool policies
 */
export function usePolicies() {
  return useQuery<ToolPolicy[]>({
    queryKey: ['policies'],
    queryFn: getPolicies,
  });
}

/**
 * Hook to create a new tool policy
 */
export function useCreatePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateToolPolicyRequest) => createPolicy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] });
    },
  });
}

/**
 * Hook to update a tool policy
 */
export function useUpdatePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateToolPolicyRequest }) =>
      updatePolicy(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] });
    },
  });
}

/**
 * Hook to delete a tool policy
 */
export function useDeletePolicy() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => deletePolicy(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['policies'] });
    },
  });
}
