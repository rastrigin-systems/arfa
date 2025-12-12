import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { ConfigsClient } from './ConfigsClient';

export default async function ConfigsPage() {
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch all agents from catalog
  const { data: agentsData, error: agentsError } = await apiClient.GET('/agents', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (agentsError) {
    throw new Error('Failed to load agents');
  }

  // Fetch organization agent configs
  const { data: orgConfigsData } = await apiClient.GET('/organizations/current/agent-configs', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Agent Configurations</h1>
        <p className="text-muted-foreground">
          View and manage all agent configurations across your organization
        </p>
      </div>

      <ConfigsClient
        initialAgents={agentsData?.agents || []}
        initialOrgConfigs={orgConfigsData?.configs || []}
      />
    </div>
  );
}
