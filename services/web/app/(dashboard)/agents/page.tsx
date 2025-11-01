import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { OrgAgentConfigsClient } from './OrgAgentConfigsClient';

export default async function AgentsPage() {
  // Fetch data from API
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
        <h1 className="text-3xl font-bold">Organization Agent Configuration</h1>
        <p className="text-muted-foreground">
          Manage AI agents available to your organization
        </p>
      </div>

      <OrgAgentConfigsClient
        initialAgents={agentsData?.agents || []}
        initialOrgConfigs={orgConfigsData?.configs || []}
      />
    </div>
  );
}
