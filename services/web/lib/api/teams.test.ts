import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getTeams, type Team } from './teams';

// Mock fetch
const mockFetch = vi.fn();
global.fetch = mockFetch;

describe('getTeams', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch teams list', async () => {
    const mockTeams: Team[] = [
      { id: 'team-1', org_id: 'org-1', name: 'Engineering', description: 'Engineering team' },
      { id: 'team-2', org_id: 'org-1', name: 'Sales', description: 'Sales team' },
      { id: 'team-3', org_id: 'org-1', name: 'Design', description: 'Design team' },
    ];

    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({ teams: mockTeams }),
    });

    const result = await getTeams();

    expect(result).toEqual(mockTeams);
    expect(mockFetch).toHaveBeenCalledWith('/api/teams', { credentials: 'include' });
  });

  it('should throw error on API failure', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.resolve({}),
    });

    await expect(getTeams())
      .rejects
      .toThrow('Failed to fetch teams');
  });
});
