import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { updateEmployee, type Employee, type EmployeesResponse } from '../api/employees';
import type { EmployeesParams, UpdateEmployeeParams } from '../api/types';

/**
 * Hook to fetch employees list with filters and pagination
 */
export function useEmployees(params: EmployeesParams) {
  return useQuery<EmployeesResponse>({
    queryKey: ['employees', params],
    queryFn: async () => {
      const searchParams = new URLSearchParams();
      searchParams.set('page', String(params.page));
      searchParams.set('limit', String(params.limit));
      if (params.search) searchParams.set('search', params.search);
      if (params.team) searchParams.set('team_id', params.team);
      if (params.role) searchParams.set('role_id', params.role);
      if (params.status) searchParams.set('status', params.status);

      const res = await fetch(`/api/employees?${searchParams.toString()}`);
      if (!res.ok) throw new Error('Failed to fetch employees');
      return res.json();
    },
  });
}

/**
 * Hook to update employee details (team, role, status)
 */
export function useUpdateEmployee() {
  const queryClient = useQueryClient();

  return useMutation<
    Employee,
    Error,
    { employeeId: string; data: UpdateEmployeeParams }
  >({
    mutationFn: ({ employeeId, data }) => updateEmployee(employeeId, data),
    onSuccess: () => {
      // Invalidate employees queries to refetch data
      queryClient.invalidateQueries({ queryKey: ['employees'] });
    },
  });
}
