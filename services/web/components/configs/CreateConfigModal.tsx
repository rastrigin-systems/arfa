'use client';

import { useState } from 'react';
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Textarea } from '@/components/ui/textarea';
import { Switch } from '@/components/ui/switch';
import { useToast } from '@/components/ui/use-toast';
import { Loader2 } from 'lucide-react';

type Agent = {
  id: string;
  name: string;
  type: string;
  description: string;
  default_config?: Record<string, unknown>;
};

type CreateConfigModalProps = {
  agents: Agent[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
};

export function CreateConfigModal({ agents, open, onOpenChange, onSuccess }: CreateConfigModalProps) {
  const [selectedAgentId, setSelectedAgentId] = useState('');
  const [assignTo, setAssignTo] = useState<'organization' | 'team' | 'employee'>('organization');
  const [configJson, setConfigJson] = useState('{}');
  const [isEnabled, setIsEnabled] = useState(true);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const { toast } = useToast();

  const selectedAgent = agents.find((a) => a.id === selectedAgentId);

  const handleSubmit = async () => {
    if (!selectedAgentId) {
      toast({
        title: 'Validation Error',
        description: 'Please select an agent',
        variant: 'destructive',
      });
      return;
    }

    // Validate JSON
    let config: Record<string, unknown>;
    try {
      config = JSON.parse(configJson);
    } catch (error) {
      toast({
        title: 'Invalid JSON',
        description: 'Please enter valid JSON configuration',
        variant: 'destructive',
      });
      return;
    }

    setIsSubmitting(true);

    try {
      // Only organization level is supported for now
      if (assignTo !== 'organization') {
        toast({
          title: 'Not Supported',
          description: 'Team and employee configurations are coming soon',
          variant: 'destructive',
        });
        setIsSubmitting(false);
        return;
      }

      const response = await fetch('/api/organizations/current/agent-configs', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({
          agent_id: selectedAgentId,
          config,
          is_enabled: isEnabled,
        }),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || 'Failed to create configuration');
      }

      toast({
        title: 'Success',
        description: 'Configuration created successfully',
        variant: 'success',
      });

      // Reset form
      setSelectedAgentId('');
      setAssignTo('organization');
      setConfigJson('{}');
      setIsEnabled(true);

      onSuccess();
    } catch (error) {
      toast({
        title: 'Error',
        description: error instanceof Error ? error.message : 'Failed to create configuration',
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAgentChange = (agentId: string) => {
    setSelectedAgentId(agentId);
    const agent = agents.find((a) => a.id === agentId);
    if (agent?.default_config) {
      setConfigJson(JSON.stringify(agent.default_config, null, 2));
    } else {
      setConfigJson('{}');
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Create New Configuration</DialogTitle>
          <DialogDescription>
            Configure an AI agent for your organization, team, or individual employee
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-6 py-4">
          {/* Agent Selection */}
          <div className="space-y-2">
            <Label htmlFor="agent">Select Agent *</Label>
            <Select value={selectedAgentId} onValueChange={handleAgentChange}>
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
          </div>

          {/* Assignment Level */}
          <div className="space-y-2">
            <Label>Assign To *</Label>
            <RadioGroup value={assignTo} onValueChange={(value: any) => setAssignTo(value)}>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="organization" id="org" />
                <Label htmlFor="org" className="font-normal">
                  Organization (applies to all employees)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="team" id="team" disabled />
                <Label htmlFor="team" className="font-normal text-muted-foreground">
                  Specific Team (coming soon)
                </Label>
              </div>
              <div className="flex items-center space-x-2">
                <RadioGroupItem value="employee" id="employee" disabled />
                <Label htmlFor="employee" className="font-normal text-muted-foreground">
                  Individual Employee (coming soon)
                </Label>
              </div>
            </RadioGroup>
          </div>

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
            Create Configuration
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
