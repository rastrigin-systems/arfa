import { describe, it, expect } from 'vitest';
import {
  getInvitations,
  createInvitation,
  resendInvitation,
  cancelInvitation,
} from './invitations';

// NOTE: These are placeholder tests since the backend endpoints are not yet implemented

describe('getInvitations', () => {
  it('should return empty list (placeholder)', async () => {
    const result = await getInvitations({ page: 1, limit: 10 });
    expect(result.invitations).toEqual([]);
    expect(result.total).toBe(0);
  });
});

describe('createInvitation', () => {
  it('should return mock invitation (placeholder)', async () => {
    const result = await createInvitation({
      email: 'sarah@example.com',
      role_id: 'role-1',
    });
    expect(result.email).toBe('sarah@example.com');
    expect(result.id).toContain('mock-');
  });
});

describe('resendInvitation', () => {
  it('should return success message (placeholder)', async () => {
    const result = await resendInvitation('inv-1');
    expect(result.message).toBe('Invitation email resent successfully');
  });
});

describe('cancelInvitation', () => {
  it('should return success message (placeholder)', async () => {
    const result = await cancelInvitation('inv-1');
    expect(result.message).toBe('Invitation cancelled successfully');
  });
});
