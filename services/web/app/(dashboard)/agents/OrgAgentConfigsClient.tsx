'use client';

import { useState } from 'react';
import { Building2, Users } from 'lucide-react';
import { ConfigurationHierarchy } from '@/components/agents/ConfigurationHierarchy';
import { ConfigurationLevelCard } from '@/components/agents/ConfigurationLevelCard';
import { AgentCatalog } from '@/components/agents/AgentCatalog';
import { OrgAgentConfigTable } from '@/components/agents/OrgAgentConfigTable';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import { useRouter } from 'next/navigation';
import { useToast } from '@/components/ui/use-toast';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

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

type ViewMode = 'overview' | 'catalog' | 'org-configs' | 'team-configs';

export function OrgAgentConfigsClient({ initialAgents, initialOrgConfigs }: OrgAgentConfigsClientProps) {
  const [agents] = useState(initialAgents);
  const [orgConfigs, setOrgConfigs] = useState(initialOrgConfigs);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [selectedConfig, setSelectedConfig] = useState<OrgAgentConfig | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [viewMode, setViewMode] = useState<ViewMode>('overview');
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
      // Use Next.js API route for auth forwarding
      const response = await fetch(`/api/organizations/current/agent-configs/${config.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          is_enabled: !config.is_enabled,
        }),
      });

      if (!response.ok) {
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
      // Use Next.js API route for auth forwarding
      const response = await fetch(`/api/organizations/current/agent-configs/${config.id}`, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) {
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

  // Overview view
  if (viewMode === 'overview') {
    return (
      <div className="space-y-6">
        {/* Hierarchy Explanation */}
        <ConfigurationHierarchy />

        {/* Configuration Levels */}
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          <ConfigurationLevelCard
            icon={Building2}
            title="Agent Catalog"
            description="Browse and enable agents"
            count={agents.length}
            countLabel="available"
            actionLabel="View Agent Catalog"
            onAction={() => setViewMode('catalog')}
            variant="primary"
          />

          <ConfigurationLevelCard
            icon={Building2}
            title="Organization Defaults"
            description="Base settings for all employees"
            count={orgConfigs.length}
            countLabel="configured"
            actionLabel="Manage Organization Defaults"
            onAction={() => setViewMode('org-configs')}
            variant="primary"
          />

          <ConfigurationLevelCard
            icon={Users}
            title="Team Overrides"
            description="Team-specific customizations"
            count={0}
            countLabel="teams"
            actionLabel="Manage Team Overrides"
            onAction={() => setViewMode('team-configs')}
            variant="secondary"
          />
        </div>

        {/* Quick Actions Card */}
        <Card>
          <CardHeader>
            <CardTitle>Quick Start</CardTitle>
            <CardDescription>Get started with agent configuration</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            <div className="flex items-start gap-3">
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                1
              </div>
              <div className="flex-1">
                <div className="font-medium">Browse the Agent Catalog</div>
                <div className="text-sm text-muted-foreground">
                  View available agents like Claude Code, Cursor, and Windsurf
                </div>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                2
              </div>
              <div className="flex-1">
                <div className="font-medium">Configure Organization Defaults</div>
                <div className="text-sm text-muted-foreground">
                  Set base configurations that apply to all employees
                </div>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex h-6 w-6 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                3
              </div>
              <div className="flex-1">
                <div className="font-medium">Add Team or Employee Overrides (Optional)</div>
                <div className="text-sm text-muted-foreground">
                  Customize settings for specific teams or individuals
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  // Agent Catalog view
  if (viewMode === 'catalog') {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Agent Catalog</h2>
            <p className="text-muted-foreground">Browse and enable agents for your organization</p>
          </div>
          <Button variant="outline" onClick={() => setViewMode('overview')}>
            Back to Overview
          </Button>
        </div>
        <div data-testid="agent-grid">
          <AgentCatalog
            agents={agents}
            enabledAgentIds={configuredAgentIds}
            onConfigure={handleConfigure}
          />
        </div>
      </div>
    );
  }

  // Organization Configs view
  if (viewMode === 'org-configs') {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Organization Defaults</h2>
            <p className="text-muted-foreground">Base agent configurations that apply to all employees</p>
          </div>
          <Button variant="outline" onClick={() => setViewMode('overview')}>
            Back to Overview
          </Button>
        </div>
        <OrgAgentConfigTable
          configs={orgConfigs}
          onEdit={handleEdit}
          onDelete={handleDelete}
          onToggleEnabled={handleToggleEnabled}
        />
        {selectedAgent && (
          <ConfigEditorModal
            agent={selectedAgent}
            existingConfig={selectedConfig}
            open={isModalOpen}
            onOpenChange={setIsModalOpen}
            onSuccess={handleModalSuccess}
          />
        )}
      </div>
    );
  }

  // Team Configs view (coming soon)
  if (viewMode === 'team-configs') {
    return (
      <div className="space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold">Team Overrides</h2>
            <p className="text-muted-foreground">Team-specific agent customizations</p>
          </div>
          <Button variant="outline" onClick={() => setViewMode('overview')}>
            Back to Overview
          </Button>
        </div>
        <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
          <Users className="h-12 w-12 text-muted-foreground" />
          <div>
            <h3 className="text-lg font-semibold">Coming Soon</h3>
            <p className="text-sm text-muted-foreground">
              Team-specific agent configurations will be available in a future release.
            </p>
          </div>
          <Button onClick={() => setViewMode('overview')}>Return to Overview</Button>
        </div>
      </div>
    );
  }

  return null;
}
