// API types for invitation endpoints
// NOTE: Team management endpoints (list, create, resend, cancel) are placeholders
// until backend implementation is complete
//
// TYPE MISMATCH DOCUMENTATION:
// The local Invitation/AcceptInvitationResponse types expect nested objects
// (organization, inviter, role, team) but the OpenAPI schema has flat fields
// (org_name, role_name, team_name). The API actually returns nested data
// that matches our local types, but the schema is incomplete.
// TODO: Update OpenAPI spec to match actual API response structure.

import { apiClient } from './client';
import { getErrorMessage } from './errors';

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
    // Preserve status for error handling in hooks
    const apiError: ApiError & { status?: number } = {
      error: getErrorMessage(error, 'Failed to validate invitation'),
      status: response?.status,
    };
    throw apiError;
  }

  // API returns nested objects matching our Invitation type, but schema is incomplete
  // See TYPE MISMATCH DOCUMENTATION at top of file
  return { invitation: data as unknown as Invitation };
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
    // Preserve status for error handling in hooks
    const apiError: ApiError & { status?: number } = {
      error: getErrorMessage(error, 'Failed to accept invitation'),
      status: response?.status,
    };
    throw apiError;
  }

  // API returns nested objects matching our AcceptInvitationResponse type
  // See TYPE MISMATCH DOCUMENTATION at top of file
  return data as unknown as AcceptInvitationResponse;
}

// ============================================================================
// Team Management Functions (PLACEHOLDERS - Backend not yet implemented)
// ============================================================================

/**
 * Get paginated list of invitations with optional status filter
 * TODO: Backend endpoint GET /invitations needs to be implemented
 */
export async function getInvitations(params: InvitationsParams): Promise<InvitationsResponse> {
  // Placeholder - returns empty list until backend endpoint is available
  return {
    invitations: [],
    total: 0,
    page: params.page,
    limit: params.limit,
  };
}

/**
 * Create a new invitation
 * TODO: Backend endpoint POST /invitations needs to be implemented
 */
export async function createInvitation(
  params: CreateInvitationParams
): Promise<CreateInvitationResponse['invitation']> {
  // Placeholder - simulates success until backend endpoint is available
  return {
    id: 'mock-' + Date.now(),
    email: params.email,
    token: 'mock-token',
    invitation_url: 'https://app.ubik.io/accept-invite?token=mock-token',
    expires_at: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
  };
}

/**
 * Resend an existing invitation email
 * TODO: Backend endpoint POST /invitations/{id}/resend needs to be implemented
 */
export async function resendInvitation(_invitationId: string): Promise<{ message: string }> {
  // Placeholder - simulates success until backend endpoint is available
  return { message: 'Invitation email resent successfully' };
}

/**
 * Cancel (delete) an invitation
 * TODO: Backend endpoint DELETE /invitations/{id} needs to be implemented
 */
export async function cancelInvitation(_invitationId: string): Promise<{ message: string }> {
  // Placeholder - simulates success until backend endpoint is available
  return { message: 'Invitation cancelled successfully' };
}
