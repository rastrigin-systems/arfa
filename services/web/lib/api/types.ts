/**
 * Client-side request parameter types for API calls.
 *
 * For entity types (Employee, Role, Team, etc.), import from:
 * - @/lib/types for centralized schema-based types
 * - Respective API modules (employees.ts, roles.ts, etc.) for backwards compatibility
 *
 * For invitation-related types, import from @/lib/api/invitations
 */

/**
 * Parameters for fetching employees list
 */
export interface EmployeesParams {
  page: number;
  limit: number;
  search?: string;
  team?: string;
  role?: string;
  status?: string;
}

/**
 * Parameters for updating an employee
 */
export interface UpdateEmployeeParams {
  team_id?: string;
  role_id?: string;
  status?: 'active' | 'inactive' | 'suspended';
}
