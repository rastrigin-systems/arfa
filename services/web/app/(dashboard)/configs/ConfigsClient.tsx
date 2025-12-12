'use client';

import { useState, useMemo } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Plus, Search } from 'lucide-react';
import { ConfigsTable } from '@/components/configs/ConfigsTable';
import { CreateConfigModal } from '@/components/configs/CreateConfigModal';
import { useRouter } from 'next/navigation';
import { useToast } from '@/components/ui/use-toast';

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

type ConfigWithLevel = OrgAgentConfig & {
  level: 'organization' | 'team' | 'employee';
  assigned_to: string;
};

type ConfigsClientProps = {
  initialAgents: Agent[];
  initialOrgConfigs: OrgAgentConfig[];
};

export function ConfigsClient({ initialAgents, initialOrgConfigs }: ConfigsClientProps) {
  const [agents] = useState(initialAgents);
  const [orgConfigs, setOrgConfigs] = useState(initialOrgConfigs);
  const [levelFilter, setLevelFilter] = useState<'all' | 'organization' | 'team' | 'employee'>('all');
  const [agentFilter, setAgentFilter] = useState<string>('all');
  const [statusFilter, setStatusFilter] = useState<'all' | 'enabled' | 'disabled'>('all');
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const router = useRouter();
  const { toast } = useToast();

  // Combine all configs with level information
  const allConfigs: ConfigWithLevel[] = useMemo(() => {
    const configs: ConfigWithLevel[] = [];

    // Add org configs
    orgConfigs.forEach((config) => {
      configs.push({
        ...config,
        level: 'organization',
        assigned_to: 'Organization',
      });
    });

    // TODO: Add team configs when API is available
    // TODO: Add employee configs when API is available

    return configs;
  }, [orgConfigs]);

  // Filter configs
  const filteredConfigs = useMemo(() => {
    return allConfigs.filter((config) => {
      // Level filter
      if (levelFilter !== 'all' && config.level !== levelFilter) {
        return false;
      }

      // Agent filter
      if (agentFilter !== 'all' && config.agent_id !== agentFilter) {
        return false;
      }

      // Status filter
      if (statusFilter === 'enabled' && !config.is_enabled) {
        return false;
      }
      if (statusFilter === 'disabled' && config.is_enabled) {
        return false;
      }

      // Search query
      if (searchQuery) {
        const query = searchQuery.toLowerCase();
        const matchesName = config.agent_name?.toLowerCase().includes(query);
        const matchesAssignedTo = config.assigned_to.toLowerCase().includes(query);
        if (!matchesName && !matchesAssignedTo) {
          return false;
        }
      }

      return true;
    });
  }, [allConfigs, levelFilter, agentFilter, statusFilter, searchQuery]);

  const handleToggleEnabled = async (config: ConfigWithLevel) => {
    if (config.level !== 'organization') {
      // TODO: Handle team and employee configs when available
      return;
    }

    const action = config.is_enabled ? 'disable' : 'enable';

    try {
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
        description: `Configuration ${action}d successfully`,
        variant: 'success',
      });

      router.refresh();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : `Failed to ${action} configuration`,
        variant: 'destructive',
      });
    }
  };

  const handleDelete = async (config: ConfigWithLevel) => {
    if (config.level !== 'organization') {
      // TODO: Handle team and employee configs when available
      return;
    }

    if (!confirm(`Are you sure you want to delete this configuration for ${config.agent_name}?`)) {
      return;
    }

    try {
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
        description: 'Configuration deleted successfully',
        variant: 'success',
      });

      router.refresh();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to delete configuration',
        variant: 'destructive',
      });
    }
  };

  return (
    <div className="space-y-6">
      {/* Filters and Search */}
      <div className="flex flex-col gap-4 md:flex-row md:items-center md:justify-between">
        <div className="flex flex-1 flex-col gap-4 md:flex-row md:items-center">
          {/* Level Filter */}
          <Select value={levelFilter} onValueChange={(value) => setLevelFilter(value as 'all' | 'organization' | 'team' | 'employee')}>
            <SelectTrigger className="w-full md:w-[180px]">
              <SelectValue placeholder="Level" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Levels</SelectItem>
              <SelectItem value="organization">Organization</SelectItem>
              <SelectItem value="team">Team</SelectItem>
              <SelectItem value="employee">Employee</SelectItem>
            </SelectContent>
          </Select>

          {/* Agent Filter */}
          <Select value={agentFilter} onValueChange={setAgentFilter}>
            <SelectTrigger className="w-full md:w-[180px]">
              <SelectValue placeholder="Agent" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Agents</SelectItem>
              {agents.map((agent) => (
                <SelectItem key={agent.id} value={agent.id}>
                  {agent.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>

          {/* Status Filter */}
          <Select value={statusFilter} onValueChange={(value) => setStatusFilter(value as 'all' | 'enabled' | 'disabled')}>
            <SelectTrigger className="w-full md:w-[180px]">
              <SelectValue placeholder="Status" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Status</SelectItem>
              <SelectItem value="enabled">Enabled</SelectItem>
              <SelectItem value="disabled">Disabled</SelectItem>
            </SelectContent>
          </Select>

          {/* Search */}
          <div className="relative flex-1">
            <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              placeholder="Search configurations..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="pl-8"
            />
          </div>
        </div>

        {/* Create Button */}
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          New Configuration
        </Button>
      </div>

      {/* Table */}
      <ConfigsTable
        configs={filteredConfigs}
        onToggleEnabled={handleToggleEnabled}
        onDelete={handleDelete}
      />

      {/* Create Modal */}
      <CreateConfigModal
        agents={agents}
        open={isCreateModalOpen}
        onOpenChange={setIsCreateModalOpen}
        onSuccess={() => {
          setIsCreateModalOpen(false);
          router.refresh();
        }}
      />
    </div>
  );
}
