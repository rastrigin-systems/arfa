// API types for invitation endpoints
// NOTE: Team management endpoints (list, create, resend, cancel) are placeholders
// until backend implementation is complete

import { apiClient } from './client';

export interface InvitationOrganization {
  id: string;
  name: string;
  slug: string;
}

export interface InvitationInviter {
  id: string;
  full_name: string;
  email: string;
  role: {
    name: string;
  };
}

export interface InvitationRole {
  id: string;
  name: string;
  description: string;
}

export interface InvitationTeam {
  id: string;
  name: string;
}

export interface Invitation {
  id: string;
  email: string;
  status: 'pending' | 'accepted' | 'expired' | 'cancelled';
  created_at: string;
  expires_at: string;
  organization: InvitationOrganization;
  inviter: InvitationInviter;
  role: InvitationRole;
  team?: InvitationTeam;
}

export interface ValidateInvitationResponse {
  invitation: Invitation;
}

export interface AcceptInvitationRequest {
  full_name: string;
  password: string;
}

export interface AcceptInvitationResponse {
  employee: {
    id: string;
    email: string;
    full_name: string;
    role: {
      id: string;
      name: string;
    };
    team?: {
      id: string;
      name: string;
    };
  };
  organization: {
    id: string;
    name: string;
    slug: string;
  };
  token: string;
}

export interface ApiError {
  error: string;
  code?: string;
  details?: { field: string; message: string }[];
  expired_at?: string;
  accepted_at?: string;
}

// Team management types (placeholders)
export interface InvitationsParams {
  page: number;
  limit: number;
  status?: 'pending' | 'accepted' | 'expired' | 'cancelled';
}

export interface InvitationsResponse {
  invitations: Invitation[];
  total: number;
  page: number;
  limit: number;
}

export interface CreateInvitationParams {
  email: string;
  role_id: string;
  team_id?: string;
  message?: string;
}

export interface CreateInvitationResponse {
  invitation: {
    id: string;
    email: string;
    token: string;
    invitation_url: string;
    expires_at: string;
  };
}

/**
 * Validate an invitation token
 * GET /api/v1/invitations/{token}
 */
export async function validateInvitation(token: string): Promise<ValidateInvitationResponse> {
  const { data, error, response } = await apiClient.GET('/invitations/{token}', {
    params: { path: { token } },
  });

  if (error) {
    throw { status: response?.status, ...(error as ApiError) };
  }

  // Transform API response to expected format
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return { invitation: data } as any as ValidateInvitationResponse;
}

/**
 * Accept an invitation
 * POST /api/v1/invitations/{token}/accept
 */
export async function acceptInvitation(
  token: string,
  requestData: AcceptInvitationRequest
): Promise<AcceptInvitationResponse> {
  const { data, error, response } = await apiClient.POST('/invitations/{token}/accept', {
    params: { path: { token } },
    body: requestData,
  });

  if (error) {
    throw { status: response?.status, ...(error as ApiError) };
  }

  // Transform API response to expected format
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return data as any as AcceptInvitationResponse;
}

// ============================================================================
// Team Management Functions (PLACEHOLDERS - Backend not yet implemented)
// ============================================================================

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
export async function resendInvitation(_invitationId: string): Promise<{ message: string }> {
  // Placeholder implementation - simulates success
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
export async function cancelInvitation(_invitationId: string): Promise<{ message: string }> {
  // Placeholder implementation - simulates success
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
