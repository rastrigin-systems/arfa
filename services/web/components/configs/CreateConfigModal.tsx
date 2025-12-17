'use client';

import { useState, useEffect } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/components/ui/use-toast';
import { Loader2 } from 'lucide-react';
import { useTeams } from '@/lib/hooks/useTeams';
import { useEmployees } from '@/lib/hooks/useEmployees';
import type { Agent, AgentConfig } from '@/lib/types';

type ConfigWithLevel = {
  id: string;
  org_id: string;
  agent_id: string;
  agent_name?: string;
  agent_type?: string;
  config: AgentConfig;
  is_enabled: boolean;
  level: 'organization' | 'team' | 'employee';
  assigned_to: string;
};

type CreateConfigModalProps = {
  agents: Agent[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  editingConfig?: ConfigWithLevel | null;
};

export function CreateConfigModal({ agents, open, onOpenChange, onSuccess, editingConfig }: CreateConfigModalProps) {
  const [selectedAgentId, setSelectedAgentId] = useState('');
  const [assignTo, setAssignTo] = useState<'organization' | 'team' | 'employee'>('organization');
  const [selectedTeamId, setSelectedTeamId] = useState<string>('');
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>('');
  const [configJson, setConfigJson] = useState('{}');
  const [isEnabled, setIsEnabled] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  // Fetch teams and employees
  const { data: teams, isLoading: teamsLoading } = useTeams();
  const { data: employeesData, isLoading: employeesLoading } = useEmployees({
    page: 1,
    limit: 100, // Get all employees for selection
    status: 'active',
  });

  const selectedAgent = agents.find((a) => a.id === selectedAgentId);
  const isEditMode = !!editingConfig;

  // Initialize form when editing
  useEffect(() => {
    if (editingConfig && open) {
      setSelectedAgentId(editingConfig.agent_id);
      setAssignTo(editingConfig.level);
      setConfigJson(JSON.stringify(editingConfig.config, null, 2));
      setIsEnabled(editingConfig.is_enabled);
    } else if (!open) {
      // Reset form when closing
      setSelectedAgentId('');
      setAssignTo('organization');
      setSelectedTeamId('');
      setSelectedEmployeeId('');
      setConfigJson('{}');
      setIsEnabled(true);
    }
  }, [editingConfig, open]);

  const handleSubmit = async () => {
    if (!selectedAgentId) {
      toast({
        title: 'Validation Error',
        description: 'Please select an agent',
        variant: 'destructive',
      });
      return;
    }

    // Validate team/employee selection
    if (assignTo === 'team' && !selectedTeamId) {
      toast({
        title: 'Validation Error',
        description: 'Please select a team',
        variant: 'destructive',
      });
      return;
    }

    if (assignTo === 'employee' && !selectedEmployeeId) {
      toast({
        title: 'Validation Error',
        description: 'Please select an employee',
        variant: 'destructive',
      });
      return;
    }

    // Validate JSON
    let config: AgentConfig;
    try {
      config = JSON.parse(configJson);
    } catch {
      toast({
        title: 'Invalid JSON',
        description: 'Please enter valid JSON configuration',
        variant: 'destructive',
      });
      return;
    }

    setIsSubmitting(true);

    try {
      // Determine the correct API endpoint based on assignment level
      let url: string;
      let method = 'POST';

      if (isEditMode && editingConfig) {
        // Edit mode - use the appropriate endpoint based on level
        if (editingConfig.level === 'team') {
          url = `/api/teams/${editingConfig.assigned_to}/agent-configs/${editingConfig.id}`;
        } else if (editingConfig.level === 'employee') {
          url = `/api/employees/${editingConfig.assigned_to}/agent-configs/${editingConfig.id}`;
        } else {
          url = `/api/organizations/current/agent-configs/${editingConfig.id}`;
        }
        method = 'PATCH';
      } else {
        // Create mode - use the appropriate endpoint based on assignTo
        if (assignTo === 'team') {
          url = `/api/teams/${selectedTeamId}/agent-configs`;
        } else if (assignTo === 'employee') {
          url = `/api/employees/${selectedEmployeeId}/agent-configs`;
        } else {
          url = '/api/organizations/current/agent-configs';
        }
      }

      // Prepare request body - config_override for team/employee, config for org
      const body: Record<string, unknown> = {
        agent_id: selectedAgentId,
        is_enabled: isEnabled,
      };

      // Use config_override for team/employee level, config for org level
      if (assignTo === 'organization') {
        body.config = config;
      } else {
        body.config_override = config;
      }

      const response = await fetch(url, {
        method,
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `Failed to ${isEditMode ? 'update' : 'create'} configuration`);
      }

      toast({
        title: 'Success',
        description: `Configuration ${isEditMode ? 'updated' : 'created'} successfully`,
        variant: 'success',
      });

      onSuccess();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : `Failed to ${isEditMode ? 'update' : 'create'} configuration`,
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAgentChange = (agentId: string) => {
    setSelectedAgentId(agentId);
    // Only set default config if not editing
    if (!isEditMode) {
      const agent = agents.find((a) => a.id === agentId);
      if (agent?.default_config) {
        setConfigJson(JSON.stringify(agent.default_config, null, 2));
      } else {
        setConfigJson('{}');
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEditMode ? 'Edit Configuration' : 'Create New Configuration'}</DialogTitle>
          <DialogDescription>
            Configure an AI agent for your organization, team, or individual employee
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Agent Selection */}
          <div className="space-y-2">
            <Label htmlFor="agent">Select Agent *</Label>
            <Select value={selectedAgentId} onValueChange={handleAgentChange} disabled={isEditMode}>
              <SelectTrigger id="agent">
                <SelectValue placeholder="Choose an agent" />
              </SelectTrigger>
              <SelectContent>
                {agents.map((agent) => (
                  <SelectItem key={agent.id} value={agent.id}>
                    <div>
                      <div className="font-medium">{agent.name}</div>
                      <div className="text-xs text-muted-foreground">{agent.type}</div>
                    </div>
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {selectedAgent && (
              <p className="text-sm text-muted-foreground">{selectedAgent.description}</p>
            )}
            {isEditMode && (
              <p className="text-xs text-muted-foreground">Agent cannot be changed when editing</p>
            )}
          </div>

          {/* Assignment Level */}
          <div className="space-y-2">
            <Label>Assign To *</Label>
            <RadioGroup value={assignTo} onValueChange={(value) => setAssignTo(value as 'organization' | 'team' | 'employee')} disabled={isEditMode}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="organization" id="org" />
                <Label htmlFor="org" className="font-normal">
                  Organization (applies to all employees)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="team" id="team" />
                <Label htmlFor="team" className="font-normal">
                  Specific Team
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="employee" id="employee" />
                <Label htmlFor="employee" className="font-normal">
                  Individual Employee
                </Label>
              </div>
            </RadioGroup>
            {isEditMode && (
              <p className="text-xs text-muted-foreground">Assignment level cannot be changed when editing</p>
            )}
          </div>

          {/* Team Selection (when assignTo === 'team') */}
          {assignTo === 'team' && !isEditMode && (
            <div className="space-y-2">
              <Label htmlFor="team-select">Select Team *</Label>
              <Select value={selectedTeamId} onValueChange={setSelectedTeamId} disabled={teamsLoading}>
                <SelectTrigger id="team-select">
                  <SelectValue placeholder={teamsLoading ? 'Loading teams...' : 'Choose a team'} />
                </SelectTrigger>
                <SelectContent>
                  {teamsLoading ? (
                    <SelectItem value="loading" disabled>
                      <div className="flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Loading teams...
                      </div>
                    </SelectItem>
                  ) : teams && teams.length > 0 ? (
                    teams.map((team) => (
                      <SelectItem key={team.id} value={team.id}>
                        {team.name}
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value="no-teams" disabled>
                      No teams available
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Employee Selection (when assignTo === 'employee') */}
          {assignTo === 'employee' && !isEditMode && (
            <div className="space-y-2">
              <Label htmlFor="employee-select">Select Employee *</Label>
              <Select value={selectedEmployeeId} onValueChange={setSelectedEmployeeId} disabled={employeesLoading}>
                <SelectTrigger id="employee-select">
                  <SelectValue placeholder={employeesLoading ? 'Loading employees...' : 'Choose an employee'} />
                </SelectTrigger>
                <SelectContent>
                  {employeesLoading ? (
                    <SelectItem value="loading" disabled>
                      <div className="flex items-center gap-2">
                        <Loader2 className="h-4 w-4 animate-spin" />
                        Loading employees...
                      </div>
                    </SelectItem>
                  ) : employeesData?.employees && employeesData.employees.length > 0 ? (
                    employeesData.employees.map((employee) => (
                      <SelectItem key={employee.id} value={employee.id}>
                        <div>
                          <div>{employee.full_name}</div>
                          <div className="text-xs text-muted-foreground">{employee.email}</div>
                        </div>
                      </SelectItem>
                    ))
                  ) : (
                    <SelectItem value="no-employees" disabled>
                      No employees available
                    </SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Configuration JSON */}
          <div className="space-y-2">
            <Label htmlFor="config">Configuration (JSON)</Label>
            <Textarea
              id="config"
              value={configJson}
              onChange={(e) => setConfigJson(e.target.value)}
              rows={8}
              className="font-mono text-sm"
              placeholder='{"key": "value"}'
            />
            <p className="text-xs text-muted-foreground">
              Enter configuration as JSON. Default values are pre-filled based on the selected agent.
            </p>
          </div>

          {/* Enabled Toggle */}
          <div className="flex items-center justify-between rounded-lg border p-4">
            <div className="space-y-0.5">
              <Label htmlFor="enabled">Enable Configuration</Label>
              <p className="text-sm text-muted-foreground">
                Make this configuration active immediately
              </p>
            </div>
            <Switch id="enabled" checked={isEnabled} onCheckedChange={setIsEnabled} />
          </div>
        </div>

        {/* Actions */}
        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            {isEditMode ? 'Update Configuration' : 'Create Configuration'}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
