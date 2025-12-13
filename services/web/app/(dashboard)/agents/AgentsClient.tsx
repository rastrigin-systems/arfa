'use client';

import { useState } from 'react';
import { CompactAgentCard } from '@/components/agents/CompactAgentCard';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
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

type AgentsClientProps = {
  initialAgents: Agent[];
  initialOrgConfigs: OrgAgentConfig[];
};

export function AgentsClient({ initialAgents, initialOrgConfigs }: AgentsClientProps) {
  const [agents] = useState(initialAgents);
  const [orgConfigs, setOrgConfigs] = useState(initialOrgConfigs);
  const [confirmDialog, setConfirmDialog] = useState<{ open: boolean; agent: Agent | null }>({
    open: false,
    agent: null,
  });
  const [loadingAgentId, setLoadingAgentId] = useState<string | null>(null);
  const router = useRouter();
  const { toast } = useToast();

  // Build set of enabled agent IDs
  const enabledAgentIds = new Set(
    orgConfigs.filter((config) => config.is_enabled).map((config) => config.agent_id)
  );

  const handleToggle = async (agentId: string, shouldEnable: boolean) => {
    const agent = agents.find((a) => a.id === agentId);
    if (!agent) return;

    // If disabling, show confirmation dialog
    if (!shouldEnable) {
      setConfirmDialog({ open: true, agent });
      return;
    }

    // If enabling, proceed directly
    await enableAgent(agentId);
  };

  const enableAgent = async (agentId: string) => {
    const agent = agents.find((a) => a.id === agentId);
    if (!agent) return;

    setLoadingAgentId(agentId);

    try {
      // Check if config already exists
      const existingConfig = orgConfigs.find((c) => c.agent_id === agentId);

      if (existingConfig) {
        // Update existing config to enable it
        const response = await fetch(`/api/organizations/current/agent-configs/${existingConfig.id}`, {
          method: 'PATCH',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({ is_enabled: true }),
        });

        if (!response.ok) {
          throw new Error('Failed to enable agent');
        }

        // Update local state
        setOrgConfigs((prev) =>
          prev.map((c) => (c.id === existingConfig.id ? { ...c, is_enabled: true } : c))
        );
      } else {
        // Create new config
        const response = await fetch('/api/organizations/current/agent-configs', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          credentials: 'include',
          body: JSON.stringify({
            agent_id: agentId,
            config: agent.default_config || {},
            is_enabled: true,
          }),
        });

        if (!response.ok) {
          throw new Error('Failed to enable agent');
        }

        const data = await response.json();
        setOrgConfigs((prev) => [...prev, data]);
      }

      toast({
        title: 'Agent enabled',
        description: `${agent.name} is now available to your organization`,
        variant: 'success',
      });

      router.refresh();
    } catch (error) {
      toast({
        title: 'Failed to enable agent',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    } finally {
      setLoadingAgentId(null);
    }
  };

  const disableAgent = async (agentId: string) => {
    const agent = agents.find((a) => a.id === agentId);
    const existingConfig = orgConfigs.find((c) => c.agent_id === agentId);

    if (!agent || !existingConfig) return;

    setLoadingAgentId(agentId);
    setConfirmDialog({ open: false, agent: null });

    try {
      const response = await fetch(`/api/organizations/current/agent-configs/${existingConfig.id}`, {
        method: 'PATCH',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ is_enabled: false }),
      });

      if (!response.ok) {
        throw new Error('Failed to disable agent');
      }

      // Update local state
      setOrgConfigs((prev) =>
        prev.map((c) => (c.id === existingConfig.id ? { ...c, is_enabled: false } : c))
      );

      toast({
        title: 'Agent disabled',
        description: `${agent.name} has been disabled`,
        variant: 'success',
      });

      router.refresh();
    } catch (error) {
      toast({
        title: 'Failed to disable agent',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    } finally {
      setLoadingAgentId(null);
    }
  };

  const handleConfigure = (agentId: string) => {
    // Navigate to configs page with agent filter
    router.push(`/configs?agent=${agentId}`);
  };

  const handleCancelDisable = () => {
    setConfirmDialog({ open: false, agent: null });
  };

  const handleConfirmDisable = () => {
    if (confirmDialog.agent) {
      disableAgent(confirmDialog.agent.id);
    }
  };

  return (
    <>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4 auto-rows-fr">
        {agents.map((agent) => (
          <CompactAgentCard
            key={agent.id}
            agent={agent}
            isEnabled={enabledAgentIds.has(agent.id)}
            onToggle={handleToggle}
            onConfigure={handleConfigure}
            isLoading={loadingAgentId === agent.id}
          />
        ))}
      </div>

      {/* Confirmation Dialog */}
      <Dialog open={confirmDialog.open} onOpenChange={(open) => !open && handleCancelDisable()}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle className="text-lg font-semibold">
              Disable {confirmDialog.agent?.name}?
            </DialogTitle>
          </DialogHeader>

          <DialogDescription className="text-sm text-gray-600">
            This will remove access for all teams and employees. Configurations will be preserved and can
            be restored.
          </DialogDescription>

          <DialogFooter className="gap-2">
            <Button variant="ghost" onClick={handleCancelDisable}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleConfirmDisable}>
              Disable Agent
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
