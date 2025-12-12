import { apiClient } from './client';
import type {
  Employee,
  EmployeesParams,
  EmployeesResponse,
  UpdateEmployeeParams,
} from './types';

/**
 * Get paginated list of employees with optional filters
 */
export async function getEmployees(params: EmployeesParams): Promise<EmployeesResponse> {
  const { data, error } = await apiClient.GET('/employees', {
    params: {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      query: params as any,
    },
  });

  if (error) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    throw new Error((error as any).message || 'Failed to fetch employees');
  }

  // Transform API response to match our types
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    employees: (data as any).employees || [],
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    total: (data as any).total || 0,
    page: params.page,
    limit: params.limit,
  };
}

/**
 * Update employee details (team, role, status)
 * Uses Next.js API route which handles auth token forwarding
 */
export async function updateEmployee(
  employeeId: string,
  params: UpdateEmployeeParams
): Promise<Employee> {
  const response = await fetch(`/api/employees/${employeeId}`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(params),
  });

  if (!response.ok) {
    const errorData = await response.json();
    throw new Error(errorData.error || 'Failed to update employee');
  }

  return response.json();
}
