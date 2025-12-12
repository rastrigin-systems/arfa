import { getCurrentEmployee, getServerToken } from '@/lib/auth';
import { redirect } from 'next/navigation';
import { apiClient } from '@/lib/api/client';
import { OrganizationSettingsClient } from './OrganizationSettingsClient';

export default async function OrganizationSettingsPage() {
  const employee = await getCurrentEmployee();
  const token = await getServerToken();

  if (!employee || !token) {
    redirect('/login');
  }

  // TODO: Add role-based access control when API supports role_name
  // For now, organization settings are accessible to all users

  // Fetch organization details
  const { data: orgData, error: orgError } = await apiClient.GET('/organizations/current', {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (orgError) {
    throw new Error('Failed to load organization details');
  }

  return (
    <div className="space-y-6">
      {/* Page Header */}
      <div>
        <nav className="text-sm text-muted-foreground mb-2">
          Settings &gt; Organization
        </nav>
        <h1 className="text-3xl font-bold tracking-tight">Organization Settings</h1>
        <p className="text-muted-foreground mt-1">
          Manage your organization details and plan information
        </p>
      </div>

      <OrganizationSettingsClient organization={orgData} />
    </div>
  );
}
