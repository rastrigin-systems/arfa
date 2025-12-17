'use client';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Edit, Trash2, Users } from 'lucide-react';
import type { Role } from '@/lib/types';

type RoleCardProps = {
  role: Role;
  onEdit: () => void;
  onDelete: () => void;
};

export function RoleCard({ role, onEdit, onDelete }: RoleCardProps) {
  return (
    <Card className="h-full flex flex-col">
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span className="truncate">{role.name}</span>
          <Badge variant="secondary" className="ml-2">
            {(role.permissions?.length ?? 0)} {(role.permissions?.length ?? 0) === 1 ? 'permission' : 'permissions'}
          </Badge>
        </CardTitle>
        <CardDescription className="line-clamp-2">
          {role.description || 'No description provided'}
        </CardDescription>
      </CardHeader>

      <CardContent className="flex-1">
        <div className="flex items-center gap-2 text-sm text-muted-foreground">
          <Users className="h-4 w-4" />
          <span>{role.employee_count || 0} employees</span>
        </div>
      </CardContent>

      <CardFooter className="flex gap-2">
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
