import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { RolesClient } from './RolesClient';

type Role = {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  employee_count?: number;
  created_at: string;
  updated_at: string;
};

export const metadata = {
  title: 'Roles | Ubik Enterprise',
  description: 'Manage organizational roles and permissions',
};

export default async function RolesPage() {
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch roles from API
  const { data } = await apiClient.GET('/roles', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  // Note: API returns { roles: Role[], total: number } but OpenAPI spec says { data: Role[] }
  // Using type assertion to match actual API response
  const roles: Role[] = (data as { roles?: Role[] })?.roles ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-100">Roles</h1>
        <p className="text-base text-gray-600 dark:text-gray-400">
          Manage roles and permissions for your organization
        </p>
      </div>

      <RolesClient initialRoles={roles} />
    </div>
  );
}
