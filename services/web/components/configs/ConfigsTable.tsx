'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Power, PowerOff, Trash2, Building2, Users, User, Edit } from 'lucide-react';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';

type ConfigWithLevel = {
  id: string;
  org_id: string;
  agent_id: string;
  agent_name?: string;
  agent_type?: string;
  config: Record<string, unknown>;
  is_enabled: boolean;
  level: 'organization' | 'team' | 'employee';
  assigned_to: string;
};

type ConfigsTableProps = {
  configs: ConfigWithLevel[];
  onEdit: (config: ConfigWithLevel) => void;
  onToggleEnabled: (config: ConfigWithLevel) => void;
  onDelete: (config: ConfigWithLevel) => void;
};

function LevelIcon({ level }: { level: string }) {
  switch (level) {
    case 'organization':
      return <Building2 className="h-4 w-4" />;
    case 'team':
      return <Users className="h-4 w-4" />;
    case 'employee':
      return <User className="h-4 w-4" />;
    default:
      return null;
  }
}

function ConfigPreview({ config }: { config: Record<string, unknown> }) {
  const entries = Object.entries(config).slice(0, 3);
  if (entries.length === 0) {
    return <span className="text-muted-foreground">No configuration</span>;
  }

  return (
    <div className="space-y-1">
      {entries.map(([key, value]) => (
        <div key={key} className="text-sm">
          <span className="font-medium">{key}:</span>{' '}
          <span className="text-muted-foreground">
            {typeof value === 'object' ? JSON.stringify(value).slice(0, 30) + '...' : String(value)}
          </span>
        </div>
      ))}
      {Object.keys(config).length > 3 && (
        <div className="text-xs text-muted-foreground">+{Object.keys(config).length - 3} more</div>
      )}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex flex-col items-center justify-center gap-4 rounded-lg border border-dashed p-12 text-center">
      <div>
        <h3 className="text-lg font-semibold">No configurations found</h3>
        <p className="text-sm text-muted-foreground">
          Try adjusting your filters or create a new configuration to get started.
        </p>
      </div>
    </div>
  );
}

export function ConfigsTable({ configs, onEdit, onToggleEnabled, onDelete }: ConfigsTableProps) {
  if (configs.length === 0) {
    return <EmptyState />;
  }

  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Agent</TableHead>
            <TableHead>Assigned To</TableHead>
            <TableHead>Configuration</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Actions</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {configs.map((config) => (
            <TableRow key={config.id}>
              <TableCell>
                <div className="font-medium">{config.agent_name || 'Unknown Agent'}</div>
                <div className="text-sm text-muted-foreground">{config.agent_type}</div>
              </TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <LevelIcon level={config.level} />
                  <div>
                    <div className="font-medium capitalize">{config.level}</div>
                    <div className="text-sm text-muted-foreground">{config.assigned_to}</div>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <ConfigPreview config={config.config} />
              </TableCell>
              <TableCell>
                <Badge variant={config.is_enabled ? 'default' : 'outline'}>
                  {config.is_enabled ? 'Enabled' : 'Disabled'}
                </Badge>
              </TableCell>
              <TableCell className="text-right">
                <div className="flex justify-end gap-2">
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onEdit(config)}
                    aria-label="Edit configuration"
                  >
                    <Edit className="h-4 w-4" />
                    <span className="ml-2">Edit</span>
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => onToggleEnabled(config)}
                    aria-label={config.is_enabled ? 'Disable configuration' : 'Enable configuration'}
                  >
                    {config.is_enabled ? <PowerOff className="h-4 w-4" /> : <Power className="h-4 w-4" />}
                    <span className="ml-2">{config.is_enabled ? 'Disable' : 'Enable'}</span>
                  </Button>
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
              </TableCell>            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
