/**
 * Authentication Test Fixtures
 *
 * Mock data for authentication-related E2E tests
 */

export const mockAuthUser = {
  id: '550e8400-e29b-41d4-a716-446655440001',
  email: 'alice@acme.com',
  name: 'Alice Johnson',
  org_id: '550e8400-e29b-41d4-a716-446655440030',
  role: 'member' as const,
};

export const mockAuthCredentials = {
  email: 'alice@acme.com',
  password: 'password123',
};

export const mockAdminCredentials = {
  email: 'bob@acme.com',
  password: 'admin123',
};

export const mockAuthTokens = {
  access_token: 'mock-jwt-token-abc123',
  refresh_token: 'mock-refresh-token-xyz789',
  expires_in: 3600,
  token_type: 'Bearer',
};

export const mockSession = {
  id: '880e8400-e29b-41d4-a716-446655440001',
  employee_id: mockAuthUser.id,
  org_id: mockAuthUser.org_id,
  token: mockAuthTokens.access_token,
  expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
  created_at: new Date().toISOString(),
};
