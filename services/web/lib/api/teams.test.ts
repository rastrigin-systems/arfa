import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getTeams } from './teams';
import type { Team } from './types';

// Mock API client
vi.mock('./client', () => ({
  apiClient: {
    GET: vi.fn(),
  },
}));

describe('getTeams', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch teams list', async () => {
    const { apiClient } = await import('./client');

    const mockTeams: Team[] = [
      { id: 'team-1', name: 'Engineering', description: 'Engineering team' },
      { id: 'team-2', name: 'Sales', description: 'Sales team' },
      { id: 'team-3', name: 'Design', description: 'Design team' },
    ];

    vi.mocked(apiClient.GET).mockResolvedValue({
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      data: { data: mockTeams } as any,
      error: undefined,
      response: { ok: true } as Response,
    });

    const result = await getTeams();

    expect(result).toEqual(mockTeams);
    expect(apiClient.GET).toHaveBeenCalledWith('/teams');
  });

  it('should throw error on API failure', async () => {
    const { apiClient } = await import('./client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: undefined,
      error: { message: 'Failed to fetch teams' },
      response: { ok: false, status: 500 } as Response,
    });

    await expect(getTeams())
      .rejects
      .toThrow('Failed to fetch teams');
  });
});
