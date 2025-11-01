import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { OrgAgentConfigsClient } from './OrgAgentConfigsClient';

export default async function OrgAgentConfigsPage() {
  // Get authentication token
  const token = await getServerToken();

  if (!token) {
    throw new Error('Unauthorized');
  }

  // Fetch organization agent configurations
  const { data: configsData, error: configsError } = await apiClient.GET('/organizations/current/agent-configs', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (configsError) {
    throw new Error('Failed to load organization agent configurations');
  }

  // Fetch all agents from catalog
  const { data: agentsData, error: agentsError } = await apiClient.GET('/agents', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  if (agentsError) {
    throw new Error('Failed to load agent catalog');
  }

  // Type assertion needed because API schema has readonly created_at/updated_at as optional,
  // but our component expects them as required strings
  const configs = (configsData?.configs || []).map((config) => ({
    ...config,
    created_at: config.created_at || new Date().toISOString(),
    updated_at: config.updated_at || new Date().toISOString(),
  }));

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Organization Agent Configurations</h1>
        <p className="text-muted-foreground">
          Configure AI agents at the organization level to make them available to your teams. Agents configured here
          will be accessible to all employees unless overridden at the team level.
        </p>
      </div>

      <OrgAgentConfigsClient initialConfigs={configs} initialAgents={agentsData?.agents || []} />
    </div>
  );
}
