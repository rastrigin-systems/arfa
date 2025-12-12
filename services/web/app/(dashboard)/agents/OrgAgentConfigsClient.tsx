'use client';

import { useState } from 'react';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { AgentCatalog } from '@/components/agents/AgentCatalog';
import { OrgAgentConfigTable } from '@/components/agents/OrgAgentConfigTable';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import { useRouter } from 'next/navigation';
import { useToast } from '@/components/ui/use-toast';
import { apiClient } from '@/lib/api/client';

type Agent = {
  id: string;
  name: string;
  type: string;
  description: string;
  provider: string;
  llm_provider: string;
  llm_model: string;
  is_public: boolean;
  capabilities?: string[];
  default_config?: Record<string, unknown>;
};

type OrgAgentConfig = {
  id: string;
  org_id: string;
  agent_id: string;
  agent_name?: string;
  agent_type?: string;
  agent_provider?: string;
  config: Record<string, unknown>;
  is_enabled: boolean;
  created_at?: string;
  updated_at?: string;
};

type OrgAgentConfigsClientProps = {
  initialAgents: Agent[];
  initialOrgConfigs: OrgAgentConfig[];
};

export function OrgAgentConfigsClient({ initialAgents, initialOrgConfigs }: OrgAgentConfigsClientProps) {
  const [agents] = useState(initialAgents);
  const [orgConfigs, setOrgConfigs] = useState(initialOrgConfigs);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [selectedConfig, setSelectedConfig] = useState<OrgAgentConfig | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  // Build set of configured agent IDs
  const configuredAgentIds = new Set(orgConfigs.map((config) => config.agent_id));

  const handleConfigure = (agentId: string) => {
    const agent = agents.find((a) => a.id === agentId);
    if (!agent) return;

    // Check if config already exists
    const existingConfig = orgConfigs.find((c) => c.agent_id === agentId);

    setSelectedAgent(agent);
    setSelectedConfig(existingConfig || null);
    setIsModalOpen(true);
  };

  const handleEdit = (config: OrgAgentConfig) => {
    const agent = agents.find((a) => a.id === config.agent_id);
    if (!agent) return;

    setSelectedAgent(agent);
    setSelectedConfig(config);
    setIsModalOpen(true);
  };

  const handleToggleEnabled = async (config: OrgAgentConfig) => {
    const action = config.is_enabled ? 'disable' : 'enable';

    try {
      const { error } = await apiClient.PATCH('/organizations/current/agent-configs/{config_id}', {
        params: { path: { config_id: config.id } },
        body: {
          is_enabled: !config.is_enabled,
        },
      });

      if (error) {
        throw new Error(`Failed to ${action} configuration`);
      }

      // Update local state
      setOrgConfigs((prev) =>
        prev.map((c) => (c.id === config.id ? { ...c, is_enabled: !c.is_enabled } : c))
      );

      toast({
        title: 'Success',
        description: `Agent ${action}d successfully`,
        variant: 'success',
      });

      // Refresh page data
      router.refresh();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : `Failed to ${action} configuration`,
        variant: 'destructive',
      });
    }
  };

  const handleDelete = async (config: OrgAgentConfig) => {
    if (!confirm(`Are you sure you want to remove ${config.agent_name}? This will affect all employees using this agent.`)) {
      return;
    }

    try {
      const { error } = await apiClient.DELETE('/organizations/current/agent-configs/{config_id}', {
        params: { path: { config_id: config.id } },
      });

      if (error) {
        throw new Error('Failed to delete configuration');
      }

      // Remove from local state
      setOrgConfigs((prev) => prev.filter((c) => c.id !== config.id));

      toast({
        title: 'Success',
        description: 'Configuration removed successfully',
        variant: 'success',
      });

      // Refresh page data
      router.refresh();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to remove configuration',
        variant: 'destructive',
      });
    }
  };

  const handleModalSuccess = () => {
    // Refresh page data
    router.refresh();

    // Close modal
    setIsModalOpen(false);
    setSelectedAgent(null);
    setSelectedConfig(null);
  };

  return (
    <>
      <Tabs defaultValue="available" className="space-y-6">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="available">Available Agents</TabsTrigger>
          <TabsTrigger value="organization">Organization Configs</TabsTrigger>
          <TabsTrigger value="teams">Team Configs</TabsTrigger>
        </TabsList>

        <TabsContent value="available" className="space-y-4">
          <div data-testid="agent-grid">
            <AgentCatalog
              agents={agents}
              enabledAgentIds={configuredAgentIds}
              onConfigure={handleConfigure}
            />
          </div>
        </TabsContent>

        <TabsContent value="organization" className="space-y-4">
          <OrgAgentConfigTable
            configs={orgConfigs}
            onEdit={handleEdit}
            onDelete={handleDelete}
            onToggleEnabled={handleToggleEnabled}
          />
        </TabsContent>

        <TabsContent value="teams" className="space-y-4">
          <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
            <div>
              <h3 className="text-lg font-semibold">Coming Soon</h3>
              <p className="text-sm text-muted-foreground">
                Team-specific agent configurations will be available in a future release.
              </p>
            </div>
          </div>
        </TabsContent>
      </Tabs>

      {selectedAgent && (
        <ConfigEditorModal
          agent={selectedAgent}
          existingConfig={selectedConfig}
          open={isModalOpen}
          onOpenChange={setIsModalOpen}
          onSuccess={handleModalSuccess}
        />
      )}
    </>
  );
}
