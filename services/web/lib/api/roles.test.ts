import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getRoles } from './roles';
import type { Role } from './types';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('getRoles', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch roles list', async () => {
    const mockRoles: Role[] = [
      { id: 'role-1', name: 'Member', description: 'Standard member' },
      { id: 'role-2', name: 'Approver', description: 'Can approve requests' },
      { id: 'role-3', name: 'Administrator', description: 'Full access' },
    ];

    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ roles: mockRoles }),
    });

    const result = await getRoles();

    expect(result).toEqual(mockRoles);
    expect(mockFetch).toHaveBeenCalledWith('/api/roles', { credentials: 'include' });
  });

  it('should throw error on API failure', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
    });

    await expect(getRoles())
      .rejects
      .toThrow('Failed to fetch roles');
  });
});
