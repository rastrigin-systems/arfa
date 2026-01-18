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
import type { ToolPolicy } from '@/lib/types';

type DeletePolicyDialogProps = {
  isOpen: boolean;
  onClose: () => void;
  onDelete: () => void;
  policy: ToolPolicy;
  isLoading?: boolean;
};

export function DeletePolicyDialog({
  isOpen,
  onClose,
  onDelete,
  policy,
  isLoading,
}: DeletePolicyDialogProps) {
  return (
    <AlertDialog open={isOpen} onOpenChange={(open) => !open && onClose()}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>Delete Policy</AlertDialogTitle>
          <AlertDialogDescription className="space-y-2">
            <p>
              Are you sure you want to delete the policy for <strong>{policy.tool_name}</strong>?
            </p>
            {policy.action === 'deny' && (
              <div className="rounded-lg bg-yellow-50 dark:bg-yellow-900/20 p-3 border border-yellow-200 dark:border-yellow-800">
                <p className="text-yellow-800 dark:text-yellow-200 font-medium">
                  Warning: This is a deny policy
                </p>
                <p className="text-yellow-700 dark:text-yellow-300 text-sm mt-1">
                  Deleting this policy will allow agents to use the {policy.tool_name} tool again.
                </p>
              </div>
            )}
            <p className="text-sm">This action cannot be undone.</p>
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel disabled={isLoading}>Cancel</AlertDialogCancel>
          <AlertDialogAction
            onClick={onDelete}
            className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
            disabled={isLoading}
          >
            {isLoading ? 'Deleting...' : 'Delete Policy'}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
