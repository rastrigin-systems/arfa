// NOTE: Invitation endpoints are not yet implemented in the backend.
// This is a placeholder implementation ready for when the backend endpoints are added.
// The frontend UI is fully implemented and tested.

import type {
  InvitationsParams,
  InvitationsResponse,
  CreateInvitationParams,
  CreateInvitationResponse,
} from './types';

/**
 * Get paginated list of invitations with optional status filter
 * TODO: Backend endpoint needs to be implemented: GET /invitations
 */
export async function getInvitations(params: InvitationsParams): Promise<InvitationsResponse> {
  // Placeholder implementation - returns empty list until backend is ready
  return {
    invitations: [],
    total: 0,
    page: params.page,
    limit: params.limit,
  };

  /* When backend is ready, uncomment this:
  const { data, error } = await apiClient.GET('/invitations', {
    params: {
      query: params as any,
    },
  });

  if (error) {
    throw new Error((error as any).message || 'Failed to fetch invitations');
  }

  return {
    invitations: (data as any).invitations || [],
    total: (data as any).total || 0,
    page: params.page,
    limit: params.limit,
  };
  */
}

/**
 * Create a new invitation
 * TODO: Backend endpoint needs to be implemented: POST /invitations
 */
export async function createInvitation(
  params: CreateInvitationParams
): Promise<CreateInvitationResponse['invitation']> {
  // Placeholder implementation - simulates success
  return {
    id: 'mock-' + Date.now(),
    email: params.email,
    token: 'mock-token',
    invitation_url: 'https://app.ubik.io/accept-invite?token=mock-token',
    expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
  };

  /* When backend is ready, uncomment this:
  const { data, error } = await apiClient.POST('/invitations', {
    body: params as any,
  });

  if (error) {
    throw new Error((error as any).message || 'Failed to create invitation');
  }

  return (data as any).invitation;
  */
}

/**
 * Resend an existing invitation email
 * TODO: Backend endpoint needs to be implemented: POST /invitations/{id}/resend
 */
export async function resendInvitation(invitationId: string): Promise<{ message: string }> {
  // Placeholder implementation - simulates success
  console.log('Resending invitation:', invitationId);
  return { message: 'Invitation email resent successfully' };

  /* When backend is ready, uncomment this:
  const { data, error } = await apiClient.POST('/invitations/{id}/resend', {
    params: {
      path: { id: invitationId },
    },
  });

  if (error) {
    throw new Error((error as any).message || 'Failed to resend invitation');
  }

  return data as { message: string };
  */
}

/**
 * Cancel (delete) an invitation
 * TODO: Backend endpoint needs to be implemented: DELETE /invitations/{id}
 */
export async function cancelInvitation(invitationId: string): Promise<{ message: string }> {
  // Placeholder implementation - simulates success
  console.log('Cancelling invitation:', invitationId);
  return { message: 'Invitation cancelled successfully' };

  /* When backend is ready, uncomment this:
  const { data, error } = await apiClient.DELETE('/invitations/{id}', {
    params: {
      path: { id: invitationId },
    },
  });

  if (error) {
    throw new Error((error as any).message || 'Failed to cancel invitation');
  }

  return data as { message: string };
  */
}
