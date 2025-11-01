'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Edit, Trash2, Power, PowerOff } from 'lucide-react';

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

type OrgAgentConfigTableProps = {
  configs: OrgAgentConfig[];
  onEdit: (config: OrgAgentConfig) => void;
  onDelete: (config: OrgAgentConfig) => void;
  onToggleEnabled?: (config: OrgAgentConfig) => void;
};

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
      <div>
        <h3 className="text-lg font-semibold">No agent configurations</h3>
        <p className="text-sm text-muted-foreground">
          You haven&apos;t configured any agents yet. Enable agents to make them available to your organization.
        </p>
      </div>
    </div>
  );
}

export function OrgAgentConfigTable({ configs, onEdit, onDelete, onToggleEnabled }: OrgAgentConfigTableProps) {
  if (configs.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Agent</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Provider</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Status</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Last Updated</th>
            <th className="h-12 px-4 text-right align-middle font-medium text-muted-foreground">Actions</th>
          </tr>
        </thead>
        <tbody>
          {configs.map((config) => (
            <tr key={config.id} className="border-b transition-colors hover:bg-muted/50">
              <td className="p-4 align-middle">
                <div className="font-medium">{config.agent_name || 'Unknown Agent'}</div>
                <div className="text-sm text-muted-foreground">{config.agent_type}</div>
              </td>
              <td className="p-4 align-middle">
                <div className="text-sm">{config.agent_provider}</div>
              </td>
              <td className="p-4 align-middle">
                <Badge variant={config.is_enabled ? 'default' : 'outline'} data-testid="status-badge">
                  {config.is_enabled ? 'Enabled' : 'Disabled'}
                </Badge>
              </td>
              <td className="p-4 align-middle">
                <div className="text-sm text-muted-foreground">
                  {config.updated_at ? new Date(config.updated_at).toLocaleDateString() : 'N/A'}
                </div>
              </td>
              <td className="p-4 align-middle text-right">
                <div className="flex justify-end gap-2">
                  <Button variant="ghost" size="sm" onClick={() => onEdit(config)} aria-label="Edit configuration">
                    <Edit className="h-4 w-4" />
                    <span className="ml-2">Edit</span>
                  </Button>
                  {onToggleEnabled && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onToggleEnabled(config)}
                      aria-label={config.is_enabled ? 'Disable agent' : 'Enable agent'}
                    >
                      {config.is_enabled ? <PowerOff className="h-4 w-4" /> : <Power className="h-4 w-4" />}
                      <span className="ml-2">{config.is_enabled ? 'Disable' : 'Enable'}</span>
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(config)}
                    aria-label="Delete configuration"
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                    <span className="ml-2">Delete</span>
                  </Button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
