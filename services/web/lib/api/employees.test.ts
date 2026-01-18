import { describe, it, expect, beforeEach, mock } from 'bun:test';
import type { Employee } from './employees';

// Types for mock returns
interface EmployeesApiResponse {
  data: { employees: Employee[]; total: number } | undefined;
  error: { message: string } | undefined;
  response: Response;
}

interface PatchApiResponse {
  data: Record<string, unknown> | undefined;
  error: { message: string } | undefined;
  response: Response;
}

// Create mock functions with proper types
const mockGET = mock<() => Promise<EmployeesApiResponse>>(() => Promise.resolve({ data: { employees: [], total: 0 }, error: undefined, response: { ok: true } as Response }));
const mockPATCH = mock<() => Promise<PatchApiResponse>>(() => Promise.resolve({ data: {}, error: undefined, response: { ok: true } as Response }));

// Mock the API client module
mock.module('./client', () => ({
  apiClient: {
    GET: mockGET,
    PATCH: mockPATCH,
  },
}));

// Import after mocking
import { getEmployees } from './employees';

describe('getEmployees', () => {
  beforeEach(() => {
    mockGET.mockClear();
    mockPATCH.mockClear();
  });

  it('should fetch employees list', async () => {
    const mockEmployees: Employee[] = [
      {
        id: 'emp-1',
        org_id: 'org-1',
        email: 'john@example.com',
        full_name: 'John Smith',
        status: 'active',
        role_id: 'role-1',
        team_id: 'team-1',
        team_name: 'Engineering',
      },
    ];

    mockGET.mockImplementation(() =>
      Promise.resolve({
        data: {
          employees: mockEmployees,
          total: 1,
        },
        error: undefined,
        response: { ok: true } as Response,
      })
    );

    const result = await getEmployees({ page: 1, limit: 10 });

    expect(result.employees).toEqual(mockEmployees);
    expect(result.total).toBe(1);
    expect(result.page).toBe(1);
    expect(result.limit).toBe(10);
    expect(mockGET).toHaveBeenCalledWith('/employees', {
      params: {
        query: { page: 1, per_page: 10, status: undefined },
      },
    });
  });

  it('should include status filter', async () => {
    mockGET.mockImplementation(() =>
      Promise.resolve({
        data: { employees: [], total: 0 },
        error: undefined,
        response: { ok: true } as Response,
      })
    );

    await getEmployees({ page: 1, limit: 10, status: 'active' });

    expect(mockGET).toHaveBeenCalledWith('/employees', {
      params: {
        query: { page: 1, per_page: 10, status: 'active' },
      },
    });
  });

  it('should throw error on API failure', async () => {
    mockGET.mockImplementation(() =>
      Promise.resolve({
        data: undefined,
        error: { message: 'Failed to fetch employees' },
        response: { ok: false, status: 500 } as Response,
      })
    );

    await expect(getEmployees({ page: 1, limit: 10 }))
      .rejects
      .toThrow('Failed to fetch employees');
  });
});

