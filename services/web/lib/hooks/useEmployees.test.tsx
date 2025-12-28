import { describe, it, expect, beforeEach, mock } from 'bun:test';
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import type { Employee } from '../api/employees';

// Create mock function
const mockUpdateEmployee = mock(() => Promise.resolve({}));
const mockFetch = mock(() => Promise.resolve({ ok: true, json: () => Promise.resolve({}) }));

// Mock API before importing
mock.module('../api/employees', () => ({
  updateEmployee: mockUpdateEmployee,
}));

// Import after mocking
import { useEmployees, useUpdateEmployee } from './useEmployees';

// Test wrapper with QueryClient
const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  });

  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
};

describe('useEmployees', () => {
  beforeEach(() => {
    mockFetch.mockClear();
    global.fetch = mockFetch as unknown as typeof fetch;
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

    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: true,
        json: async () => mockResponse,
      } as Response)
    );

    const { result } = renderHook(
      () => useEmployees({ page: 1, limit: 10 }),
      { wrapper: createWrapper() }
    );

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.employees).toEqual(mockEmployees);
    expect(result.current.data?.total).toBe(1);
  });

  it('should handle errors', async () => {
    mockFetch.mockImplementation(() =>
      Promise.resolve({
        ok: false,
        statusText: 'Server Error',
      } as Response)
    );

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
    mockUpdateEmployee.mockClear();
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

    mockUpdateEmployee.mockImplementation(() => Promise.resolve(updatedEmployee));

    const { result } = renderHook(() => useUpdateEmployee(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ employeeId: 'emp-1', data: { team_id: 'team-2' } });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data).toEqual(updatedEmployee);
  });

  it('should handle update errors', async () => {
    mockUpdateEmployee.mockImplementation(() =>
      Promise.reject(new Error('Failed to update employee'))
    );

    const { result } = renderHook(() => useUpdateEmployee(), {
      wrapper: createWrapper(),
    });

    result.current.mutate({ employeeId: 'emp-1', data: { team_id: 'team-2' } });

    await waitFor(() => expect(result.current.isError).toBe(true));

    expect(result.current.error).toBeDefined();
  });
});
