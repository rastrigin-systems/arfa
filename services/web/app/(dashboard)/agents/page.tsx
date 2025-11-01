import { apiClient } from '@/lib/api/client';
import { getServerToken } from '@/lib/auth';
import { AgentCatalogClient } from './AgentCatalogClient';

export default async function AgentsPage() {
  // Fetch agents from API
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

  // Fetch enabled agents for the organization
  const { data: orgConfigsData } = await apiClient.GET('/organizations/current/agent-configs', {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  });

  // Build set of enabled agent IDs
  const enabledAgentIds = new Set(orgConfigsData?.configs?.map((config) => config.agent_id) || []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold">Agent Catalog</h1>
        <p className="text-muted-foreground">
          Browse and configure AI agents for your organization. Enable agents to make them available to your teams.
        </p>
      </div>

      <AgentCatalogClient initialAgents={agentsData?.agents || []} initialEnabledAgentIds={enabledAgentIds} />
    </div>
  );
}
