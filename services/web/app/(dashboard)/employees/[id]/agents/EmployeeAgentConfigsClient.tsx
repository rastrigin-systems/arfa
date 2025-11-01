'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { useToast } from '@/components/ui/use-toast';
import { AgentCatalog } from '@/components/agents/AgentCatalog';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import { formatDistanceToNow } from 'date-fns';

type Employee = {
  readonly id: string;
  org_id: string;
  team_id?: string | null;
  readonly team_name?: string | null;
  role_id: string;
  email: string;
  full_name: string;
  status: 'active' | 'suspended' | 'inactive';
  preferences?: Record<string, unknown>;
  readonly has_personal_claude_token?: boolean;
  readonly last_login_at?: string | null;
  readonly created_at?: string;
  readonly updated_at?: string;
};

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

type EmployeeAgentConfig = {
  readonly id: string;
  employee_id: string;
  agent_id: string;
  readonly agent_name?: string;
  readonly agent_type?: string;
  readonly agent_provider?: string;
  config_override: Record<string, unknown>;
  is_enabled: boolean;
  sync_token?: string | null;
  readonly last_synced_at?: string | null;
  readonly created_at?: string;
  readonly updated_at?: string;
};

type EmployeeAgentConfigsClientProps = {
  employee: Employee;
  initialConfigs: EmployeeAgentConfig[];
  initialAgents: Agent[];
};

export function EmployeeAgentConfigsClient({
  employee,
  initialConfigs,
  initialAgents,
}: EmployeeAgentConfigsClientProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [configs] = useState(initialConfigs);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [editingConfig, setEditingConfig] = useState<EmployeeAgentConfig | null>(null);
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

  const handleEditConfig = (config: EmployeeAgentConfig) => {
    const agent = initialAgents.find((a) => a.id === config.agent_id);

    if (agent) {
      setSelectedAgent(agent);
      setEditingConfig(config);
      setIsModalOpen(true);
    }
  };

  const handleDeleteConfig = async (config: EmployeeAgentConfig) => {
    if (!confirm(`Are you sure you want to delete the configuration for ${config.agent_name}?`)) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/employees/${employee.id}/agent-configs/${config.id}`, {
        method: 'DELETE',
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

  const formatSyncTime = (lastSynced?: string | null) => {
    if (!lastSynced) return 'Never';
    try {
      return formatDistanceToNow(new Date(lastSynced), { addSuffix: true });
    } catch {
      return 'Invalid date';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <div className="flex items-center gap-2 mb-2">
            <Link href="/employees" className="text-sm text-muted-foreground hover:text-foreground">
              Employees
            </Link>
            <span className="text-sm text-muted-foreground">/</span>
            <Link
              href={`/employees/${employee.id}`}
              className="text-sm text-muted-foreground hover:text-foreground"
            >
              {employee.full_name}
            </Link>
            <span className="text-sm text-muted-foreground">/</span>
            <span className="text-sm font-medium">Agents</span>
          </div>
          <h1 className="text-3xl font-bold">Agent Configurations</h1>
        </div>
      </div>

      <Tabs defaultValue="configured" className="space-y-4">
        <TabsList>
          <TabsTrigger value="configured">Configured Agents ({configs.length})</TabsTrigger>
          <TabsTrigger value="available">Available Agents ({initialAgents.length})</TabsTrigger>
        </TabsList>

        <TabsContent value="configured" className="space-y-4">
          <Card>
            <CardHeader>
              <div className="flex items-center justify-between">
                <div>
                  <CardTitle>Employee Agent Configurations</CardTitle>
                  <CardDescription>
                    Agent overrides for {employee.full_name}. These override team and org settings.
                  </CardDescription>
                </div>
                <Button
                  onClick={() => {
                    setSelectedAgent(null);
                    setEditingConfig(null);
                    setIsModalOpen(true);
                  }}
                >
                  Add Configuration
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              {configs.length === 0 ? (
                <p className="text-sm text-muted-foreground text-center py-8">
                  No agent configurations yet. Click &quot;Add Configuration&quot; to get started.
                </p>
              ) : (
                <div className="space-y-4">
                  {configs.map((config) => (
                    <div
                      key={config.id}
                      className="border rounded-lg p-4 space-y-3"
                      data-testid={`agent-config-${config.agent_id}`}
                    >
                      <div className="flex items-start justify-between">
                        <div>
                          <div className="flex items-center gap-2">
                            <h3 className="font-semibold text-lg">{config.agent_name || 'Unknown Agent'}</h3>
                            <Badge variant={config.is_enabled ? 'default' : 'secondary'}>
                              {config.is_enabled ? 'Enabled' : 'Disabled'}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {config.agent_provider} • {config.agent_type}
                          </p>
                        </div>
                        <div className="flex gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => handleEditConfig(config)}
                            aria-label={`Edit ${config.agent_name} configuration`}
                          >
                            Edit
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => handleDeleteConfig(config)}
                            aria-label={`Delete ${config.agent_name} configuration`}
                          >
                            Delete
                          </Button>
                        </div>
                      </div>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-sm">
                        <div>
                          <dt className="font-medium text-muted-foreground">Last Sync</dt>
                          <dd>{formatSyncTime(config.last_synced_at)}</dd>
                        </div>
                        <div>
                          <dt className="font-medium text-muted-foreground">Sync Token</dt>
                          <dd className="font-mono text-xs">
                            {config.sync_token ? `${config.sync_token.substring(0, 8)}...` : 'Not generated'}
                          </dd>
                        </div>
                      </div>

                      {config.config_override && Object.keys(config.config_override).length > 0 && (
                        <div>
                          <dt className="font-medium text-muted-foreground text-sm mb-1">Configuration Override</dt>
                          <dd className="bg-muted rounded p-2">
                            <pre className="text-xs overflow-x-auto">
                              {JSON.stringify(config.config_override, null, 2)}
                            </pre>
                          </dd>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              )}
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
      </Tabs>

      {selectedAgent && (
        <ConfigEditorModal
          agent={selectedAgent}
          existingConfig={editingConfig}
          open={isModalOpen}
          onOpenChange={setIsModalOpen}
          onSuccess={handleModalSuccess}
          employeeId={employee.id}
          scope="employee"
        />
      )}
    </div>
  );
}
