import { describe, it, expect, beforeEach, mock } from 'bun:test';

// Create mock functions
const mockGetCurrentEmployee = mock(() => Promise.resolve(null));
const mockRedirect = mock(() => {});

// Mock getCurrentEmployee before importing the component
mock.module('@/lib/auth', () => ({
  getCurrentEmployee: mockGetCurrentEmployee,
}));

// Mock redirect
mock.module('next/navigation', () => ({
  redirect: mockRedirect,
}));

// Import after mocking
import SettingsPage from './page';

describe('SettingsPage redirect logic', () => {
  beforeEach(() => {
    mockGetCurrentEmployee.mockClear();
    mockRedirect.mockClear();
  });

  it('should redirect to /login if no employee', async () => {
    mockGetCurrentEmployee.mockImplementation(() => Promise.resolve(null));

    await SettingsPage();

    expect(mockRedirect).toHaveBeenCalledWith('/login');
  });

  it('should redirect to /settings/profile for authenticated users', async () => {
    mockGetCurrentEmployee.mockImplementation(() =>
      Promise.resolve({
        id: '1',
        email: 'user@test.com',
        full_name: 'Test User',
        org_id: 'org-1',
        role_id: 'role-1',
        status: 'active' as const,
      })
    );

    await SettingsPage();

    expect(mockRedirect).toHaveBeenCalledWith('/settings/profile');
  });
});
