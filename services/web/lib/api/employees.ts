import { apiClient } from './client';
import { getErrorMessage } from './errors';
import type { components } from './schema';
import type { EmployeesParams, UpdateEmployeeParams } from './types';

// Use schema types for API responses
type SchemaEmployee = components['schemas']['Employee'];

export interface EmployeesResponse {
  employees: SchemaEmployee[];
  total: number;
  page: number;
  limit: number;
}

// Re-export for backwards compatibility
export type Employee = SchemaEmployee;

/**
 * Get paginated list of employees with optional filters
 */
export async function getEmployees(params: EmployeesParams): Promise<EmployeesResponse> {
  const { data, error } = await apiClient.GET('/employees', {
    params: {
      query: {
        page: params.page,
        per_page: params.limit,
        status: params.status as 'active' | 'inactive' | 'suspended' | undefined,
      },
    },
  });

  if (error) {
    throw new Error(getErrorMessage(error, 'Failed to fetch employees'));
  }

  return {
    employees: data?.employees || [],
    total: data?.total || 0,
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
