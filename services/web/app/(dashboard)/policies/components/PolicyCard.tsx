'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Edit, Trash2, Building2, Users, User, Settings } from 'lucide-react';
import type { ToolPolicy } from '@/lib/types';

type PolicyCardProps = {
  policy: ToolPolicy;
  onEdit: () => void;
  onDelete: () => void;
};

function getScopeIcon(scope: string) {
  switch (scope) {
    case 'organization':
      return <Building2 className="h-4 w-4" />;
    case 'team':
      return <Users className="h-4 w-4" />;
    case 'employee':
      return <User className="h-4 w-4" />;
    default:
      return <Settings className="h-4 w-4" />;
  }
}

function getScopeLabel(scope: string) {
  switch (scope) {
    case 'organization':
      return 'Organization-wide';
    case 'team':
      return 'Team-level';
    case 'employee':
      return 'Employee-specific';
    default:
      return scope;
  }
}

function getConditionCount(conditions: object | null | undefined): number {
  if (!conditions) return 0;
  const cond = conditions as { any?: unknown[]; all?: unknown[] };
  if (cond.any && Array.isArray(cond.any)) return cond.any.length;
  if (cond.all && Array.isArray(cond.all)) return cond.all.length;
  return Object.keys(conditions).length;
}

export function PolicyCard({ policy, onEdit, onDelete }: PolicyCardProps) {
  const conditionCount = getConditionCount(policy.conditions);

  return (
    <Card className="h-full flex flex-col">
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <CardTitle className="flex items-center gap-2 text-lg">
            <span className="truncate">{policy.tool_name}</span>
          </CardTitle>
          <Badge
            variant={policy.action === 'deny' ? 'destructive' : 'secondary'}
            className={
              policy.action === 'audit'
                ? 'bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-300'
                : ''
            }
          >
            {policy.action === 'deny' ? 'Deny' : 'Audit'}
          </Badge>
        </div>
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          {getScopeIcon(policy.scope)}
          <span>{getScopeLabel(policy.scope)}</span>
        </div>
      </CardHeader>

      <CardContent className="flex-1 pb-3">
        {policy.reason && (
          <p className="text-sm text-muted-foreground line-clamp-2 mb-3">{policy.reason}</p>
        )}
        {conditionCount > 0 && (
          <div className="flex items-center gap-2 text-xs text-muted-foreground">
            <Settings className="h-3 w-3" />
            <span>
              {conditionCount} condition{conditionCount !== 1 ? 's' : ''}
            </span>
          </div>
        )}
      </CardContent>

      <CardFooter className="flex gap-2 pt-0">
        <Button variant="outline" size="sm" onClick={onEdit} className="flex-1">
          <Edit className="h-4 w-4 mr-2" />
          Edit
        </Button>
        <Button variant="outline" size="sm" onClick={onDelete} className="flex-1">
          <Trash2 className="h-4 w-4 mr-2" />
          Delete
        </Button>
      </CardFooter>
    </Card>
  );
}
