import { describe, it, expect, beforeEach, mock } from 'bun:test';

// Custom error class to simulate Next.js redirect behavior
class RedirectError extends Error {
  constructor(public path: string) {
    super(`NEXT_REDIRECT:${path}`);
    this.name = 'RedirectError';
  }
}

// Employee type for mock return value
interface MockEmployee {
  id: string;
  email: string;
  full_name: string;
  org_id: string;
  role_id: string;
  status: 'active' | 'inactive' | 'suspended';
}

// Create mock functions with proper types
const mockGetCurrentEmployee = mock<() => Promise<MockEmployee | null>>(() => Promise.resolve(null));
const mockRedirect = mock<(path: string) => never>((path: string) => { throw new RedirectError(path); });

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

    try {
      await SettingsPage();
    } catch (error) {
      // Expected: redirect throws
      expect(error).toBeInstanceOf(RedirectError);
      expect((error as RedirectError).path).toBe('/login');
    }

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

    try {
      await SettingsPage();
    } catch (error) {
      // Expected: redirect throws
      expect(error).toBeInstanceOf(RedirectError);
      expect((error as RedirectError).path).toBe('/settings/profile');
    }

    expect(mockRedirect).toHaveBeenCalledWith('/settings/profile');
  });
});
