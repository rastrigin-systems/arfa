import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  getInvitations,
  createInvitation,
  resendInvitation,
  cancelInvitation,
} from '../api/invitations';
import type {
  InvitationsParams,
  InvitationsResponse,
  CreateInvitationParams,
  CreateInvitationResponse,
} from '../api/invitations';

/**
 * Hook to fetch invitations list with filters and pagination
 */
export function useInvitations(params: InvitationsParams) {
  return useQuery<InvitationsResponse>({
    queryKey: ['invitations', params],
    queryFn: () => getInvitations(params),
  });
}

/**
 * Hook to create a new invitation
 */
export function useCreateInvitation() {
  const queryClient = useQueryClient();

  return useMutation<
    CreateInvitationResponse['invitation'],
    Error,
    CreateInvitationParams
  >({
    mutationFn: createInvitation,
    onSuccess: () => {
      // Invalidate invitations queries to refetch data
      queryClient.invalidateQueries({ queryKey: ['invitations'] });
    },
  });
}

/**
 * Hook to resend an invitation
 */
export function useResendInvitation() {
  return useMutation<{ message: string }, Error, string>({
    mutationFn: resendInvitation,
  });
}

/**
 * Hook to cancel an invitation
 */
export function useCancelInvitation() {
  const queryClient = useQueryClient();

  return useMutation<{ message: string }, Error, string>({
    mutationFn: cancelInvitation,
    onSuccess: () => {
      // Invalidate invitations queries to refetch data
      queryClient.invalidateQueries({ queryKey: ['invitations'] });
    },
  });
}
