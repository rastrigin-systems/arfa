import { getServerToken } from '@/lib/auth';
import { apiClient } from '@/lib/api/client';
import { PoliciesClient } from './PoliciesClient';
import type { ToolPolicy } from '@/lib/types';

export const metadata = {
  title: 'Policies | Arfa',
  description: 'Manage tool policies for your organization',
};

export default async function PoliciesPage() {
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch policies from API
  const { data } = await apiClient.GET('/policies', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  const policies: ToolPolicy[] = data?.policies ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-100">Policies</h1>
        <p className="text-base text-gray-600 dark:text-gray-400">
          Manage tool policies to control agent behavior in your organization
        </p>
      </div>

      <PoliciesClient initialPolicies={policies} />
    </div>
  );
}
