'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { OrgAgentConfigTable } from '@/components/agents/OrgAgentConfigTable';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import { TeamAgentConfigsTab } from '@/components/agents/TeamAgentConfigsTab';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useToast } from '@/components/ui/use-toast';
import { AgentCatalog } from '@/components/agents/AgentCatalog';

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
  initialConfigs: OrgAgentConfig[];
  initialAgents: Agent[];
};

export function OrgAgentConfigsClient({ initialConfigs, initialAgents }: OrgAgentConfigsClientProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [configs] = useState(initialConfigs);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [editingConfig, setEditingConfig] = useState<OrgAgentConfig | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Build set of enabled agent IDs
  const enabledAgentIds = new Set(configs.map((config) => config.agent_id));

  const handleEnableAgent = (agent: Agent) => {
    setSelectedAgent(agent);
    setEditingConfig(null);
    setIsModalOpen(true);
  };

  const handleConfigureAgent = (agentId: string) => {
    const config = configs.find((c) => c.agent_id === agentId);
    const agent = initialAgents.find((a) => a.id === agentId);

    if (config && agent) {
      setSelectedAgent(agent);
      setEditingConfig(config);
      setIsModalOpen(true);
    }
  };

  const handleEditConfig = (config: OrgAgentConfig) => {
    const agent = initialAgents.find((a) => a.id === config.agent_id);

    if (agent) {
      setSelectedAgent(agent);
      setEditingConfig(config);
      setIsModalOpen(true);
    }
  };

  const handleDeleteConfig = async (config: OrgAgentConfig) => {
    if (!confirm(`Are you sure you want to delete the configuration for ${config.agent_name}?`)) {
      return;
    }

    try {
      // Use Next.js API route for auth forwarding
      const response = await fetch(`/api/organizations/current/agent-configs/${config.id}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) {
        throw new Error('Failed to delete configuration');
      }

      toast({
        title: 'Success',
        description: `Configuration for ${config.agent_name} deleted successfully`,
        variant: 'success',
      });

      // Refresh data
      router.refresh();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    }
  };

  const handleModalSuccess = () => {
    // Refresh the page to get updated data
    router.refresh();
  };

  return (
    <div className="space-y-6">
      <Tabs defaultValue="configured" className="space-y-4">
        <TabsList>
          <TabsTrigger value="configured">Configured Agents ({configs.length})</TabsTrigger>
          <TabsTrigger value="available">Available Agents ({initialAgents.length})</TabsTrigger>
          <TabsTrigger value="teams">Team Configs</TabsTrigger>
        </TabsList>

        <TabsContent value="configured" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Organization Agent Configurations</CardTitle>
              <CardDescription>
                Manage agent configurations for your organization. These settings apply to all employees.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <OrgAgentConfigTable configs={configs} onEdit={handleEditConfig} onDelete={handleDeleteConfig} />
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="available" className="space-y-4">
          <AgentCatalog
            agents={initialAgents}
            enabledAgentIds={enabledAgentIds}
            onEnable={(agentId: string) => {
              const agent = initialAgents.find((a) => a.id === agentId);
              if (agent) handleEnableAgent(agent);
            }}
            onConfigure={handleConfigureAgent}
          />
        </TabsContent>

        <TabsContent value="teams" className="space-y-4">
          <TeamAgentConfigsTab orgConfigs={configs} agents={initialAgents} />
        </TabsContent>
      </Tabs>

      {selectedAgent && (
        <ConfigEditorModal
          agent={selectedAgent}
          existingConfig={editingConfig}
          open={isModalOpen}
          onOpenChange={setIsModalOpen}
          onSuccess={handleModalSuccess}
        />
      )}
    </div>
  );
}
