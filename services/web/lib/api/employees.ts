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
 */
export async function updateEmployee(
  employeeId: string,
  params: UpdateEmployeeParams
): Promise<Employee> {
  const { data, error } = await apiClient.PATCH('/employees/{employee_id}', {
    params: {
      path: { employee_id: employeeId },
    },
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    body: params as any,
  });

  if (error) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    throw new Error((error as any).message || 'Failed to update employee');
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (data as any) as Employee;
}
