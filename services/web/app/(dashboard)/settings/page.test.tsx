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

  it('should redirect admin to /settings/organization', async () => {
    vi.mocked(getCurrentEmployee).mockResolvedValue({
      id: '1',
      email: 'admin@test.com',
      full_name: 'Admin User',
      role_name: 'Admin',
      org_id: 'org-1',
      role_id: 'role-1',
      status: 'active',
    } as any);

    // Need to re-import for each test since redirect is called
    vi.resetModules();

    const { default: SettingsPage } = await import('./page');
    await SettingsPage();

    expect(redirect).toHaveBeenCalledWith('/settings/organization');
  });

  it('should redirect non-admin to /settings/profile', async () => {
    vi.mocked(getCurrentEmployee).mockResolvedValue({
      id: '2',
      email: 'user@test.com',
      full_name: 'Regular User',
      role_name: 'Employee',
      org_id: 'org-1',
      role_id: 'role-2',
      status: 'active',
    } as any);

    vi.resetModules();

    const { default: SettingsPage } = await import('./page');
    await SettingsPage();

    expect(redirect).toHaveBeenCalledWith('/settings/profile');
  });
});
