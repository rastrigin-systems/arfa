/**
 * Employee Test Fixtures
 *
 * Mock data for employee-related E2E tests
 */

export const mockEmployee = {
  id: '550e8400-e29b-41d4-a716-446655440001',
  email: 'alice@acme.com',
  name: 'Alice Johnson',
  status: 'active' as const,
  role_id: '550e8400-e29b-41d4-a716-446655440010',
  role_name: 'Member',
  team_id: '550e8400-e29b-41d4-a716-446655440020',
  team_name: 'Engineering',
  org_id: '550e8400-e29b-41d4-a716-446655440030',
  created_at: '2024-01-15T10:00:00Z',
  updated_at: '2024-01-15T10:00:00Z',
};

export const mockEmployees = [
  mockEmployee,
  {
    id: '550e8400-e29b-41d4-a716-446655440002',
    email: 'bob@acme.com',
    name: 'Bob Smith',
    status: 'active' as const,
    role_id: '550e8400-e29b-41d4-a716-446655440011',
    role_name: 'Approver',
    team_id: '550e8400-e29b-41d4-a716-446655440020',
    team_name: 'Engineering',
    org_id: '550e8400-e29b-41d4-a716-446655440030',
    created_at: '2024-01-16T10:00:00Z',
    updated_at: '2024-01-16T10:00:00Z',
  },
  {
    id: '550e8400-e29b-41d4-a716-446655440003',
    email: 'charlie@acme.com',
    name: 'Charlie Brown',
    status: 'inactive' as const,
    role_id: '550e8400-e29b-41d4-a716-446655440010',
    role_name: 'Member',
    team_id: '550e8400-e29b-41d4-a716-446655440021',
    team_name: 'Product',
    org_id: '550e8400-e29b-41d4-a716-446655440030',
    created_at: '2024-01-17T10:00:00Z',
    updated_at: '2024-01-17T10:00:00Z',
  },
];

export const mockEmployeeCreateRequest = {
  email: 'newemployee@acme.com',
  name: 'New Employee',
  role_id: '550e8400-e29b-41d4-a716-446655440010',
  team_id: '550e8400-e29b-41d4-a716-446655440020',
};

export const mockEmployeeUpdateRequest = {
  name: 'Updated Name',
  role_id: '550e8400-e29b-41d4-a716-446655440011',
  team_id: '550e8400-e29b-41d4-a716-446655440021',
  status: 'inactive' as const,
};

export const mockCreateEmployeeResponse = {
  ...mockEmployeeCreateRequest,
  id: '550e8400-e29b-41d4-a716-446655440099',
  status: 'active' as const,
  role_name: 'Member',
  team_name: 'Engineering',
  org_id: '550e8400-e29b-41d4-a716-446655440030',
  created_at: '2024-02-01T10:00:00Z',
  updated_at: '2024-02-01T10:00:00Z',
  temporary_password: 'temp-password-123',
};
