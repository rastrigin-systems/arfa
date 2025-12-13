'use client';

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';

type Role = {
  id: string;
  name: string;
  description: string;
  permissions: string[];
  employee_count?: number;
  created_at: string;
  updated_at: string;
};

type DeleteRoleDialogProps = {
  isOpen: boolean;
  onClose: () => void;
  onDelete: () => void;
  role: Role;
};

export function DeleteRoleDialog({ isOpen, onClose, onDelete, role }: DeleteRoleDialogProps) {
  const handleDelete = () => {
    onDelete();
  };

  // Check if role is assigned to employees
  const hasEmployees = role.employee_count && role.employee_count > 0;

  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Role</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              Are you sure you want to delete the role <strong>{role.name}</strong>?
            </p>
            {hasEmployees && (
              <div className="rounded-lg bg-yellow-50 dark:bg-yellow-900/20 p-3 border border-yellow-200 dark:border-yellow-800">
                <p className="text-yellow-800 dark:text-yellow-200 font-medium">
                  ⚠️ Warning: This role is assigned to {role.employee_count} employee
                  {role.employee_count !== 1 ? 's' : ''}
                </p>
                <p className="text-yellow-700 dark:text-yellow-300 text-sm mt-1">
                  Deleting this role will affect their access permissions. Consider reassigning employees to another role first.
                </p>
              </div>
            )}
            <p className="text-sm">This action cannot be undone.</p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={handleDelete}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
          >
            Delete Role
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
