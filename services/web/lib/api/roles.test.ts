import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getRoles } from './roles';
import type { Role } from './types';

// Mock API client
vi.mock('./client', () => ({
  apiClient: {
    GET: vi.fn(),
  },
}));

describe('getRoles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch roles list', async () => {
    const { apiClient } = await import('./client');

    const mockRoles: Role[] = [
      { id: 'role-1', name: 'Member', description: 'Standard member' },
      { id: 'role-2', name: 'Approver', description: 'Can approve requests' },
      { id: 'role-3', name: 'Administrator', description: 'Full access' },
    ];

    vi.mocked(apiClient.GET).mockResolvedValue({
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data: { data: mockRoles } as any,
      error: undefined,
      response: { ok: true } as Response,
    });

    const result = await getRoles();

    expect(result).toEqual(mockRoles);
    expect(apiClient.GET).toHaveBeenCalledWith('/roles');
  });

  it('should throw error on API failure', async () => {
    const { apiClient } = await import('./client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: undefined,
      error: { message: 'Failed to fetch roles' },
      response: { ok: false, status: 500 } as Response,
    });

    await expect(getRoles())
      .rejects
      .toThrow('Failed to fetch roles');
  });
});
