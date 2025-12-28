import { describe, it, expect, beforeEach, mock } from 'bun:test';
import { getTeams, type Team } from './teams';

// Mock fetch
const mockFetch = mock(() => Promise.resolve({ ok: true, json: () => Promise.resolve({}) }));
global.fetch = mockFetch as unknown as typeof fetch;

describe('getTeams', () => {
  beforeEach(() => {
    mockFetch.mockClear();
  });

  it('should fetch teams list', async () => {
    const mockTeams: Team[] = [
      { id: 'team-1', org_id: 'org-1', name: 'Engineering', description: 'Engineering team' },
      { id: 'team-2', org_id: 'org-1', name: 'Sales', description: 'Sales team' },
      { id: 'team-3', org_id: 'org-1', name: 'Design', description: 'Design team' },
    ];

    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: () => Promise.resolve({ teams: mockTeams }),
      } as Response)
    );

    const result = await getTeams();

    expect(result).toEqual(mockTeams);
    expect(mockFetch).toHaveBeenCalledWith('/api/teams', { credentials: 'include' });
  });

  it('should throw error on API failure', async () => {
    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: false,
        status: 500,
        json: () => Promise.resolve({}),
      } as Response)
    );

    await expect(getTeams()).rejects.toThrow('Failed to fetch teams');
  });
});
