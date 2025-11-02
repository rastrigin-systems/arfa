'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { useToast } from '@/components/ui/use-toast';

// Dynamically import Monaco Editor to avoid SSR issues
const MonacoEditor = dynamic(() => import('@monaco-editor/react'), { ssr: false });

type Agent = {
  id: string;
  name: string;
  type: string;
  provider: string;
  default_config?: Record<string, unknown>;
};

type OrgAgentConfig = {
  id: string;
  agent_id: string;
  config: Record<string, unknown>;
  is_enabled: boolean;
};

type ConfigEditorModalProps = {
  agent: Agent;
  existingConfig: OrgAgentConfig | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSuccess: () => void;
  mode?: 'org' | 'team'; // Defaults to 'org'
  teamId?: string; // Required when mode is 'team'
};

export function ConfigEditorModal({ agent, existingConfig, open, onOpenChange, onSuccess, mode = 'org', teamId }: ConfigEditorModalProps) {
  const { toast } = useToast();

  // Initialize config value
  const initialConfig = existingConfig?.config || agent.default_config || {};
  const [configValue, setConfigValue] = useState(JSON.stringify(initialConfig, null, 2));
  const [validationError, setValidationError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const isEditing = existingConfig !== null;

  const handleSubmit = async () => {
    // Validate JSON
    let parsedConfig: Record<string, unknown>;
    try {
      parsedConfig = JSON.parse(configValue);
    } catch {
      setValidationError('Invalid JSON syntax. Please fix the configuration.');
      return;
    }

    setValidationError(null);
    setIsSubmitting(true);

    try {
      let url: string;
      let body: Record<string, unknown>;
      const method = isEditing ? 'PATCH' : 'POST';

      if (mode === 'team') {
        if (!teamId) {
          throw new Error('Team ID is required for team mode');
        }
        url = isEditing
          ? `/api/v1/teams/${teamId}/agent-configs/${existingConfig.id}`
          : `/api/v1/teams/${teamId}/agent-configs`;

        body = isEditing
          ? { config_override: parsedConfig, is_enabled: true }
          : { agent_id: agent.id, config_override: parsedConfig, is_enabled: true };
      } else {
        url = isEditing
          ? `/api/v1/organizations/current/agent-configs/${existingConfig.id}`
          : '/api/v1/organizations/current/agent-configs';

        body = isEditing
          ? { config: parsedConfig, is_enabled: true }
          : { agent_id: agent.id, config: parsedConfig, is_enabled: true };
      }

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to save configuration');
      }

      toast({
        title: 'Success',
        description: `Configuration ${isEditing ? 'updated' : 'created'} successfully`,
        variant: 'success',
      });

      onSuccess();
      onOpenChange(false);
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
      setValidationError(`Failed to save configuration: ${errorMessage}`);
      toast({
        title: 'Error',
        description: errorMessage,
        variant: 'destructive',
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-4xl max-h-[80vh] flex flex-col" aria-labelledby="dialog-title" aria-describedby="dialog-description">
        <DialogHeader>
          <DialogTitle id="dialog-title">Configure {agent.name}</DialogTitle>
          <DialogDescription id="dialog-description">
            {mode === 'team'
              ? isEditing
                ? `Update the team-level configuration override for ${agent.name}. This will override the organization-level settings for this team.`
                : `Create a team-level configuration override for ${agent.name}. This will customize the agent for this specific team.`
              : isEditing
                ? `Update the configuration for ${agent.name}. Changes will apply to all employees in your organization.`
                : `Create a new configuration for ${agent.name}. This will make the agent available to all employees.`}
          </DialogDescription>
        </DialogHeader>

        <div className="flex-1 min-h-0">
          <div className="h-[400px] border rounded-md overflow-hidden">
            <MonacoEditor
              height="100%"
              defaultLanguage="json"
              theme="vs-dark"
              value={configValue}
              onChange={(value) => {
                setConfigValue(value || '');
                setValidationError(null);
              }}
              options={{
                minimap: { enabled: false },
                scrollBeyondLastLine: false,
                fontSize: 14,
                tabSize: 2,
                formatOnPaste: true,
                formatOnType: true,
              }}
            />
          </div>

          {validationError && (
            <div className="mt-2 text-sm text-destructive" role="alert">
              {validationError}
            </div>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={isSubmitting}>
            Cancel
          </Button>
          <Button onClick={handleSubmit} disabled={isSubmitting}>
            {isSubmitting ? 'Saving...' : isEditing ? 'Update Configuration' : 'Create Configuration'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
