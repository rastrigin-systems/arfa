import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getEmployees, updateEmployee } from '../api/employees';
import type {
  EmployeesParams,
  EmployeesResponse,
  UpdateEmployeeParams,
  Employee,
} from '../api/types';

/**
 * Hook to fetch employees list with filters and pagination
 */
export function useEmployees(params: EmployeesParams) {
  return useQuery<EmployeesResponse>({
    queryKey: ['employees', params],
    queryFn: () => getEmployees(params),
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
