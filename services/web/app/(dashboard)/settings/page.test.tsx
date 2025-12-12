import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock getCurrentEmployee before importing the component
vi.mock('@/lib/auth', () => ({
  getCurrentEmployee: vi.fn(),
}));

// Mock redirect
vi.mock('next/navigation', () => ({
  redirect: vi.fn(),
}));

import { redirect } from 'next/navigation';
import { getCurrentEmployee } from '@/lib/auth';

describe('SettingsPage redirect logic', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should redirect to /login if no employee', async () => {
    vi.mocked(getCurrentEmployee).mockResolvedValue(null);

    // Dynamically import the page to trigger redirect
    const { default: SettingsPage } = await import('./page');
    await SettingsPage();

    expect(redirect).toHaveBeenCalledWith('/login');
  });

  it('should redirect to /settings/profile for authenticated users', async () => {
    vi.mocked(getCurrentEmployee).mockResolvedValue({
      id: '1',
      email: 'user@test.com',
      full_name: 'Test User',
      org_id: 'org-1',
      role_id: 'role-1',
      status: 'active' as const,
    });

    vi.resetModules();

    const { default: SettingsPage } = await import('./page');
    await SettingsPage();

    expect(redirect).toHaveBeenCalledWith('/settings/profile');
  });
});
