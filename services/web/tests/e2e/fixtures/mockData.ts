/**
 * Mock data for E2E tests
 *
 * This file contains all mock entities used in Playwright E2E tests.
 * Mock data is based on the OpenAPI schema types.
 */

import type { components } from '../../../lib/api/schema';

// Type aliases for convenience
type Employee = components['schemas']['Employee'];
type Organization = components['schemas']['Organization'];
type Team = components['schemas']['Team'];
type Role = components['schemas']['Role'];

// Mock Organization
export const mockOrganization: Organization = {
  id: '01234567-89ab-cdef-0123-456789abcdef',
  name: 'ACME Corporation',
  slug: 'acme',
  plan: 'professional',
  max_employees: 50,
  max_agents_per_employee: 5,
  has_claude_token: true,
  settings: {
    theme: 'light',
    notifications: true,
  },
  created_at: '2025-01-01T00:00:00Z',
  updated_at: '2025-01-15T10:30:00Z',
};

// Mock Roles
export const mockRoles: Role[] = [
  {
    id: 'role-member-uuid',
    name: 'Member',
    description: 'Standard employee with basic permissions',
    permissions: ['read:own_profile', 'read:logs', 'sync:configs'],
    created_at: '2025-01-01T00:00:00Z',
    employee_count: 8,
  },
  {
    id: 'role-approver-uuid',
    name: 'Approver',
    description: 'Can approve requests and manage team configs',
    permissions: [
      'read:own_profile',
      'read:logs',
      'sync:configs',
      'approve:requests',
      'manage:team_configs',
    ],
    created_at: '2025-01-01T00:00:00Z',
    employee_count: 2,
  },
];

// Mock Teams
export const mockTeams: Team[] = [
  {
    id: 'team-engineering-uuid',
    org_id: mockOrganization.id,
    name: 'Engineering',
    description: 'Software development team',
    member_count: 5,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
  {
    id: 'team-product-uuid',
    org_id: mockOrganization.id,
    name: 'Product',
    description: 'Product management team',
    member_count: 3,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
  {
    id: 'team-design-uuid',
    org_id: mockOrganization.id,
    name: 'Design',
    description: 'UI/UX design team',
    member_count: 2,
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T10:30:00Z',
  },
];

// Mock Employees
export const mockEmployees: Employee[] = [
  {
    id: 'employee-alice-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[1].id,
    email: 'alice@acme.com',
    full_name: 'Alice Johnson',
    status: 'active',
    has_personal_claude_token: true,
    preferences: {
      theme: 'dark',
      notifications: true,
    },
    last_login_at: '2025-01-15T09:00:00Z',
    created_at: '2025-01-01T00:00:00Z',
    updated_at: '2025-01-15T09:00:00Z',
  },
  {
    id: 'employee-bob-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[0].id,
    email: 'bob@acme.com',
    full_name: 'Bob Smith',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {
      theme: 'light',
    },
    last_login_at: '2025-01-14T16:30:00Z',
    created_at: '2025-01-02T00:00:00Z',
    updated_at: '2025-01-14T16:30:00Z',
  },
  {
    id: 'employee-charlie-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[1].id,
    team_name: mockTeams[1].name,
    role_id: mockRoles[0].id,
    email: 'charlie@acme.com',
    full_name: 'Charlie Davis',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {},
    last_login_at: '2025-01-15T08:15:00Z',
    created_at: '2025-01-03T00:00:00Z',
    updated_at: '2025-01-15T08:15:00Z',
  },
  {
    id: 'employee-diana-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[2].id,
    team_name: mockTeams[2].name,
    role_id: mockRoles[0].id,
    email: 'diana@acme.com',
    full_name: 'Diana Martinez',
    status: 'active',
    has_personal_claude_token: false,
    preferences: {
      theme: 'dark',
    },
    last_login_at: '2025-01-13T11:45:00Z',
    created_at: '2025-01-04T00:00:00Z',
    updated_at: '2025-01-13T11:45:00Z',
  },
  {
    id: 'employee-evan-uuid',
    org_id: mockOrganization.id,
    team_id: mockTeams[0].id,
    team_name: mockTeams[0].name,
    role_id: mockRoles[0].id,
    email: 'evan@acme.com',
    full_name: 'Evan Wilson',
    status: 'suspended',
    has_personal_claude_token: false,
    preferences: {},
    last_login_at: '2025-01-10T14:20:00Z',
    created_at: '2025-01-05T00:00:00Z',
    updated_at: '2025-01-12T10:00:00Z',
  },
];

// Helper function to get mock employee by email
export function getMockEmployeeByEmail(email: string): Employee | undefined {
  return mockEmployees.find((emp) => emp.email === email);
}

// Helper function to filter employees by status
export function filterEmployeesByStatus(
  status: 'active' | 'suspended' | 'inactive'
): Employee[] {
  return mockEmployees.filter((emp) => emp.status === status);
}

// Helper function to search employees
export function searchEmployees(query: string): Employee[] {
  const lowerQuery = query.toLowerCase();
  return mockEmployees.filter(
    (emp) =>
      emp.full_name.toLowerCase().includes(lowerQuery) ||
      emp.email.toLowerCase().includes(lowerQuery)
  );
}
