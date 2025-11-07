// API types for invitation endpoints
// These will be replaced with generated types when backend is implemented

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

/**
 * Validate an invitation token
 * GET /api/v1/invitations/{token}
 */
export async function validateInvitation(token: string): Promise<ValidateInvitationResponse> {
  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api/v1';
  const response = await fetch(`${API_URL}/invitations/${token}`, {
    method: 'GET',
    headers: {
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    const error: ApiError = await response.json();
    throw { status: response.status, ...error };
  }

  return response.json();
}

/**
 * Accept an invitation
 * POST /api/v1/invitations/{token}/accept
 */
export async function acceptInvitation(
  token: string,
  data: AcceptInvitationRequest
): Promise<AcceptInvitationResponse> {
  const API_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:3001/api/v1';
  const response = await fetch(`${API_URL}/invitations/${token}/accept`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const error: ApiError = await response.json();
    throw { status: response.status, ...error };
  }

  return response.json();
}
