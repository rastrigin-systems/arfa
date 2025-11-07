import { useState, useEffect } from 'react';
import {
  validateInvitation,
  acceptInvitation,
  type Invitation,
  type AcceptInvitationRequest,
  type AcceptInvitationResponse,
  type ApiError,
} from '@/lib/api/invitations';

export type InvitationPageState =
  | 'loading'
  | 'valid'
  | 'invalid'
  | 'expired'
  | 'accepted'
  | 'error';

export interface UseInvitationReturn {
  state: InvitationPageState;
  invitation: Invitation | null;
  error: string | null;
  errorCode: string | null;
}

/**
 * Hook to validate an invitation token
 */
export function useInvitation(token: string | null): UseInvitationReturn {
  const [state, setState] = useState<InvitationPageState>('loading');
  const [invitation, setInvitation] = useState<Invitation | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [errorCode, setErrorCode] = useState<string | null>(null);

  useEffect(() => {
    if (!token) {
      setState('invalid');
      setError('Invitation token is missing');
      return;
    }

    async function loadInvitation() {
      setState('loading');
      try {
        const response = await validateInvitation(token!);
        setInvitation(response.invitation);
        setState('valid');
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
      } catch (err: any) {
        const apiError = err as ApiError & { status?: number };

        if (apiError.status === 404) {
          setState('invalid');
          setError(apiError.error || 'Invitation not found');
          setErrorCode(apiError.code || 'INVITATION_NOT_FOUND');
        } else if (apiError.status === 410) {
          setState('expired');
          setError(apiError.error || 'Invitation has expired');
          setErrorCode(apiError.code || 'INVITATION_EXPIRED');
        } else if (apiError.status === 409) {
          setState('accepted');
          setError(apiError.error || 'Invitation already accepted');
          setErrorCode(apiError.code || 'INVITATION_ALREADY_ACCEPTED');
        } else {
          setState('error');
          setError(apiError.error || 'Failed to validate invitation');
          setErrorCode(apiError.code || 'UNKNOWN_ERROR');
        }
      }
    }

    loadInvitation();
  }, [token]);

  return { state, invitation, error, errorCode };
}

export interface UseAcceptInvitationReturn {
  acceptInvitation: (data: AcceptInvitationRequest) => Promise<AcceptInvitationResponse>;
  isSubmitting: boolean;
  error: string | null;
  errorCode: string | null;
}

/**
 * Hook to accept an invitation
 */
export function useAcceptInvitation(token: string): UseAcceptInvitationReturn {
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [errorCode, setErrorCode] = useState<string | null>(null);

  const accept = async (data: AcceptInvitationRequest): Promise<AcceptInvitationResponse> => {
    setIsSubmitting(true);
    setError(null);
    setErrorCode(null);

    try {
      const response = await acceptInvitation(token, data);
      return response;
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      const apiError = err as ApiError & { status?: number };

      if (apiError.status === 409 && apiError.code === 'EMAIL_EXISTS') {
        setError('This email is already registered');
        setErrorCode('EMAIL_EXISTS');
      } else if (apiError.status === 400) {
        const errorMessage = apiError.details
          ? apiError.details.map((d) => d.message).join(', ')
          : apiError.error || 'Validation failed';
        setError(errorMessage);
        setErrorCode(apiError.code || 'VALIDATION_ERROR');
      } else if (apiError.status === 410) {
        setError('Invitation has expired');
        setErrorCode('INVITATION_EXPIRED');
      } else {
        setError(apiError.error || 'Failed to accept invitation');
        setErrorCode(apiError.code || 'UNKNOWN_ERROR');
      }

      throw err;
    } finally {
      setIsSubmitting(false);
    }
  };

  return { acceptInvitation: accept, isSubmitting, error, errorCode };
}
