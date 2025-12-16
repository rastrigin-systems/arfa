'use client';

import { useState, useEffect, useMemo } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';
import {
  clientListTeamAgentConfigs,
  clientCreateTeamAgentConfig,
  clientDeleteTeamAgentConfig,
  type TeamAgentConfig,
} from '@/lib/api/team-configs';
import type { OrgAgentConfig } from '@/lib/api/org-configs';
import type { Agent, Team } from '@/lib/types';

interface TeamAgentConfigsTabProps {
  orgConfigs: OrgAgentConfig[];
  agents: Agent[];
}

// Extended TeamAgentConfig with additional display fields
type TeamAgentConfigWithDisplay = TeamAgentConfig & {
  agent_name?: string;
  agent_type?: string;
  agent_provider?: string;
  team_name?: string;
};

export function TeamAgentConfigsTab({ orgConfigs, agents }: TeamAgentConfigsTabProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeamId, setSelectedTeamId] = useState<string>('all');
  const [teamConfigs, setTeamConfigs] = useState<TeamAgentConfigWithDisplay[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAssigning, setIsAssigning] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [editingConfig, setEditingConfig] = useState<TeamAgentConfigWithDisplay | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState<{ open: boolean; config: TeamAgentConfigWithDisplay | null }>({
    open: false,
    config: null,
  });
  const [isDeleting, setIsDeleting] = useState(false);

  // Fetch teams on mount using Next.js API route
  useEffect(() => {
    const fetchTeams = async () => {
      try {
        const response = await fetch('/api/teams', { credentials: 'include' });
        if (!response.ok) throw new Error('Failed to fetch teams');
        const data = await response.json();
        setTeams(data?.teams || []);
      } catch {
        toast({
          title: 'Error',
          description: 'Failed to load teams',
          variant: 'destructive',
        });
      }
    };

    fetchTeams();
  }, [toast]);

  // Fetch team agent configs when team selection changes
  useEffect(() => {
    const fetchTeamConfigs = async () => {
      if (selectedTeamId === 'all') {
        setIsLoading(false);
        setTeamConfigs([]);
        return;
      }

      setIsLoading(true);
      try {
        const configs = await clientListTeamAgentConfigs(selectedTeamId);
        // Enrich with agent display info
        const enrichedConfigs: TeamAgentConfigWithDisplay[] = configs.map((config) => {
          const agent = agents.find((a) => a.id === config.agent_id);
          return {
            ...config,
            agent_name: agent?.name,
            agent_type: agent?.type,
            agent_provider: agent?.provider,
          };
        });
        setTeamConfigs(enrichedConfigs);
      } catch {
        toast({
          title: 'Error',
          description: 'Failed to load team agent configurations',
          variant: 'destructive',
        });
      } finally {
        setIsLoading(false);
      }
    };

    fetchTeamConfigs();
  }, [selectedTeamId, agents, toast]);

  const handleAssignToTeam = async (orgConfigId: string) => {
    if (selectedTeamId === 'all') {
      toast({
        title: 'Select a team',
        description: 'Please select a team first to assign this agent configuration',
        variant: 'destructive',
      });
      return;
    }

    const orgConfig = orgConfigs.find((c) => c.id === orgConfigId);
    const agent = agents.find((a) => a.id === orgConfig?.agent_id);

    if (!orgConfig || !agent) return;

    setIsAssigning(true);
    try {
      await clientCreateTeamAgentConfig(selectedTeamId, {
        agent_id: orgConfig.agent_id,
        config_override: {},
        is_enabled: true,
      });

      toast({
        title: 'Success',
        description: `${agent.name} assigned to team successfully`,
      });

      // Refresh team configs
      const configs = await clientListTeamAgentConfigs(selectedTeamId);
      const enrichedConfigs: TeamAgentConfigWithDisplay[] = configs.map((config) => {
        const a = agents.find((ag) => ag.id === config.agent_id);
        return {
          ...config,
          agent_name: a?.name,
          agent_type: a?.type,
          agent_provider: a?.provider,
        };
      });
      setTeamConfigs(enrichedConfigs);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsAssigning(false);
    }
  };

  const handleEditTeamConfig = (config: TeamAgentConfigWithDisplay) => {
    const agent = agents.find((a) => a.id === config.agent_id);
    if (agent) {
      setSelectedAgent(agent);
      setEditingConfig(config);
      setIsModalOpen(true);
    }
  };

  const handleDeleteTeamConfig = (config: TeamAgentConfigWithDisplay) => {
    setDeleteConfirm({ open: true, config });
  };

  const confirmDelete = async () => {
    if (!deleteConfirm.config) return;

    setIsDeleting(true);
    try {
      await clientDeleteTeamAgentConfig(deleteConfirm.config.team_id, deleteConfirm.config.id);

      toast({
        title: 'Success',
        description: `${deleteConfirm.config.agent_name} removed from team successfully`,
      });

      setTeamConfigs(teamConfigs.filter((c) => c.id !== deleteConfirm.config!.id));
      setDeleteConfirm({ open: false, config: null });
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

  const handleModalSuccess = async () => {
    // Refresh team configs after editing
    if (selectedTeamId !== 'all') {
      const configs = await clientListTeamAgentConfigs(selectedTeamId);
      const enrichedConfigs: TeamAgentConfigWithDisplay[] = configs.map((config) => {
        const agent = agents.find((a) => a.id === config.agent_id);
        return {
          ...config,
          agent_name: agent?.name,
          agent_type: agent?.type,
          agent_provider: agent?.provider,
        };
      });
      setTeamConfigs(enrichedConfigs);
    }
    router.refresh();
  };

  // Get assigned agent IDs for selected team
  const assignedAgentIds = useMemo(() => new Set(teamConfigs.map((c) => c.agent_id)), [teamConfigs]);

  return (
    <div className="space-y-6">
      {/* Team Selection */}
      <Card>
        <CardHeader>
          <CardTitle>Team Selection</CardTitle>
          <CardDescription>Select a team to view and manage their agent configurations</CardDescription>
        </CardHeader>
        <CardContent>
          <Select value={selectedTeamId} onValueChange={setSelectedTeamId}>
            <SelectTrigger className="w-[300px]" aria-label="Select team">
              <SelectValue placeholder="Select a team" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Teams</SelectItem>
              {teams.map((team) => (
                <SelectItem key={team.id} value={team.id}>
                  {team.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </CardContent>
      </Card>

      {/* Available Org Configs to Assign */}
      {selectedTeamId !== 'all' && (
        <Card>
          <CardHeader>
            <CardTitle>Assign Organization Agents to Team</CardTitle>
            <CardDescription>
              Select agents configured at the organization level to assign to this team
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-2">
              {orgConfigs.length === 0 ? (
                <p className="text-sm text-muted-foreground">No organization agent configurations available</p>
              ) : (
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Agent</TableHead>
                      <TableHead>Provider</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Actions</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {orgConfigs.map((config) => {
                      const isAssigned = assignedAgentIds.has(config.agent_id);
                      return (
                        <TableRow key={config.id}>
                          <TableCell className="font-medium">{config.agent_name}</TableCell>
                          <TableCell>{config.agent_provider}</TableCell>
                          <TableCell>
                            {isAssigned ? (
                              <Badge variant="success">Assigned</Badge>
                            ) : (
                              <Badge variant="secondary">Not Assigned</Badge>
                            )}
                          </TableCell>
                          <TableCell>
                            {!isAssigned && (
                              <Button
                                size="sm"
                                onClick={() => handleAssignToTeam(config.id)}
                                disabled={isAssigning}
                              >
                                Assign to Team
                              </Button>
                            )}
                          </TableCell>
                        </TableRow>
                      );
                    })}
                  </TableBody>
                </Table>
              )}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Team Agent Configs */}
      {selectedTeamId !== 'all' && (
        <Card>
          <CardHeader>
            <CardTitle>Team Agent Configurations</CardTitle>
            <CardDescription>Manage agents assigned to this team with custom overrides</CardDescription>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <p className="text-sm text-muted-foreground">Loading team configurations...</p>
            ) : teamConfigs.length === 0 ? (
              <p className="text-sm text-muted-foreground">No agents assigned to this team yet</p>
            ) : (
              <Table>
                <TableCaption>Showing {teamConfigs.length} agent configuration(s) for this team</TableCaption>
                <TableHeader>
                  <TableRow>
                    <TableHead>Agent</TableHead>
                    <TableHead>Provider</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Actions</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {teamConfigs.map((config) => (
                    <TableRow key={config.id}>
                      <TableCell className="font-medium">{config.agent_name}</TableCell>
                      <TableCell>{config.agent_provider}</TableCell>
                      <TableCell>
                        <Badge variant={config.is_enabled ? 'success' : 'secondary'}>
                          {config.is_enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                      </TableCell>
                      <TableCell className="space-x-2">
                        <Button size="sm" variant="outline" onClick={() => handleEditTeamConfig(config)}>
                          Edit Override
                        </Button>
                        <Button
                          size="sm"
                          variant="destructive"
                          onClick={() => handleDeleteTeamConfig(config)}
                        >
                          Remove
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      )}

      {/* Config Editor Modal */}
      {selectedAgent && editingConfig && (
        <ConfigEditorModal
          agent={selectedAgent}
          existingConfig={{
            id: editingConfig.id,
            agent_id: editingConfig.agent_id,
            config: editingConfig.config_override,
            is_enabled: editingConfig.is_enabled,
          }}
          open={isModalOpen}
          onOpenChange={setIsModalOpen}
          onSuccess={handleModalSuccess}
          mode="team"
          teamId={selectedTeamId !== 'all' ? selectedTeamId : undefined}
        />
      )}

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteConfirm.open} onOpenChange={(open) => !open && setDeleteConfirm({ open: false, config: null })}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Remove Agent from Team?</DialogTitle>
          </DialogHeader>
          <DialogDescription>
            Are you sure you want to remove <strong>{deleteConfirm.config?.agent_name}</strong> from this team?
            The team will inherit the organization-level configuration.
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
              {isDeleting ? 'Removing...' : 'Remove'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
