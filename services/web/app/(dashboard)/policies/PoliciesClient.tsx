'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Search, Plus, ShieldCheck } from 'lucide-react';
import { PolicyCard } from './components/PolicyCard';
import { useToast } from '@/components/ui/use-toast';
import {
  useCreatePolicy,
  useUpdatePolicy,
  useDeletePolicy,
} from '@/lib/hooks/usePolicies';
import type { ToolPolicy, CreateToolPolicyRequest, UpdateToolPolicyRequest } from '@/lib/types';

// Dynamic imports to avoid hydration issues with modals
const CreatePolicyModal = dynamic(
  () => import('./components/CreatePolicyModal').then((mod) => mod.CreatePolicyModal),
  { ssr: false }
);
const EditPolicyModal = dynamic(
  () => import('./components/EditPolicyModal').then((mod) => mod.EditPolicyModal),
  { ssr: false }
);
const DeletePolicyDialog = dynamic(
  () => import('./components/DeletePolicyDialog').then((mod) => mod.DeletePolicyDialog),
  { ssr: false }
);

type PoliciesClientProps = {
  initialPolicies: ToolPolicy[];
};

export function PoliciesClient({ initialPolicies }: PoliciesClientProps) {
  const [policies, setPolicies] = useState<ToolPolicy[]>(initialPolicies);
  const [searchQuery, setSearchQuery] = useState('');
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingPolicy, setEditingPolicy] = useState<ToolPolicy | null>(null);
  const [deletingPolicy, setDeletingPolicy] = useState<ToolPolicy | null>(null);
  const { toast } = useToast();

  const createMutation = useCreatePolicy();
  const updateMutation = useUpdatePolicy();
  const deleteMutation = useDeletePolicy();

  // Filter policies based on search query
  const filteredPolicies = policies.filter((policy) => {
    return (
      policy.tool_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      policy.reason?.toLowerCase().includes(searchQuery.toLowerCase())
    );
  });

  const handleCreatePolicy = async (data: CreateToolPolicyRequest) => {
    try {
      const newPolicy = await createMutation.mutateAsync(data);
      setPolicies([...policies, newPolicy]);
      setIsCreateModalOpen(false);

      toast({
        title: 'Policy created',
        description: `Policy for ${data.tool_name} has been created successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to create policy',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    }
  };

  const handleEditPolicy = async (data: UpdateToolPolicyRequest) => {
    if (!editingPolicy) return;

    try {
      const updatedPolicy = await updateMutation.mutateAsync({
        id: editingPolicy.id,
        data,
      });
      setPolicies(policies.map((p) => (p.id === editingPolicy.id ? updatedPolicy : p)));
      setEditingPolicy(null);

      toast({
        title: 'Policy updated',
        description: `Policy for ${updatedPolicy.tool_name} has been updated successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to update policy',
        description: error instanceof Error ? error.message : 'Please try again',
        variant: 'destructive',
      });
    }
  };

  const handleDeletePolicy = async () => {
    if (!deletingPolicy) return;

    try {
      await deleteMutation.mutateAsync(deletingPolicy.id);
      setPolicies(policies.filter((p) => p.id !== deletingPolicy.id));
      setDeletingPolicy(null);

      toast({
        title: 'Policy deleted',
        description: `Policy for ${deletingPolicy.tool_name} has been deleted successfully`,
        variant: 'success',
      });
    } catch (error) {
      toast({
        title: 'Failed to delete policy',
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
            placeholder="Search policies..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Create Policy
        </Button>
      </div>

      {/* Policies Grid */}
      {filteredPolicies.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-96 rounded-lg border-2 border-dashed border-gray-300 dark:border-gray-700">
          <ShieldCheck className="h-12 w-12 text-gray-400 dark:text-gray-500 mb-4" />
          <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">
            {searchQuery ? 'No policies found' : 'No policies yet'}
          </h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            {searchQuery
              ? 'Try adjusting your search'
              : 'Get started by creating your first policy'}
          </p>
          {!searchQuery && (
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="h-4 w-4 mr-2" />
              Create Policy
            </Button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {filteredPolicies.map((policy) => (
            <PolicyCard
              key={policy.id}
              policy={policy}
              onEdit={() => setEditingPolicy(policy)}
              onDelete={() => setDeletingPolicy(policy)}
            />
          ))}
        </div>
      )}

      {/* Modals - only render when open to avoid hydration issues */}
      {isCreateModalOpen && (
        <CreatePolicyModal
          isOpen={isCreateModalOpen}
          onClose={() => setIsCreateModalOpen(false)}
          onCreate={handleCreatePolicy}
          isLoading={createMutation.isPending}
        />
      )}

      {editingPolicy && (
        <EditPolicyModal
          isOpen={!!editingPolicy}
          onClose={() => setEditingPolicy(null)}
          onEdit={handleEditPolicy}
          policy={editingPolicy}
          isLoading={updateMutation.isPending}
        />
      )}

      {deletingPolicy && (
        <DeletePolicyDialog
          isOpen={!!deletingPolicy}
          onClose={() => setDeletingPolicy(null)}
          onDelete={handleDeletePolicy}
          policy={deletingPolicy}
          isLoading={deleteMutation.isPending}
        />
      )}
    </>
  );
}
