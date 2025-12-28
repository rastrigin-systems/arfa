import { describe, it, expect, beforeEach, mock } from 'bun:test';
import { getRoles, type Role } from './roles';

// Mock fetch
const mockFetch = mock(() => Promise.resolve({ ok: true, json: () => Promise.resolve({}) }));
global.fetch = mockFetch as unknown as typeof fetch;

describe('getRoles', () => {
  beforeEach(() => {
    mockFetch.mockClear();
  });

  it('should fetch roles list', async () => {
    const mockRoles: Role[] = [
      { id: 'role-1', name: 'Member', description: 'Standard member' },
      { id: 'role-2', name: 'Approver', description: 'Can approve requests' },
      { id: 'role-3', name: 'Administrator', description: 'Full access' },
    ];

    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ roles: mockRoles }),
      } as Response)
    );

    const result = await getRoles();

    expect(result).toEqual(mockRoles);
    expect(mockFetch).toHaveBeenCalledWith('/api/roles', { credentials: 'include' });
  });

  it('should throw error on API failure', async () => {
    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: false,
        status: 500,
        json: () => Promise.resolve({}),
      } as Response)
    );

    await expect(getRoles())
      .rejects
      .toThrow('Failed to fetch roles');
  });
});
