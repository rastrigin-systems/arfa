'use client';

import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import { Checkbox } from '@/components/ui/checkbox';
import type { Role } from '@/lib/types';

type EditRoleModalProps = {
  isOpen: boolean;
  onClose: () => void;
  onEdit: (data: { name: string; description: string; permissions: string[] }) => void;
  role: Role;
};

// Available permissions (same as CreateRoleModal)
const AVAILABLE_PERMISSIONS = [
  { id: 'agents.create', label: 'Create Agents', description: 'Create new agent configurations' },
  { id: 'agents.edit', label: 'Edit Agents', description: 'Modify existing agent configurations' },
  { id: 'agents.delete', label: 'Delete Agents', description: 'Remove agent configurations' },
  { id: 'agents.view', label: 'View Agents', description: 'View agent configurations' },
  { id: 'employees.create', label: 'Create Employees', description: 'Add new employees' },
  { id: 'employees.edit', label: 'Edit Employees', description: 'Modify employee details' },
  { id: 'employees.delete', label: 'Delete Employees', description: 'Remove employees' },
  { id: 'employees.view', label: 'View Employees', description: 'View employee information' },
  { id: 'teams.create', label: 'Create Teams', description: 'Create new teams' },
  { id: 'teams.edit', label: 'Edit Teams', description: 'Modify team details' },
  { id: 'teams.delete', label: 'Delete Teams', description: 'Remove teams' },
  { id: 'teams.view', label: 'View Teams', description: 'View team information' },
  { id: 'roles.create', label: 'Create Roles', description: 'Create new roles' },
  { id: 'roles.edit', label: 'Edit Roles', description: 'Modify roles' },
  { id: 'roles.delete', label: 'Delete Roles', description: 'Remove roles' },
  { id: 'roles.view', label: 'View Roles', description: 'View roles' },
];

export function EditRoleModal({ isOpen, onClose, onEdit, role }: EditRoleModalProps) {
  const [name, setName] = useState(role.name);
  const [description, setDescription] = useState(role.description ?? '');
  const [selectedPermissions, setSelectedPermissions] = useState<string[]>(role.permissions ?? []);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Update form when role changes
  useEffect(() => {
    setName(role.name);
    setDescription(role.description ?? '');
    setSelectedPermissions(role.permissions ?? []);
  }, [role]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    // Validation
    const newErrors: Record<string, string> = {};
    if (!name.trim()) {
      newErrors.name = 'Role name is required';
    }
    if (selectedPermissions.length === 0) {
      newErrors.permissions = 'At least one permission must be selected';
    }

    if (Object.keys(newErrors).length > 0) {
      setErrors(newErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      await onEdit({
        name: name.trim(),
        description: description.trim(),
        permissions: selectedPermissions,
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  const togglePermission = (permissionId: string) => {
    setSelectedPermissions((prev) =>
      prev.includes(permissionId)
        ? prev.filter((p) => p !== permissionId)
        : [...prev, permissionId]
    );
  };

  const handleClose = () => {
    setErrors({});
    onClose();
  };

  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && handleClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Edit Role</DialogTitle>
          <DialogDescription>
            Update role details and permissions
            {role.employee_count && role.employee_count > 0 && (
              <span className="block mt-2 text-yellow-600 dark:text-yellow-500">
                ⚠️ This role is assigned to {role.employee_count} employee
                {role.employee_count !== 1 ? 's' : ''}. Changes will affect their permissions.
              </span>
            )}
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Role Name */}
          <div className="space-y-2">
            <Label htmlFor="name">
              Role Name <span className="text-red-500">*</span>
            </Label>
            <Input
              id="name"
              placeholder="e.g., Admin, Developer, Manager"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className={errors.name ? 'border-red-500' : ''}
            />
            {errors.name && <p className="text-sm text-red-500">{errors.name}</p>}
          </div>

          {/* Description */}
          <div className="space-y-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Describe the purpose of this role..."
              value={description ?? ''}
              onChange={(e) => setDescription(e.target.value)}
              rows={3}
            />
          </div>

          {/* Permissions */}
          <div className="space-y-2">
            <Label>
              Permissions <span className="text-red-500">*</span>
            </Label>
            <div className="rounded-lg border p-4 max-h-64 overflow-y-auto space-y-3">
              {AVAILABLE_PERMISSIONS.map((permission) => (
                <div key={permission.id} className="flex items-start space-x-3">
                  <Checkbox
                    id={`edit-${permission.id}`}
                    checked={selectedPermissions.includes(permission.id)}
                    onCheckedChange={() => togglePermission(permission.id)}
                  />
                  <div className="flex-1">
                    <label
                      htmlFor={`edit-${permission.id}`}
                      className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70 cursor-pointer"
                    >
                      {permission.label}
                    </label>
                    <p className="text-xs text-muted-foreground">{permission.description}</p>
                  </div>
                </div>
              ))}
            </div>
            {errors.permissions && <p className="text-sm text-red-500">{errors.permissions}</p>}
            <p className="text-xs text-muted-foreground">
              {selectedPermissions.length} permission{selectedPermissions.length !== 1 ? 's' : ''} selected
            </p>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" onClick={handleClose} disabled={isSubmitting}>
              Cancel
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? 'Saving...' : 'Save Changes'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
