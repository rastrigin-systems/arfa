'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Search, Plus, Shield } from 'lucide-react';
import { RoleCard } from './components/RoleCard';
import { CreateRoleModal } from './components/CreateRoleModal';
import { EditRoleModal } from './components/EditRoleModal';
import { DeleteRoleDialog } from './components/DeleteRoleDialog';
import { useToast } from '@/components/ui/use-toast';
import type { Role } from '@/lib/types';

type RolesClientProps = {
  initialRoles: Role[];
};

export function RolesClient({ initialRoles }: RolesClientProps) {
  const [roles, setRoles] = useState<Role[]>(initialRoles);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | null>(null);
  const [deletingRole, setDeletingRole] = useState<Role | null>(null);
  const { toast } = useToast();

  // Filter roles based on search query
  const filteredRoles = roles.filter(
    (role) =>
      role.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      role.description?.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleCreateRole = async (data: { name: string; description: string; permissions: string[] }) => {
    try {
      // TODO: Call API to create role
      const newRole: Role = {
        id: crypto.randomUUID(),
        name: data.name,
        description: data.description,
        permissions: data.permissions,
        employee_count: 0,
        created_at: new Date().toISOString(),
      };

      setRoles([...roles, newRole]);
      setIsCreateModalOpen(false);

      toast({
        title: 'Role created',
        description: `${data.name} has been created successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to create role',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    }
  };

  const handleEditRole = async (data: { name: string; description: string; permissions: string[] }) => {
    if (!editingRole) return;

    try {
      // TODO: Call API to update role
      const updatedRole: Role = {
        ...editingRole,
        name: data.name,
        description: data.description,
        permissions: data.permissions,
      };

      setRoles(roles.map((r) => (r.id === editingRole.id ? updatedRole : r)));
      setEditingRole(null);

      toast({
        title: 'Role updated',
        description: `${data.name} has been updated successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to update role',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    }
  };

  const handleDeleteRole = async () => {
    if (!deletingRole) return;

    try {
      // TODO: Call API to delete role
      setRoles(roles.filter((r) => r.id !== deletingRole.id));
      setDeletingRole(null);

      toast({
        title: 'Role deleted',
        description: `${deletingRole.name} has been deleted successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to delete role',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    }
  };

  return (
    <>
      {/* Search and Create */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search roles..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Role
        </Button>
      </div>

      {/* Roles Grid */}
      {filteredRoles.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-96 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-700">
          <Shield className="h-12 w-12 text-gray-400 dark:text-gray-500 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
            {searchQuery ? 'No roles found' : 'No roles yet'}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            {searchQuery
              ? 'Try adjusting your search query'
              : 'Get started by creating your first role'}
          </p>
          {!searchQuery && (
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Create Role
            </Button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredRoles.map((role) => (
            <RoleCard
              key={role.id}
              role={role}
              onEdit={() => setEditingRole(role)}
              onDelete={() => setDeletingRole(role)}
            />
          ))}
        </div>
      )}

      {/* Modals */}
      <CreateRoleModal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        onCreate={handleCreateRole}
      />

      {editingRole && (
        <EditRoleModal
          isOpen={!!editingRole}
          onClose={() => setEditingRole(null)}
          onEdit={handleEditRole}
          role={editingRole}
        />
      )}

      {deletingRole && (
        <DeleteRoleDialog
          isOpen={!!deletingRole}
          onClose={() => setDeletingRole(null)}
          onDelete={handleDeleteRole}
          role={deletingRole}
        />
      )}
    </>
  );
}
