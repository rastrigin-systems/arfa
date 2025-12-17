'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Edit, Trash2, Power, PowerOff } from 'lucide-react';
import type { AgentConfig } from '@/lib/types';

export type EmployeeAgentOverride = {
  id: string;
  employee_id: string;
  agent_id: string;
  agent_name?: string;
  agent_type?: string;
  agent_provider?: string;
  config_override: AgentConfig;
  override_reason?: string;
  is_enabled: boolean;
  created_at?: string;
  updated_at?: string;
  updated_by?: string;
  org_config?: AgentConfig;
  team_config?: AgentConfig;
};

type EmployeeAgentOverridesTableProps = {
  overrides: EmployeeAgentOverride[];
  onEdit: (override: EmployeeAgentOverride) => void;
  onDelete: (override: EmployeeAgentOverride) => void;
  onToggleEnabled?: (override: EmployeeAgentOverride) => void;
};

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
      <div>
        <h3 className="text-lg font-semibold">No employee overrides</h3>
        <p className="text-sm text-muted-foreground">
          Employee is using default organization and team configurations.
        </p>
      </div>
    </div>
  );
}

function formatDate(dateString?: string | null): string {
  if (!dateString) return 'N/A';
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

function formatConfigValue(key: string, value: unknown): string {
  if (key === 'rate_limit' && typeof value === 'number') {
    return `${value} req/day`;
  }
  if (key === 'cost_limit' && typeof value === 'number') {
    return `$${value}/month`;
  }
  return String(value);
}

function renderConfigComparison(
  key: string,
  overrideValue: unknown,
  orgValue: unknown | undefined
): JSX.Element | null {
  if (orgValue === undefined) return null;

  const overrideStr = formatConfigValue(key, overrideValue);
  const orgStr = formatConfigValue(key, orgValue);

  if (overrideStr === orgStr) return null;

  return (
    <div className="text-xs text-muted-foreground">
      (org: {orgStr})
    </div>
  );
}

export function EmployeeAgentOverridesTable({
  overrides,
  onEdit,
  onDelete,
  onToggleEnabled,
}: EmployeeAgentOverridesTableProps) {
  if (overrides.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead>
          <tr className="border-b bg-muted/50">
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Agent</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Override Reason</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Config</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Status</th>
            <th className="h-12 px-4 text-left align-middle font-medium text-muted-foreground">Last Updated</th>
            <th className="h-12 px-4 text-right align-middle font-medium text-muted-foreground">Actions</th>
          </tr>
        </thead>
        <tbody>
          {overrides.map((override) => (
            <tr key={override.id} className="border-b transition-colors hover:bg-muted/50">
              <td className="p-4 align-middle">
                <div className="font-medium">{override.agent_name || 'Unknown Agent'}</div>
                <div className="text-sm text-muted-foreground">{override.agent_type}</div>
                <div className="text-sm text-muted-foreground">{override.agent_provider}</div>
              </td>
              <td className="p-4 align-middle">
                <div className="text-sm max-w-xs truncate" title={override.override_reason}>
                  {override.override_reason || 'No reason provided'}
                </div>
              </td>
              <td className="p-4 align-middle">
                <div className="text-sm space-y-1">
                  {Object.entries(override.config_override).map(([key, value]) => (
                    <div key={key}>
                      <span className="font-medium">{formatConfigValue(key, value)}</span>
                      {override.org_config && renderConfigComparison(key, value, override.org_config[key])}
                    </div>
                  ))}
                </div>
              </td>
              <td className="p-4 align-middle">
                <Badge variant={override.is_enabled ? 'default' : 'outline'} data-testid="status-badge">
                  {override.is_enabled ? 'Enabled' : 'Disabled'}
                </Badge>
              </td>
              <td className="p-4 align-middle">
                <div className="text-sm text-muted-foreground">
                  {formatDate(override.updated_at)}
                </div>
                {override.updated_by && (
                  <div className="text-xs text-muted-foreground">
                    by {override.updated_by}
                  </div>
                )}
              </td>
              <td className="p-4 align-middle text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEdit(override)}
                    aria-label="Edit configuration"
                  >
                    <Edit className="h-4 w-4" />
                    <span className="ml-2">Edit</span>
                  </Button>
                  {onToggleEnabled && (
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => onToggleEnabled(override)}
                      aria-label={override.is_enabled ? 'Disable agent' : 'Enable agent'}
                    >
                      {override.is_enabled ? <PowerOff className="h-4 w-4" /> : <Power className="h-4 w-4" />}
                      <span className="ml-2">{override.is_enabled ? 'Disable' : 'Enable'}</span>
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onDelete(override)}
                    aria-label="Remove configuration"
                    className="text-destructive hover:text-destructive"
                  >
                    <Trash2 className="h-4 w-4" />
                    <span className="ml-2">Remove</span>
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
