import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useEmployees, useUpdateEmployee } from './useEmployees';
import * as employeesApi from '../api/employees';
import type { Employee } from '../api/employees';

// Mock API for updateEmployee which still uses the api client
vi.mock('../api/employees', () => ({
  updateEmployee: vi.fn(),
}));

// Test wrapper with QueryClient
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  // eslint-disable-next-line react/display-name
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useEmployees', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    global.fetch = vi.fn();
  });

  it('should fetch employees', async () => {
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

    const mockResponse = {
      employees: mockEmployees,
      total: 1,
      page: 1,
      limit: 10,
    };

    vi.mocked(global.fetch).mockResolvedValue({
      ok: true,
      json: async () => mockResponse,
    } as Response);

    const { result } = renderHook(
      () => useEmployees({ page: 1, limit: 10 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.employees).toEqual(mockEmployees);
    expect(result.current.data?.total).toBe(1);
  });

  it('should handle errors', async () => {
    vi.mocked(global.fetch).mockResolvedValue({
      ok: false,
      statusText: 'Server Error',
    } as Response);

    const { result } = renderHook(
      () => useEmployees({ page: 1, limit: 10 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });
});

describe('useUpdateEmployee', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should update employee', async () => {
    const updatedEmployee: Employee = {
      id: 'emp-1',
      org_id: 'org-1',
      email: 'john@example.com',
      full_name: 'John Smith',
      status: 'active',
      role_id: 'role-1',
      team_id: 'team-2',
      team_name: 'Sales',
    };

    vi.mocked(employeesApi.updateEmployee).mockResolvedValue(updatedEmployee);

    const { result } = renderHook(() => useUpdateEmployee(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ employeeId: 'emp-1', data: { team_id: 'team-2' } });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(updatedEmployee);
  });

  it('should handle update errors', async () => {
    vi.mocked(employeesApi.updateEmployee).mockRejectedValue(
      new Error('Failed to update employee')
    );

    const { result } = renderHook(() => useUpdateEmployee(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ employeeId: 'emp-1', data: { team_id: 'team-2' } });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });
});
