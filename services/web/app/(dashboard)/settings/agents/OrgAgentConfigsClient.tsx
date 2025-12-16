'use client';

import { useState, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { OrgAgentConfigTable } from '@/components/agents/OrgAgentConfigTable';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import { TeamAgentConfigsTab } from '@/components/agents/TeamAgentConfigsTab';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { AgentCatalog } from '@/components/agents/AgentCatalog';
import { clientDeleteOrgAgentConfig, type OrgAgentConfig } from '@/lib/api/org-configs';
import type { Agent } from '@/lib/types';

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
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; config: OrgAgentConfig | null }>({
    open: false,
    config: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  // Build set of enabled agent IDs
  const enabledAgentIds = useMemo(() => new Set(configs.map((config) => config.agent_id)), [configs]);

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

  const handleDeleteConfig = (config: OrgAgentConfig) => {
    setDeleteConfirm({ open: true, config });
  };

  const confirmDelete = async () => {
    if (!deleteConfirm.config) return;

    setIsDeleting(true);
    try {
      await clientDeleteOrgAgentConfig(deleteConfirm.config.id);

      toast({
        title: 'Success',
        description: `Configuration for ${deleteConfirm.config.agent_name} deleted successfully`,
        variant: 'success',
      });

      setDeleteConfirm({ open: false, config: null });
      router.refresh();
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsDeleting(false);
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

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteConfirm.open} onOpenChange={(open) => !open && setDeleteConfirm({ open: false, config: null })}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Delete Configuration?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            Are you sure you want to delete the configuration for{' '}
            <strong>{deleteConfirm.config?.agent_name}</strong>? This action cannot be undone.
          </DialogDescription>
          <DialogFooter className="gap-2">
            <Button
              variant="outline"
              onClick={() => setDeleteConfirm({ open: false, config: null })}
              disabled={isDeleting}
            >
              Cancel
            </Button>
            <Button variant="destructive" onClick={confirmDelete} disabled={isDeleting}>
              {isDeleting ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
