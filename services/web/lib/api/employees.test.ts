import { describe, it, expect, vi, beforeEach } from 'vitest';
import { getEmployees, updateEmployee } from './employees';
import type { Employee } from './types';

// Mock API client
vi.mock('./client', () => ({
  apiClient: {
    GET: vi.fn(),
    PATCH: vi.fn(),
  },
}));

describe('getEmployees', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should fetch employees list', async () => {
    const { apiClient } = await import('./client');

    const mockEmployees: Employee[] = [
      {
        id: 'emp-1',
        email: 'john@example.com',
        full_name: 'John Smith',
        status: 'active',
        role: { id: 'role-1', name: 'Admin' },
        team: { id: 'team-1', name: 'Engineering' },
        created_at: '2024-01-01T00:00:00Z',
      },
    ];

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: {
        employees: mockEmployees,
        total: 1,
        page: 1,
        limit: 10,
      },
      error: undefined,
      response: { ok: true } as Response,
    });

    const result = await getEmployees({ page: 1, limit: 10 });

    expect(result.employees).toEqual(mockEmployees);
    expect(result.total).toBe(1);
    expect(apiClient.GET).toHaveBeenCalledWith('/employees', {
      params: {
        query: { page: 1, limit: 10 },
      },
    });
  });

  it('should include search parameter', async () => {
    const { apiClient } = await import('./client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { employees: [], total: 0, page: 1, limit: 10 },
      error: undefined,
      response: { ok: true } as Response,
    });

    await getEmployees({ page: 1, limit: 10, search: 'john' });

    expect(apiClient.GET).toHaveBeenCalledWith('/employees', {
      params: {
        query: { page: 1, limit: 10, search: 'john' },
      },
    });
  });

  it('should include filter parameters', async () => {
    const { apiClient } = await import('./client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: { employees: [], total: 0, page: 1, limit: 10 },
      error: undefined,
      response: { ok: true } as Response,
    });

    await getEmployees({
      page: 1,
      limit: 10,
      team: 'team-1',
      role: 'role-1',
      status: 'active',
    });

    expect(apiClient.GET).toHaveBeenCalledWith('/employees', {
      params: {
        query: {
          page: 1,
          limit: 10,
          team: 'team-1',
          role: 'role-1',
          status: 'active',
        },
      },
    });
  });

  it('should throw error on API failure', async () => {
    const { apiClient } = await import('./client');

    vi.mocked(apiClient.GET).mockResolvedValue({
      data: undefined,
      error: { message: 'Failed to fetch employees' },
      response: { ok: false, status: 500 } as Response,
    });

    await expect(getEmployees({ page: 1, limit: 10 }))
      .rejects
      .toThrow('Failed to fetch employees');
  });
});

describe('updateEmployee', () => {
  const mockFetch = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = mockFetch;
  });

  it('should update employee team', async () => {
    const updatedEmployee: Employee = {
      id: 'emp-1',
      email: 'john@example.com',
      full_name: 'John Smith',
      status: 'active',
      role: { id: 'role-1', name: 'Admin' },
      team: { id: 'team-2', name: 'Sales' },
      created_at: '2024-01-01T00:00:00Z',
    };

    mockFetch.mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(updatedEmployee),
    });

    const result = await updateEmployee('emp-1', { team_id: 'team-2' });

    expect(result).toEqual(updatedEmployee);
    expect(mockFetch).toHaveBeenCalledWith('/api/employees/emp-1', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ team_id: 'team-2' }),
    });
  });

  it('should throw error on update failure', async () => {
    mockFetch.mockResolvedValue({
      ok: false,
      json: () => Promise.resolve({ error: 'Failed to update employee' }),
    });

    await expect(updateEmployee('emp-1', { team_id: 'team-2' }))
      .rejects
      .toThrow('Failed to update employee');
  });
});
