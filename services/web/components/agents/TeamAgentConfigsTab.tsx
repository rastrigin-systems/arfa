'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';
import { Badge } from '@/components/ui/badge';
import { apiClient } from '@/lib/api/client';
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
import { ConfigEditorModal } from '@/components/agents/ConfigEditorModal';

type Team = {
  id: string;
  name: string;
  description?: string;
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
};

type TeamAgentConfig = {
  id: string;
  team_id: string;
  agent_id: string;
  agent_name?: string;
  agent_type?: string;
  agent_provider?: string;
  config_override: Record<string, unknown>;
  is_enabled: boolean;
  team_name?: string;
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

interface TeamAgentConfigsTabProps {
  orgConfigs: OrgAgentConfig[];
  agents: Agent[];
}

export function TeamAgentConfigsTab({ orgConfigs, agents }: TeamAgentConfigsTabProps) {
  const router = useRouter();
  const { toast } = useToast();
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeamId, setSelectedTeamId] = useState<string>('all');
  const [teamConfigs, setTeamConfigs] = useState<TeamAgentConfig[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isAssigning, setIsAssigning] = useState(false);
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [editingConfig, setEditingConfig] = useState<TeamAgentConfig | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  // Fetch teams on mount
  useEffect(() => {
    const fetchTeams = async () => {
      try {
        const { data, error } = await apiClient.GET('/teams');
        if (error) throw new Error('Failed to fetch teams');

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setTeams((data as any)?.teams || []);
      } catch (error) {
        console.error('Error fetching teams:', error);
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
        const { data, error } = await apiClient.GET('/teams/{team_id}/agent-configs', {
          params: { path: { team_id: selectedTeamId } },
        });
        if (error) throw new Error('Failed to fetch team agent configs');

        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setTeamConfigs((data as any)?.configs || []);
      } catch (error) {
        console.error('Error fetching team configs:', error);
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
  }, [selectedTeamId, toast]);

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
      const { error } = await apiClient.POST('/teams/{team_id}/agent-configs', {
        params: { path: { team_id: selectedTeamId } },
        body: {
          agent_id: orgConfig.agent_id,
          config_override: {}, // Start with empty override
          is_enabled: true,
        },
      });

      if (error) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        throw new Error((error as any).message || 'Failed to assign agent to team');
      }

      toast({
        title: 'Success',
        description: `${agent.name} assigned to team successfully`,
      });

      // Refresh team configs
      const { data: configsData } = await apiClient.GET('/teams/{team_id}/agent-configs', {
        params: { path: { team_id: selectedTeamId } },
      });
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setTeamConfigs((configsData as any)?.configs || []);
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

  const handleEditTeamConfig = (config: TeamAgentConfig) => {
    const agent = agents.find((a) => a.id === config.agent_id);
    if (agent) {
      setSelectedAgent(agent);
      setEditingConfig(config);
      setIsModalOpen(true);
    }
  };

  const handleDeleteTeamConfig = async (config: TeamAgentConfig) => {
    if (!confirm(`Are you sure you want to remove ${config.agent_name} from this team?`)) {
      return;
    }

    try {
      const { error } = await apiClient.DELETE('/teams/{team_id}/agent-configs/{config_id}', {
        params: { path: { team_id: config.team_id, config_id: config.id } },
      });

      if (error) {
        throw new Error('Failed to delete team agent configuration');
      }

      toast({
        title: 'Success',
        description: `${config.agent_name} removed from team successfully`,
      });

      // Refresh team configs
      setTeamConfigs(teamConfigs.filter((c) => c.id !== config.id));
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    }
  };

  const handleModalSuccess = async () => {
    // Refresh team configs after editing
    if (selectedTeamId !== 'all') {
      const { data } = await apiClient.GET('/teams/{team_id}/agent-configs', {
        params: { path: { team_id: selectedTeamId } },
      });
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setTeamConfigs((data as any)?.configs || []);
    }
    router.refresh();
  };

  // Get assigned agent IDs for selected team
  const assignedAgentIds = new Set(teamConfigs.map((c) => c.agent_id));

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
    </div>
  );
}
